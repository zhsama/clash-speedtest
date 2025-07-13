package utils

import (
	"bufio"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"log/slog"

	"github.com/faceair/clash-speedtest/logger"
)

// TUNStatus 表示 TUN 模式的状态信息
type TUNStatus struct {
	Enabled           bool                   `json:"enabled"`             // TUN 模式是否启用
	Interfaces        []TUNInterface         `json:"interfaces"`          // TUN 接口列表
	ActiveInterface   *TUNInterface          `json:"active_interface"`    // 当前活动的 TUN 接口
	ProxyProcesses    []ProxyProcess         `json:"proxy_processes"`     // 检测到的代理进程
	DefaultRoute      *RouteInfo             `json:"default_route"`       // 默认路由信息
	DetectionTime     time.Time              `json:"detection_time"`      // 检测时间
	SystemInfo        SystemInfo             `json:"system_info"`         // 系统信息
	AdditionalDetails map[string]any `json:"additional_details"`  // 额外的检测信息
}

// TUNInterface 表示 TUN 网络接口信息
type TUNInterface struct {
	Name         string   `json:"name"`          // 接口名称
	Type         string   `json:"type"`          // 接口类型（TUN/TAP）
	IPAddresses  []string `json:"ip_addresses"`  // IP 地址列表
	IsUp         bool     `json:"is_up"`         // 接口是否启用
	MTU          int      `json:"mtu"`           // MTU 大小
	IsDefault    bool     `json:"is_default"`    // 是否为默认路由接口
	AssociatedPID int     `json:"associated_pid"`// 关联的进程 PID
}

// ProxyProcess 表示代理进程信息
type ProxyProcess struct {
	Name        string `json:"name"`         // 进程名称
	PID         int    `json:"pid"`          // 进程 ID
	Command     string `json:"command"`      // 完整命令
	ProcessType string `json:"process_type"` // 进程类型（clash/surge/other）
}

// RouteInfo 表示路由信息
type RouteInfo struct {
	Destination string `json:"destination"` // 目标网络
	Gateway     string `json:"gateway"`     // 网关
	Interface   string `json:"interface"`   // 接口
	Metric      int    `json:"metric"`      // 路由优先级
}

// SystemInfo 表示系统信息
type SystemInfo struct {
	OS           string `json:"os"`            // 操作系统
	Architecture string `json:"architecture"`  // 系统架构
	Hostname     string `json:"hostname"`      // 主机名
}

// CheckTUNMode 检测系统是否启用了 TUN 模式
func CheckTUNMode() *TUNStatus {
	logger.Logger.Info("开始检测 TUN 模式状态")
	
	status := &TUNStatus{
		DetectionTime:     time.Now(),
		AdditionalDetails: make(map[string]any),
	}
	
	// 获取系统信息
	status.SystemInfo = getSystemInfo()
	
	// 检测网络接口
	interfaces := getTUNInterfaces()
	status.Interfaces = interfaces
	
	// 检测活动的 TUN 接口
	activeInterface := getActiveTUNInterface(interfaces)
	if activeInterface != nil {
		status.ActiveInterface = activeInterface
		status.Enabled = true
	}
	
	// 检测代理进程
	processes := getProxyProcesses()
	status.ProxyProcesses = processes
	
	// 获取默认路由信息
	defaultRoute := getDefaultRoute()
	status.DefaultRoute = defaultRoute
	
	// 基于多个条件判断 TUN 模式是否启用
	status.Enabled = determineTUNModeStatus(status)
	
	// 记录检测结果
	logger.Logger.Info("TUN 模式检测完成",
		slog.Bool("enabled", status.Enabled),
		slog.Int("tun_interfaces", len(status.Interfaces)),
		slog.Int("proxy_processes", len(status.ProxyProcesses)),
	)
	
	return status
}

// getSystemInfo 获取系统信息
func getSystemInfo() SystemInfo {
	hostname, _ := os.Hostname()
	
	return SystemInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Hostname:     hostname,
	}
}

