package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestSpotify 测试 Spotify 解锁情况
func TestSpotify(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Spotify",
	}

	req, err := http.NewRequest("GET", "https://open.spotify.com/", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Network Connection Error"
		return result
	}
	defer resp.Body.Close()

	// 检查是否被重定向到不可用页面
	finalURL := resp.Request.URL.String()
	if strings.Contains(finalURL, "unavailable") {
		result.Status = "Failed"
		result.Info = "Not Available"
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Response Error"
		return result
	}

	htmlContent := string(body)

	// 检查是否被阻止
	if strings.Contains(htmlContent, "not available") ||
		strings.Contains(htmlContent, "blocked") {
		result.Status = "Failed"
		result.Info = "Blocked"
		return result
	}

	// 检查是否正常显示
	if strings.Contains(htmlContent, "spotify") &&
		(strings.Contains(htmlContent, "sign up") ||
			strings.Contains(htmlContent, "login") ||
			strings.Contains(htmlContent, "premium")) {
		// 提取地区信息
		region := extractSpotifyRegion(htmlContent)
		result.Status = "Success"
		result.Region = region
		if region == "" {
			result.Region = "Available"
		}
		return result
	}

	result.Status = "Failed"
	result.Info = "Unknown Error"
	return result
}

// extractSpotifyRegion 从Spotify响应中提取地区信息
func extractSpotifyRegion(body string) string {
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

func init() {
	// 注册 Spotify 测试
	unlock.StreamTests = append(unlock.StreamTests, TestSpotify)
}