package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/zhsama/clash-speedtest/unlock"
)

// TestNetflix 测试 Netflix 解锁情况
func TestNetflix(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Netflix",
	}

	req, err := http.NewRequest("GET", "https://www.netflix.com/title/81280792", nil)
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
	case strings.Contains(htmlContent, "Netflix hasn't come to this country yet"):
		result.Status = "Failed"
		result.Info = "Not Available"
		return result

	case strings.Contains(htmlContent, "Sorry, we are unable to process your request"):
		result.Status = "Failed"
		result.Info = "Error"
		return result

	case strings.Contains(htmlContent, "page-404"):
		fallthrough
	case strings.Contains(htmlContent, "NSEZ-403"):
		result.Status = "Failed"
		result.Info = "Blocked"
		return result
	}

	// 尝试获取地区信息
	if strings.Contains(htmlContent, `"requestCountry":`) {
		start := strings.Index(htmlContent, `"requestCountry":"`) + 17
		end := strings.Index(htmlContent[start:], `"`) + start
		if end > start {
			result.Status = "Success"
			result.Region = htmlContent[start:end]
			return result
		}
	}

	// 检查是否显示播放界面
	if strings.Contains(htmlContent, "watch-video") ||
		strings.Contains(htmlContent, "video-title") ||
		strings.Contains(htmlContent, "player-title-link") {
		result.Status = "Success"
		result.Region = "Available"
		return result
	}

	result.Status = "Failed"
	result.Info = "Unknown Error"
	return result
}

func init() {
	// 注册 Netflix 测试
	unlock.StreamTests = append(unlock.StreamTests, TestNetflix)
}