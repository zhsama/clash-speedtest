package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/faceair/clash-speedtest/logger"
	"github.com/faceair/clash-speedtest/speedtester"
	"github.com/faceair/clash-speedtest/unlock"
	"github.com/faceair/clash-speedtest/utils"
	"github.com/faceair/clash-speedtest/websocket"
	"github.com/metacubex/mihomo/log"
)

type TestRequest struct {
	ConfigPaths      string   `json:"configPaths"`
	FilterRegex      string   `json:"filterRegex"`
	IncludeNodes     []string `json:"includeNodes"`
	ExcludeNodes     []string `json:"excludeNodes"`
	ProtocolFilter   []string `json:"protocolFilter"`
	ServerURL        string   `json:"serverUrl"`
	DownloadSize     int      `json:"downloadSize"`
	UploadSize       int      `json:"uploadSize"`
	Timeout          int      `json:"timeout"`
	Concurrent       int      `json:"concurrent"`
	MaxLatency       int      `json:"maxLatency"`
	MinDownloadSpeed float64  `json:"minDownloadSpeed"`
	MinUploadSpeed   float64  `json:"minUploadSpeed"`
	StashCompatible  bool     `json:"stashCompatible"`
	// 新增字段
	FastMode         bool     `json:"fastMode"`        // 快速模式：只测试延迟
	RenameNodes      bool     `json:"renameNodes"`     // 节点重命名：添加地理位置信息
	ExportFormat     string   `json:"exportFormat"`    // 导出格式：json, csv, yaml, clash
	ExportPath       string   `json:"exportPath"`      // 导出路径
	// 解锁检测相关字段
	TestMode         string   `json:"testMode"`        // 测试模式：speed_only, unlock_only, both
	UnlockEnabled    bool     `json:"unlockEnabled"`   // 是否启用解锁检测
	UnlockPlatforms  []string `json:"unlockPlatforms"` // 要检测的平台列表
	UnlockConcurrent int      `json:"unlockConcurrent"` // 解锁检测并发数
	UnlockTimeout    int      `json:"unlockTimeout"`   // 解锁检测超时时间
	UnlockRetry      bool     `json:"unlockRetry"`     // 解锁检测失败时是否重试
}

type TestResponse struct {
	Success bool                  `json:"success"`
	Error   string                `json:"error,omitempty"`
	Results []*speedtester.Result `json:"results,omitempty"`
}

var wsHub *websocket.Hub
var testCancelFunc context.CancelFunc
var testCancelMutex sync.RWMutex

// 任务管理
type TestTask struct {
	ID         string
	Config     TestRequest
	Context    context.Context
	CancelFunc context.CancelFunc
	Status     string // pending, running, completed, cancelled
	StartTime  time.Time
}

var (
	testTasks      = make(map[string]*TestTask)
	testTasksMutex sync.RWMutex
)

// createUnlockConfig 根据TestRequest创建解锁检测配置
func createUnlockConfig(req TestRequest) *unlock.UnlockTestConfig {
	if !req.UnlockEnabled {
		return &unlock.UnlockTestConfig{
			Enabled: false,
		}
	}
	
	// 设置默认值
	platforms := req.UnlockPlatforms
	if len(platforms) == 0 {
		platforms = []string{"Netflix", "YouTube", "Disney+", "ChatGPT", "Spotify", "Bilibili"}
	}
	
	concurrent := req.UnlockConcurrent
	if concurrent <= 0 {
		concurrent = 5
	}
	
	timeout := req.UnlockTimeout
	if timeout <= 0 {
		timeout = 10
	}
	
	return &unlock.UnlockTestConfig{
		Enabled:       true,
		Platforms:     platforms,
		Concurrent:    concurrent,
		Timeout:       timeout,
		RetryOnError:  req.UnlockRetry,
		IncludeIPInfo: true,
	}
}

