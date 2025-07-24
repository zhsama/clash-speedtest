package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/zhsama/clash-speedtest/logger"
)

// contextKey 用于在 context 中存储值的键类型
type contextKey string

const (
	// RequestIDKey 请求 ID 的 context 键
	RequestIDKey contextKey = "request_id"
	// StartTimeKey 请求开始时间的 context 键
	StartTimeKey contextKey = "start_time"
)

// loggingResponseWriter HTTP响应写入器，用于记录响应状态
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader 记录HTTP状态码
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	if !lrw.written {
		lrw.statusCode = code
		lrw.written = true
		lrw.ResponseWriter.WriteHeader(code)
	}
}

// Write 记录写入状态
func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	if !lrw.written {
		lrw.WriteHeader(http.StatusOK)
	}
	return lrw.ResponseWriter.Write(data)
}

// generateRequestID 生成请求 ID
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// RequestID 请求 ID 中间件
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 生成请求 ID
		requestID := generateRequestID()
		
		// 将请求 ID 添加到 context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, StartTimeKey, time.Now())
		
		// 添加到响应头
		w.Header().Set("X-Request-ID", requestID)
		
		// 使用新的 context 继续处理
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logging 日志中间件
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 从 context 获取请求 ID
		requestID := GetRequestID(r.Context())
		
		// 跳过 WebSocket 路由的日志包装，以允许连接劫持
		if r.URL.Path == "/ws" {
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			
			logger.Logger.InfoContext(r.Context(), "WebSocket connection established",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("request_id", requestID),
				slog.String("duration", duration.String()),
			)
			return
		}
		
		// 创建自定义 ResponseWriter 以捕获状态码
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			written:        false,
		}
		
		// 记录请求开始
		logger.Logger.InfoContext(r.Context(), "Request started",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("request_id", requestID),
			slog.String("user_agent", r.UserAgent()),
		)
		
		// 处理请求
		next.ServeHTTP(lrw, r)
		
		// 记录请求完成
		duration := time.Since(start)
		logger.Logger.InfoContext(r.Context(), "Request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("request_id", requestID),
			slog.Int("status_code", lrw.statusCode),
			slog.String("duration", duration.String()),
		)
	})
}

// GetRequestID 从 context 获取请求 ID
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetStartTime 从 context 获取请求开始时间
func GetStartTime(ctx context.Context) time.Time {
	if startTime, ok := ctx.Value(StartTimeKey).(time.Time); ok {
		return startTime
	}
	return time.Time{}
}

// Recovery 恢复中间件，处理 panic
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(r.Context())
				
				logger.Logger.ErrorContext(r.Context(), "Panic recovered",
					slog.String("request_id", requestID),
					slog.String("error", err.(error).Error()),
					slog.String("path", r.URL.Path),
				)
				
				// 返回内部服务器错误
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}