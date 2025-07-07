package speedtester

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/faceair/clash-speedtest/logger"
	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/adapter/provider"
	"github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/log"
	"gopkg.in/yaml.v3"
)

// VlessTestError 代表 vless 测试过程中的详细错误信息
type VlessTestError struct {
	Stage   string `json:"stage"`   // "dns", "connect", "handshake", "transfer"
	Code    string `json:"code"`    // 错误代码
	Message string `json:"message"` // 详细错误信息
	ProxyName string `json:"proxy_name"` // 代理名称
}

// 错误阶段常量
const (
	StageValidation = "validation"
	StageDNS        = "dns"
	StageConnect    = "connect"
	StageHandshake  = "handshake"
	StageTransfer   = "transfer"
)

// 错误代码常量
const (
	ErrorInvalidConfig     = "INVALID_CONFIG"
	ErrorDNSResolution     = "DNS_RESOLUTION_FAILED"
	ErrorConnectionRefused = "CONNECTION_REFUSED"
	ErrorConnectionTimeout = "CONNECTION_TIMEOUT"
	ErrorHandshakeTimeout  = "HANDSHAKE_TIMEOUT"
	ErrorProtocolError     = "PROTOCOL_ERROR"
	ErrorAuthFailed        = "AUTHENTICATION_FAILED"
	ErrorTransferTimeout   = "TRANSFER_TIMEOUT"
	ErrorUnknown          = "UNKNOWN_ERROR"
)

// NewVlessTestError 创建新的 vless 测试错误
func NewVlessTestError(stage, code, message, proxyName string) *VlessTestError {
	return &VlessTestError{
		Stage:     stage,
		Code:      code,
		Message:   message,
		ProxyName: proxyName,
	}
}

// Error 实现 error 接口
func (e *VlessTestError) Error() string {
	return fmt.Sprintf("[%s:%s] %s - %s", e.Stage, e.Code, e.ProxyName, e.Message)
}

type Config struct {
	ConfigPaths      string
	FilterRegex      string
	IncludeNodes     []string
	ExcludeNodes     []string
	ProtocolFilter   []string
	ServerURL        string
	DownloadSize     int
	UploadSize       int
	Timeout          time.Duration
	Concurrent       int
	MaxLatency       time.Duration
	MinDownloadSpeed float64
	MinUploadSpeed   float64
}

type SpeedTester struct {
	config *Config
}

func New(config *Config) *SpeedTester {
	if config.Concurrent <= 0 {
		config.Concurrent = 1
	}
	if config.DownloadSize < 0 {
		config.DownloadSize = 100 * 1024 * 1024
	}
	if config.UploadSize < 0 {
		config.UploadSize = 10 * 1024 * 1024
	}
	return &SpeedTester{
		config: config,
	}
}

type CProxy struct {
	constant.Proxy
	Config map[string]any
}

type RawConfig struct {
	Providers map[string]map[string]any `yaml:"proxy-providers"`
	Proxies   []map[string]any          `yaml:"proxies"`
}