func main() {
	// Initialize custom logging configuration
	logConfig := logger.DefaultLogConfig()
	
	// Override from environment variables
	if logDir := os.Getenv("LOG_DIR"); logDir != "" {
		logConfig.LogDir = logDir
	}
	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		logConfig.LogFileName = logFile
	}
	if os.Getenv("LOG_TO_FILE") == "false" {
		logConfig.OutputToFile = false
	}
	if os.Getenv("LOG_TO_CONSOLE") == "false" {
		logConfig.EnableConsole = false
	}
	
	// Initialize logger with custom config
	logger.InitLoggerWithConfig(logConfig)
	
	// Ensure proper cleanup on exit
	defer logger.Cleanup()
	
	// Enable mihomo logs for debugging
	log.SetLevel(log.DEBUG)
	
	logger.Logger.Info("Starting Clash SpeedTest API Server",
		slog.String("version", "2.0.0"),
		slog.String("port", "8080"),
	)

	// Initialize WebSocket hub
	wsHub = websocket.NewHub()
	// Set message handler for stop test
	wsHub.SetMessageHandler(handleWebSocketMessage)
	go wsHub.Run()

	http.HandleFunc("/api/test", loggingMiddleware(handleTestWithWebSocket))
	http.HandleFunc("/api/test/async", loggingMiddleware(handleTestAsync))
	http.HandleFunc("/api/nodes", loggingMiddleware(handleGetNodes))
	http.HandleFunc("/api/protocols", loggingMiddleware(handleGetProtocols))
	http.HandleFunc("/api/export", loggingMiddleware(handleExportResults))
	http.HandleFunc("/api/logs", loggingMiddleware(handleLogManagement))
	http.HandleFunc("/api/health", loggingMiddleware(handleHealth))
	http.HandleFunc("/ws", loggingMiddleware(wsHub.HandleWebSocket))

	// Enable CORS
	handler := corsMiddleware(http.DefaultServeMux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() {
		logger.Logger.Info("HTTP server starting", slog.String("address", server.Addr))
		fmt.Println("Speed test API server starting on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError("Failed to start server", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Logger.Info("Received shutdown signal, gracefully shutting down server")
	fmt.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.LogError("Server forced to shutdown", err)
		os.Exit(1)
	}
	
	logger.Logger.Info("Server gracefully shut down")
	fmt.Println("Server exited")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Skip logging wrapper for WebSocket routes to allow hijacking
		if r.URL.Path == "/ws" {
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, 200, duration.String())
			return
		}
		
		// Create a custom ResponseWriter to capture status code for non-WebSocket routes
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(lrw, r)
		
		duration := time.Since(start)
		logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, lrw.statusCode, duration.String())
	}
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Debug("Health check requested")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// 生成任务ID
func generateTaskID() string {
	return fmt.Sprintf("task-%d-%s", time.Now().Unix(), time.Now().Format("150405"))
}

// 异步测试处理
func handleTestAsync(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Async test request received")
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.LogError("Failed to decode request body", err)
		sendError(w, "Invalid request body: "+err.Error())
		return
	}
	
	// 设置默认值
	if req.FilterRegex == "" {
		req.FilterRegex = ".+"
	}
	if req.ServerURL == "" {
		req.ServerURL = "https://speed.cloudflare.com"
	}
	if req.DownloadSize == 0 {
		req.DownloadSize = 50
	}
	if req.UploadSize == 0 {
		req.UploadSize = 20
	}
	if req.Timeout == 0 {
		req.Timeout = 5
	}
	if req.Concurrent == 0 {
		req.Concurrent = 4
	}
	if req.MaxLatency == 0 {
		req.MaxLatency = 800
	}
	// 保持用户设置的0值，表示不限制速度
	
	// 创建任务
	taskID := generateTaskID()
	ctx, cancel := context.WithCancel(context.Background())
	
	task := &TestTask{
		ID:         taskID,
		Config:     req,
		Context:    ctx,
		CancelFunc: cancel,
		Status:     "pending",
		StartTime:  time.Now(),
	}
	
	testTasksMutex.Lock()
	testTasks[taskID] = task
	testTasksMutex.Unlock()
	
	// 异步执行测试
	go runTestTask(task)
	
	// 返回任务ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"taskId":  taskID,
		"message": "Test task created successfully",
	})
}