// getTUNInterfaces 获取所有 TUN 接口
func getTUNInterfaces() []TUNInterface {
	var tunInterfaces []TUNInterface
	
	interfaces, err := net.Interfaces()
	if err != nil {
		logger.Logger.Error("获取网络接口失败", slog.String("error", err.Error()))
		return tunInterfaces
	}
	
	tunPattern := regexp.MustCompile(`^(tun|utun|tap)\d*$`)
	
	for _, iface := range interfaces {
		if tunPattern.MatchString(iface.Name) {
			tunIface := TUNInterface{
				Name:  iface.Name,
				Type:  detectInterfaceType(iface.Name),
				IsUp:  iface.Flags&net.FlagUp != 0,
				MTU:   iface.MTU,
			}
			
			// 获取接口的 IP 地址
			addrs, err := iface.Addrs()
			if err == nil {
				for _, addr := range addrs {
					tunIface.IPAddresses = append(tunIface.IPAddresses, addr.String())
				}
			}
			
			// 检查是否为默认路由接口
			tunIface.IsDefault = isDefaultRouteInterface(iface.Name)
			
			tunInterfaces = append(tunInterfaces, tunIface)
			
			logger.Logger.Debug("发现 TUN 接口",
				slog.String("name", tunIface.Name),
				slog.String("type", tunIface.Type),
				slog.Bool("is_up", tunIface.IsUp),
				slog.Bool("is_default", tunIface.IsDefault),
			)
		}
	}
	
	return tunInterfaces
}

// detectInterfaceType 检测接口类型
func detectInterfaceType(name string) string {
	if strings.HasPrefix(name, "tun") || strings.HasPrefix(name, "utun") {
		return "TUN"
	} else if strings.HasPrefix(name, "tap") {
		return "TAP"
	}
	return "UNKNOWN"
}

// getActiveTUNInterface 获取活动的 TUN 接口
func getActiveTUNInterface(interfaces []TUNInterface) *TUNInterface {
	for _, iface := range interfaces {
		if iface.IsUp && len(iface.IPAddresses) > 0 {
			// 优先选择默认路由接口
			if iface.IsDefault {
				return &iface
			}
		}
	}
	
	// 如果没有默认路由接口，选择第一个启用的接口
	for _, iface := range interfaces {
		if iface.IsUp && len(iface.IPAddresses) > 0 {
			return &iface
		}
	}
	
	return nil
}

// getProxyProcesses 获取代理进程信息
func getProxyProcesses() []ProxyProcess {
	var processes []ProxyProcess
	
	// 根据操作系统使用不同的命令
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd = exec.Command("ps", "aux")
	case "windows":
		cmd = exec.Command("tasklist", "/v")
	default:
		logger.Logger.Warn("不支持的操作系统", slog.String("os", runtime.GOOS))
		return processes
	}
	
	output, err := cmd.Output()
	if err != nil {
		logger.Logger.Error("获取进程列表失败", slog.String("error", err.Error()))
		return processes
	}
	
	lines := strings.Split(string(output), "\n")
	
	// 要检测的代理应用关键词
	proxyKeywords := []string{"clash", "surge", "shadowsocks", "v2ray", "quantumult", "proxyman"}
	
	for _, line := range lines {
		line = strings.ToLower(line)
		for _, keyword := range proxyKeywords {
			if strings.Contains(line, keyword) {
				process := parseProcessLine(line, keyword)
				if process.Name != "" {
					processes = append(processes, process)
				}
				break
			}
		}
	}
	
	logger.Logger.Debug("检测到代理进程", slog.Int("count", len(processes)))
	
	return processes
}

// parseProcessLine 解析进程行信息
func parseProcessLine(line, keyword string) ProxyProcess {
	// 这里简化处理，实际可以根据操作系统的 ps 输出格式进行更精确的解析
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return ProxyProcess{}
	}
	
	return ProxyProcess{
		Name:        keyword,
		ProcessType: detectProxyType(keyword),
		Command:     line,
	}
}

// detectProxyType 检测代理类型
func detectProxyType(keyword string) string {
	switch {
	case strings.Contains(keyword, "clash"):
		return "clash"
	case strings.Contains(keyword, "surge"):
		return "surge"
	case strings.Contains(keyword, "shadowsocks"):
		return "shadowsocks"
	case strings.Contains(keyword, "v2ray"):
		return "v2ray"
	default:
		return "other"
	}
}