func (st *SpeedTester) LoadProxies(stashCompatible bool) (map[string]*CProxy, error) {
	logger.Logger.Info("Starting proxy loading", 
		slog.String("config_paths", st.config.ConfigPaths),
		slog.Bool("stash_compatible", stashCompatible),
	)
	
	allProxies := make(map[string]*CProxy)
	configPaths := strings.Split(st.config.ConfigPaths, ",")
	
	logger.Logger.Debug("Processing config paths", slog.Int("path_count", len(configPaths)))

	for i, configPath := range configPaths {
		// Trim空格并移除可能的引号
		configPath = strings.TrimSpace(configPath)
		if (strings.HasPrefix(configPath, "\"") && strings.HasSuffix(configPath, "\"")) ||
			(strings.HasPrefix(configPath, "'") && strings.HasSuffix(configPath, "'")) {
			configPath = configPath[1 : len(configPath)-1]
		}
		
		if configPath == "" {
			logger.Logger.Debug("Skipping empty config path", slog.Int("index", i))
			continue
		}

		logger.Logger.Info("Loading config from path", 
			slog.String("path", configPath),
			slog.Int("index", i),
		)

		var body []byte
		var err error
		if strings.HasPrefix(configPath, "http") {
			logger.Logger.Debug("Fetching config via HTTP", slog.String("url", configPath))
			var resp *http.Response
			resp, err = http.Get(configPath)
			if err != nil {
				logger.LogError("Failed to fetch config", err, slog.String("url", configPath))
				log.Warnln("failed to fetch config: %s", err)
				continue
			}
			defer resp.Body.Close()
			
			logger.Logger.Debug("HTTP response received", 
				slog.Int("status_code", resp.StatusCode),
				slog.String("content_type", resp.Header.Get("Content-Type")),
			)
			
			body, err = io.ReadAll(resp.Body)
		} else {
			logger.Logger.Debug("Reading config from file", slog.String("file", configPath))
			body, err = os.ReadFile(configPath)
		}
		if err != nil {
			logger.LogError("Failed to read config", err, slog.String("path", configPath))
			log.Warnln("failed to read config: %s", err)
			continue
		}

		logger.Logger.Debug("Config content loaded", 
			slog.String("path", configPath),
			slog.Int("size_bytes", len(body)),
		)

		// 尝试检测并解码base64编码的配置
		if strings.TrimSpace(string(body)) != "" {
			if decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(body))); err == nil {
				// 检查解码后的内容是否是有效的YAML
				if strings.Contains(string(decoded), "proxies:") || strings.Contains(string(decoded), "proxy-providers:") {
					logger.Logger.Debug("Detected base64 encoded config, decoding")
					body = decoded
				}
			}
		}

		rawCfg := &RawConfig{
			Proxies: []map[string]any{},
		}
		if err := yaml.Unmarshal(body, rawCfg); err != nil {
			logger.LogError("Failed to parse YAML config", err, slog.String("path", configPath))
			return nil, err
		}
		
		logger.Logger.Info("Config parsed successfully",
			slog.String("path", configPath),
			slog.Int("proxy_count", len(rawCfg.Proxies)),
			slog.Int("provider_count", len(rawCfg.Providers)),
		)
		
		proxies := make(map[string]*CProxy)
		proxiesConfig := rawCfg.Proxies
		providersConfig := rawCfg.Providers

		// Process direct proxies
		for i, config := range proxiesConfig {
			proxy, err := adapter.ParseProxy(config)
			if err != nil {
				logger.LogError("Failed to parse proxy", err, 
					slog.Int("proxy_index", i),
					slog.String("config_path", configPath),
				)
				return nil, fmt.Errorf("proxy %d: %w", i, err)
			}

			if _, exist := proxies[proxy.Name()]; exist {
				logger.Logger.Error("Duplicate proxy name found", 
					slog.String("proxy_name", proxy.Name()),
					slog.String("config_path", configPath),
				)
				return nil, fmt.Errorf("proxy %s is the duplicate name", proxy.Name())
			}
			proxies[proxy.Name()] = &CProxy{Proxy: proxy, Config: config}
		}
		
		// Process proxy providers
		for name, config := range providersConfig {
			if name == provider.ReservedName {
				logger.Logger.Error("Reserved provider name used", 
					slog.String("provider_name", name),
					slog.String("reserved_name", provider.ReservedName),
				)
				return nil, fmt.Errorf("can not defined a provider called `%s`", provider.ReservedName)
			}
			
			logger.Logger.Debug("Processing proxy provider", 
				slog.String("provider_name", name),
				slog.String("config_path", configPath),
			)
			
			pd, err := provider.ParseProxyProvider(name, config)
			if err != nil {
				logger.LogError("Failed to parse proxy provider", err, 
					slog.String("provider_name", name),
					slog.String("config_path", configPath),
				)
				return nil, fmt.Errorf("parse proxy provider %s error: %w", name, err)
			}
			if err := pd.Initial(); err != nil {
				logger.LogError("Failed to initialize proxy provider", err, 
					slog.String("provider_name", name),
				)
				return nil, fmt.Errorf("initial proxy provider %s error: %w", pd.Name(), err)
			}

			resp, err := http.Get(config["url"].(string))
			if err != nil {
				logger.LogError("Failed to fetch provider config", err, 
					slog.String("provider_name", name),
					slog.String("provider_url", config["url"].(string)),
				)
				log.Warnln("failed to fetch config: %s", err)
				continue
			}
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				logger.LogError("Failed to read provider response", err, 
					slog.String("provider_name", name),
				)
				return nil, err
			}
			pdRawCfg := &RawConfig{
				Proxies: []map[string]any{},
			}
			if err := yaml.Unmarshal(body, pdRawCfg); err != nil {
				logger.LogError("Failed to parse provider YAML", err, 
					slog.String("provider_name", name),
				)
				return nil, err
			}
			pdProxies := make(map[string]map[string]any)
			for _, pdProxy := range pdRawCfg.Proxies {
				pdProxies[pdProxy["name"].(string)] = pdProxy
			}
			
			providerProxyCount := 0
			for _, proxy := range pd.Proxies() {
				proxies[fmt.Sprintf("[%s] %s", name, proxy.Name())] = &CProxy{
					Proxy:  proxy,
					Config: pdProxies[proxy.Name()],
				}
				providerProxyCount++
			}
			
			logger.Logger.Info("Provider proxies loaded", 
				slog.String("provider_name", name),
				slog.Int("proxy_count", providerProxyCount),
			)
		}
		
		// Filter and add proxies to allProxies
		addedCount := 0
		for k, p := range proxies {
			switch p.Type() {
			case constant.Shadowsocks, constant.ShadowsocksR, constant.Snell, constant.Socks5, constant.Http,
				constant.Vmess, constant.Vless, constant.Trojan, constant.Hysteria, constant.Hysteria2,
				constant.WireGuard, constant.Tuic, constant.Ssh, constant.Mieru, constant.AnyTLS:
			default:
				logger.Logger.Debug("Skipping unsupported proxy type", 
					slog.String("proxy_name", k),
					slog.String("proxy_type", p.Type().String()),
				)
				continue
			}
			if server, ok := p.Config["server"]; ok {
				p.Config["server"] = convertMappedIPv6ToIPv4(server.(string))
			}
			if stashCompatible && !isStashCompatible(p) {
				logger.Logger.Debug("Skipping proxy not compatible with Stash", 
					slog.String("proxy_name", k),
					slog.String("proxy_type", p.Type().String()),
				)
				continue
			}
			if _, ok := allProxies[k]; !ok {
				allProxies[k] = p
				addedCount++
			}
		}
		
		logger.Logger.Info("Proxies processed from config", 
			slog.String("config_path", configPath),
			slog.Int("loaded_count", len(proxies)),
			slog.Int("added_count", addedCount),
		)
	}

	filterRegexp := regexp.MustCompile(st.config.FilterRegex)
	filteredProxies := make(map[string]*CProxy)
	matchedCount := 0
	
	for name := range allProxies {
		proxy := allProxies[name]
		
		// Apply regex filter
		if !filterRegexp.MatchString(name) {
			continue
		}
		
		// Apply include nodes filter
		if len(st.config.IncludeNodes) > 0 {
			includeMatch := false
			for _, include := range st.config.IncludeNodes {
				if strings.TrimSpace(include) == "" {
					continue
				}
				if strings.Contains(strings.ToLower(name), strings.ToLower(strings.TrimSpace(include))) {
					includeMatch = true
					break
				}
			}
			if !includeMatch {
				continue
			}
		}
		
		// Apply exclude nodes filter
		if len(st.config.ExcludeNodes) > 0 {
			excludeMatch := false
			for _, exclude := range st.config.ExcludeNodes {
				if strings.TrimSpace(exclude) == "" {
					continue
				}
				if strings.Contains(strings.ToLower(name), strings.ToLower(strings.TrimSpace(exclude))) {
					excludeMatch = true
					break
				}
			}
			if excludeMatch {
				continue
			}
		}
		
		// Apply protocol filter
		if len(st.config.ProtocolFilter) > 0 {
			protocolMatch := false
			proxyType := proxy.Type().String()
			for _, protocol := range st.config.ProtocolFilter {
				if strings.TrimSpace(protocol) == "" {
					continue
				}
				if strings.EqualFold(proxyType, strings.TrimSpace(protocol)) {
					protocolMatch = true
					break
				}
			}
			if !protocolMatch {
				continue
			}
		}
		
		filteredProxies[name] = proxy
		matchedCount++
	}
	
	logger.Logger.Info("Proxy loading completed",
		slog.Int("total_loaded", len(allProxies)),
		slog.Int("after_filter", len(filteredProxies)),
		slog.String("filter_regex", st.config.FilterRegex),
		slog.Int("matched_filter", matchedCount),
	)
	
	return filteredProxies, nil
}