// 执行测试任务
func runTestTask(task *TestTask) {
	// 更新任务状态
	testTasksMutex.Lock()
	task.Status = "running"
	testTasksMutex.Unlock()
	
	logger.Logger.Info("Starting test task", slog.String("task_id", task.ID))
	
	// 创建SpeedTester实例
	unlockConfig := createUnlockConfig(task.Config)
	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths:      task.Config.ConfigPaths,
		FilterRegex:      task.Config.FilterRegex,
		IncludeNodes:     task.Config.IncludeNodes,
		ExcludeNodes:     task.Config.ExcludeNodes,
		ProtocolFilter:   task.Config.ProtocolFilter,
		ServerURL:        task.Config.ServerURL,
		DownloadSize:     task.Config.DownloadSize * 1024 * 1024,
		UploadSize:       task.Config.UploadSize * 1024 * 1024,
		Timeout:          time.Duration(task.Config.Timeout) * time.Second,
		Concurrent:       task.Config.Concurrent,
		MaxLatency:       time.Duration(task.Config.MaxLatency) * time.Millisecond,
		MinDownloadSpeed: task.Config.MinDownloadSpeed * 1024 * 1024,
		MinUploadSpeed:   task.Config.MinUploadSpeed * 1024 * 1024,
		FastMode:         task.Config.FastMode,
		RenameNodes:      task.Config.RenameNodes,
		TestMode:         task.Config.TestMode,
		UnlockConfig:     unlockConfig,
	})
	
	// 加载代理
	allProxies, err := speedTester.LoadProxies(task.Config.StashCompatible)
	if err != nil {
		logger.LogError("Failed to load proxies", err)
		wsHub.BroadcastMessage(websocket.MessageTypeError, websocket.ErrorData{
			Message: "Failed to load proxies: " + err.Error(),
			Code:    "PROXY_LOAD_ERROR",
		})
		
		testTasksMutex.Lock()
		task.Status = "failed"
		testTasksMutex.Unlock()
		return
	}
	
	if len(allProxies) == 0 {
		wsHub.BroadcastMessage(websocket.MessageTypeError, websocket.ErrorData{
			Message: "No proxies found",
			Code:    "NO_PROXIES_FOUND",
		})
		
		testTasksMutex.Lock()
		task.Status = "failed"
		testTasksMutex.Unlock()
		return
	}
	
	// 发送测试开始消息
	testStartData := websocket.TestStartData{
		TotalProxies: len(allProxies),
	}
	testStartData.Config.ConfigPaths = task.Config.ConfigPaths
	testStartData.Config.FilterRegex = task.Config.FilterRegex
	testStartData.Config.ServerURL = task.Config.ServerURL
	testStartData.Config.DownloadSize = task.Config.DownloadSize
	testStartData.Config.UploadSize = task.Config.UploadSize
	testStartData.Config.Timeout = task.Config.Timeout
	testStartData.Config.Concurrent = task.Config.Concurrent
	testStartData.Config.MaxLatency = task.Config.MaxLatency
	testStartData.Config.MinDownloadSpeed = task.Config.MinDownloadSpeed
	testStartData.Config.MinUploadSpeed = task.Config.MinUploadSpeed
	testStartData.Config.StashCompatible = task.Config.StashCompatible
	
	wsHub.BroadcastMessage(websocket.MessageTypeTestStart, testStartData)
	
	// 执行测试
	results := make([]*speedtester.Result, 0)
	completed := 0
	successful := 0
	failed := 0
	
	err = speedTester.TestProxiesWithContext(task.Context, allProxies, func(result *speedtester.Result) {
		results = append(results, result)
		completed++
		
		// 判断结果状态
		status := "success"
		if result.PacketLoss == 100 || result.Latency > time.Duration(task.Config.MaxLatency)*time.Millisecond {
			status = "failed"
			failed++
		} else if result.DownloadSpeed < task.Config.MinDownloadSpeed*1024*1024 || result.UploadSpeed < task.Config.MinUploadSpeed*1024*1024 {
			status = "failed"
			failed++
		} else {
			successful++
		}
		
		// 发送进度更新
		progressData := websocket.TestProgressData{
			CurrentProxy:    result.ProxyName,
			CompletedCount:  completed,
			TotalCount:      len(allProxies),
			ProgressPercent: float64(completed) / float64(len(allProxies)) * 100,
			Status:          status,
		}
		wsHub.BroadcastMessage(websocket.MessageTypeTestProgress, progressData)
		
		// 发送单个结果
		resultData := websocket.TestResultData{
			ProxyName:         result.ProxyName,
			ProxyType:         result.ProxyType,
			ProxyIP:           result.ProxyIP,
			Latency:           result.Latency.Milliseconds(),
			Jitter:            result.Jitter.Milliseconds(),
			PacketLoss:        result.PacketLoss,
			DownloadSpeed:     result.DownloadSpeed,
			UploadSpeed:       result.UploadSpeed,
			DownloadSpeedMbps: result.DownloadSpeed / (1024 * 1024),
			UploadSpeedMbps:   result.UploadSpeed / (1024 * 1024),
			Status:            status,
		}
		
		if result.TestError != nil {
			resultData.ErrorStage = result.TestError.Stage
			resultData.ErrorCode = result.TestError.Code
			resultData.ErrorMessage = result.TestError.Message
		} else if result.FailureStage != "" {
			resultData.ErrorStage = result.FailureStage
			resultData.ErrorMessage = result.FailureReason
		}
		
		wsHub.BroadcastMessage(websocket.MessageTypeTestResult, resultData)
	})
	
	testDuration := time.Since(task.StartTime)
	
	// 检查是否被取消
	if err != nil && err == context.Canceled {
		logger.Logger.Info("Test task cancelled", slog.String("task_id", task.ID))
		
		cancelData := websocket.TestCancelledData{
			Message:         "测试已被用户取消",
			CompletedTests:  completed,
			TotalTests:      len(allProxies),
			PartialDuration: testDuration.String(),
		}
		wsHub.BroadcastMessage(websocket.MessageTypeTestCancelled, cancelData)
		
		testTasksMutex.Lock()
		task.Status = "cancelled"
		testTasksMutex.Unlock()
		return
	}
	
	// 发送测试完成消息
	var totalLatency, totalDownload, totalUpload float64
	bestProxy := ""
	bestDownloadSpeed := 0.0
	
	for _, result := range results {
		if result.PacketLoss < 100 {
			totalLatency += float64(result.Latency.Milliseconds())
			totalDownload += result.DownloadSpeed / (1024 * 1024)
			totalUpload += result.UploadSpeed / (1024 * 1024)
			
			downloadMbps := result.DownloadSpeed / (1024 * 1024)
			if downloadMbps > bestDownloadSpeed {
				bestDownloadSpeed = downloadMbps
				bestProxy = result.ProxyName
			}
		}
	}
	
	avgLatency := 0.0
	avgDownload := 0.0
	avgUpload := 0.0
	if successful > 0 {
		avgLatency = totalLatency / float64(successful)
		avgDownload = totalDownload / float64(successful)
		avgUpload = totalUpload / float64(successful)
	}
	
	completeData := websocket.TestCompleteData{
		TotalTested:       len(results),
		SuccessfulTests:   successful,
		FailedTests:       failed,
		TotalDuration:     testDuration.String(),
		AverageLatency:    avgLatency,
		AverageDownload:   avgDownload,
		AverageUpload:     avgUpload,
		BestProxy:         bestProxy,
		BestDownloadSpeed: bestDownloadSpeed,
	}
	wsHub.BroadcastMessage(websocket.MessageTypeTestComplete, completeData)
	
	testTasksMutex.Lock()
	task.Status = "completed"
	testTasksMutex.Unlock()
	
	logger.Logger.Info("Test task completed", 
		slog.String("task_id", task.ID),
		slog.Int("total_tested", len(results)),
		slog.Int("successful", successful),
		slog.Int("failed", failed),
		slog.String("duration", testDuration.String()),
	)
}

