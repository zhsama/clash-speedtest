package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/zhsama/clash-speedtest/logger"
	"github.com/zhsama/clash-speedtest/server/common"
	"github.com/zhsama/clash-speedtest/server/response"
	"github.com/zhsama/clash-speedtest/speedtester"
	"github.com/zhsama/clash-speedtest/unlock"
)

// Handler 处理器基础结构
type Handler struct {
	// 可以在这里添加依赖注入的字段
}

// NewHandler 创建新的处理器
func NewHandler() *Handler {
	return &Handler{}
}

// parseTestRequest 解析测试请求
func (h *Handler) parseTestRequest(r *http.Request) (*common.TestRequest, error) {
	var req common.TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, response.NewValidationError("Invalid request body", err)
	}
	
	// 设置默认值
	common.SetRequestDefaults(&req)
	
	// 验证请求
	if err := common.ValidateRequest(&req); err != nil {
		return nil, err
	}
	
	return &req, nil
}

// createUnlockConfig 根据TestRequest创建解锁检测配置
func (h *Handler) createUnlockConfig(req *common.TestRequest) *unlock.UnlockTestConfig {
	logger.Logger.DebugContext(context.Background(), "Creating unlock config",
		slog.Bool("unlock_enabled", req.UnlockEnabled),
		slog.String("test_mode", req.TestMode),
		slog.Any("platforms", req.UnlockPlatforms),
		slog.Int("concurrent", req.UnlockConcurrent),
		slog.Int("timeout", req.UnlockTimeout),
	)

	needsUnlock := req.TestMode == "unlock_only" || req.TestMode == "both"
	
	if !req.UnlockEnabled && !needsUnlock {
		logger.Logger.DebugContext(context.Background(), "Unlock detection disabled by request")
		return &unlock.UnlockTestConfig{
			Enabled: false,
		}
	}

	if needsUnlock {
		logger.Logger.DebugContext(context.Background(), "Auto-enabling unlock detection for test mode", 
			slog.String("test_mode", req.TestMode))
	}

	platforms := req.UnlockPlatforms
	if len(platforms) == 0 {
		platforms = []string{"Netflix", "YouTube", "Disney+", "ChatGPT", "Spotify", "Bilibili"}
		logger.Logger.DebugContext(context.Background(), "Using default platforms", 
			slog.Any("platforms", platforms))
	}

	concurrent := req.UnlockConcurrent
	if concurrent <= 0 {
		concurrent = 5
	}

	timeout := req.UnlockTimeout
	if timeout <= 0 {
		timeout = 10
	}

	config := &unlock.UnlockTestConfig{
		Enabled:       true,
		Platforms:     platforms,
		Concurrent:    concurrent,
		Timeout:       timeout,
		RetryOnError:  req.UnlockRetry,
		IncludeIPInfo: true,
	}

	return config
}

// createSpeedTester 创建速度测试器
func (h *Handler) createSpeedTester(req *common.TestRequest) *speedtester.SpeedTester {
	unlockConfig := h.createUnlockConfig(req)
	
	return speedtester.New(&speedtester.Config{
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
}

// handleMethodNotAllowed 处理不允许的方法
func (h *Handler) handleMethodNotAllowed(ctx context.Context, w http.ResponseWriter, r *http.Request, allowedMethods ...string) {
	methods := "GET, POST, OPTIONS"
	if len(allowedMethods) > 0 {
		methods = ""
		for i, method := range allowedMethods {
			if i > 0 {
				methods += ", "
			}
			methods += method
		}
	}
	
	w.Header().Set("Allow", methods)
	response.SendError(ctx, w, http.StatusMethodNotAllowed, "Method not allowed")
}

// HandleHealth 处理健康检查请求
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	logger.Logger.DebugContext(ctx, "Health check requested")
	
	response.SendSuccess(ctx, w, map[string]string{
		"status": "ok",
		"service": "clash-speedtest",
		"version": "2.0.0",
	})
}