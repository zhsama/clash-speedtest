package detectors

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/faceair/clash-speedtest/unlock"
	"github.com/metacubex/mihomo/constant"
)

// OpenAIDetector ChatGPT/OpenAI检测器
type OpenAIDetector struct {
	*unlock.BaseDetector
}

// NewOpenAIDetector 创建OpenAI检测器
func NewOpenAIDetector() *OpenAIDetector {
	return &OpenAIDetector{
		BaseDetector: unlock.NewBaseDetector("ChatGPT", 1), // 高优先级
	}
}

// Detect 检测ChatGPT/OpenAI解锁状态
func (d *OpenAIDetector) Detect(ctx context.Context, proxy constant.Proxy) *unlock.UnlockResult {
	d.LogDetectionStart(proxy)

	client := unlock.CreateHTTPClient(ctx, proxy)

	// 方法1: 检查API合规端点
	result1 := d.checkAPICompliance(ctx, client)
	if result1.Status == unlock.StatusUnlocked {
		result1.CheckedAt = time.Now()
		d.LogDetectionResult(proxy, result1)
		return result1
	}

	// 方法2: 检查iOS ChatGPT端点
	result2 := d.checkiOSEndpoint(ctx, client)

	result2.CheckedAt = time.Now()
	d.LogDetectionResult(proxy, result2)
	return result2
}

// checkAPICompliance 检查API合规端点
func (d *OpenAIDetector) checkAPICompliance(ctx context.Context, client *http.Client) *unlock.UnlockResult {
	resp, err := unlock.MakeRequest(ctx, client, "GET", "https://api.openai.com/compliance/cookie_requirements", nil)
	if err != nil {
		return d.CreateErrorResult("Failed to connect to OpenAI API", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return d.CreateErrorResult("Failed to read OpenAI API response", err)
	}

	bodyStr := string(body)

	if strings.Contains(bodyStr, "unsupported_country") ||
		strings.Contains(bodyStr, "vpn") {
		return d.CreateResult(unlock.StatusLocked, "", "ChatGPT not available in this region")
	}

	// 尝试解析JSON响应获取地区信息
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err == nil {
		if country, ok := apiResponse["country"].(string); ok {
			return d.CreateResult(unlock.StatusUnlocked, strings.ToUpper(country), "ChatGPT available")
		}
	}

	if resp.StatusCode == 200 {
		return d.CreateResult(unlock.StatusUnlocked, "", "ChatGPT available")
	}

	return d.CreateResult(unlock.StatusFailed, "", "Unable to determine ChatGPT status")
}

// checkiOSEndpoint 检查iOS ChatGPT端点
func (d *OpenAIDetector) checkiOSEndpoint(ctx context.Context, client *http.Client) *unlock.UnlockResult {
	resp, err := unlock.MakeRequest(ctx, client, "GET", "https://ios.chat.openai.com/", nil)
	if err != nil {
		return d.CreateErrorResult("Failed to connect to ChatGPT iOS", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return d.CreateErrorResult("Failed to read ChatGPT iOS response", err)
	}

	bodyStr := string(body)

	if strings.Contains(bodyStr, "unsupported_country") ||
		strings.Contains(bodyStr, "vpn") ||
		strings.Contains(bodyStr, "blocked") {
		return d.CreateResult(unlock.StatusLocked, "", "ChatGPT blocked in this region")
	}

	if resp.StatusCode == 200 && !strings.Contains(bodyStr, "error") {
		return d.CreateResult(unlock.StatusUnlocked, "", "ChatGPT available")
	}

	return d.CreateResult(unlock.StatusFailed, "", "Unable to determine ChatGPT status")
}

// init 函数用于自动注册检测器
func init() {
	unlock.Register(NewOpenAIDetector())
}