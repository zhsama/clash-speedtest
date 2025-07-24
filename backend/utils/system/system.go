package system

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/zhsama/clash-speedtest/logger"
)

// TUNStatus 表示 TUN 模式的状态信息
type TUNStatus struct {
	Enabled           bool           `json:"enabled"`            // TUN 模式是否启用
	Interfaces        []TUNInterface `json:"interfaces"`         // TUN 接口列表
	ActiveInterface   *TUNInterface  `json:"active_interface"`   // 当前活动的 TUN 接口
	ProxyProcesses    []ProxyProcess `json:"proxy_processes"`    // 检测到的代理进程
	DefaultRoute      *RouteInfo     `json:"default_route"`      // 默认路由信息
	DetectionTime     time.Time      `json:"detection_time"`     // 检测时间
	SystemInfo        SystemInfo     `json:"system_info"`        // 系统信息
	AdditionalDetails map[string]any `json:"additional_details"` // 额外的检测信息
}

// TUNInterface 表示 TUN 网络接口信息
type TUNInterface struct {
	Name          string   `json:"name"`           // 接口名称
	Type          string   `json:"type"`           // 接口类型（TUN/TAP）
	IPAddresses   []string `json:"ip_addresses"`   // IP 地址列表
	IsUp          bool     `json:"is_up"`          // 接口是否启用
	MTU           int      `json:"mtu"`            // MTU 大小
	IsDefault     bool     `json:"is_default"`     // 是否为默认路由接口
	AssociatedPID int      `json:"associated_pid"` // 关联的进程 PID
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
	OS           string `json:"os"`           // 操作系统
	Architecture string `json:"architecture"` // 系统架构
	Hostname     string `json:"hostname"`     // 主机名
}

// 检测系统是否启用了 TUN 模式，使用多阶段检测策略：优先使用 Fake-IP 检测，降级到传统接口检测
func CheckTUNMode() *TUNStatus {
	logger.Logger.Info("Starting TUN mode detection")

	status := &TUNStatus{
		DetectionTime:     time.Now(),
		AdditionalDetails: make(map[string]any),
	}

	status.SystemInfo = getSystemInfo()

	interfaces := getTUNInterfaces()
	status.Interfaces = interfaces

	activeInterface := getActiveTUNInterface(interfaces)
	if activeInterface != nil {
		status.ActiveInterface = activeInterface
		status.Enabled = true
	}

	processes := getProxyProcesses()
	status.ProxyProcesses = processes

	defaultRoute := getDefaultRoute()
	status.DefaultRoute = defaultRoute

	// 核心检测逻辑：结合接口状态、进程信息和路由表判断
	status.Enabled = determineTUNModeStatus(status)

	// Log detection results
	logger.Logger.Info("TUN mode detection completed",
		slog.Bool("enabled", status.Enabled),
		slog.Int("tun_interfaces", len(status.Interfaces)),
		slog.Int("proxy_processes", len(status.ProxyProcesses)),
	)

	return status
}

// 获取系统信息
func getSystemInfo() SystemInfo {
	hostname, _ := os.Hostname()

	return SystemInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Hostname:     hostname,
	}
}

