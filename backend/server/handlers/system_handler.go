package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/faceair/clash-speedtest/logger"
	"github.com/faceair/clash-speedtest/server/response"
	"github.com/faceair/clash-speedtest/utils/system"
)

// SystemHandler 系统处理器
type SystemHandler struct {
	*Handler
}

// NewSystemHandler 创建新的系统处理器
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		Handler: NewHandler(),
	}
}

// HandleTUNCheck 处理 TUN 模式检测请求
func (h *SystemHandler) HandleTUNCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if r.Method != http.MethodGet {
		h.handleMethodNotAllowed(ctx, w, r, "GET")
		return
	}
	
	logger.Logger.InfoContext(ctx, "TUN mode detection request")
	
	status := system.CheckTUNMode()
	
	responseData := map[string]interface{}{
		"success":    true,
		"tun_status": status,
		"warning":    "",
	}
	
	// If TUN mode is detected, add warning information
	if status.Enabled {
		warning := "TUN mode is already enabled on the system!"
		if status.ActiveInterface != nil {
			warning += fmt.Sprintf(" Active interface: %s", status.ActiveInterface.Name)
		}
		if len(status.ProxyProcesses) > 0 {
			warning += fmt.Sprintf(" Proxy process detected: %s", status.ProxyProcesses[0].Name)
		}
		warning += " It is recommended to disable TUN mode before speed testing for more accurate results."
		
		responseData["warning"] = warning
		
		logger.Logger.WarnContext(ctx, "TUN mode detected as enabled",
			slog.String("active_interface", func() string {
				if status.ActiveInterface != nil {
					return status.ActiveInterface.Name
				}
				return "unknown"
			}()),
			slog.Int("proxy_processes", len(status.ProxyProcesses)),
		)
	} else {
		logger.Logger.InfoContext(ctx, "TUN mode not detected")
	}
	
	response.SendJSON(ctx, w, http.StatusOK, responseData)
}

// HandleLogManagement 处理日志管理请求
func (h *SystemHandler) HandleLogManagement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	switch r.Method {
	case http.MethodGet:
		h.handleGetLogInfo(ctx, w, r)
	case http.MethodPost:
		h.handleLogAction(ctx, w, r)
	default:
		h.handleMethodNotAllowed(ctx, w, r, "GET", "POST")
	}
}

// handleGetLogInfo 处理获取日志信息请求
func (h *SystemHandler) handleGetLogInfo(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logInfo := map[string]interface{}{
		"level":           logger.Logger.Enabled(ctx, slog.LevelDebug),
		"file_logging":    true, // 基于当前配置
		"console_logging": true,
		"log_dir":         "logs",
		"log_file":        "clash-speedtest.log",
		"max_size_mb":     10,
		"max_files":       5,
	}
	
	response.SendJSON(ctx, w, http.StatusOK, map[string]interface{}{
		"success": true,
		"config":  logInfo,
	})
}

// handleLogAction 处理日志操作请求
func (h *SystemHandler) handleLogAction(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req struct {
		Action string `json:"action"` // "rotate", "set_level"
		Level  string `json:"level,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(ctx, w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	
	switch req.Action {
	case "rotate":
		if err := logger.RotateLogNow(); err != nil {
			response.SendError(ctx, w, http.StatusInternalServerError, "Failed to rotate log: "+err.Error())
			return
		}
		
		response.SendJSON(ctx, w, http.StatusOK, map[string]interface{}{
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
			response.SendError(ctx, w, http.StatusBadRequest, 
				"Invalid log level. Use DEBUG, INFO, WARN, or ERROR")
			return
		}
		
		logger.SetLevel(level)
		
		response.SendJSON(ctx, w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Log level set to %s", req.Level),
		})
		
	default:
		response.SendError(ctx, w, http.StatusBadRequest, 
			"Invalid action. Use 'rotate' or 'set_level'")
	}
}