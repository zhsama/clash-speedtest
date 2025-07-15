package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/faceair/clash-speedtest/logger"
	"github.com/faceair/clash-speedtest/speedtester"
	"github.com/faceair/clash-speedtest/utils"
)

// generateTaskID generates a unique task ID
func generateTaskID() string {
	return fmt.Sprintf("task-%d-%s", time.Now().Unix(), time.Now().Format("150405"))
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Debug("Health check requested")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// sendError sends an error response
func sendError(w http.ResponseWriter, message string) {
	logger.Logger.Error("Sending error response", slog.String("error", message))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(TestResponse{
		Success: false,
		Error:   message,
	})
}

// sendSuccess sends a success response
func sendSuccess(w http.ResponseWriter, results []*speedtester.Result) {
	logger.Logger.Info("Sending successful response", slog.Int("result_count", len(results)))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TestResponse{
		Success: true,
		Results: results,
	})
}

// handleTUNCheck handles TUN mode detection requests
func handleTUNCheck(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("TUN 模式检测请求")

	status := utils.CheckTUNMode()

	w.Header().Set("Content-Type", "application/json")

	response := map[string]any{
		"success":    true,
		"tun_status": status,
		"warning":    "",
	}

	// 如果检测到 TUN 模式启用，添加警告信息
	if status.Enabled {
		warning := "检测到系统已启用 TUN 模式！"
		if status.ActiveInterface != nil {
			warning += fmt.Sprintf(" 活动接口: %s", status.ActiveInterface.Name)
		}
		if len(status.ProxyProcesses) > 0 {
			warning += fmt.Sprintf(" 检测到代理进程: %s", status.ProxyProcesses[0].Name)
		}
		warning += " 建议在进行速度测试前先关闭 TUN 模式，以获得更准确的测试结果。"

		response["warning"] = warning

		logger.Logger.Warn("检测到 TUN 模式已启用",
			slog.String("active_interface", func() string {
				if status.ActiveInterface != nil {
					return status.ActiveInterface.Name
				}
				return "unknown"
			}()),
			slog.Int("proxy_processes", len(status.ProxyProcesses)),
		)
	} else {
		logger.Logger.Info("未检测到 TUN 模式")
	}

	json.NewEncoder(w).Encode(response)
}