// GetAvailableProtocols returns all unique protocols from loaded proxies
func (st *SpeedTester) GetAvailableProtocols(proxies map[string]*CProxy) []string {
	protocolSet := make(map[string]bool)
	for _, proxy := range proxies {
		protocolSet[proxy.Type().String()] = true
	}
	
	protocols := make([]string, 0, len(protocolSet))
	for protocol := range protocolSet {
		protocols = append(protocols, protocol)
	}
	
	return protocols
}

func isStashCompatible(proxy *CProxy) bool {
	switch proxy.Type() {
	case constant.Shadowsocks:
		cipher, ok := proxy.Config["cipher"]
		if ok {
			switch cipher {
			case "aes-128-gcm", "aes-192-gcm", "aes-256-gcm",
				"aes-128-cfb", "aes-192-cfb", "aes-256-cfb",
				"aes-128-ctr", "aes-192-ctr", "aes-256-ctr",
				"rc4-md5", "chacha20", "chacha20-ietf", "xchacha20",
				"chacha20-ietf-poly1305", "xchacha20-ietf-poly1305",
				"2022-blake3-aes-128-gcm", "2022-blake3-aes-256-gcm":
			default:
				return false
			}
		}
	case constant.ShadowsocksR:
		if obfs, ok := proxy.Config["obfs"]; ok {
			switch obfs {
			case "plain", "http_simple", "http_post", "random_head",
				"tls1.2_ticket_auth", "tls1.2_ticket_fastauth":
			default:
				return false
			}
		}
		if protocol, ok := proxy.Config["protocol"]; ok {
			switch protocol {
			case "origin", "auth_sha1_v4", "auth_aes128_md5",
				"auth_aes128_sha1", "auth_chain_a", "auth_chain_b":
			default:
				return false
			}
		}
	case constant.Snell:
		if obfsOpts, ok := proxy.Config["obfs-opts"]; ok {
			if obfsOptsMap, ok := obfsOpts.(map[string]any); ok {
				if mode, ok := obfsOptsMap["mode"]; ok {
					switch mode {
					case "http", "tls":
					default:
						return false
					}
				}
			}
		}
	case constant.Socks5, constant.Http:
	case constant.Vmess:
		if cipher, ok := proxy.Config["cipher"]; ok {
			switch cipher {
			case "auto", "aes-128-gcm", "chacha20-poly1305", "none":
			default:
				return false
			}
		}
		if network, ok := proxy.Config["network"]; ok {
			switch network {
			case "ws", "h2", "http", "grpc":
			default:
				return false
			}
		}
	case constant.Vless:
		if flow, ok := proxy.Config["flow"]; ok {
			switch flow {
			case "xtls-rprx-origin", "xtls-rprx-direct", "xtls-rprx-splice", "xtls-rprx-vision":
			default:
				return false
			}
		}
	case constant.Trojan:
		if network, ok := proxy.Config["network"]; ok {
			switch network {
			case "ws", "grpc":
			default:
				return false
			}
		}
	case constant.Hysteria, constant.Hysteria2:
	case constant.WireGuard:
	case constant.Tuic:
	case constant.Ssh:
	default:
		return false
	}
	return true
}

