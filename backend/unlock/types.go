package unlock

import (
	"time"

	"github.com/metacubex/mihomo/constant"
)

// TestMode 测试模式枚举
type TestMode string

const (
	TestModeSpeedOnly  TestMode = "speed_only"  // 仅测速
	TestModeUnlockOnly TestMode = "unlock_only" // 仅解锁
	TestModeBoth       TestMode = "both"        // 两者都测（默认）
)

// UnlockStatus 解锁状态枚举
type UnlockStatus string

const (
	StatusUnlocked UnlockStatus = "unlocked" // 已解锁
	StatusLocked   UnlockStatus = "locked"   // 被锁定/不可用
	StatusFailed   UnlockStatus = "failed"   // 检测失败
	StatusError    UnlockStatus = "error"    // 检测错误
)

// UnlockResult 单个平台的解锁检测结果
type UnlockResult struct {
	Platform  string       `json:"platform"`   // 平台名称
	Status    UnlockStatus `json:"status"`     // 状态
	Region    string       `json:"region"`     // 解锁地区
	Message   string       `json:"message"`    // 额外信息
	Latency   int64        `json:"latency_ms"` // 检测延迟
	CheckedAt time.Time    `json:"checked_at"` // 检测时间
}

// UnlockTestConfig 解锁检测配置
type UnlockTestConfig struct {
	Enabled       bool     `json:"enabled"`         // 是否启用
	Platforms     []string `json:"platforms"`       // 要检测的平台列表
	Concurrent    int      `json:"concurrent"`      // 并发检测数
	Timeout       int      `json:"timeout"`         // 单个检测超时（秒）
	RetryOnError  bool     `json:"retry_on_error"`  // 错误时重试
	IncludeIPInfo bool     `json:"include_ip_info"` // 包含 IP 信息
}

// IPInfo IP 信息结构
type IPInfo struct {
	IP        string `json:"ip"`
	Country   string `json:"country"`
	City      string `json:"city"`
	ISP       string `json:"isp"`
	RiskScore int    `json:"risk_score"` // IP 风险值 (0-100)
}

// UnlockDetector 解锁检测器接口
type UnlockDetector interface {
	Detect(proxy constant.Proxy, timeout time.Duration) *UnlockResult
	GetPlatformName() string
	GetPriority() int // 检测优先级 (1=高, 2=中, 3=低)
}

// DetectorRegistry 检测器注册表接口
type DetectorRegistry interface {
	RegisterDetector(detector UnlockDetector)
	GetDetector(platform string) (UnlockDetector, bool)
	GetAllDetectors() []UnlockDetector
	GetDetectorsByPriority() []UnlockDetector
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Result    *UnlockResult
	ExpiresAt time.Time
}

// UnlockCache 解锁缓存接口
type UnlockCache interface {
	Get(key string) *UnlockResult
	Set(key string, result *UnlockResult, duration time.Duration)
	Delete(key string)
	Clear()
	Stats() CacheStats
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Hits      int64     `json:"hits"`
	Misses    int64     `json:"misses"`
	Entries   int       `json:"entries"`
	HitRatio  float64   `json:"hit_ratio"`
	CreatedAt time.Time `json:"created_at"`
}

// PlatformCategory 平台分类
type PlatformCategory string

const (
	CategoryStreaming PlatformCategory = "streaming" // 流媒体
	CategorySocial    PlatformCategory = "social"    // 社交
	CategoryRegional  PlatformCategory = "regional"  // 地区性
	CategoryGaming    PlatformCategory = "gaming"    // 游戏
	CategoryOther     PlatformCategory = "other"     // 其他
)

// PlatformInfo 平台信息
type PlatformInfo struct {
	Name        string           `json:"name"`
	DisplayName string           `json:"display_name"`
	Category    PlatformCategory `json:"category"`
	Priority    int              `json:"priority"`
	Enabled     bool             `json:"enabled"`
	Description string           `json:"description"`
}

// ConcurrencyController 并发控制器
type ConcurrencyController struct {
	semaphore chan struct{}
}

// NewConcurrencyController 创建并发控制器
func NewConcurrencyController(maxConcurrent int) *ConcurrencyController {
	return &ConcurrencyController{
		semaphore: make(chan struct{}, maxConcurrent),
	}
}

// Acquire 获取信号量
func (c *ConcurrencyController) Acquire() {
	c.semaphore <- struct{}{}
}

// Release 释放信号量
func (c *ConcurrencyController) Release() {
	<-c.semaphore
}