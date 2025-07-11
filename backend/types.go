package main

import (
	"context"
	"net/http"
	"time"

	"github.com/faceair/clash-speedtest/speedtester"
)

// TestRequest 表示测试请求的结构
type TestRequest struct {
	ConfigPaths      string   `json:"configPaths"`
	FilterRegex      string   `json:"filterRegex"`
	IncludeNodes     []string `json:"includeNodes"`
	ExcludeNodes     []string `json:"excludeNodes"`
	ProtocolFilter   []string `json:"protocolFilter"`
	ServerURL        string   `json:"serverUrl"`
	DownloadSize     int      `json:"downloadSize"`
	UploadSize       int      `json:"uploadSize"`
	Timeout          int      `json:"timeout"`
	Concurrent       int      `json:"concurrent"`
	MaxLatency       int      `json:"maxLatency"`
	MinDownloadSpeed float64  `json:"minDownloadSpeed"`
	MinUploadSpeed   float64  `json:"minUploadSpeed"`
	StashCompatible  bool     `json:"stashCompatible"`
	// 新增字段
	FastMode     bool   `json:"fastMode"`     // 快速模式：只测试延迟
	RenameNodes  bool   `json:"renameNodes"`  // 节点重命名：添加地理位置信息
	ExportFormat string `json:"exportFormat"` // 导出格式：json, csv, yaml, clash
	ExportPath   string `json:"exportPath"`   // 导出路径
	// 解锁检测相关字段
	TestMode         string   `json:"testMode"`         // 测试模式：speed_only, unlock_only, both
	UnlockEnabled    bool     `json:"unlockEnabled"`    // 是否启用解锁检测
	UnlockPlatforms  []string `json:"unlockPlatforms"`  // 要检测的平台列表
	UnlockConcurrent int      `json:"unlockConcurrent"` // 解锁检测并发数
	UnlockTimeout    int      `json:"unlockTimeout"`    // 解锁检测超时时间
	UnlockRetry      bool     `json:"unlockRetry"`      // 解锁检测失败时是否重试
}

// TestResponse 表示测试响应的结构
type TestResponse struct {
	Success bool                  `json:"success"`
	Error   string                `json:"error,omitempty"`
	Results []*speedtester.Result `json:"results,omitempty"`
}

// TestTask 表示测试任务的结构
type TestTask struct {
	ID         string
	Config     TestRequest
	Context    context.Context
	CancelFunc context.CancelFunc
	Status     string // pending, running, completed, cancelled
	StartTime  time.Time
}

// loggingResponseWriter HTTP响应写入器，用于记录响应状态
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader 记录HTTP状态码
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// ProtocolsResponse 协议列表响应结构
type ProtocolsResponse struct {
	Success   bool     `json:"success"`
	Error     string   `json:"error,omitempty"`
	Protocols []string `json:"protocols,omitempty"`
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

// NodesResponse 节点列表响应结构
type NodesResponse struct {
	Success bool       `json:"success"`
	Error   string     `json:"error,omitempty"`
	Nodes   []NodeInfo `json:"nodes,omitempty"`
}
