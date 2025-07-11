package unlock

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/faceair/clash-speedtest/logger"
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

// Detector 主检测器
type Detector struct {
	config    *UnlockTestConfig
	detectors map[string]UnlockDetector
	cache     *UnlockCache
}

// NewDetector 创建新的解锁检测器
func NewDetector(config *UnlockTestConfig) *Detector {
	detector := &Detector{
		config:    config,
		detectors: make(map[string]UnlockDetector),
		cache:     NewUnlockCache(),
	}

	// 注册默认检测器
	detector.registerDefaultDetectors()

	return detector
}

// registerDefaultDetectors 注册默认的平台检测器
func (d *Detector) registerDefaultDetectors() {
	// 注册各个平台的检测器
	d.Register("Netflix", NewNetflixDetector())
	d.Register("YouTube", NewYouTubeDetector())
	d.Register("Disney+", NewDisneyDetector())
	d.Register("ChatGPT", NewOpenAIDetector())
	d.Register("Spotify", NewSpotifyDetector())
	d.Register("Bilibili", NewBilibiliDetector())
	// 可以继续添加更多平台...
}

// Register 注册平台检测器
func (d *Detector) Register(name string, detector UnlockDetector) {
	d.detectors[name] = detector
	logger.Logger.Debug("Registered unlock detector",
		slog.String("platform", name),
		slog.Int("priority", detector.GetPriority()),
	)
}

// GetSupportedPlatforms 获取支持的平台列表
func (d *Detector) GetSupportedPlatforms() []string {
	platforms := make([]string, 0, len(d.detectors))
	for name := range d.detectors {
		platforms = append(platforms, name)
	}
	return platforms
}

// DetectAll 检测所有指定平台
func (d *Detector) DetectAll(proxy constant.Proxy, platforms []string) []UnlockResult {
	if !d.config.Enabled {
		return []UnlockResult{}
	}

	logger.Logger.Info("Starting unlock detection",
		slog.String("proxy_name", proxy.Name()),
		slog.Int("platform_count", len(platforms)),
		slog.Int("concurrent", d.config.Concurrent),
	)

	start := time.Now()
	results := d.executeDetection(proxy, platforms)

	logger.Logger.Info("Unlock detection completed",
		slog.String("proxy_name", proxy.Name()),
		slog.Int("detected_platforms", len(results)),
		slog.String("duration", time.Since(start).String()),
	)

	return results
}

// executeDetection 执行并发检测
func (d *Detector) executeDetection(proxy constant.Proxy, platforms []string) []UnlockResult {
	results := make([]UnlockResult, 0, len(platforms))
	resultsChan := make(chan UnlockResult, len(platforms))

	var wg sync.WaitGroup
	controller := NewConcurrencyController(d.config.Concurrent)

	// 按优先级排序平台
	sortedPlatforms := d.sortByPriority(platforms)

	for _, platform := range sortedPlatforms {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			controller.Acquire()
			defer controller.Release()

			result := d.detectPlatform(proxy, p)
			resultsChan <- *result
		}(platform)
	}

	wg.Wait()
	close(resultsChan)

	// 收集结果
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

// detectPlatform 检测单个平台
func (d *Detector) detectPlatform(proxy constant.Proxy, platform string) *UnlockResult {
	// 检查缓存
	cacheKey := fmt.Sprintf("%s:%s", proxy.Name(), platform)
	if cached := d.cache.Get(cacheKey); cached != nil {
		logger.Logger.Debug("Using cached unlock result",
			slog.String("proxy_name", proxy.Name()),
			slog.String("platform", platform),
		)
		return cached
	}

	detector, exists := d.detectors[platform]
	if !exists {
		return &UnlockResult{
			Platform:  platform,
			Status:    StatusError,
			Message:   "Platform detector not found",
			CheckedAt: time.Now(),
		}
	}

	logger.Logger.Debug("Starting platform detection",
		slog.String("proxy_name", proxy.Name()),
		slog.String("platform", platform),
	)

	start := time.Now()
	timeout := time.Duration(d.config.Timeout) * time.Second

	var result *UnlockResult
	if d.config.RetryOnError {
		result = d.detectWithRetry(detector, proxy, timeout, 2)
	} else {
		result = detector.Detect(proxy, timeout)
	}

	result.Latency = time.Since(start).Milliseconds()
	result.CheckedAt = time.Now()

	// 缓存结果
	d.cache.Set(cacheKey, result)

	logger.Logger.Debug("Platform detection completed",
		slog.String("proxy_name", proxy.Name()),
		slog.String("platform", platform),
		slog.String("status", string(result.Status)),
		slog.String("region", result.Region),
		slog.Int64("latency_ms", result.Latency),
	)

	return result
}