func handleTest(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Speed test request received")
	
	if r.Method != http.MethodPost {
		logger.Logger.Warn("Invalid method for speed test", slog.String("method", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.LogError("Failed to decode request body", err)
		sendError(w, "Invalid request body: "+err.Error())
		return
	}

	logger.Logger.Info("Speed test configuration",
		slog.String("config_paths", req.ConfigPaths),
		slog.String("filter_regex", req.FilterRegex),
		slog.String("server_url", req.ServerURL),
		slog.Int("download_size_mb", req.DownloadSize),
		slog.Int("upload_size_mb", req.UploadSize),
		slog.Int("concurrent", req.Concurrent),
		slog.Bool("stash_compatible", req.StashCompatible),
	)

	// Set defaults
	if req.FilterRegex == "" {
		req.FilterRegex = ".+"
	}
	if req.ServerURL == "" {
		req.ServerURL = "https://speed.cloudflare.com"
	}
	if req.DownloadSize == 0 {
		req.DownloadSize = 50
	}
	if req.UploadSize == 0 {
		req.UploadSize = 20
	}
	if req.Timeout == 0 {
		req.Timeout = 5
	}
	if req.Concurrent == 0 {
		req.Concurrent = 4
	}
	if req.MaxLatency == 0 {
		req.MaxLatency = 800
	}
	// 保持用户设置的0值，表示不限制速度

	unlockConfig := createUnlockConfig(req)
	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths:      req.ConfigPaths,
		FilterRegex:      req.FilterRegex,
		IncludeNodes:     req.IncludeNodes,
		ExcludeNodes:     req.ExcludeNodes,
		ProtocolFilter:   req.ProtocolFilter,
		ServerURL:        req.ServerURL,
		DownloadSize:     req.DownloadSize * 1024 * 1024,
		UploadSize:       req.UploadSize * 1024 * 1024,
		Timeout:          time.Duration(req.Timeout) * time.Second,
		Concurrent:       req.Concurrent,
		MaxLatency:       time.Duration(req.MaxLatency) * time.Millisecond,
		MinDownloadSpeed: req.MinDownloadSpeed * 1024 * 1024,
		MinUploadSpeed:   req.MinUploadSpeed * 1024 * 1024,
		FastMode:         req.FastMode,
		RenameNodes:      req.RenameNodes,
		TestMode:         req.TestMode,
		UnlockConfig:     unlockConfig,
	})

	logger.Logger.Info("Loading proxies", slog.String("config_paths", req.ConfigPaths))
	allProxies, err := speedTester.LoadProxies(req.StashCompatible)
	if err != nil {
		logger.LogError("Failed to load proxies", err, slog.String("config_paths", req.ConfigPaths))
		sendError(w, "Failed to load proxies: "+err.Error())
		return
	}

	logger.Logger.Info("Proxies loaded successfully", slog.Int("proxy_count", len(allProxies)))

	if len(allProxies) == 0 {
		logger.Logger.Warn("No proxies found after loading and filtering")
		sendError(w, "No proxies found. Check your configuration path and filter regex.")
		return
	}

	logger.Logger.Info("Starting speed tests", slog.Int("proxy_count", len(allProxies)))
	startTime := time.Now()
	
	results := make([]*speedtester.Result, 0)
	speedTester.TestProxies(allProxies, func(result *speedtester.Result) {
		results = append(results, result)
		logger.Logger.Debug("Proxy test completed",
			slog.String("proxy_name", result.ProxyName),
			slog.String("proxy_type", result.ProxyType),
			slog.Float64("download_speed_mbps", result.DownloadSpeed/(1024*1024)),
			slog.Float64("upload_speed_mbps", result.UploadSpeed/(1024*1024)),
			slog.Int64("latency_ms", result.Latency.Milliseconds()),
		)
	})

	testDuration := time.Since(startTime)
	logger.Logger.Info("Speed tests completed",
		slog.Int("total_results", len(results)),
		slog.String("duration", testDuration.String()),
	)

	// Filter and sort results
	filteredResults := make([]*speedtester.Result, 0)
	for _, result := range results {
		if req.MaxLatency > 0 && result.Latency > time.Duration(req.MaxLatency)*time.Millisecond {
			continue
		}
		if req.MinDownloadSpeed > 0 && result.DownloadSpeed < req.MinDownloadSpeed*1024*1024 {
			continue
		}
		if req.MinUploadSpeed > 0 && result.UploadSpeed < req.MinUploadSpeed*1024*1024 {
			continue
		}
		filteredResults = append(filteredResults, result)
	}

	logger.Logger.Info("Results filtered",
		slog.Int("original_count", len(results)),
		slog.Int("filtered_count", len(filteredResults)),
	)

	// Sort by download speed
	sort.Slice(filteredResults, func(i, j int) bool {
		return filteredResults[i].DownloadSpeed > filteredResults[j].DownloadSpeed
	})

	sendSuccess(w, filteredResults)
}

