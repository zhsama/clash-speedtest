package unlock

import (
	"io"
	"strings"
	"time"

	"github.com/metacubex/mihomo/constant"
)

// NetflixDetector Netflix检测器
type NetflixDetector struct {
	*BaseDetector
}

// NewNetflixDetector 创建Netflix检测器
func NewNetflixDetector() *NetflixDetector {
	return &NetflixDetector{
		BaseDetector: NewBaseDetector("Netflix", 1), // 高优先级
	}
}

// Detect 检测Netflix解锁状态
func (d *NetflixDetector) Detect(proxy constant.Proxy, timeout time.Duration) *UnlockResult {
	d.logDetectionStart(proxy)

	client := createHTTPClient(proxy, timeout)

	// 访问Netflix原创内容页面进行检测
	resp, err := makeRequest(client, "GET", "https://www.netflix.com/title/81280792", nil)
	if err != nil {
		result := d.createErrorResult("Failed to connect to Netflix", err)
		d.logDetectionResult(proxy, result)
		return result
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result := d.createErrorResult("Failed to read Netflix response", err)
		d.logDetectionResult(proxy, result)
		return result
	}

	bodyStr := string(body)

	// 分析响应内容
	var result *UnlockResult
	if strings.Contains(bodyStr, "Not Available") ||
		strings.Contains(bodyStr, "page-404") ||
		strings.Contains(bodyStr, "NSEZ-403") {
		result = d.createResult(StatusLocked, "", "Netflix content not available in this region")
	} else if strings.Contains(bodyStr, "requestCountry") {
		// 尝试提取国家代码
		region := d.extractCountryCode(bodyStr)
		result = d.createResult(StatusUnlocked, region, "Netflix accessible")
	} else if resp.StatusCode == 200 && strings.Contains(bodyStr, "netflix") {
		result = d.createResult(StatusUnlocked, "", "Netflix accessible")
	} else {
		result = d.createResult(StatusFailed, "", "Unable to determine Netflix status")
	}

	d.logDetectionResult(proxy, result)
	return result
}

// extractCountryCode 从响应中提取国家代码
func (d *NetflixDetector) extractCountryCode(body string) string {
	// 简单的国家代码提取逻辑
	if strings.Contains(body, `"country":"US"`) || strings.Contains(body, `"requestCountry":"US"`) {
		return "US"
	}
	if strings.Contains(body, `"country":"JP"`) || strings.Contains(body, `"requestCountry":"JP"`) {
		return "JP"
	}
	if strings.Contains(body, `"country":"GB"`) || strings.Contains(body, `"requestCountry":"GB"`) {
		return "GB"
	}
	// 可以继续添加更多国家代码的识别逻辑
	return ""
}
