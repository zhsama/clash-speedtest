package server

import (
	"net/http"

	"github.com/faceair/clash-speedtest/server/handlers"
	"github.com/faceair/clash-speedtest/server/middleware"
	"github.com/faceair/clash-speedtest/websocket"
)

// Router 服务器路由器
type Router struct {
	mux           *http.ServeMux
	testHandler   *handlers.TestHandler
	configHandler *handlers.ConfigHandler
	systemHandler *handlers.SystemHandler
	wsHub         *websocket.Hub
}

// NewRouter 创建新的路由器
func NewRouter(wsHub *websocket.Hub) *Router {
	return &Router{
		mux:           http.NewServeMux(),
		testHandler:   handlers.NewTestHandler(wsHub),
		configHandler: handlers.NewConfigHandler(),
		systemHandler: handlers.NewSystemHandler(),
		wsHub:         wsHub,
	}
}

// Setup 设置路由
func (r *Router) Setup() {
	// 应用中间件
	r.mux.HandleFunc("/health", r.withMiddleware(r.testHandler.HandleHealth))
	
	// 系统相关路由
	r.mux.HandleFunc("/api/tun-check", r.withMiddleware(r.systemHandler.HandleTUNCheck))
	r.mux.HandleFunc("/system/tun-check", r.withMiddleware(r.systemHandler.HandleTUNCheck))
	r.mux.HandleFunc("/system/logs", r.withMiddleware(r.systemHandler.HandleLogManagement))
	
	// 测试相关路由
	r.mux.HandleFunc("/test", r.withMiddleware(r.testHandler.HandleTest))
	r.mux.HandleFunc("/test/async", r.withMiddleware(r.testHandler.HandleTestAsync))
	r.mux.HandleFunc("/api/test/async", r.withMiddleware(r.testHandler.HandleTestAsync))
	r.mux.HandleFunc("/test/websocket", r.withMiddleware(r.handleWebSocket))
	r.mux.HandleFunc("/ws", r.withMiddleware(r.handleWebSocket))
	
	// 配置相关路由
	r.mux.HandleFunc("/config/protocols", r.withMiddleware(r.configHandler.HandleGetProtocols))
	r.mux.HandleFunc("/api/protocols", r.withMiddleware(r.configHandler.HandleGetProtocols))
	r.mux.HandleFunc("/config/nodes", r.withMiddleware(r.configHandler.HandleGetNodes))
	r.mux.HandleFunc("/api/nodes", r.withMiddleware(r.configHandler.HandleGetNodes))
	r.mux.HandleFunc("/config/export", r.withMiddleware(r.configHandler.HandleExportResults))
}

// withMiddleware 应用中间件
func (r *Router) withMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	corsMiddleware := middleware.CORS(middleware.DefaultCORSConfig())
	return func(w http.ResponseWriter, req *http.Request) {
		corsMiddleware(
			middleware.RequestID(
				middleware.Logging(
					middleware.Recovery(http.HandlerFunc(handler)),
				),
			),
		).ServeHTTP(w, req)
	}
}

// handleWebSocket 处理 WebSocket 连接
func (r *Router) handleWebSocket(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	r.wsHub.HandleWebSocket(w, req)
}

// GetMux 获取 HTTP 多路复用器
func (r *Router) GetMux() *http.ServeMux {
	return r.mux
}