func (st *SpeedTester) TestProxies(proxies map[string]*CProxy, tester func(result *Result)) {
	for name, proxy := range proxies {
		tester(st.testProxy(name, proxy))
	}
}

// TestProxiesWithCallback is an alias for TestProxies for clarity in WebSocket context
func (st *SpeedTester) TestProxiesWithCallback(proxies map[string]*CProxy, callback func(result *Result)) {
	st.TestProxies(proxies, callback)
}

// TestProxiesWithContext tests proxies with context cancellation support
func (st *SpeedTester) TestProxiesWithContext(ctx context.Context, proxies map[string]*CProxy, callback func(result *Result)) error {
	for name, proxy := range proxies {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			logger.Logger.Info("Proxy testing cancelled", slog.String("reason", ctx.Err().Error()))
			return ctx.Err()
		default:
			// Continue with the test
		}
		
		callback(st.testProxy(name, proxy))
	}
	return nil
}

type testJob struct {
	name  string
	proxy *CProxy
}

type Result struct {
	ProxyName     string         `json:"proxy_name"`
	ProxyType     string         `json:"proxy_type"`
	ProxyConfig   map[string]any `json:"proxy_config"`
	Latency       time.Duration  `json:"latency"`
	Jitter        time.Duration  `json:"jitter"`
	PacketLoss    float64        `json:"packet_loss"`
	DownloadSize  float64        `json:"download_size"`
	DownloadTime  time.Duration  `json:"download_time"`
	DownloadSpeed float64        `json:"download_speed"`
	UploadSize    float64        `json:"upload_size"`
	UploadTime    time.Duration  `json:"upload_time"`
	UploadSpeed   float64        `json:"upload_speed"`
	// 新增错误诊断字段
	TestError     *VlessTestError `json:"test_error,omitempty"` // 测试错误详情
	FailureStage  string          `json:"failure_stage,omitempty"` // 失败阶段
	FailureReason string          `json:"failure_reason,omitempty"` // 失败原因
}

func (r *Result) FormatDownloadSpeed() string {
	return formatSpeed(r.DownloadSpeed)
}

