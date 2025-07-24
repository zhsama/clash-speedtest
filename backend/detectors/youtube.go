package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/zhsama/clash-speedtest/unlock"
)

// TestYouTube 测试 YouTube Premium 解锁情况
func TestYouTube(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "YouTube",
	}

	req, err := http.NewRequest("GET", "https://www.youtube.com/premium", nil)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Response Error"
		return result
	}

	htmlContent := string(body)

	// 检查是否不可用
	if strings.Contains(htmlContent, "Premium is not available") ||
		strings.Contains(htmlContent, "isn't available") {
		result.Status = "Failed"
		result.Info = "Not Available"
		return result
	}

	// 检查是否可用并提取地区信息
	if strings.Contains(htmlContent, "countryCode") {
		region := extractYouTubeRegion(htmlContent)
		result.Status = "Success"
		result.Region = region
		if region == "" {
			result.Region = "Available"
		}
		return result
	}

	// 如果页面正常显示
	if resp.StatusCode == 200 && strings.Contains(htmlContent, "youtube") {
		result.Status = "Success"
		result.Region = "Available"
		return result
	}

	result.Status = "Failed"
	result.Info = "Unknown Error"
	return result
}

// extractYouTubeRegion 从YouTube响应中提取地区信息
func extractYouTubeRegion(body string) string {
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

func init() {
	// 注册 YouTube 测试
	unlock.StreamTests = append(unlock.StreamTests, TestYouTube)
}