func sendError(w http.ResponseWriter, message string) {
	logger.Logger.Error("Sending error response", slog.String("error", message))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(TestResponse{
		Success: false,
		Error:   message,
	})
}

func sendSuccess(w http.ResponseWriter, results []*speedtester.Result) {
	logger.Logger.Info("Sending successful response", slog.Int("result_count", len(results)))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TestResponse{
		Success: true,
		Results: results,
	})
}

func handleTestWithWebSocket(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Speed test request received (WebSocket enabled)")
	
	if r.Method != http.MethodPost {
		logger.Logger.Warn("Invalid method for speed test", slog.String("method", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.LogError("Failed to decode request body", err)
		sendError(w, "Invalid request body: "+err.Error())
		return
	}

	logger.Logger.Info("Speed test configuration",
		slog.String("config_paths", req.ConfigPaths),
		slog.String("filter_regex", req.FilterRegex),
		slog.String("server_url", req.ServerURL),
		slog.Int("download_size_mb", req.DownloadSize),
		slog.Int("upload_size_mb", req.UploadSize),
		slog.Int("concurrent", req.Concurrent),
		slog.Bool("stash_compatible", req.StashCompatible),
	)

	// Set defaults
	if req.FilterRegex == "" {
		req.FilterRegex = ".+"
	}
	if req.ServerURL == "" {
		req.ServerURL = "https://speed.cloudflare.com"
	}
	if req.DownloadSize == 0 {
		req.DownloadSize = 50
	}
	if req.UploadSize == 0 {
		req.UploadSize = 20
	}
	if req.Timeout == 0 {
		req.Timeout = 5
	}
	if req.Concurrent == 0 {
		req.Concurrent = 4
	}
	if req.MaxLatency == 0 {
		req.MaxLatency = 800
	}
	// 保持用户设置的0值，表示不限制速度

	unlockConfig := createUnlockConfig(req)
	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths:      req.ConfigPaths,
		FilterRegex:      req.FilterRegex,
		IncludeNodes:     req.IncludeNodes,
		ExcludeNodes:     req.ExcludeNodes,
		ProtocolFilter:   req.ProtocolFilter,
		ServerURL:        req.ServerURL,
		DownloadSize:     req.DownloadSize * 1024 * 1024,
		UploadSize:       req.UploadSize * 1024 * 1024,
		Timeout:          time.Duration(req.Timeout) * time.Second,
		Concurrent:       req.Concurrent,
		MaxLatency:       time.Duration(req.MaxLatency) * time.Millisecond,
		MinDownloadSpeed: req.MinDownloadSpeed * 1024 * 1024,
		MinUploadSpeed:   req.MinUploadSpeed * 1024 * 1024,
		FastMode:         req.FastMode,
		RenameNodes:      req.RenameNodes,
		TestMode:         req.TestMode,
		UnlockConfig:     unlockConfig,
	})

	logger.Logger.Info("Loading proxies", slog.String("config_paths", req.ConfigPaths))
	allProxies, err := speedTester.LoadProxies(req.StashCompatible)
	if err != nil {
		logger.LogError("Failed to load proxies", err, slog.String("config_paths", req.ConfigPaths))
		
		// Send error via WebSocket
		wsHub.BroadcastMessage(websocket.MessageTypeError, websocket.ErrorData{
			Message: "Failed to load proxies: " + err.Error(),
			Code:    "PROXY_LOAD_ERROR",
		})
		
		sendError(w, "Failed to load proxies: "+err.Error())
		return
	}

	logger.Logger.Info("Proxies loaded successfully", slog.Int("proxy_count", len(allProxies)))

	if len(allProxies) == 0 {
		logger.Logger.Warn("No proxies found after loading and filtering")
		
		// Send error via WebSocket
		wsHub.BroadcastMessage(websocket.MessageTypeError, websocket.ErrorData{
			Message: "No proxies found. Check your configuration path and filter regex.",
			Code:    "NO_PROXIES_FOUND",
		})
		
		sendError(w, "No proxies found. Check your configuration path and filter regex.")
		return
	}

	// Send test start message via WebSocket
	testStartData := websocket.TestStartData{
		TotalProxies: len(allProxies),
	}
	testStartData.Config.ConfigPaths = req.ConfigPaths
	testStartData.Config.FilterRegex = req.FilterRegex
	testStartData.Config.ServerURL = req.ServerURL
	testStartData.Config.DownloadSize = req.DownloadSize
	testStartData.Config.UploadSize = req.UploadSize
	testStartData.Config.Timeout = req.Timeout
	testStartData.Config.Concurrent = req.Concurrent
	testStartData.Config.MaxLatency = req.MaxLatency
	testStartData.Config.MinDownloadSpeed = req.MinDownloadSpeed
	testStartData.Config.MinUploadSpeed = req.MinUploadSpeed
	testStartData.Config.StashCompatible = req.StashCompatible

	wsHub.BroadcastMessage(websocket.MessageTypeTestStart, testStartData)

	// Create cancellable context for the test
	ctx, cancel := context.WithCancel(context.Background())
	testCancelMutex.Lock()
	testCancelFunc = cancel
	testCancelMutex.Unlock()

	// Clean up the cancel function when done
	defer func() {
		testCancelMutex.Lock()
		testCancelFunc = nil
		testCancelMutex.Unlock()
	}()

	logger.Logger.Info("Starting speed tests", slog.Int("proxy_count", len(allProxies)))
	startTime := time.Now()
	
	results := make([]*speedtester.Result, 0)
	completed := 0
	successful := 0
	failed := 0
	
	// Test proxies with WebSocket updates and context cancellation
	err = speedTester.TestProxiesWithContext(ctx, allProxies, func(result *speedtester.Result) {
		results = append(results, result)
		completed++
		
		// Determine status
		status := "success"
		if result.PacketLoss == 100 || result.Latency > time.Duration(req.MaxLatency)*time.Millisecond {
			status = "failed"
			failed++
		} else if (req.MinDownloadSpeed > 0 && result.DownloadSpeed < req.MinDownloadSpeed*1024*1024) || (req.MinUploadSpeed > 0 && result.UploadSpeed < req.MinUploadSpeed*1024*1024) {
			status = "failed"
			failed++
		} else {
			successful++
		}

		// Send progress update
		progressData := websocket.TestProgressData{
			CurrentProxy:    result.ProxyName,
			CompletedCount:  completed,
			TotalCount:      len(allProxies),
			ProgressPercent: float64(completed) / float64(len(allProxies)) * 100,
			Status:          status,
		}
		wsHub.BroadcastMessage(websocket.MessageTypeTestProgress, progressData)

		// Send individual result
		resultData := websocket.TestResultData{
			ProxyName:         result.ProxyName,
			ProxyType:         result.ProxyType,
			ProxyIP:           result.ProxyIP,
			Latency:           result.Latency.Milliseconds(),
			Jitter:            result.Jitter.Milliseconds(),
			PacketLoss:        result.PacketLoss,
			DownloadSpeed:     result.DownloadSpeed,
			UploadSpeed:       result.UploadSpeed,
			DownloadSpeedMbps: result.DownloadSpeed / (1024 * 1024),
			UploadSpeedMbps:   result.UploadSpeed / (1024 * 1024),
			Status:            status,
		}
		
		// 如果有错误详情，添加到WebSocket消息中
		if result.TestError != nil {
			resultData.ErrorStage = result.TestError.Stage
			resultData.ErrorCode = result.TestError.Code
			resultData.ErrorMessage = result.TestError.Message
		} else if result.FailureStage != "" {
			resultData.ErrorStage = result.FailureStage
			resultData.ErrorMessage = result.FailureReason
		}
		
		wsHub.BroadcastMessage(websocket.MessageTypeTestResult, resultData)

		logger.Logger.Debug("Proxy test completed",
			slog.String("proxy_name", result.ProxyName),
			slog.String("proxy_type", result.ProxyType),
			slog.Float64("download_speed_mbps", result.DownloadSpeed/(1024*1024)),
			slog.Float64("upload_speed_mbps", result.UploadSpeed/(1024*1024)),
			slog.Int64("latency_ms", result.Latency.Milliseconds()),
			slog.String("status", status),
		)
	})

	testDuration := time.Since(startTime)
	
	// Check if the test was cancelled
	if err != nil && err == context.Canceled {
		logger.Logger.Info("Speed tests cancelled",
			slog.Int("completed_tests", completed),
			slog.Int("total_tests", len(allProxies)),
			slog.String("duration", testDuration.String()),
		)
		
		// Update the cancellation data with actual progress
		cancelData := websocket.TestCancelledData{
			Message:         "测试已被用户取消",
			CompletedTests:  completed,
			TotalTests:      len(allProxies),
			PartialDuration: testDuration.String(),
		}
		wsHub.BroadcastMessage(websocket.MessageTypeTestCancelled, cancelData)
		
		// Still return filtered results for the completed tests
		filteredResults := make([]*speedtester.Result, 0)
		for _, result := range results {
			if req.MaxLatency > 0 && result.Latency > time.Duration(req.MaxLatency)*time.Millisecond {
				continue
			}
			if req.MinDownloadSpeed > 0 && result.DownloadSpeed < req.MinDownloadSpeed*1024*1024 {
				continue
			}
			if req.MinUploadSpeed > 0 && result.UploadSpeed < req.MinUploadSpeed*1024*1024 {
				continue
			}
			filteredResults = append(filteredResults, result)
		}
		
		sendSuccess(w, filteredResults)
		return
	}

	logger.Logger.Info("Speed tests completed",
		slog.Int("total_results", len(results)),
		slog.Int("successful", successful),
		slog.Int("failed", failed),
		slog.String("duration", testDuration.String()),
	)

	// Calculate summary statistics
	var totalLatency, totalDownload, totalUpload float64
	bestProxy := ""
	bestDownloadSpeed := 0.0
	
	for _, result := range results {
		if result.PacketLoss < 100 {
			totalLatency += float64(result.Latency.Milliseconds())
			totalDownload += result.DownloadSpeed / (1024 * 1024)
			totalUpload += result.UploadSpeed / (1024 * 1024)
			
			downloadMbps := result.DownloadSpeed / (1024 * 1024)
			if downloadMbps > bestDownloadSpeed {
				bestDownloadSpeed = downloadMbps
				bestProxy = result.ProxyName
			}
		}
	}

	avgLatency := 0.0
	avgDownload := 0.0
	avgUpload := 0.0
	if successful > 0 {
		avgLatency = totalLatency / float64(successful)
		avgDownload = totalDownload / float64(successful)
		avgUpload = totalUpload / float64(successful)
	}

	// Send test complete message
	completeData := websocket.TestCompleteData{
		TotalTested:       len(results),
		SuccessfulTests:   successful,
		FailedTests:       failed,
		TotalDuration:     testDuration.String(),
		AverageLatency:    avgLatency,
		AverageDownload:   avgDownload,
		AverageUpload:     avgUpload,
		BestProxy:         bestProxy,
		BestDownloadSpeed: bestDownloadSpeed,
	}
	wsHub.BroadcastMessage(websocket.MessageTypeTestComplete, completeData)

	// Filter and sort results for HTTP response
	filteredResults := make([]*speedtester.Result, 0)
	for _, result := range results {
		if req.MaxLatency > 0 && result.Latency > time.Duration(req.MaxLatency)*time.Millisecond {
			continue
		}
		if req.MinDownloadSpeed > 0 && result.DownloadSpeed < req.MinDownloadSpeed*1024*1024 {
			continue
		}
		if req.MinUploadSpeed > 0 && result.UploadSpeed < req.MinUploadSpeed*1024*1024 {
			continue
		}
		filteredResults = append(filteredResults, result)
	}

	logger.Logger.Info("Results filtered",
		slog.Int("original_count", len(results)),
		slog.Int("filtered_count", len(filteredResults)),
	)

	// Sort by download speed
	sort.Slice(filteredResults, func(i, j int) bool {
		return filteredResults[i].DownloadSpeed > filteredResults[j].DownloadSpeed
	})

	sendSuccess(w, filteredResults)
}