// AnalyzeError 分析错误信息并创建 VlessTestError
func AnalyzeError(err error, proxyName string, defaultStage string) *VlessTestError {
	if err == nil {
		return nil
	}

	errorMsg := err.Error()
	stage := defaultStage
	code := ErrorUnknown
	
	// 根据错误信息确定错误类型和阶段
	switch {
	case strings.Contains(errorMsg, "no such host") || strings.Contains(errorMsg, "dns"):
		stage = StageDNS
		code = ErrorDNSResolution
	case strings.Contains(errorMsg, "connection refused"):
		stage = StageConnect
		code = ErrorConnectionRefused
	case strings.Contains(errorMsg, "connect: connection timed out") || strings.Contains(errorMsg, "i/o timeout"):
		stage = StageConnect
		code = ErrorConnectionTimeout
	case strings.Contains(errorMsg, "handshake") || strings.Contains(errorMsg, "tls"):
		stage = StageHandshake
		code = ErrorHandshakeTimeout
	case strings.Contains(errorMsg, "authentication") || strings.Contains(errorMsg, "auth"):
		stage = StageHandshake
		code = ErrorAuthFailed
	case strings.Contains(errorMsg, "protocol") || strings.Contains(errorMsg, "unexpected"):
		stage = StageHandshake
		code = ErrorProtocolError
	case strings.Contains(errorMsg, "read") || strings.Contains(errorMsg, "write") || strings.Contains(errorMsg, "transfer"):
		stage = StageTransfer
		code = ErrorTransferTimeout
	}
	
	return NewVlessTestError(stage, code, errorMsg, proxyName)
}

// IsVlessProtocol 检查是否为 vless 协议
func IsVlessProtocol(proxyType constant.AdapterType) bool {
	return proxyType == constant.Vless
}

func (r *Result) FormatLatency() string {
	if r.Latency == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%dms", r.Latency.Milliseconds())
}

func (r *Result) FormatJitter() string {
	if r.Jitter == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%dms", r.Jitter.Milliseconds())
}

func (r *Result) FormatPacketLoss() string {
	return fmt.Sprintf("%.1f%%", r.PacketLoss)
}

func (r *Result) FormatUploadSpeed() string {
	return formatSpeed(r.UploadSpeed)
}

func formatSpeed(bytesPerSecond float64) string {
	units := []string{"B/s", "KB/s", "MB/s", "GB/s", "TB/s"}
	unit := 0
	speed := bytesPerSecond
	for speed >= 1024 && unit < len(units)-1 {
		speed /= 1024
		unit++
	}
	return fmt.Sprintf("%.2f%s", speed, units[unit])
}

