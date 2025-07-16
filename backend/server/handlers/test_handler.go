package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/faceair/clash-speedtest/logger"
	"github.com/faceair/clash-speedtest/server/common"
	"github.com/faceair/clash-speedtest/server/response"
	"github.com/faceair/clash-speedtest/speedtester"
	"github.com/faceair/clash-speedtest/websocket"
)

// TestHandler 测试处理器
type TestHandler struct {
	*Handler
	wsHub *websocket.Hub
	
	// 任务管理
	testTasks      map[string]*TestTask
	testTasksMutex sync.RWMutex
	
	// 全局测试取消
	testCancelFunc context.CancelFunc
	testCancelMutex sync.RWMutex
}

// TestTask 测试任务结构
type TestTask struct {
	ID         string
	Config     *common.TestRequest
	Context    context.Context
	CancelFunc context.CancelFunc
	Status     string // pending, running, completed, cancelled
	StartTime  time.Time
}

// NewTestHandler 创建新的测试处理器
func NewTestHandler(wsHub *websocket.Hub) *TestHandler {
	return &TestHandler{
		Handler:   NewHandler(),
		wsHub:     wsHub,
		testTasks: make(map[string]*TestTask),
	}
}

// generateTaskID 生成任务 ID
func (h *TestHandler) generateTaskID() string {
	return fmt.Sprintf("task-%d-%s", time.Now().Unix(), time.Now().Format("150405"))
}

// HandleTest 处理同步测试请求
func (h *TestHandler) HandleTest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if r.Method != http.MethodPost {
		h.handleMethodNotAllowed(ctx, w, r, "POST")
		return
	}
	
	logger.Logger.InfoContext(ctx, "Speed test request received")
	
	// 解析请求
	req, err := h.parseTestRequest(r)
	if err != nil {
		response.HandleError(ctx, w, err)
		return
	}
	
	// 记录测试配置
	logger.Logger.InfoContext(ctx, "Speed test configuration",
		slog.String("config_paths", req.ConfigPaths),
		slog.String("filter_regex", req.FilterRegex),
		slog.String("server_url", req.ServerURL),
		slog.Int("download_size_mb", req.DownloadSize),
		slog.Int("upload_size_mb", req.UploadSize),
		slog.Int("concurrent", req.Concurrent),
		slog.Bool("stash_compatible", req.StashCompatible),
	)
	
	// 创建速度测试器
	speedTester := h.createSpeedTester(req)
	
	// 加载代理
	logger.Logger.InfoContext(ctx, "Loading proxies", 
		slog.String("config_paths", req.ConfigPaths))
	
	allProxies, err := speedTester.LoadProxies(req.StashCompatible)
	if err != nil {
		logger.Logger.ErrorContext(ctx, "Failed to load proxies", 
			slog.String("error", err.Error()),
			slog.String("config_paths", req.ConfigPaths))
		response.SendTestError(ctx, w, "Failed to load proxies: "+err.Error())
		return
	}
	
	logger.Logger.InfoContext(ctx, "Proxies loaded successfully", 
		slog.Int("proxy_count", len(allProxies)))
	
	if len(allProxies) == 0 {
		logger.Logger.WarnContext(ctx, "No proxies found after loading and filtering")
		response.SendTestError(ctx, w, "No proxies found. Check your configuration path and filter regex.")
		return
	}
	
	// 执行测试
	logger.Logger.InfoContext(ctx, "Starting speed tests", 
		slog.Int("proxy_count", len(allProxies)))
	
	startTime := time.Now()
	results := make([]*speedtester.Result, 0)
	
	speedTester.TestProxies(allProxies, func(result *speedtester.Result) {
		results = append(results, result)
		logger.Logger.DebugContext(ctx, "Proxy test completed",
			slog.String("proxy_name", result.ProxyName),
			slog.String("proxy_type", result.ProxyType),
			slog.Float64("download_speed_mbps", result.DownloadSpeed/(1024*1024)),
			slog.Float64("upload_speed_mbps", result.UploadSpeed/(1024*1024)),
			slog.Int64("latency_ms", result.Latency.Milliseconds()),
		)
	})
	
	testDuration := time.Since(startTime)
	logger.Logger.InfoContext(ctx, "Speed tests completed",
		slog.Int("total_results", len(results)),
		slog.String("duration", testDuration.String()),
	)
	
	// 过滤和排序结果
	filteredResults := h.filterResults(results, req)
	
	logger.Logger.InfoContext(ctx, "Results filtered",
		slog.Int("original_count", len(results)),
		slog.Int("filtered_count", len(filteredResults)),
	)
	
	// 按下载速度排序
	sort.Slice(filteredResults, func(i, j int) bool {
		return filteredResults[i].DownloadSpeed > filteredResults[j].DownloadSpeed
	})
	
	response.SendTestSuccess(ctx, w, filteredResults)
}

