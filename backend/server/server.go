package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/faceair/clash-speedtest/logger"
	"github.com/faceair/clash-speedtest/websocket"
)

// Server 服务器结构
type Server struct {
	httpServer *http.Server
	wsHub      *websocket.Hub
	router     *Router
	port       int
}

// NewServer 创建新的服务器
func NewServer(port int) *Server {
	wsHub := websocket.NewHub()
	router := NewRouter(wsHub)
	
	server := &Server{
		wsHub:  wsHub,
		router: router,
		port:   port,
	}
	
	// 设置路由
	router.Setup()
	
	// 创建 HTTP 服务器
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router.GetMux(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	return server
}

// Start 启动服务器
func (s *Server) Start() error {
	// 启动 WebSocket Hub
	go s.wsHub.Run()
	
	// 启动 HTTP 服务器
	go func() {
		logger.Logger.Info("Starting HTTP server",
			slog.Int("port", s.port),
			slog.String("address", s.httpServer.Addr))
		
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("HTTP server error", 
				slog.String("error", err.Error()))
		}
	}()
	
	// 等待停止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Logger.Info("Shutting down server...")
	
	// 优雅关闭
	return s.Shutdown()
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// WebSocket Hub 会在连接关闭时自动清理
	
	// 关闭 HTTP 服务器
	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.Logger.Error("Failed to shutdown server gracefully", 
			slog.String("error", err.Error()))
		return err
	}
	
	logger.Logger.Info("Server shutdown complete")
	return nil
}

// GetPort 获取服务器端口
func (s *Server) GetPort() int {
	return s.port
}