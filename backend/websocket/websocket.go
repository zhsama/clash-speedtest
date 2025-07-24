package websocket

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/zhsama/clash-speedtest/logger"
	"github.com/zhsama/clash-speedtest/speedtester"
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
	// 新增细化的进度消息类型
	MessageTypeLatencyStart   MessageType = "latency_start"
	MessageTypeLatencyResult  MessageType = "latency_result"
	MessageTypeDownloadStart  MessageType = "download_start"
	MessageTypeDownloadResult MessageType = "download_result"
	MessageTypeUploadStart    MessageType = "upload_start"
	MessageTypeUploadResult   MessageType = "upload_result"
	MessageTypeProxySkipped   MessageType = "proxy_skipped"
)

// WebSocketMessage represents a message sent via WebSocket
type WebSocketMessage struct {
	Type      MessageType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      any         `json:"data"`
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
	// 新增详细进度信息
	CurrentStage  string  `json:"current_stage,omitempty"`  // "latency", "download", "upload"
	StageProgress float64 `json:"stage_progress,omitempty"` // 当前阶段进度 0-100
	EstimatedTime int     `json:"estimated_time,omitempty"` // 预计剩余时间(秒)
}

// LatencyTestData contains latency test specific data
type LatencyTestData struct {
	ProxyName      string  `json:"proxy_name"`
	ProxyType      string  `json:"proxy_type"`
	AttemptCount   int     `json:"attempt_count"`    // 当前尝试次数
	TotalAttempts  int     `json:"total_attempts"`   // 总尝试次数
	CurrentLatency int64   `json:"current_latency"`  // 当前延迟(ms)
	AverageLatency int64   `json:"average_latency"`  // 平均延迟(ms)
	PacketLossRate float64 `json:"packet_loss_rate"` // 当前丢包率
}

// BandwidthTestData contains bandwidth test specific data
type BandwidthTestData struct {
	ProxyName        string  `json:"proxy_name"`
	ProxyType        string  `json:"proxy_type"`
	TestType         string  `json:"test_type"`         // "download" or "upload"
	BytesTransferred int64   `json:"bytes_transferred"` // 已传输字节数
	TotalBytes       int64   `json:"total_bytes"`       // 总字节数
	CurrentSpeed     float64 `json:"current_speed"`     // 当前速度(bytes/s)
	ElapsedTime      int64   `json:"elapsed_time"`      // 已用时间(ms)
	Concurrent       int     `json:"concurrent"`        // 并发数
}

// ProxySkippedData contains information about skipped proxies
type ProxySkippedData struct {
	ProxyName string `json:"proxy_name"`
	ProxyType string `json:"proxy_type"`
	Reason    string `json:"reason"` // "latency_failed", "min_speed_not_met", "timeout", etc.
	Details   string `json:"details,omitempty"`
}

// TestResultData contains the result of a single proxy test
type TestResultData struct {
	ProxyName         string  `json:"proxy_name"`
	ProxyType         string  `json:"proxy_type"`
	ProxyIP           string  `json:"proxy_ip,omitempty"` // 新增代理IP地址
	Latency           int64   `json:"latency_ms"`
	Jitter            int64   `json:"jitter_ms"`
	PacketLoss        float64 `json:"packet_loss"`
	DownloadSpeed     float64 `json:"download_speed"`
	UploadSpeed       float64 `json:"upload_speed"`
	DownloadSpeedMbps float64 `json:"download_speed_mbps"`
	UploadSpeedMbps   float64 `json:"upload_speed_mbps"`
	Status            string  `json:"status"` // "success", "failed", "timeout"
	// 新增错误诊断字段
	ErrorStage   string `json:"error_stage,omitempty"`   // 错误阶段
	ErrorCode    string `json:"error_code,omitempty"`    // 错误代码
	ErrorMessage string `json:"error_message,omitempty"` // 错误消息
	UnlockResults []UnlockResult   `json:"unlock_results,omitempty"` // 解锁检测结果
	UnlockSummary *UnlockSummary   `json:"unlock_summary,omitempty"` // 解锁摘要
}

// UnlockResult 前端期望的解锁结果格式
type UnlockResult struct {
	Platform     string `json:"platform"`
	Supported    bool   `json:"supported"`
	Region       string `json:"region,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// UnlockSummary 前端期望的解锁摘要格式
type UnlockSummary struct {
	SupportedPlatforms   []string `json:"supported_platforms"`
	UnsupportedPlatforms []string `json:"unsupported_platforms"`
	TotalTested          int      `json:"total_tested"`
	TotalSupported       int      `json:"total_supported"`
}

// ConvertSpeedtesterUnlockResults 将speedtester的unlock结果转换为websocket格式
func ConvertSpeedtesterUnlockResults(results []speedtester.FrontendUnlockResult) []UnlockResult {
	converted := make([]UnlockResult, len(results))
	for i, result := range results {
		converted[i] = UnlockResult{
			Platform:     result.Platform,
			Supported:    result.Supported,
			Region:       result.Region,
			ErrorMessage: result.ErrorMessage,
		}
	}
	return converted
}

// ConvertSpeedtesterUnlockSummary 将speedtester的unlock摘要转换为websocket格式
func ConvertSpeedtesterUnlockSummary(summary speedtester.FrontendUnlockSummary) *UnlockSummary {
	return &UnlockSummary{
		SupportedPlatforms:   summary.SupportedPlatforms,
		UnsupportedPlatforms: summary.UnsupportedPlatforms,
		TotalTested:          summary.TotalTested,
		TotalSupported:       summary.TotalSupported,
	}
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
func (h *Hub) BroadcastMessage(msgType MessageType, data any) {
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