// 扫描系统中的 TUN/TAP 类型网络接口，支持标准的 tun/utun/tap 接口以及 Clash 专用接口
func getTUNInterfaces() []TUNInterface {
	var tunInterfaces []TUNInterface

	interfaces, err := net.Interfaces()
	if err != nil {
		logger.Logger.Error("Failed to get network interfaces", slog.String("error", err.Error()))
		return tunInterfaces
	}

	// 匹配 TUN/TAP 接口命名模式
	tunPattern := regexp.MustCompile(`^(tun|utun|tap|clash)\d*$`)

	for _, iface := range interfaces {
		if tunPattern.MatchString(iface.Name) {
			tunIface := TUNInterface{
				Name: iface.Name,
				Type: detectInterfaceType(iface.Name),
				IsUp: iface.Flags&net.FlagUp != 0,
				MTU:  iface.MTU,
			}

			addrs, err := iface.Addrs()
			if err == nil {
				for _, addr := range addrs {
					tunIface.IPAddresses = append(tunIface.IPAddresses, addr.String())
				}
			}

			tunIface.IsDefault = isDefaultRouteInterface(iface.Name)

			tunInterfaces = append(tunInterfaces, tunIface)
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
	} else if strings.HasPrefix(name, "clash") {
		return "CLASH_TUN"
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

// getProxyProcesses 检测系统中运行的代理应用进程
// 通过进程名模式匹配和开发工具过滤，准确识别真实的代理进程
func getProxyProcesses() []ProxyProcess {
	var processes []ProxyProcess
	seenProcesses := make(map[string]bool)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd = exec.Command("ps", "aux")
	case "windows":
		cmd = exec.Command("tasklist", "/v")
	default:
		logger.Logger.Warn("Unsupported operating system", slog.String("os", runtime.GOOS))
		return processes
	}

	output, err := cmd.Output()
	if err != nil {
		logger.Logger.Error("Failed to get process list", slog.String("error", err.Error()))
		return processes
	}

	lines := strings.Split(string(output), "\n")

	// 代理应用特征模式匹配表
	proxyPatterns := map[string][]string{
		"clash": {
			"clash-verge",
			"verge-mihomo",
			"clash-premium",
			"clash-meta",
			"clash.exe",
			"clash-darwin",
			"clash-linux",
			"clash-service",
			"/clash", // 路径包含 clash 但不是开发工具
		},
		"surge": {
			"surge",
			"surge-cli",
		},
		"shadowsocks": {
			"ss-local",
			"ss-server",
			"shadowsocks",
		},
		"v2ray": {
			"v2ray",
			"v2fly",
		},
	}

	for _, line := range lines {
		lineLower := strings.ToLower(line)

		// 跳过明显的开发工具进程
		if isDevToolProcess(lineLower) {
			continue
		}

		for proxyType, patterns := range proxyPatterns {
			for _, pattern := range patterns {
				if strings.Contains(lineLower, pattern) {
					process := parseProcessLine(line, proxyType)
					if process.Name != "" {
						// 使用进程命令作为去重key
						processKey := process.Command
						if !seenProcesses[processKey] {
							seenProcesses[processKey] = true
							processes = append(processes, process)
						}
					}
					break
				}
			}
		}
	}

	logger.Logger.Debug("Detected proxy processes", slog.Int("count", len(processes)))

	return processes
}

// isDevToolProcess 过滤开发工具和构建工具进程
// 避免将前端构建工具、包管理器等误识别为代理进程
func isDevToolProcess(line string) bool {
	devToolKeywords := []string{
		"turbo", "esbuild", "astro", "node_modules", "go-build",
		"npm", "yarn", "pnpm", "webpack", "vite", "rollup",
		"clash-speedtest", // 排除自身测试服务器
	}

	for _, keyword := range devToolKeywords {
		if strings.Contains(line, keyword) {
			return true
		}
	}

	return false
}

// parseProcessLine 解析进程行信息
func parseProcessLine(line, proxyType string) ProxyProcess {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return ProxyProcess{}
	}

	// 提取 PID (第二个字段)
	pid, _ := strconv.Atoi(fields[1])

	// 简化进程名提取：从命令行路径中提取最后部分
	var processName = proxyType
	if len(fields) > 10 {
		commandField := strings.Join(fields[10:], " ")
		if strings.Contains(commandField, "/") {
			parts := strings.Split(commandField, "/")
			name := parts[len(parts)-1]
			// 移除参数部分
			if spaceIndex := strings.Index(name, " "); spaceIndex != -1 {
				name = name[:spaceIndex]
			}
			if name != "" {
				processName = name
			}
		}
	}

	return ProxyProcess{
		Name:        processName,
		PID:         pid,
		ProcessType: proxyType,
		Command:     line,
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
		logger.Logger.Warn("Unsupported operating system for route detection", slog.String("os", runtime.GOOS))
		return nil
	}

	output, err := cmd.Output()
	if err != nil {
		logger.Logger.Error("Failed to get default route", slog.String("error", err.Error()))
		return nil
	}

	return parseDefaultRoute(string(output), runtime.GOOS)
}

// parseDefaultRoute 解析不同操作系统的路由命令输出
// 提取默认路由的网关和接口信息，用于判断流量路由策略
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
		// Windows 路由格式解析
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
		logger.Logger.Debug("Default route information",
			slog.String("gateway", route.Gateway),
			slog.String("interface", route.Interface),
		)
		return route
	}

	return nil
}

