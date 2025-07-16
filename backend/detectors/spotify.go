package detectors

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/faceair/clash-speedtest/unlock"
	"github.com/metacubex/mihomo/constant"
)

// SpotifyDetector Spotify检测器
type SpotifyDetector struct {
	*unlock.BaseDetector
}

// NewSpotifyDetector 创建Spotify检测器
func NewSpotifyDetector() *SpotifyDetector {
	return &SpotifyDetector{
		BaseDetector: unlock.NewBaseDetector("Spotify", 2), // 中优先级
	}
}

// Detect 检测Spotify解锁状态
func (d *SpotifyDetector) Detect(ctx context.Context, proxy constant.Proxy) *unlock.UnlockResult {
	d.LogDetectionStart(proxy)

	client := unlock.CreateHTTPClient(ctx, proxy)

	// 访问Spotify主页
	resp, err := unlock.MakeRequest(ctx, client, "GET", "https://open.spotify.com/", nil)
	if err != nil {
		result := d.CreateErrorResult("Failed to connect to Spotify", err)
		d.LogDetectionResult(proxy, result)
		return result
	}
	defer resp.Body.Close()

	// 检查是否被重定向到不可用页面
	finalURL := resp.Request.URL.String()
	if strings.Contains(finalURL, "unavailable") {
		result := d.CreateResult(unlock.StatusLocked, "", "Spotify not available in this region")
		d.LogDetectionResult(proxy, result)
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result := d.CreateErrorResult("Failed to read Spotify response", err)
		d.LogDetectionResult(proxy, result)
		return result
	}

	bodyStr := string(body)

	var result *unlock.UnlockResult
	if strings.Contains(bodyStr, "not available") ||
		strings.Contains(bodyStr, "blocked") {
		result = d.CreateResult(unlock.StatusLocked, "", "Spotify blocked in this region")
	} else if strings.Contains(bodyStr, "spotify") &&
		(strings.Contains(bodyStr, "sign up") ||
			strings.Contains(bodyStr, "login") ||
			strings.Contains(bodyStr, "premium")) {
		// 提取地区信息
		region := d.extractSpotifyRegion(bodyStr)
		result = d.CreateResult(unlock.StatusUnlocked, region, "Spotify available")
	} else {
		result = d.CreateResult(unlock.StatusFailed, "", "Unable to determine Spotify status")
	}

	result.CheckedAt = time.Now()
	d.LogDetectionResult(proxy, result)
	return result
}

// extractSpotifyRegion 从Spotify响应中提取地区信息
func (d *SpotifyDetector) extractSpotifyRegion(body string) string {
	// 从响应中查找地区标识
	regions := map[string]string{
		`"country":"US"`:   "US",
		`"country":"GB"`:   "GB",
		`"country":"DE"`:   "DE",
		`"country":"JP"`:   "JP",
		`"country":"CA"`:   "CA",
		`"country":"AU"`:   "AU",
		`"country":"KR"`:   "KR",
		`"country":"TW"`:   "TW",
		`"country":"HK"`:   "HK",
		`"country":"SG"`:   "SG",
		`"locale":"en-US"`: "US",
		`"locale":"en-GB"`: "GB",
		`"locale":"de-DE"`: "DE",
		`"locale":"ja-JP"`: "JP",
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
	unlock.Register(NewSpotifyDetector())
}