// HandleTestAsync 处理异步测试请求
func (h *TestHandler) HandleTestAsync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if r.Method != http.MethodPost {
		h.handleMethodNotAllowed(ctx, w, r, "POST")
		return
	}
	
	logger.Logger.InfoContext(ctx, "Async test request received")
	
	// 解析请求
	req, err := h.parseTestRequest(r)
	if err != nil {
		response.HandleError(ctx, w, err)
		return
	}
	
	// 创建任务
	taskID := h.generateTaskID()
	taskCtx, cancel := context.WithCancel(context.Background())
	
	task := &TestTask{
		ID:         taskID,
		Config:     req,
		Context:    taskCtx,
		CancelFunc: cancel,
		Status:     "pending",
		StartTime:  time.Now(),
	}
	
	h.testTasksMutex.Lock()
	h.testTasks[taskID] = task
	h.testTasksMutex.Unlock()
	
	// 异步执行测试
	go h.runTestTask(task)
	
	// 返回任务ID
	response.SendSuccess(ctx, w, map[string]interface{}{
		"taskId":  taskID,
		"message": "Test task created successfully",
	})
}

// runTestTask 执行测试任务
func (h *TestHandler) runTestTask(task *TestTask) {
	ctx := task.Context
	
	// 更新任务状态
	h.testTasksMutex.Lock()
	task.Status = "running"
	h.testTasksMutex.Unlock()
	
	logger.Logger.InfoContext(ctx, "Starting test task", 
		slog.String("task_id", task.ID))
	
	// 创建速度测试器
	speedTester := h.createSpeedTester(task.Config)
	
	// 加载代理
	allProxies, err := speedTester.LoadProxies(task.Config.StashCompatible)
	if err != nil {
		logger.Logger.ErrorContext(ctx, "Failed to load proxies", 
			slog.String("error", err.Error()))
		
		h.wsHub.BroadcastMessage(websocket.MessageTypeError, websocket.ErrorData{
			Message: "Failed to load proxies: " + err.Error(),
			Code:    "PROXY_LOAD_ERROR",
		})
		
		h.testTasksMutex.Lock()
		task.Status = "failed"
		h.testTasksMutex.Unlock()
		return
	}
	
	if len(allProxies) == 0 {
		h.wsHub.BroadcastMessage(websocket.MessageTypeError, websocket.ErrorData{
			Message: "No proxies found",
			Code:    "NO_PROXIES_FOUND",
		})
		
		h.testTasksMutex.Lock()
		task.Status = "failed"
		h.testTasksMutex.Unlock()
		return
	}
	
	// 发送测试开始消息
	h.sendTestStartMessage(task, len(allProxies))
	
	// 执行测试
	results := make([]*speedtester.Result, 0)
	completed := 0
	successful := 0
	failed := 0
	
	err = speedTester.TestProxiesWithContext(ctx, allProxies, func(result *speedtester.Result) {
		results = append(results, result)
		completed++
		
		// 判断结果状态
		status := h.determineResultStatus(result, task.Config)
		if status == "success" {
			successful++
		} else {
			failed++
		}
		
		// 发送进度更新
		h.sendProgressUpdate(task, result, completed, len(allProxies), status)
	})
	
	testDuration := time.Since(task.StartTime)
	
	// 检查是否被取消
	if err != nil && err == context.Canceled {
		logger.Logger.InfoContext(ctx, "Test task cancelled", 
			slog.String("task_id", task.ID))
		
		h.sendTestCancelledMessage(task, completed, len(allProxies), testDuration)
		
		h.testTasksMutex.Lock()
		task.Status = "cancelled"
		h.testTasksMutex.Unlock()
		return
	}
	
	// 发送测试完成消息
	h.sendTestCompleteMessage(task, results, successful, failed, testDuration)
	
	h.testTasksMutex.Lock()
	task.Status = "completed"
	h.testTasksMutex.Unlock()
	
	logger.Logger.InfoContext(ctx, "Test task completed",
		slog.String("task_id", task.ID),
		slog.Int("total_tested", len(results)),
		slog.Int("successful", successful),
		slog.Int("failed", failed),
		slog.String("duration", testDuration.String()),
	)
}

