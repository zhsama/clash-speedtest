package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/zhsama/clash-speedtest/unlock"
)

// TestHotstar 测试 Hotstar 解锁情况
func TestHotstar(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Hotstar",
	}

	req, err := http.NewRequest("GET", "https://www.hotstar.com/", nil)
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

	// 检查是否解锁
	switch {
	case strings.Contains(htmlContent, "Not Available"):
		fallthrough
	case strings.Contains(htmlContent, "unavailable"):
		result.Status = "Failed"
		result.Info = "Not Available"
		return result

	case strings.Contains(htmlContent, "blocked"):
		fallthrough
	case strings.Contains(htmlContent, "restricted"):
		result.Status = "Failed"
		result.Info = "Blocked"
		return result

	case strings.Contains(htmlContent, "geo-blocked"):
		fallthrough
	case strings.Contains(htmlContent, "location"):
		result.Status = "Failed"
		result.Info = "Geo-blocked"
		return result

	case strings.Contains(htmlContent, "redirect"):
		// 检查是否重定向到特定地区
		if strings.Contains(htmlContent, "hotstar") {
			result.Status = "Success"
			result.Region = "Available"
			return result
		}
		result.Status = "Failed"
		result.Info = "Redirected"
		return result
	}

	// 尝试获取地区信息
	if strings.Contains(htmlContent, "market") ||
		strings.Contains(htmlContent, "region") ||
		strings.Contains(htmlContent, "country") {
		result.Status = "Success"
		result.Region = "Available"
		return result
	}

	// 检查是否显示正常内容
	if strings.Contains(htmlContent, "hotstar") ||
		strings.Contains(htmlContent, "shows") ||
		strings.Contains(htmlContent, "movies") ||
		strings.Contains(htmlContent, "sports") {
		result.Status = "Success"
		result.Region = "Available"
		return result
	}

	result.Status = "Failed"
	result.Info = "Unknown Error"
	return result
}

func init() {
	// 注册 Hotstar 测试
	unlock.StreamTests = append(unlock.StreamTests, TestHotstar)
}