// 检查指定接口是否为系统默认路由接口，TUN 接口作为默认路由时通常表示全局代理模式
func isDefaultRouteInterface(interfaceName string) bool {
	defaultRoute := getDefaultRoute()
	if defaultRoute == nil {
		return false
	}

	return defaultRoute.Interface == interfaceName
}

// 核心检测逻辑：结合接口、进程和路由信息判断 TUN 模式，优先使用 Fake-IP 检测（更准确），失败时降级到传统检测
func determineTUNModeStatus(status *TUNStatus) bool {
	hasTUNInterface := status.ActiveInterface != nil
	hasProxyProcess := len(status.ProxyProcesses) > 0

	status.AdditionalDetails["has_tun_interface"] = hasTUNInterface
	status.AdditionalDetails["has_proxy_process"] = hasProxyProcess

	// 优先策略：Fake-IP 检测 - 通过检查 198.18.x.x 网段配置和路由
	if fakeIPEnabled := checkFakeIPMode(status); fakeIPEnabled {
		status.AdditionalDetails["detection_method"] = "fake_ip"
		return true
	}

	// 降级策略：传统检测 - 基于接口状态和进程存在性
	if traditionalEnabled := checkTraditionalTUNMode(status); traditionalEnabled {
		status.AdditionalDetails["detection_method"] = "traditional"
		return true
	}

	status.AdditionalDetails["detection_method"] = "none"
	return false
}

// 检测 Fake-IP 模式 - 通过检查 198.18.0.0/16 网段的接口配置和对应路由表项
func checkFakeIPMode(status *TUNStatus) bool {
	enabled, details, err := CheckTUNModeWithFakeIP(DefaultFakeIPCIDR)

	if err != nil {
		logger.Logger.Debug("Fake-IP detection failed", slog.String("error", err.Error()))
		return false
	}

	if enabled {
		status.AdditionalDetails["fake_ip_interface"] = details["suspect_interfaces"]
		status.AdditionalDetails["fake_ip_cidr"] = DefaultFakeIPCIDR
	}

	return enabled
}

// 传统 TUN 检测 - 基于接口存在性和代理进程，要求同时满足：TUN 接口活跃 + 代理进程运行 + (默认路由指向TUN 或 非链路本地IP)
func checkTraditionalTUNMode(status *TUNStatus) bool {
	hasTUNInterface := status.ActiveInterface != nil
	hasProxyProcess := len(status.ProxyProcesses) > 0

	if !hasTUNInterface || !hasProxyProcess {
		return false
	}

	defaultRoutesToTUN := false
	if status.DefaultRoute != nil && status.ActiveInterface != nil {
		defaultRoutesToTUN = status.DefaultRoute.Interface == status.ActiveInterface.Name
	}

	// 检查是否有非链路本地地址的 TUN 接口
	hasNonLinkLocalIP := hasNonLinkLocalIPAddress(status.Interfaces)

	result := hasNonLinkLocalIP || defaultRoutesToTUN
	status.AdditionalDetails["default_routes_to_tun"] = defaultRoutesToTUN
	status.AdditionalDetails["has_non_link_local_ip"] = hasNonLinkLocalIP

	return result
}

// 检查是否有非链路本地地址
func hasNonLinkLocalIPAddress(interfaces []TUNInterface) bool {
	for _, iface := range interfaces {
		if !iface.IsUp {
			continue
		}

		for _, addr := range iface.IPAddresses {
			ipStr := addr
			if strings.Contains(addr, "/") {
				ipStr = strings.Split(addr, "/")[0]
			}

			ip := net.ParseIP(strings.TrimSpace(ipStr))
			if ip != nil && ip.To4() != nil && !ip.IsLinkLocalUnicast() {
				return true
			}
		}
	}
	return false
}

const (
	// DefaultFakeIPCIDR 是默认的 Fake-IP 网段
	DefaultFakeIPCIDR = "198.18.0.0/16"
)

