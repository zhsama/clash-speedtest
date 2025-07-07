package websocket

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/faceair/clash-speedtest/logger"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// MessageType defines the types of WebSocket messages
type MessageType string

const (
	MessageTypeTestStart     MessageType = "test_start"
	MessageTypeTestProgress  MessageType = "test_progress"
	MessageTypeTestResult    MessageType = "test_result"
	MessageTypeTestComplete  MessageType = "test_complete"
	MessageTypeTestCancelled MessageType = "test_cancelled"
	MessageTypeError         MessageType = "error"
	MessageTypeStopTest      MessageType = "stop_test"
)

// WebSocketMessage represents a message sent via WebSocket
type WebSocketMessage struct {
	Type      MessageType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// TestStartData contains information about test initialization
type TestStartData struct {
	TotalProxies int `json:"total_proxies"`
	Config       struct {
		ConfigPaths      string  `json:"config_paths"`
		FilterRegex      string  `json:"filter_regex"`
		ServerURL        string  `json:"server_url"`
		DownloadSize     int     `json:"download_size"`
		UploadSize       int     `json:"upload_size"`
		Timeout          int     `json:"timeout"`
		Concurrent       int     `json:"concurrent"`
		MaxLatency       int     `json:"max_latency"`
		MinDownloadSpeed float64 `json:"min_download_speed"`
		MinUploadSpeed   float64 `json:"min_upload_speed"`
		StashCompatible  bool    `json:"stash_compatible"`
	} `json:"config"`
}

// TestProgressData contains information about current testing progress
type TestProgressData struct {
	CurrentProxy    string  `json:"current_proxy"`
	CompletedCount  int     `json:"completed_count"`
	TotalCount      int     `json:"total_count"`
	ProgressPercent float64 `json:"progress_percent"`
	Status          string  `json:"status"`
}

// TestResultData contains the result of a single proxy test
type TestResultData struct {
	ProxyName         string  `json:"proxy_name"`
	ProxyType         string  `json:"proxy_type"`
	Latency           int64   `json:"latency_ms"`
	Jitter            int64   `json:"jitter_ms"`
	PacketLoss        float64 `json:"packet_loss"`
	DownloadSpeed     float64 `json:"download_speed"`
	UploadSpeed       float64 `json:"upload_speed"`
	DownloadSpeedMbps float64 `json:"download_speed_mbps"`
	UploadSpeedMbps   float64 `json:"upload_speed_mbps"`
	Status            string  `json:"status"` // "success", "failed", "timeout"
	// 新增错误诊断字段
	ErrorStage        string  `json:"error_stage,omitempty"`   // 错误阶段
	ErrorCode         string  `json:"error_code,omitempty"`    // 错误代码  
	ErrorMessage      string  `json:"error_message,omitempty"` // 错误消息
}

// TestCompleteData contains summary information when all tests are done
type TestCompleteData struct {
	TotalTested       int     `json:"total_tested"`
	SuccessfulTests   int     `json:"successful_tests"`
	FailedTests       int     `json:"failed_tests"`
	TotalDuration     string  `json:"total_duration"`
	AverageLatency    float64 `json:"average_latency"`
	AverageDownload   float64 `json:"average_download_mbps"`
	AverageUpload     float64 `json:"average_upload_mbps"`
	BestProxy         string  `json:"best_proxy"`
	BestDownloadSpeed float64 `json:"best_download_speed_mbps"`
}

// TestCancelledData contains information when tests are cancelled
type TestCancelledData struct {
	Message         string `json:"message"`
	CompletedTests  int    `json:"completed_tests"`
	TotalTests      int    `json:"total_tests"`
	PartialDuration string `json:"partial_duration"`
}

// ErrorData contains error information
type ErrorData struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Client represents a WebSocket client connection
type Client struct {
	conn   net.Conn
	id     string
	closed bool
	mu     sync.Mutex
}

// Hub manages WebSocket connections
type Hub struct {
	clients        map[string]*Client
	register       chan *Client
	unregister     chan *Client
	broadcast      chan []byte
	messageHandler func(string, []byte)
	mu             sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
	}
}

// SetMessageHandler sets the message handler function
func (h *Hub) SetMessageHandler(handler func(string, []byte)) {
	h.messageHandler = handler
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.id] = client
			h.mu.Unlock()
			logger.Logger.Info("WebSocket client connected", 
				slog.String("client_id", client.id),
				slog.Int("total_clients", len(h.clients)),
			)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.id]; ok {
				delete(h.clients, client.id)
				client.close()
			}
			h.mu.Unlock()
			logger.Logger.Info("WebSocket client disconnected", 
				slog.String("client_id", client.id),
				slog.Int("total_clients", len(h.clients)),
			)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				if !client.sendMessage(message) {
					// Client disconnected, schedule for removal
					go func(c *Client) {
						h.unregister <- c
					}(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastMessage sends a message to all connected clients
func (h *Hub) BroadcastMessage(msgType MessageType, data interface{}) {
	message := WebSocketMessage{
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      data,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		logger.LogError("Failed to marshal WebSocket message", err)
		return
	}

	logger.Logger.Debug("Broadcasting WebSocket message", 
		slog.String("message_type", string(msgType)),
		slog.Int("clients_count", len(h.clients)),
	)

	select {
	case h.broadcast <- jsonData:
	default:
		logger.Logger.Warn("WebSocket broadcast channel full, dropping message")
	}
}

// sendMessage sends a message to a specific client
func (c *Client) sendMessage(message []byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return false
	}

	err := wsutil.WriteServerMessage(c.conn, ws.OpText, message)
	if err != nil {
		logger.Logger.Debug("Failed to send WebSocket message", 
			slog.String("client_id", c.id),
			slog.String("error", err.Error()),
		)
		return false
	}
	return true
}

// close closes the client connection
func (c *Client) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.closed {
		c.closed = true
		c.conn.Close()
	}
}

// HandleWebSocket handles WebSocket upgrade and connection
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		logger.LogError("Failed to upgrade WebSocket connection", err)
		http.Error(w, "Failed to upgrade connection", http.StatusBadRequest)
		return
	}

	clientID := generateClientID()
	client := &Client{
		conn: conn,
		id:   clientID,
	}

	h.register <- client

	// Start a goroutine to handle client disconnection and messages
	go func() {
		defer func() {
			h.unregister <- client
		}()

		// Read messages from client
		for {
			msgData, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				logger.Logger.Debug("WebSocket client disconnected", 
					slog.String("client_id", clientID),
					slog.String("error", err.Error()),
				)
				break
			}

			// Try to parse and handle the message
			if h.messageHandler != nil && len(msgData) > 0 {
				var msg struct {
					Type string `json:"type"`
				}
				if err := json.Unmarshal(msgData, &msg); err == nil {
					h.messageHandler(msg.Type, msgData)
				}
			}
		}
	}()
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + 
		   time.Now().Format("000000")[3:]
}