type ProtocolsResponse struct {
	Success   bool     `json:"success"`
	Error     string   `json:"error,omitempty"`
	Protocols []string `json:"protocols,omitempty"`
}

func handleGetProtocols(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Get protocols request received")
	
	if r.Method != http.MethodPost {
		logger.Logger.Warn("Invalid method for get protocols", slog.String("method", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ConfigPaths string `json:"configPaths"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.LogError("Failed to decode request body", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ProtocolsResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.ConfigPaths == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ProtocolsResponse{
			Success: false,
			Error:   "Config paths are required",
		})
		return
	}

	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths: req.ConfigPaths,
		FilterRegex: ".+",
	})

	logger.Logger.Info("Loading proxies for protocol discovery", slog.String("config_paths", req.ConfigPaths))
	allProxies, err := speedTester.LoadProxies(false)
	if err != nil {
		logger.LogError("Failed to load proxies", err, slog.String("config_paths", req.ConfigPaths))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ProtocolsResponse{
			Success: false,
			Error:   "Failed to load proxies: " + err.Error(),
		})
		return
	}

	protocols := speedTester.GetAvailableProtocols(allProxies)
	logger.Logger.Info("Protocols discovered", slog.Int("protocol_count", len(protocols)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProtocolsResponse{
		Success:   true,
		Protocols: protocols,
	})
}

// handleWebSocketMessage handles incoming WebSocket messages
func handleWebSocketMessage(msgType string, data []byte) {
	logger.Logger.Debug("Received WebSocket message", 
		slog.String("type", msgType),
		slog.String("data", string(data)),
	)
	
	switch msgType {
	case "stop_test":
		// 解析消息获取任务ID
		var msg struct {
			TaskID string `json:"taskId"`
		}
		if err := json.Unmarshal(data, &msg); err == nil && msg.TaskID != "" {
			// 取消特定任务
			testTasksMutex.RLock()
			if task, ok := testTasks[msg.TaskID]; ok {
				logger.Logger.Info("Stopping test task via WebSocket", slog.String("task_id", msg.TaskID))
				task.CancelFunc()
			}
			testTasksMutex.RUnlock()
		} else {
			// 兼容旧版本：取消全局任务
			testCancelMutex.RLock()
			if testCancelFunc != nil {
				logger.Logger.Info("Stopping test via WebSocket command")
				testCancelFunc()
				
				// Send cancellation message to all WebSocket clients
				cancelData := websocket.TestCancelledData{
					Message: "测试已被用户取消",
				}
				wsHub.BroadcastMessage(websocket.MessageTypeTestCancelled, cancelData)
			}
			testCancelMutex.RUnlock()
		}
	}
}

// NodeInfo represents basic node information without test results
type NodeInfo struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Server string `json:"server"`
	Port   int    `json:"port"`
}

type NodesResponse struct {
	Success bool       `json:"success"`
	Error   string     `json:"error,omitempty"`
	Nodes   []NodeInfo `json:"nodes,omitempty"`
}

func handleGetNodes(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Get nodes request received")
	
	if r.Method != http.MethodPost {
		logger.Logger.Warn("Invalid method for get nodes", slog.String("method", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ConfigPaths     string   `json:"configPaths"`
		IncludeNodes    []string `json:"includeNodes"`
		ExcludeNodes    []string `json:"excludeNodes"`
		ProtocolFilter  []string `json:"protocolFilter"`
		StashCompatible bool     `json:"stashCompatible"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.LogError("Failed to decode request body", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(NodesResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths:    req.ConfigPaths,
		FilterRegex:    ".+",
		IncludeNodes:   req.IncludeNodes,
		ExcludeNodes:   req.ExcludeNodes,
		ProtocolFilter: req.ProtocolFilter,
	})

	logger.Logger.Info("Loading nodes", slog.String("config_paths", req.ConfigPaths))
	allProxies, err := speedTester.LoadProxies(req.StashCompatible)
	if err != nil {
		logger.LogError("Failed to load proxies", err, slog.String("config_paths", req.ConfigPaths))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(NodesResponse{
			Success: false,
			Error:   "Failed to load proxies: " + err.Error(),
		})
		return
	}

	// Convert proxies to NodeInfo
	nodes := make([]NodeInfo, 0, len(allProxies))
	for name, proxy := range allProxies {
		nodeInfo := NodeInfo{
			Name: name,
			Type: proxy.Type().String(),
		}
		
		// Extract server and port from config
		if server, ok := proxy.Config["server"]; ok {
			nodeInfo.Server = server.(string)
		}
		if port, ok := proxy.Config["port"]; ok {
			switch p := port.(type) {
			case int:
				nodeInfo.Port = p
			case float64:
				nodeInfo.Port = int(p)
			}
		}
		
		nodes = append(nodes, nodeInfo)
	}

	logger.Logger.Info("Nodes loaded successfully", slog.Int("node_count", len(nodes)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(NodesResponse{
		Success: true,
		Nodes:   nodes,
	})
}

// handleExportResults handles exporting test results in various formats
func handleExportResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var exportReq struct {
		TaskID  string              `json:"taskId"`
		Options utils.ExportOptions `json:"options"`
	}

	if err := json.NewDecoder(r.Body).Decode(&exportReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate export options
	if err := utils.ValidateExportOptions(exportReq.Options); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid export options: " + err.Error(),
		})
		return
	}

	// TODO: Get test results from the task ID
	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Export functionality will be implemented",
		"format":  exportReq.Options.Format,
		"path":    exportReq.Options.OutputPath,
	})
}