// detectWithRetry 带重试的检测
func (d *Detector) detectWithRetry(detector UnlockDetector, proxy constant.Proxy, timeout time.Duration, maxRetries int) *UnlockResult {
	var lastResult *UnlockResult

	for i := 0; i <= maxRetries; i++ {
		result := detector.Detect(proxy, timeout)

		if result.Status != StatusError {
			return result
		}

		lastResult = result

		// 指数退避
		if i < maxRetries {
			sleepDuration := time.Duration(1<<uint(i)) * time.Second
			logger.Logger.Debug("Retrying platform detection",
				slog.String("platform", detector.GetPlatformName()),
				slog.Int("attempt", i+1),
				slog.String("sleep_duration", sleepDuration.String()),
			)
			time.Sleep(sleepDuration)
		}
	}

	lastResult.Message = fmt.Sprintf("Failed after %d retries: %s", maxRetries, lastResult.Message)
	return lastResult
}

// sortByPriority 按优先级排序平台
func (d *Detector) sortByPriority(platforms []string) []string {
	type platformPriority struct {
		name     string
		priority int
	}

	prioritized := make([]platformPriority, 0, len(platforms))

	for _, platform := range platforms {
		priority := 3 // 默认低优先级
		if detector, exists := d.detectors[platform]; exists {
			priority = detector.GetPriority()
		}
		prioritized = append(prioritized, platformPriority{
			name:     platform,
			priority: priority,
		})
	}

	// 排序：优先级数字越小越优先
	for i := 0; i < len(prioritized)-1; i++ {
		for j := i + 1; j < len(prioritized); j++ {
			if prioritized[i].priority > prioritized[j].priority {
				prioritized[i], prioritized[j] = prioritized[j], prioritized[i]
			}
		}
	}

	result := make([]string, len(prioritized))
	for i, p := range prioritized {
		result[i] = p.name
	}

	return result
}

// GetUnlockSummary 生成解锁结果摘要
func GetUnlockSummary(results []UnlockResult) string {
	if len(results) == 0 {
		return "N/A"
	}

	var unlocked []string
	for _, result := range results {
		if result.Status == StatusUnlocked {
			if result.Region != "" {
				unlocked = append(unlocked, fmt.Sprintf("%s:%s", result.Platform, result.Region))
			} else {
				unlocked = append(unlocked, result.Platform)
			}
		}
	}

	if len(unlocked) == 0 {
		return "None"
	}

	// 限制摘要长度
	summary := ""
	for i, item := range unlocked {
		if i > 0 {
			summary += ", "
		}
		summary += item

		// 如果摘要太长，截断并添加省略号
		if len(summary) > 100 && i < len(unlocked)-1 {
			summary += "..."
			break
		}
	}

	return summary
}

// DefaultUnlockConfig 返回默认的解锁检测配置
func DefaultUnlockConfig() *UnlockTestConfig {
	return &UnlockTestConfig{
		Enabled:       true,
		Platforms:     []string{"Netflix", "YouTube", "Disney+", "ChatGPT", "Spotify", "Bilibili"},
		Concurrent:    5,
		Timeout:       10,
		RetryOnError:  true,
		IncludeIPInfo: true,
	}
}