// 使用 Fake-IP 检测算法验证 TUN 模式状态，通过检查 198.18.0.0/16 网段的接口配置和对应路由表项来确认 TUN 模式
func CheckTUNModeWithFakeIP(fakeCIDR string) (bool, map[string]any, error) {
	details := make(map[string]any)
	details["fake_ip_cidr"] = fakeCIDR

	// 获取所有 TUN 接口
	interfaces := getTUNInterfaces()
	details["total_tun_interfaces"] = len(interfaces)

	// 查找具有 Fake-IP 地址的接口
	suspectIfaces, err := FindFakeIPInterfaces(interfaces, fakeCIDR)
	if err != nil {
		return false, details, err
	}

	details["suspect_interfaces"] = MapKeysToSlice(suspectIfaces)

	if len(suspectIfaces) == 0 {
		details["reason"] = "no_fake_ip_interface"
		return false, details, nil
	}

	// 检查路由表
	routeExists, err := CheckRouteToFakeIP(fakeCIDR, suspectIfaces)
	if err != nil {
		details["route_check_error"] = err.Error()
		// 路由检查失败时，如果找到了 Fake-IP 接口，可能仍然是启用状态
		if len(suspectIfaces) > 0 {
			details["reason"] = "route_check_failed_but_fake_ip_interface_exists"
			return true, details, nil
		}
		return false, details, err
	}

	details["route_to_fake_ip_exists"] = routeExists

	if !routeExists {
		details["reason"] = "no_route_to_fake_ip"
		return false, details, nil
	}

	details["reason"] = "fake_ip_interface_and_route_detected"
	return true, details, nil
}

// 识别配置了 Fake-IP 地址段的网络接口，扫描所有 TUN 接口的 IP 配置，查找 198.18.0.0/16 网段地址，支持多种地址格式：CIDR、点对点连接等
func FindFakeIPInterfaces(interfaces []TUNInterface, fakeCIDR string) (map[string]bool, error) {
	suspectIfaces := make(map[string]bool)

	_, fakeNet, err := net.ParseCIDR(fakeCIDR)
	if err != nil {
		return nil, fmt.Errorf("解析 Fake-IP 网段失败: %w", err)
	}

	for _, iface := range interfaces {
		for _, addrStr := range iface.IPAddresses {
			// 处理 CIDR 格式的地址 (如 "192.168.1.1/24")
			ipStr := addrStr
			if strings.Contains(addrStr, "/") {
				ipStr = strings.Split(addrStr, "/")[0]
			}

			// 处理点对点格式的地址 (如 "198.18.0.1 --> 198.18.0.1")
			if strings.Contains(ipStr, " --> ") {
				ipStr = strings.Split(ipStr, " --> ")[0]
			}

			// 解析 IP 地址
			ip := net.ParseIP(strings.TrimSpace(ipStr))
			if ip == nil || ip.To4() == nil {
				continue
			}

			// 检查是否在 Fake-IP 网段内
			if fakeNet.Contains(ip) {
				suspectIfaces[iface.Name] = true
				logger.Logger.Debug("Found possible TUN interface",
					slog.String("name", iface.Name),
					slog.String("ip", ipStr),
				)
			}
		}
	}

	return suspectIfaces, nil
}

