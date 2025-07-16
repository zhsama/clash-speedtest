package detectors

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/faceair/clash-speedtest/unlock"
	"github.com/metacubex/mihomo/constant"
)

// YouTubeDetector YouTube Premium检测器
type YouTubeDetector struct {
	*unlock.BaseDetector
}

// NewYouTubeDetector 创建YouTube检测器
func NewYouTubeDetector() *YouTubeDetector {
	return &YouTubeDetector{
		BaseDetector: unlock.NewBaseDetector("YouTube", 1), // 高优先级
	}
}

// Detect 检测YouTube Premium解锁状态
func (d *YouTubeDetector) Detect(ctx context.Context, proxy constant.Proxy) *unlock.UnlockResult {
	d.LogDetectionStart(proxy)

	client := unlock.CreateHTTPClient(ctx, proxy)

	// 访问YouTube Premium页面
	resp, err := unlock.MakeRequest(ctx, client, "GET", "https://www.youtube.com/premium", nil)
	if err != nil {
		result := d.CreateErrorResult("Failed to connect to YouTube", err)
		d.LogDetectionResult(proxy, result)
		return result
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result := d.CreateErrorResult("Failed to read YouTube response", err)
		d.LogDetectionResult(proxy, result)
		return result
	}

	bodyStr := string(body)

	var result *unlock.UnlockResult
	if strings.Contains(bodyStr, "Premium is not available") ||
		strings.Contains(bodyStr, "isn't available") {
		result = d.CreateResult(unlock.StatusLocked, "", "YouTube Premium not available in this region")
	} else if strings.Contains(bodyStr, "countryCode") {
		// 尝试提取国家代码
		region := d.extractYouTubeRegion(bodyStr)
		result = d.CreateResult(unlock.StatusUnlocked, region, "YouTube Premium available")
	} else if resp.StatusCode == 200 && strings.Contains(bodyStr, "youtube") {
		result = d.CreateResult(unlock.StatusUnlocked, "", "YouTube Premium available")
	} else {
		result = d.CreateResult(unlock.StatusFailed, "", "Unable to determine YouTube Premium status")
	}

	result.CheckedAt = time.Now()
	d.LogDetectionResult(proxy, result)
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

// init 函数用于自动注册检测器
func init() {
	unlock.Register(NewYouTubeDetector())
}