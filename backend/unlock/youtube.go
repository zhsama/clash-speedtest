package unlock

import (
	"io"
	"strings"
	"time"

	"github.com/metacubex/mihomo/constant"
)

// YouTubeDetector YouTube Premium检测器
type YouTubeDetector struct {
	*BaseDetector
}

// NewYouTubeDetector 创建YouTube检测器
func NewYouTubeDetector() *YouTubeDetector {
	return &YouTubeDetector{
		BaseDetector: NewBaseDetector("YouTube", 1), // 高优先级
	}
}

// Detect 检测YouTube Premium解锁状态
func (d *YouTubeDetector) Detect(proxy constant.Proxy, timeout time.Duration) *UnlockResult {
	d.logDetectionStart(proxy)

	client := createHTTPClient(proxy, timeout)

	// 访问YouTube Premium页面
	resp, err := makeRequest(client, "GET", "https://www.youtube.com/premium", nil)
	if err != nil {
		result := d.createErrorResult("Failed to connect to YouTube", err)
		d.logDetectionResult(proxy, result)
		return result
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result := d.createErrorResult("Failed to read YouTube response", err)
		d.logDetectionResult(proxy, result)
		return result
	}

	bodyStr := string(body)

	var result *UnlockResult
	if strings.Contains(bodyStr, "Premium is not available") ||
		strings.Contains(bodyStr, "isn't available") {
		result = d.createResult(StatusLocked, "", "YouTube Premium not available in this region")
	} else if strings.Contains(bodyStr, "countryCode") {
		// 尝试提取国家代码
		region := d.extractYouTubeRegion(bodyStr)
		result = d.createResult(StatusUnlocked, region, "YouTube Premium available")
	} else if resp.StatusCode == 200 && strings.Contains(bodyStr, "youtube") {
		result = d.createResult(StatusUnlocked, "", "YouTube Premium available")
	} else {
		result = d.createResult(StatusFailed, "", "Unable to determine YouTube Premium status")
	}

	d.logDetectionResult(proxy, result)
	return result
}

// extractYouTubeRegion 从YouTube响应中提取地区信息
func (d *YouTubeDetector) extractYouTubeRegion(body string) string {
	// 简单的地区提取逻辑
	regions := map[string]string{
		`"countryCode":"US"`: "US",
		`"countryCode":"JP"`: "JP",
		`"countryCode":"GB"`: "GB",
		`"countryCode":"DE"`: "DE",
		`"countryCode":"CA"`: "CA",
		`"countryCode":"AU"`: "AU",
		`"countryCode":"KR"`: "KR",
		`"countryCode":"TW"`: "TW",
		`"countryCode":"HK"`: "HK",
	}

	for pattern, region := range regions {
		if strings.Contains(body, pattern) {
			return region
		}
	}

	return ""
}