// 跨平台路由表检查入口函数，根据操作系统类型调用相应的路由检查实现
func CheckRouteToFakeIP(fakeCIDR string, suspectIfaces map[string]bool) (bool, error) {
	switch runtime.GOOS {
	case "darwin":
		return CheckRouteToFakeIPDarwin(fakeCIDR, suspectIfaces)
	case "linux":
		return CheckRouteToFakeIPLinux(fakeCIDR, suspectIfaces)
	case "windows":
		return CheckRouteToFakeIPWindows(fakeCIDR, suspectIfaces)
	default:
		return false, fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// CheckRouteToFakeIPDarwin macOS 系统路由表分析
// 使用 netstat 命令解析路由表，查找指向 Fake-IP 网段的路由条目
// 支持 CIDR 路由和单IP路由的检测，确认 TUN 接口的路由配置
func CheckRouteToFakeIPDarwin(fakeCIDR string, suspectIfaces map[string]bool) (bool, error) {
	_, fakeNet, err := net.ParseCIDR(fakeCIDR)
	if err != nil {
		return false, fmt.Errorf("解析 Fake-IP 网段失败: %w", err)
	}

	cmd := exec.Command("netstat", "-rn", "-f", "inet")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("获取路由表失败: %w", err)
	}

	logger.Logger.Debug("Starting route table check",
		slog.String("fake_cidr", fakeCIDR),
		slog.Any("suspect_interfaces", MapKeysToSlice(suspectIfaces)),
	)

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		dest := fields[0]
		gateway := fields[1]
		// macOS netstat 格式: Destination Gateway Flags Netif [Expire]
		iface := fields[3]

		// 检查接口是否为可疑接口
		if !suspectIfaces[iface] {
			continue
		}

		logger.Logger.Debug("Checking route entry",
			slog.String("destination", dest),
			slog.String("gateway", gateway),
			slog.String("interface", iface),
		)

		// 检查目标地址是否在 Fake-IP 段内
		isRelated := false

		// 处理不同格式的路由条目
		if strings.Contains(dest, "/") {
			// CIDR 格式，如 "128.0/1"
			if destNet, err := parseCIDRRoute(dest); err == nil {
				// 检查这个网段是否与 Fake-IP 段有重叠
				if fakeNet.Contains(destNet.IP) || destNet.Contains(fakeNet.IP) {
					isRelated = true
					logger.Logger.Debug("Found overlapping subnet route",
						slog.String("dest_cidr", dest),
						slog.String("interface", iface),
					)
				}
			}
		} else {
			// 单个IP地址
			if ip := net.ParseIP(dest); ip != nil {
				if fakeNet.Contains(ip) {
					isRelated = true
					logger.Logger.Debug("Found single IP route to Fake-IP",
						slog.String("dest_ip", dest),
						slog.String("interface", iface),
					)
				}
			}
		}

		// 检查网关是否在 Fake-IP 段内（常见的 TUN 配置）
		if gatewayIP := net.ParseIP(gateway); gatewayIP != nil {
			if fakeNet.Contains(gatewayIP) {
				isRelated = true
				logger.Logger.Debug("Found route using Fake-IP gateway",
					slog.String("gateway", gateway),
					slog.String("interface", iface),
				)
			}
		}

		if isRelated {
			logger.Logger.Info("Confirmed route to Fake-IP found",
				slog.String("destination", dest),
				slog.String("gateway", gateway),
				slog.String("interface", iface),
			)
			return true, nil
		}
	}

	logger.Logger.Debug("No route to Fake-IP found")
	return false, nil
}

// 解析路由表条目中的网络地址表示法，将路由表中的 "128.0/1" 格式转换为标准的 IPNet 结构
func parseCIDRRoute(route string) (*net.IPNet, error) {
	parts := strings.Split(route, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid CIDR format: %s", route)
	}

	ip := net.ParseIP(parts[0])
	if ip == nil {
		return nil, fmt.Errorf("invalid IP: %s", parts[0])
	}

	prefixLen, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid prefix length: %s", parts[1])
	}

	mask := net.CIDRMask(prefixLen, 32)
	return &net.IPNet{IP: ip.Mask(mask), Mask: mask}, nil
}

// Linux 系统路由表检查，使用 ip route 命令分析路由配置，查找 Fake-IP 相关路由
func CheckRouteToFakeIPLinux(fakeCIDR string, suspectIfaces map[string]bool) (bool, error) {
	cmd := exec.Command("ip", "route", "show")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("获取路由表失败: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, fakeCIDR) {
			for ifaceName := range suspectIfaces {
				if strings.Contains(line, " dev "+ifaceName+" ") {
					logger.Logger.Debug("Found route to Fake-IP (Linux)",
						slog.String("route", line),
						slog.String("interface", ifaceName),
					)
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// Windows 系统路由表检查，使用 route print 命令解析路由信息，匹配 Fake-IP 段路由
func CheckRouteToFakeIPWindows(fakeCIDR string, suspectIfaces map[string]bool) (bool, error) {
	cmd := exec.Command("route", "print")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("获取路由表失败: %w", err)
	}

	lines := strings.Split(string(output), "\n")

	fakeCIDRIP := strings.Split(fakeCIDR, "/")[0]

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, fakeCIDRIP) {
			for ifaceName := range suspectIfaces {
				if strings.Contains(line, ifaceName) {
					logger.Logger.Debug("Found route to Fake-IP (macOS)",
						slog.String("route", line),
						slog.String("interface", ifaceName),
					)
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// 检查一个网段是否为另一个网段的子网，用于路由分析时判断网段包含关系
func IsSubnetOf(cidr, parentCIDR string) (bool, error) {
	if !strings.Contains(cidr, "/") {
		cidr = cidr + "/32"
	}

	_, child, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, err
	}

	_, parent, err := net.ParseCIDR(parentCIDR)
	if err != nil {
		return false, err
	}

	childIP := child.IP
	return parent.Contains(childIP), nil
}

// 提取字符串映射的所有键值，辅助函数，用于将接口名称集合转换为数组格式
func MapKeysToSlice(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