// handleLogManagement handles log management operations
func handleLogManagement(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetLogInfo(w, r)
	case http.MethodPost:
		handleLogAction(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// handleGetLogInfo returns current log configuration and status
func handleGetLogInfo(w http.ResponseWriter, r *http.Request) {
	logInfo := map[string]any{
		"level":          logger.Logger.Enabled(context.Background(), slog.LevelDebug),
		"file_logging":   true, // Based on current config
		"console_logging": true,
		"log_dir":        "logs",
		"log_file":       "clash-speedtest.log",
		"max_size_mb":    10,
		"max_files":      5,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"config":  logInfo,
	})
}

// handleLogAction handles log actions like rotation, level change
func handleLogAction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Action string `json:"action"` // "rotate", "set_level"
		Level  string `json:"level,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}
	
	switch req.Action {
	case "rotate":
		if err := logger.RotateLogNow(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to rotate log: " + err.Error(),
			})
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"message": "Log rotated successfully",
		})
		
	case "set_level":
		var level slog.Level
		switch strings.ToUpper(req.Level) {
		case "DEBUG":
			level = slog.LevelDebug
		case "INFO":
			level = slog.LevelInfo
		case "WARN":
			level = slog.LevelWarn
		case "ERROR":
			level = slog.LevelError
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid log level. Use DEBUG, INFO, WARN, or ERROR",
			})
			return
		}
		
		logger.SetLevel(level)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"message": fmt.Sprintf("Log level set to %s", req.Level),
		})
		
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid action. Use 'rotate' or 'set_level'",
		})
	}
}