// filterResults 过滤测试结果
func (h *TestHandler) filterResults(results []*speedtester.Result, req *common.TestRequest) []*speedtester.Result {
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
	
	return filteredResults
}

// determineResultStatus 判断结果状态
func (h *TestHandler) determineResultStatus(result *speedtester.Result, config *common.TestRequest) string {
	if config.TestMode == "unlock_only" {
		if result.UnlockSummary.TotalSupported > 0 {
			return "success"
		}
		return "failed"
	}
	
	if result.PacketLoss == 100 || result.Latency > time.Duration(config.MaxLatency)*time.Millisecond {
		return "failed"
	}
	
	if result.DownloadSpeed < config.MinDownloadSpeed*1024*1024 || result.UploadSpeed < config.MinUploadSpeed*1024*1024 {
		return "failed"
	}
	
	return "success"
}

// sendTestStartMessage 发送测试开始消息
func (h *TestHandler) sendTestStartMessage(task *TestTask, totalProxies int) {
	testStartData := websocket.TestStartData{
		TotalProxies: totalProxies,
	}
	
	// 设置配置信息
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
	
	h.wsHub.BroadcastMessage(websocket.MessageTypeTestStart, testStartData)
}

// sendProgressUpdate 发送进度更新
func (h *TestHandler) sendProgressUpdate(task *TestTask, result *speedtester.Result, completed, total int, status string) {
	progressData := websocket.TestProgressData{
		CurrentProxy:    result.ProxyName,
		CompletedCount:  completed,
		TotalCount:      total,
		ProgressPercent: float64(completed) / float64(total) * 100,
		Status:          status,
	}
	h.wsHub.BroadcastMessage(websocket.MessageTypeTestProgress, progressData)
	
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
		UnlockResults:     websocket.ConvertSpeedtesterUnlockResults(result.UnlockResults),
		UnlockSummary:     websocket.ConvertSpeedtesterUnlockSummary(result.UnlockSummary),
	}
	
	if result.TestError != nil {
		resultData.ErrorStage = result.TestError.Stage
		resultData.ErrorCode = result.TestError.Code
		resultData.ErrorMessage = result.TestError.Message
	} else if result.FailureStage != "" {
		resultData.ErrorStage = result.FailureStage
		resultData.ErrorMessage = result.FailureReason
	}
	
	h.wsHub.BroadcastMessage(websocket.MessageTypeTestResult, resultData)
}

// sendTestCancelledMessage 发送测试取消消息
func (h *TestHandler) sendTestCancelledMessage(task *TestTask, completed, total int, duration time.Duration) {
	cancelData := websocket.TestCancelledData{
		Message:         "测试已被用户取消",
		CompletedTests:  completed,
		TotalTests:      total,
		PartialDuration: duration.String(),
	}
	h.wsHub.BroadcastMessage(websocket.MessageTypeTestCancelled, cancelData)
}

// sendTestCompleteMessage 发送测试完成消息
func (h *TestHandler) sendTestCompleteMessage(task *TestTask, results []*speedtester.Result, successful, failed int, duration time.Duration) {
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
		TotalDuration:     duration.String(),
		AverageLatency:    avgLatency,
		AverageDownload:   avgDownload,
		AverageUpload:     avgUpload,
		BestProxy:         bestProxy,
		BestDownloadSpeed: bestDownloadSpeed,
	}
	h.wsHub.BroadcastMessage(websocket.MessageTypeTestComplete, completeData)
}