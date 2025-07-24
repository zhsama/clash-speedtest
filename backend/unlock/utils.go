package unlock

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/zhsama/clash-speedtest/logger"
	"github.com/metacubex/mihomo/constant"
)

// CreateHTTPClient 创建通过代理的HTTP客户端（导出版本）
func CreateHTTPClient(ctx context.Context, proxy constant.Proxy) *http.Client {
	return createHTTPClient(ctx, proxy)
}

// MakeRequest 发送HTTP请求（导出版本）
func MakeRequest(ctx context.Context, client *http.Client, method, url string, headers map[string]string) (*http.Response, error) {
	return makeRequest(ctx, client, method, url, headers)
}

// createHTTPClient 创建通过代理的HTTP客户端
func createHTTPClient(ctx context.Context, proxy constant.Proxy) *http.Client {
	// 从context中获取超时，如果没有设置则使用默认值
	timeout := 30 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			var u16Port uint16
			if portNum, err := strconv.ParseUint(port, 10, 16); err == nil {
				u16Port = uint16(portNum)
			}

			return proxy.DialContext(ctx, &constant.Metadata{
				Host:    host,
				DstPort: u16Port,
			})
		},
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 允许最多5次重定向
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

// makeRequest 发送HTTP请求
func makeRequest(ctx context.Context, client *http.Client, method, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	// 设置默认的User-Agent
	req.Header.Set("User-Agent", getRandomUserAgent())

	// 设置额外的请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 设置通用请求头以模拟真实浏览器
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	return client.Do(req)
}

// getRandomUserAgent 获取随机User-Agent
func getRandomUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	}

	// 简单的时间基准选择
	index := int(time.Now().UnixNano()) % len(userAgents)
	return userAgents[index]
}

// BaseDetector 基础检测器实现
type BaseDetector struct {
	platformName string
	priority     int
}

// NewBaseDetector 创建基础检测器
func NewBaseDetector(platformName string, priority int) *BaseDetector {
	return &BaseDetector{
		platformName: platformName,
		priority:     priority,
	}
}

// GetPlatformName 获取平台名称
func (b *BaseDetector) GetPlatformName() string {
	return b.platformName
}

// GetPriority 获取优先级
func (b *BaseDetector) GetPriority() int {
	return b.priority
}

// LogDetectionStart 记录检测开始日志（导出版本）
func (b *BaseDetector) LogDetectionStart(proxy constant.Proxy) {
	b.logDetectionStart(proxy)
}

// LogDetectionResult 记录检测结果日志（导出版本）
func (b *BaseDetector) LogDetectionResult(proxy constant.Proxy, result *UnlockResult) {
	b.logDetectionResult(proxy, result)
}

// CreateErrorResult 创建错误结果（导出版本）
func (b *BaseDetector) CreateErrorResult(message string, err error) *UnlockResult {
	return b.createErrorResult(message, err)
}

// CreateResult 创建检测结果（导出版本）
func (b *BaseDetector) CreateResult(status UnlockStatus, region, message string) *UnlockResult {
	return b.createResult(status, region, message)
}

// logDetectionStart 记录检测开始日志
func (b *BaseDetector) logDetectionStart(proxy constant.Proxy) {
	logger.Logger.Debug("Platform detection started",
		slog.String("platform", b.platformName),
		slog.String("proxy_name", proxy.Name()),
		slog.String("proxy_type", proxy.Type().String()),
	)
}

// logDetectionResult 记录检测结果日志
func (b *BaseDetector) logDetectionResult(proxy constant.Proxy, result *UnlockResult) {
	logger.Logger.Debug("Platform detection result",
		slog.String("platform", b.platformName),
		slog.String("proxy_name", proxy.Name()),
		slog.String("status", string(result.Status)),
		slog.String("region", result.Region),
		slog.String("message", result.Message),
	)
}

// createErrorResult 创建错误结果
func (b *BaseDetector) createErrorResult(message string, err error) *UnlockResult {
	fullMessage := message
	if err != nil {
		fullMessage = fmt.Sprintf("%s: %v", message, err)
	}

	return &UnlockResult{
		Platform: b.platformName,
		Status:   StatusError,
		Message:  fullMessage,
	}
}

// createResult 创建检测结果
func (b *BaseDetector) createResult(status UnlockStatus, region, message string) *UnlockResult {
	return &UnlockResult{
		Platform: b.platformName,
		Status:   status,
		Region:   region,
		Message:  message,
	}
}
