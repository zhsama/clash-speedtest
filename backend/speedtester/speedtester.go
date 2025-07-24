package speedtester

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/zhsama/clash-speedtest/logger"
	"github.com/zhsama/clash-speedtest/unlock"
	"github.com/metacubex/mihomo/constant"
)

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

	st := &SpeedTester{
		config: config,
	}

	if config.UnlockConfig != nil && config.UnlockConfig.Enabled {
		logger.Logger.Debug("Initializing unlock detector",
			slog.Int("platforms", len(config.UnlockConfig.Platforms)),
			slog.Int("concurrent", config.UnlockConfig.Concurrent),
		)
		
		st.unlockDetector = unlock.NewDetector(config.UnlockConfig)
		
		if st.unlockDetector != nil {
			logger.Logger.Info("Unlock detector initialized successfully",
				slog.Int("platforms", len(config.UnlockConfig.Platforms)),
				slog.Int("concurrent", config.UnlockConfig.Concurrent),
			)
		} else {
			logger.Logger.Error("Failed to initialize unlock detector")
		}
	}

	return st
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

// FrontendUnlockResult 前端期望的解锁结果格式
type FrontendUnlockResult struct {
	Platform     string `json:"platform"`
	Supported    bool   `json:"supported"`
	Region       string `json:"region,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// FrontendUnlockSummary 前端期望的解锁摘要格式
type FrontendUnlockSummary struct {
	SupportedPlatforms   []string `json:"supported_platforms"`
	UnsupportedPlatforms []string `json:"unsupported_platforms"`
	TotalTested          int      `json:"total_tested"`
	TotalSupported       int      `json:"total_supported"`
}

type Result struct {
	ProxyName     string         `json:"proxy_name"`
	ProxyType     string         `json:"proxy_type"`
	ProxyConfig   map[string]any `json:"proxy_config"`
	ProxyIP       string         `json:"proxy_ip"` // 新增代理IP地址
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
	TestError     *VlessTestError `json:"test_error,omitempty"`     // 测试错误详情
	FailureStage  string          `json:"failure_stage,omitempty"`  // 失败阶段
	FailureReason string          `json:"failure_reason,omitempty"` // 失败原因
	// 新增解锁检测结果字段 - 前端兼容格式
	UnlockResults []FrontendUnlockResult `json:"unlock_results,omitempty"` // 解锁检测结果（前端格式）
	UnlockSummary FrontendUnlockSummary  `json:"unlock_summary,omitempty"` // 解锁摘要（前端格式）
}

func (r *Result) FormatDownloadSpeed() string {
	return formatSpeed(r.DownloadSpeed)
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
		slog.String("test_mode", st.config.TestMode),
	)

	result := &Result{
		ProxyName:   name,
		ProxyType:   proxy.Type().String(),
		ProxyConfig: proxy.Config,
	}

	// Extract proxy IP address from config
	if server, ok := proxy.Config["server"]; ok {
		result.ProxyIP = server.(string)
	}

	// 检查是否为 vless 协议，如果是则启用详细错误诊断
	isVless := IsVlessProtocol(proxy.Type())
	if isVless {
		logger.Logger.Debug("Detected vless protocol, enabling enhanced error diagnostics",
			slog.String("proxy_name", name),
		)
	}

	// 根据测试模式执行不同的测试
	testMode := st.config.TestMode
	if testMode == "" {
		testMode = "both" // 默认两者都测试
	}

	// 1. 延迟测试（除非是仅解锁模式）
	if testMode != "unlock_only" {
		latencyResult := st.testLatencyWithErrors(proxy, st.config.MaxLatency, isVless)
		result.Latency = latencyResult.avgLatency
		result.Jitter = latencyResult.jitter
		result.PacketLoss = latencyResult.packetLoss

		// 如果延迟测试失败，且不是快速模式或解锁优先模式，则跳过后续测试
		if testMode == "speed_only" && (result.PacketLoss == 100 || result.Latency > st.config.MaxLatency) {
			logger.Logger.Info("Proxy failed latency test, skipping speed tests",
				slog.String("proxy_name", name),
				slog.Float64("packet_loss", result.PacketLoss),
				slog.Int64("latency_ms", result.Latency.Milliseconds()),
				slog.Int64("max_latency_ms", st.config.MaxLatency.Milliseconds()),
			)
			return result
		}
	}

	// 2. 解锁检测（除非是仅测速模式）
	if testMode != "speed_only" && st.unlockDetector != nil {
		unlockResults := st.unlockDetector.DetectAll(proxy.Proxy, st.config.UnlockConfig.Platforms)
		result.UnlockResults = convertToFrontendUnlockResults(unlockResults)
		result.UnlockSummary = generateFrontendUnlockSummary(unlockResults)

		logger.Logger.Info("Unlock detection completed",
			slog.String("proxy_name", name),
			slog.Int("detected_platforms", len(unlockResults)),
			slog.Int("supported_platforms", result.UnlockSummary.TotalSupported),
		)

		// 如果是仅解锁模式，直接返回结果
		if testMode == "unlock_only" {
			return result
		}
	} else if testMode != "speed_only" {
		logger.Logger.Warn("Unlock detection requested but detector not initialized",
			slog.String("proxy_name", name),
			slog.String("test_mode", testMode),
			slog.Bool("detector_nil", st.unlockDetector == nil),
		)
	}

	// 3. 速度测试（除非是仅解锁模式或快速模式）
	if testMode != "unlock_only" && !st.config.FastMode {
		// 检查延迟是否满足要求（如果进行了延迟测试）
		if testMode != "unlock_only" && (result.PacketLoss == 100 || result.Latency > st.config.MaxLatency) {
			logger.Logger.Info("Proxy failed latency test, skipping speed tests",
				slog.String("proxy_name", name),
				slog.Float64("packet_loss", result.PacketLoss),
				slog.Int64("latency_ms", result.Latency.Milliseconds()),
				slog.Int64("max_latency_ms", st.config.MaxLatency.Milliseconds()),
			)
			return result
		}

		// 进行速度测试
		st.performSpeedTests(proxy, result, isVless, name)
	}

	logger.Logger.Info("Proxy test completed successfully",
		slog.String("proxy_name", name),
		slog.Int64("latency_ms", result.Latency.Milliseconds()),
		slog.Float64("download_speed_mbps", result.DownloadSpeed/(1024*1024)),
		slog.Float64("upload_speed_mbps", result.UploadSpeed/(1024*1024)),
		slog.Float64("packet_loss", result.PacketLoss),
		slog.Int("supported_platforms", result.UnlockSummary.TotalSupported),
	)

	return result
}

// performSpeedTests 执行速度测试
func (st *SpeedTester) performSpeedTests(proxy *CProxy, result *Result, isVless bool, name string) {
	// 并发进行下载和上传测试
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
			return
		}
	}

	uploadChunkSize := st.config.UploadSize / st.config.Concurrent
	if uploadChunkSize > 0 {
		uploadConcurrent := st.config.Concurrent
		if isVless && uploadConcurrent > 3 {
			uploadConcurrent = 3
			uploadChunkSize = st.config.UploadSize / uploadConcurrent
			logger.Logger.Debug("Adjusting upload concurrency for VLESS",
				slog.String("proxy_name", name),
				slog.Int("concurrent", uploadConcurrent),
			)
		}

		logger.Logger.Debug("Starting upload test",
			slog.String("proxy_name", name),
			slog.Int("chunk_size_mb", uploadChunkSize/(1024*1024)),
			slog.Int("concurrent", uploadConcurrent),
		)

		uploadResults := make(chan *downloadResult, uploadConcurrent)

		for i := 0; i < uploadConcurrent; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				uploadResults <- st.testUpload(proxy, uploadChunkSize, st.config.Timeout)
			}()
		}
		wg.Wait()

		var failedUploads int
		for i := 0; i < uploadConcurrent; i++ {
			if ur := <-uploadResults; ur != nil {
				totalUploadBytes += ur.bytes
				totalUploadTime += ur.duration
				uploadCount++
			} else {
				failedUploads++
			}
		}

		logger.Logger.Info("Upload test results summary",
			slog.String("proxy_name", name),
			slog.String("proxy_type", proxy.Type().String()),
			slog.Int("total_attempts", uploadConcurrent),
			slog.Int("successful_uploads", uploadCount),
			slog.Int("failed_uploads", failedUploads),
			slog.Int("chunk_size_mb", uploadChunkSize/(1024*1024)),
		)
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
		} else {
			logger.Logger.Warn("All upload tests failed",
				slog.String("proxy_name", name),
				slog.String("proxy_type", proxy.Type().String()),
				slog.Int("total_attempts", uploadConcurrent),
				slog.Int("chunk_size_mb", uploadChunkSize/(1024*1024)),
				slog.String("server_url", st.config.ServerURL),
				slog.String("timeout", st.config.Timeout.String()),
				slog.String("possible_causes", "network timeout, proxy connection issues, server errors, or protocol incompatibility"),
			)
		}

		if result.UploadSpeed < st.config.MinUploadSpeed {
			logger.Logger.Info("Proxy failed minimum upload speed requirement",
				slog.String("proxy_name", name),
				slog.Float64("actual_speed_mbps", result.UploadSpeed/(1024*1024)),
				slog.Float64("min_speed_mbps", st.config.MinUploadSpeed/(1024*1024)),
			)
			return
		}
	}
}

type latencyResult struct {
	avgLatency time.Duration
	jitter     time.Duration
	packetLoss float64
	lastError  error // 添加最后一次错误信息
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

	for i := range pingAttempts {
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

	// 对于VLESS代理，使用更保守的上传策略
	isVless := proxy.Type() == constant.Vless
	var reader io.Reader

	if isVless {
		chunkSize := 256 * 1024
		delayBetween := 1 * time.Millisecond
		reader = NewChunkedZeroReader(size, chunkSize, delayBetween)
		logger.Logger.Debug("Using chunked reader for VLESS upload",
			slog.String("proxy_name", proxy.Name()),
			slog.Int("chunk_size", chunkSize),
			slog.String("delay", delayBetween.String()),
		)
	} else {
		reader = NewZeroReader(size)
	}

	logger.Logger.Debug("Starting upload test request",
		slog.String("proxy_name", proxy.Name()),
		slog.String("proxy_type", proxy.Type().String()),
		slog.String("server_url", st.config.ServerURL),
		slog.Int("size_bytes", size),
		slog.String("timeout", timeout.String()),
	)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/__up", st.config.ServerURL), reader)
	if err != nil {
		logger.Logger.Warn("Failed to create upload request",
			slog.String("proxy_name", proxy.Name()),
			slog.String("proxy_type", proxy.Type().String()),
			slog.String("error", err.Error()),
			slog.String("error_type", fmt.Sprintf("%T", err)),
			slog.Int("size_bytes", size),
		)
		return nil
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", strconv.Itoa(size))

	if isVless {
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Transfer-Encoding", "")
		req.Header.Set("Expect", "")
		logger.Logger.Debug("Using VLESS-optimized upload settings",
			slog.String("proxy_name", proxy.Name()),
		)
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		logger.Logger.Warn("Upload test request failed",
			slog.String("proxy_name", proxy.Name()),
			slog.String("proxy_type", proxy.Type().String()),
			slog.String("error", err.Error()),
			slog.String("error_type", fmt.Sprintf("%T", err)),
			slog.Int("size_bytes", size),
			slog.String("server_url", st.config.ServerURL),
			slog.String("timeout", timeout.String()),
		)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody := make([]byte, 256)
		n, _ := resp.Body.Read(respBody)
		logger.Logger.Warn("Upload test received non-200 status",
			slog.String("proxy_name", proxy.Name()),
			slog.String("proxy_type", proxy.Type().String()),
			slog.Int("status_code", resp.StatusCode),
			slog.String("response", string(respBody[:n])),
			slog.Int("size_bytes", size),
			slog.String("server_url", st.config.ServerURL),
		)
		return nil
	}

	duration := time.Since(start)
	var uploadedBytes int64
	if isVless {
		if czr, ok := reader.(*ChunkedZeroReader); ok {
			uploadedBytes = czr.WrittenBytes()
		} else {
			uploadedBytes = int64(size)
		}
	} else {
		if zr, ok := reader.(*ZeroReader); ok {
			uploadedBytes = zr.WrittenBytes()
		} else {
			uploadedBytes = int64(size)
		}
	}

	logger.Logger.Debug("Upload test completed",
		slog.String("proxy_name", proxy.Name()),
		slog.String("proxy_type", proxy.Type().String()),
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
					slog.String("error_type", fmt.Sprintf("%T", err)),
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

// convertToFrontendUnlockResults 将后端unlock结果转换为前端期望的格式
func convertToFrontendUnlockResults(backendResults []unlock.UnlockResult) []FrontendUnlockResult {
	frontendResults := make([]FrontendUnlockResult, len(backendResults))
	for i, result := range backendResults {
		frontendResults[i] = FrontendUnlockResult{
			Platform:     result.Platform,
			Supported:    result.Status == unlock.StatusUnlocked,
			Region:       result.Region,
			ErrorMessage: result.Message,
		}
	}
	return frontendResults
}

// generateFrontendUnlockSummary 生成前端期望的unlock摘要格式
func generateFrontendUnlockSummary(backendResults []unlock.UnlockResult) FrontendUnlockSummary {
	var supported, unsupported []string
	
	for _, result := range backendResults {
		if result.Status == unlock.StatusUnlocked {
			if result.Region != "" {
				supported = append(supported, fmt.Sprintf("%s:%s", result.Platform, result.Region))
			} else {
				supported = append(supported, result.Platform)
			}
		} else {
			unsupported = append(unsupported, result.Platform)
		}
	}
	
	return FrontendUnlockSummary{
		SupportedPlatforms:   supported,
		UnsupportedPlatforms: unsupported,
		TotalTested:          len(backendResults),
		TotalSupported:       len(supported),
	}
}
