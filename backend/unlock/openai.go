package unlock

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/metacubex/mihomo/constant"
)

// OpenAIDetector ChatGPT/OpenAI检测器
type OpenAIDetector struct {
	*BaseDetector
}

// NewOpenAIDetector 创建OpenAI检测器
func NewOpenAIDetector() *OpenAIDetector {
	return &OpenAIDetector{
		BaseDetector: NewBaseDetector("ChatGPT", 1), // 高优先级
	}
}

// Detect 检测ChatGPT/OpenAI解锁状态
func (d *OpenAIDetector) Detect(proxy constant.Proxy, timeout time.Duration) *UnlockResult {
	d.logDetectionStart(proxy)

	client := createHTTPClient(proxy, timeout)

	// 方法1: 检查API合规端点
	result1 := d.checkAPICompliance(client)
	if result1.Status == StatusUnlocked {
		d.logDetectionResult(proxy, result1)
		return result1
	}

	// 方法2: 检查iOS ChatGPT端点
	result2 := d.checkiOSEndpoint(client)

	d.logDetectionResult(proxy, result2)
	return result2
}

// checkAPICompliance 检查API合规端点
func (d *OpenAIDetector) checkAPICompliance(client *http.Client) *UnlockResult {
	resp, err := makeRequest(client, "GET", "https://api.openai.com/compliance/cookie_requirements", nil)
	if err != nil {
		return d.createErrorResult("Failed to connect to OpenAI API", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return d.createErrorResult("Failed to read OpenAI API response", err)
	}

	bodyStr := string(body)

	if strings.Contains(bodyStr, "unsupported_country") ||
		strings.Contains(bodyStr, "vpn") {
		return d.createResult(StatusLocked, "", "ChatGPT not available in this region")
	}

	// 尝试解析JSON响应获取地区信息
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err == nil {
		if country, ok := apiResponse["country"].(string); ok {
			return d.createResult(StatusUnlocked, strings.ToUpper(country), "ChatGPT available")
		}
	}

	if resp.StatusCode == 200 {
		return d.createResult(StatusUnlocked, "", "ChatGPT available")
	}

	return d.createResult(StatusFailed, "", "Unable to determine ChatGPT status")
}

// checkiOSEndpoint 检查iOS ChatGPT端点
func (d *OpenAIDetector) checkiOSEndpoint(client *http.Client) *UnlockResult {
	resp, err := makeRequest(client, "GET", "https://ios.chat.openai.com/", nil)
	if err != nil {
		return d.createErrorResult("Failed to connect to ChatGPT iOS", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return d.createErrorResult("Failed to read ChatGPT iOS response", err)
	}

	bodyStr := string(body)

	if strings.Contains(bodyStr, "unsupported_country") ||
		strings.Contains(bodyStr, "vpn") ||
		strings.Contains(bodyStr, "blocked") {
		return d.createResult(StatusLocked, "", "ChatGPT blocked in this region")
	}

	if resp.StatusCode == 200 && !strings.Contains(bodyStr, "error") {
		return d.createResult(StatusUnlocked, "", "ChatGPT available")
	}

	return d.createResult(StatusFailed, "", "Unable to determine ChatGPT status")
}