func (st *SpeedTester) testProxy(name string, proxy *CProxy) *Result {
	logger.Logger.Debug("Starting proxy test", 
		slog.String("proxy_name", name),
		slog.String("proxy_type", proxy.Type().String()),
	)
	
	result := &Result{
		ProxyName:   name,
		ProxyType:   proxy.Type().String(),
		ProxyConfig: proxy.Config,
	}

	// 检查是否为 vless 协议，如果是则启用详细错误诊断
	isVless := IsVlessProtocol(proxy.Type())
	if isVless {
		logger.Logger.Debug("Detected vless protocol, enabling enhanced error diagnostics", 
			slog.String("proxy_name", name),
		)
	}

	// 1. 首先进行延迟测试
	logger.Logger.Debug("Testing proxy latency", slog.String("proxy_name", name))
	
	// For VLESS proxies, use a longer timeout for latency testing
	latencyTimeout := st.config.MaxLatency
	if isVless {
		// Use the general timeout setting for VLESS, but ensure it's at least 10 seconds
		if st.config.Timeout > latencyTimeout {
			latencyTimeout = st.config.Timeout
		}
		if latencyTimeout < 10*time.Second {
			latencyTimeout = 10 * time.Second
		}
		logger.Logger.Debug("Using extended timeout for VLESS latency test", 
			slog.String("proxy_name", name),
			slog.String("timeout", latencyTimeout.String()),
		)
	}
	
	latencyResult := st.testLatencyWithErrors(proxy, latencyTimeout, isVless)
	result.Latency = latencyResult.avgLatency
	result.Jitter = latencyResult.jitter
	result.PacketLoss = latencyResult.packetLoss
	
	// 如果是 vless 协议且有错误，记录错误详情
	if isVless && latencyResult.lastError != nil {
		result.TestError = AnalyzeError(latencyResult.lastError, name, StageDNS)
		result.FailureStage = result.TestError.Stage
		result.FailureReason = result.TestError.Message
		
		logger.Logger.Info("Vless latency test failed with detailed error", 
			slog.String("proxy_name", name),
			slog.String("error_stage", result.TestError.Stage),
			slog.String("error_code", result.TestError.Code),
			slog.String("error_message", result.TestError.Message),
		)
	}

	logger.Logger.Debug("Latency test completed", 
		slog.String("proxy_name", name),
		slog.Int64("latency_ms", result.Latency.Milliseconds()),
		slog.Float64("packet_loss", result.PacketLoss),
		slog.Int64("jitter_ms", result.Jitter.Milliseconds()),
	)

	if result.PacketLoss == 100 || result.Latency > st.config.MaxLatency {
		logger.Logger.Info("Proxy failed latency test, skipping speed tests", 
			slog.String("proxy_name", name),
			slog.Float64("packet_loss", result.PacketLoss),
			slog.Int64("latency_ms", result.Latency.Milliseconds()),
			slog.Int64("max_latency_ms", st.config.MaxLatency.Milliseconds()),
		)
		return result
	}

	// 2. 并发进行下载和上传测试
	logger.Logger.Debug("Starting speed tests", 
		slog.String("proxy_name", name),
		slog.Int("concurrent", st.config.Concurrent),
	)

	var wg sync.WaitGroup

	var totalDownloadBytes, totalUploadBytes int64
	var totalDownloadTime, totalUploadTime time.Duration
	var downloadCount, uploadCount int

	downloadChunkSize := st.config.DownloadSize / st.config.Concurrent
	if downloadChunkSize > 0 {
		logger.Logger.Debug("Starting download test", 
			slog.String("proxy_name", name),
			slog.Int("chunk_size_mb", downloadChunkSize/(1024*1024)),
			slog.Int("concurrent", st.config.Concurrent),
		)
		
		downloadResults := make(chan *downloadResult, st.config.Concurrent)

		for i := 0; i < st.config.Concurrent; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				downloadResults <- st.testDownload(proxy, downloadChunkSize, st.config.Timeout)
			}()
		}
		wg.Wait()

		for range st.config.Concurrent {
			if dr := <-downloadResults; dr != nil {
				totalDownloadBytes += dr.bytes
				totalDownloadTime += dr.duration
				downloadCount++
			}
		}
		close(downloadResults)

		if downloadCount > 0 {
			result.DownloadSize = float64(totalDownloadBytes)
			result.DownloadTime = totalDownloadTime / time.Duration(downloadCount)
			result.DownloadSpeed = float64(totalDownloadBytes) / result.DownloadTime.Seconds()
			
			logger.Logger.Debug("Download test completed", 
				slog.String("proxy_name", name),
				slog.Int64("total_bytes", totalDownloadBytes),
				slog.Float64("speed_mbps", result.DownloadSpeed/(1024*1024)),
				slog.Int("successful_downloads", downloadCount),
			)
		}

		if result.DownloadSpeed < st.config.MinDownloadSpeed {
			logger.Logger.Info("Proxy failed minimum download speed requirement", 
				slog.String("proxy_name", name),
				slog.Float64("actual_speed_mbps", result.DownloadSpeed/(1024*1024)),
				slog.Float64("min_speed_mbps", st.config.MinDownloadSpeed/(1024*1024)),
			)
			return result
		}
	}

	uploadChunkSize := st.config.UploadSize / st.config.Concurrent
	if uploadChunkSize > 0 {
		logger.Logger.Debug("Starting upload test", 
			slog.String("proxy_name", name),
			slog.Int("chunk_size_mb", uploadChunkSize/(1024*1024)),
			slog.Int("concurrent", st.config.Concurrent),
		)
		
		uploadResults := make(chan *downloadResult, st.config.Concurrent)

		for i := 0; i < st.config.Concurrent; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				uploadResults <- st.testUpload(proxy, uploadChunkSize, st.config.Timeout)
			}()
		}
		wg.Wait()

		for i := 0; i < st.config.Concurrent; i++ {
			if ur := <-uploadResults; ur != nil {
				totalUploadBytes += ur.bytes
				totalUploadTime += ur.duration
				uploadCount++
			}
		}
		close(uploadResults)

		if uploadCount > 0 {
			result.UploadSize = float64(totalUploadBytes)
			result.UploadTime = totalUploadTime / time.Duration(uploadCount)
			result.UploadSpeed = float64(totalUploadBytes) / result.UploadTime.Seconds()
			
			logger.Logger.Debug("Upload test completed", 
				slog.String("proxy_name", name),
				slog.Int64("total_bytes", totalUploadBytes),
				slog.Float64("speed_mbps", result.UploadSpeed/(1024*1024)),
				slog.Int("successful_uploads", uploadCount),
			)
		}

		if result.UploadSpeed < st.config.MinUploadSpeed {
			logger.Logger.Info("Proxy failed minimum upload speed requirement", 
				slog.String("proxy_name", name),
				slog.Float64("actual_speed_mbps", result.UploadSpeed/(1024*1024)),
				slog.Float64("min_speed_mbps", st.config.MinUploadSpeed/(1024*1024)),
			)
			return result
		}
	}

	logger.Logger.Info("Proxy test completed successfully", 
		slog.String("proxy_name", name),
		slog.Int64("latency_ms", result.Latency.Milliseconds()),
		slog.Float64("download_speed_mbps", result.DownloadSpeed/(1024*1024)),
		slog.Float64("upload_speed_mbps", result.UploadSpeed/(1024*1024)),
		slog.Float64("packet_loss", result.PacketLoss),
	)

	return result
}

