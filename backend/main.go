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
	"sync"
	"syscall"
	"time"

	"github.com/faceair/clash-speedtest/logger"
	"github.com/faceair/clash-speedtest/speedtester"
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
}

type TestResponse struct {
	Success bool                  `json:"success"`
	Error   string                `json:"error,omitempty"`
	Results []*speedtester.Result `json:"results,omitempty"`
}

var wsHub *websocket.Hub
var testCancelFunc context.CancelFunc
var testCancelMutex sync.RWMutex

func main() {
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
	http.HandleFunc("/api/protocols", loggingMiddleware(handleGetProtocols))
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
	if req.MinDownloadSpeed == 0 {
		req.MinDownloadSpeed = 5
	}
	if req.MinUploadSpeed == 0 {
		req.MinUploadSpeed = 2
	}

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
	if req.MinDownloadSpeed == 0 {
		req.MinDownloadSpeed = 5
	}
	if req.MinUploadSpeed == 0 {
		req.MinUploadSpeed = 2
	}

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
		} else if result.DownloadSpeed < req.MinDownloadSpeed*1024*1024 || result.UploadSpeed < req.MinUploadSpeed*1024*1024 {
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
