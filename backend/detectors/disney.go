package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestDisney 测试 Disney+ 解锁情况
func TestDisney(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Disney+",
	}

	req, err := http.NewRequest("GET", "https://www.disneyplus.com", nil)
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

	// 检查重定向URL
	finalURL := resp.Request.URL.String()
	if strings.Contains(finalURL, "/unavailable") ||
		strings.Contains(finalURL, "/blocked") ||
		strings.Contains(finalURL, "/unsupported") {
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
		strings.Contains(htmlContent, "access denied") {
		result.Status = "Failed"
		result.Info = "Blocked"
		return result
	}

	// 检查是否正常显示
	if strings.Contains(htmlContent, "sign up") ||
		strings.Contains(htmlContent, "subscribe") ||
		strings.Contains(htmlContent, "bundle") {
		// 提取地区信息
		region := extractDisneyRegion(finalURL, htmlContent)
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

// extractDisneyRegion 从Disney+响应中提取地区信息
func extractDisneyRegion(url, body string) string {
	// 从URL中提取地区
	if strings.Contains(url, "disneyplus.com") {
		return "US"
	}
	if strings.Contains(url, ".jp/") {
		return "JP"
	}
	if strings.Contains(url, ".co.uk/") {
		return "GB"
	}
	if strings.Contains(url, ".ca/") {
		return "CA"
	}
	if strings.Contains(url, ".com.au/") {
		return "AU"
	}

	// 从内容中提取地区信息
	if strings.Contains(body, `"market":"US"`) {
		return "US"
	}
	if strings.Contains(body, `"market":"JP"`) {
		return "JP"
	}
	if strings.Contains(body, `"market":"GB"`) {
		return "GB"
	}

	return ""
}

func init() {
	// 注册 Disney+ 测试
	unlock.StreamTests = append(unlock.StreamTests, TestDisney)
}