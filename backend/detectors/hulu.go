package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/zhsama/clash-speedtest/unlock"
)

// TestHulu 测试 Hulu 解锁情况
func TestHulu(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Hulu",
	}

	req, err := http.NewRequest("GET", "https://www.hulu.com/", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Network Connection Error"
		return result
	}
	defer resp.Body.Close()

	// 检查重定向
	location := resp.Request.URL.String()
	if strings.Contains(location, "/geo-block") {
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Response Error"
		return result
	}

	htmlContent := string(body)

	switch {
	case strings.Contains(htmlContent, "geo-not-available"):
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	case strings.Contains(htmlContent, "start-watching") ||
		strings.Contains(htmlContent, "watch-live-tv") ||
		strings.Contains(htmlContent, "welcome-page"):
		result.Status = "Success"
		result.Region = "US"
		return result
	}

	result.Status = "Failed"
	result.Info = "Not Available"
	return result
}

func init() {
	// 注册 Hulu 测试
	unlock.StreamTests = append(unlock.StreamTests, TestHulu)
}