type latencyResult struct {
	avgLatency time.Duration
	jitter     time.Duration
	packetLoss float64
	lastError  error // 添加最后一次错误信息
}

func (st *SpeedTester) testLatency(proxy constant.Proxy, minLatency time.Duration) *latencyResult {
	client := st.createClient(proxy, minLatency)
	latencies := make([]time.Duration, 0, 6)
	failedPings := 0

	for i := 0; i < 6; i++ {
		time.Sleep(100 * time.Millisecond)

		start := time.Now()
		resp, err := client.Get(fmt.Sprintf("%s/__down?bytes=0", st.config.ServerURL))
		if err != nil {
			logger.Logger.Debug("Latency test failed", 
				slog.String("proxy_name", proxy.Name()),
				slog.String("proxy_type", proxy.Type().String()),
				slog.Int("attempt", i+1),
				slog.String("error", err.Error()),
			)
			failedPings++
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			latencies = append(latencies, time.Since(start))
		} else {
			logger.Logger.Debug("Latency test received bad status", 
				slog.String("proxy_name", proxy.Name()),
				slog.String("proxy_type", proxy.Type().String()),
				slog.Int("attempt", i+1),
				slog.Int("status_code", resp.StatusCode),
			)
			failedPings++
		}
	}

	return calculateLatencyStats(latencies, failedPings, 6)
}

// testLatencyWithErrors 增强版延迟测试，包含详细错误信息
func (st *SpeedTester) testLatencyWithErrors(proxy constant.Proxy, minLatency time.Duration, captureErrors bool) *latencyResult {
	client := st.createClient(proxy, minLatency)
	latencies := make([]time.Duration, 0, 6)
	failedPings := 0
	var lastError error
	
	// For VLESS proxies, reduce the number of ping attempts to avoid overwhelming slow connections
	pingAttempts := 6
	if captureErrors && proxy.Type() == constant.Vless {
		pingAttempts = 3 // Reduce to 3 attempts for VLESS
		logger.Logger.Debug("Using reduced ping attempts for VLESS", 
			slog.String("proxy_name", proxy.Name()),
			slog.Int("attempts", pingAttempts),
		)
	}

	for i := 0; i < pingAttempts; i++ {
		time.Sleep(100 * time.Millisecond)

		start := time.Now()
		resp, err := client.Get(fmt.Sprintf("%s/__down?bytes=0", st.config.ServerURL))
		if err != nil {
			if captureErrors {
				lastError = err // 保存最后一次错误用于详细分析
				logger.Logger.Debug("Enhanced latency test failed", 
					slog.String("proxy_name", proxy.Name()),
					slog.String("proxy_type", proxy.Type().String()),
					slog.Int("attempt", i+1),
					slog.String("error", err.Error()),
					slog.String("error_type", fmt.Sprintf("%T", err)),
				)
			} else {
				logger.Logger.Debug("Latency test failed", 
					slog.String("proxy_name", proxy.Name()),
					slog.String("proxy_type", proxy.Type().String()),
					slog.Int("attempt", i+1),
					slog.String("error", err.Error()),
				)
			}
			failedPings++
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			latencies = append(latencies, time.Since(start))
		} else {
			if captureErrors && lastError == nil {
				lastError = fmt.Errorf("HTTP status %d", resp.StatusCode)
			}
			logger.Logger.Debug("Latency test received bad status", 
				slog.String("proxy_name", proxy.Name()),
				slog.String("proxy_type", proxy.Type().String()),
				slog.Int("attempt", i+1),
				slog.Int("status_code", resp.StatusCode),
			)
			failedPings++
		}
	}

	result := calculateLatencyStats(latencies, failedPings, pingAttempts)
	if captureErrors {
		result.lastError = lastError
	}
	return result
}

type downloadResult struct {
	bytes    int64
	duration time.Duration
}

func (st *SpeedTester) testDownload(proxy constant.Proxy, size int, timeout time.Duration) *downloadResult {
	client := st.createClient(proxy, timeout)
	start := time.Now()

	logger.Logger.Debug("Starting download test request", 
		slog.String("server_url", st.config.ServerURL),
		slog.Int("size_bytes", size),
		slog.String("timeout", timeout.String()),
	)

	resp, err := client.Get(fmt.Sprintf("%s/__down?bytes=%d", st.config.ServerURL, size))
	if err != nil {
		logger.Logger.Debug("Download test request failed", 
			slog.String("error", err.Error()),
			slog.Int("size_bytes", size),
		)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Logger.Debug("Download test received non-200 status", 
			slog.Int("status_code", resp.StatusCode),
			slog.Int("size_bytes", size),
		)
		return nil
	}

	downloadBytes, _ := io.Copy(io.Discard, resp.Body)
	duration := time.Since(start)

	logger.Logger.Debug("Download test completed", 
		slog.Int64("downloaded_bytes", downloadBytes),
		slog.String("duration", duration.String()),
		slog.Float64("speed_mbps", float64(downloadBytes)/duration.Seconds()/(1024*1024)),
	)

	return &downloadResult{
		bytes:    downloadBytes,
		duration: duration,
	}
}