// getDefaultRoute 获取默认路由信息
func getDefaultRoute() *RouteInfo {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("route", "-n", "get", "default")
	case "linux":
		cmd = exec.Command("ip", "route", "show", "default")
	case "windows":
		cmd = exec.Command("route", "print", "0.0.0.0")
	default:
		logger.Logger.Warn("不支持的操作系统路由检测", slog.String("os", runtime.GOOS))
		return nil
	}
	
	output, err := cmd.Output()
	if err != nil {
		logger.Logger.Error("获取默认路由失败", slog.String("error", err.Error()))
		return nil
	}
	
	return parseDefaultRoute(string(output), runtime.GOOS)
}

// parseDefaultRoute 解析默认路由信息
func parseDefaultRoute(output, osType string) *RouteInfo {
	route := &RouteInfo{}
	
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	switch osType {
	case "darwin":
		// macOS 路由格式解析
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.Contains(line, "gateway:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					route.Gateway = parts[1]
				}
			} else if strings.Contains(line, "interface:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					route.Interface = parts[1]
				}
			}
		}
		
	case "linux":
		// Linux 路由格式解析
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "default") {
				parts := strings.Fields(line)
				if len(parts) >= 5 {
					route.Gateway = parts[2]
					route.Interface = parts[4]
				}
			}
		}
		
	case "windows":
		// Windows 路由格式解析（简化）
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.Contains(line, "0.0.0.0") && strings.Contains(line, "255.255.255.255") {
				// 简化的 Windows 路由解析
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					route.Gateway = fields[2]
				}
			}
		}
	}
	
	route.Destination = "0.0.0.0/0"
	
	if route.Gateway != "" || route.Interface != "" {
		logger.Logger.Debug("默认路由信息",
			slog.String("gateway", route.Gateway),
			slog.String("interface", route.Interface),
		)
		return route
	}
	
	return nil
}

// isDefaultRouteInterface 检查接口是否为默认路由接口
func isDefaultRouteInterface(interfaceName string) bool {
	defaultRoute := getDefaultRoute()
	if defaultRoute == nil {
		return false
	}
	
	return defaultRoute.Interface == interfaceName
}

// determineTUNModeStatus 基于多个条件判断 TUN 模式状态
func determineTUNModeStatus(status *TUNStatus) bool {
	// 条件1：存在活动的 TUN 接口
	hasTUNInterface := status.ActiveInterface != nil
	
	// 条件2：检测到代理进程
	hasProxyProcess := len(status.ProxyProcesses) > 0
	
	// 条件3：默认路由指向 TUN 接口
	defaultRoutesToTUN := false
	if status.DefaultRoute != nil && status.ActiveInterface != nil {
		defaultRoutesToTUN = status.DefaultRoute.Interface == status.ActiveInterface.Name
	}
	
	// 记录详细检测信息
	status.AdditionalDetails["has_tun_interface"] = hasTUNInterface
	status.AdditionalDetails["has_proxy_process"] = hasProxyProcess
	status.AdditionalDetails["default_routes_to_tun"] = defaultRoutesToTUN
	
	logger.Logger.Debug("TUN 模式状态判断",
		slog.Bool("has_tun_interface", hasTUNInterface),
		slog.Bool("has_proxy_process", hasProxyProcess),
		slog.Bool("default_routes_to_tun", defaultRoutesToTUN),
	)
	
	// TUN 模式被认为启用的条件：
	// 1. 存在活动的 TUN 接口，并且
	// 2. 要么有代理进程运行，要么默认路由指向 TUN 接口
	return hasTUNInterface && (hasProxyProcess || defaultRoutesToTUN)
}

// GetTUNModeDetails 获取 TUN 模式的详细信息（简化版本）
func GetTUNModeDetails() map[string]any {
	status := CheckTUNMode()
	
	return map[string]any{
		"enabled":            status.Enabled,
		"interface_count":    len(status.Interfaces),
		"active_interface":   status.ActiveInterface,
		"proxy_process_count": len(status.ProxyProcesses),
		"system_os":          status.SystemInfo.OS,
		"detection_time":     status.DetectionTime,
	}
}