package response

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/faceair/clash-speedtest/logger"
	"github.com/faceair/clash-speedtest/speedtester"
)

// Response 统一响应格式
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    int         `json:"code,omitempty"`
}

// TestResponse 测试响应的结构
type TestResponse struct {
	Success bool                  `json:"success"`
	Error   string                `json:"error,omitempty"`
	Results []*speedtester.Result `json:"results,omitempty"`
}

// ProtocolsResponse 协议列表响应结构
type ProtocolsResponse struct {
	Success   bool     `json:"success"`
	Error     string   `json:"error,omitempty"`
	Protocols []string `json:"protocols,omitempty"`
}

// NodesResponse 节点列表响应结构
type NodesResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Nodes   []NodeInfo  `json:"nodes,omitempty"`
}

// NodeInfo 节点信息结构
type NodeInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Password string `json:"password,omitempty"`
	Cipher   string `json:"cipher,omitempty"`
}

// SendJSON 发送 JSON 响应
func SendJSON(ctx context.Context, w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Logger.ErrorContext(ctx, "Failed to encode JSON response", 
			slog.String("error", err.Error()))
	}
}

// SendSuccess 发送成功响应
func SendSuccess(ctx context.Context, w http.ResponseWriter, data interface{}) {
	SendJSON(ctx, w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// SendError 发送错误响应
func SendError(ctx context.Context, w http.ResponseWriter, statusCode int, message string) {
	logger.Logger.ErrorContext(ctx, "Sending error response", 
		slog.String("error", message),
		slog.Int("status_code", statusCode))
	
	SendJSON(ctx, w, statusCode, Response{
		Success: false,
		Error:   message,
		Code:    statusCode,
	})
}

// SendTestSuccess 发送测试成功响应
func SendTestSuccess(ctx context.Context, w http.ResponseWriter, results []*speedtester.Result) {
	logger.Logger.InfoContext(ctx, "Sending successful test response", 
		slog.Int("result_count", len(results)))
	
	SendJSON(ctx, w, http.StatusOK, TestResponse{
		Success: true,
		Results: results,
	})
}

// SendTestError 发送测试错误响应
func SendTestError(ctx context.Context, w http.ResponseWriter, message string) {
	SendError(ctx, w, http.StatusBadRequest, message)
}