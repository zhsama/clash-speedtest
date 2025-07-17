package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestPrimeVideo 测试 Prime Video 解锁情况
func TestPrimeVideo(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Prime Video",
	}

	req, err := http.NewRequest("GET", "https://www.primevideo.com", nil)
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

	// 检查是否有地区限制信息
	if strings.Contains(htmlContent, "not available in your location") ||
		strings.Contains(htmlContent, "isn't available in your country") {
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	}

	// 检查是否显示订阅界面
	if strings.Contains(htmlContent, "prime-header") ||
		strings.Contains(htmlContent, "dv-signup") ||
		strings.Contains(htmlContent, "primevideo-button") {
		result.Status = "Success"
		// 尝试获取地区信息
		if strings.Contains(htmlContent, `"currentTerritory":"`) {
			start := strings.Index(htmlContent, `"currentTerritory":"`) + 19
			end := strings.Index(htmlContent[start:], `"`) + start
			if end > start {
				result.Region = htmlContent[start:end]
				return result
			}
		}
		result.Region = "Available"
		return result
	}

	result.Status = "Failed"
	result.Info = "Not Available"
	return result
}

func init() {
	// 注册 Prime Video 测试
	unlock.StreamTests = append(unlock.StreamTests, TestPrimeVideo)
}