func (st *SpeedTester) testUpload(proxy constant.Proxy, size int, timeout time.Duration) *downloadResult {
	client := st.createClient(proxy, timeout)
	reader := NewZeroReader(size)

	logger.Logger.Debug("Starting upload test request", 
		slog.String("server_url", st.config.ServerURL),
		slog.Int("size_bytes", size),
		slog.String("timeout", timeout.String()),
	)

	start := time.Now()
	resp, err := client.Post(
		fmt.Sprintf("%s/__up", st.config.ServerURL),
		"application/octet-stream",
		reader,
	)
	if err != nil {
		logger.Logger.Debug("Upload test request failed", 
			slog.String("error", err.Error()),
			slog.Int("size_bytes", size),
		)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Logger.Debug("Upload test received non-200 status", 
			slog.Int("status_code", resp.StatusCode),
			slog.Int("size_bytes", size),
		)
		return nil
	}

	duration := time.Since(start)
	uploadedBytes := reader.WrittenBytes()

	logger.Logger.Debug("Upload test completed", 
		slog.Int64("uploaded_bytes", uploadedBytes),
		slog.String("duration", duration.String()),
		slog.Float64("speed_mbps", float64(uploadedBytes)/duration.Seconds()/(1024*1024)),
	)

	return &downloadResult{
		bytes:    uploadedBytes,
		duration: duration,
	}
}

func (st *SpeedTester) createClient(proxy constant.Proxy, timeout time.Duration) *http.Client {
	// For VLESS proxies, configure more conservative connection settings
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				logger.Logger.Debug("Failed to parse address", 
					slog.String("proxy_name", proxy.Name()),
					slog.String("proxy_type", proxy.Type().String()),
					slog.String("addr", addr),
					slog.String("error", err.Error()),
				)
				return nil, err
			}
			var u16Port uint16
			if port, err := strconv.ParseUint(port, 10, 16); err == nil {
				u16Port = uint16(port)
			}
			
			logger.Logger.Debug("Attempting connection via proxy", 
				slog.String("proxy_name", proxy.Name()),
				slog.String("proxy_type", proxy.Type().String()),
				slog.String("target_host", host),
				slog.Int("target_port", int(u16Port)),
			)
			
			conn, err := proxy.DialContext(ctx, &constant.Metadata{
				Host:    host,
				DstPort: u16Port,
			})
			
			if err != nil {
				logger.Logger.Debug("Connection failed via proxy", 
					slog.String("proxy_name", proxy.Name()),
					slog.String("proxy_type", proxy.Type().String()),
					slog.String("target_host", host),
					slog.Int("target_port", int(u16Port)),
					slog.String("error", err.Error()),
				)
				return nil, err
			}
			
			logger.Logger.Debug("Connection successful via proxy", 
				slog.String("proxy_name", proxy.Name()),
				slog.String("proxy_type", proxy.Type().String()),
				slog.String("target_host", host),
				slog.Int("target_port", int(u16Port)),
			)
			
			return conn, nil
		},
	}
	
	// For VLESS proxies, use more conservative timeouts
	if proxy.Type() == constant.Vless {
		transport.TLSHandshakeTimeout = timeout
		transport.ResponseHeaderTimeout = timeout
		transport.ExpectContinueTimeout = timeout / 2
		
		logger.Logger.Debug("Using VLESS-optimized transport settings", 
			slog.String("proxy_name", proxy.Name()),
			slog.String("timeout", timeout.String()),
		)
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

func calculateLatencyStats(latencies []time.Duration, failedPings int, totalAttempts int) *latencyResult {
	result := &latencyResult{
		packetLoss: float64(failedPings) / float64(totalAttempts) * 100,
	}

	if len(latencies) == 0 {
		return result
	}

	// 计算平均延迟
	var total time.Duration
	for _, l := range latencies {
		total += l
	}
	result.avgLatency = total / time.Duration(len(latencies))

	// 计算抖动
	var variance float64
	for _, l := range latencies {
		diff := float64(l - result.avgLatency)
		variance += diff * diff
	}
	variance /= float64(len(latencies))
	result.jitter = time.Duration(math.Sqrt(variance))

	return result
}

func convertMappedIPv6ToIPv4(server string) string {
	ip := net.ParseIP(server)
	if ip == nil {
		return server
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		return ipv4.String()
	}
	return server
}
