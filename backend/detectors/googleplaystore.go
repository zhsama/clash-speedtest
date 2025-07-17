package detectors

import (
	"io"
	"net/http"
	"regexp"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestGooglePlayStore 测试 Google Play Store 区域限制
func TestGooglePlayStore(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "GooglePlayStore",
	}

	req, err := http.NewRequest("GET", "https://play.google.com/", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	// 设置请求头
	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US;q=0.9")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="131", "Not_A Brand";v="24", "Google Chrome";v="131"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "Windows")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

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

	// 使用正则表达式匹配区域信息
	re := regexp.MustCompile(`<div class="yVZQTb">([^<(]+)`)
	matches := re.FindSubmatch(body)
	if len(matches) > 1 {
		result.Status = "Success"
		result.Region = string(matches[1])
		return result
	}

	result.Status = "Failed"
	result.Info = "Region Not Found"
	return result
}

func init() {
	// 注册 Google Play Store 测试
	unlock.StreamTests = append(unlock.StreamTests, TestGooglePlayStore)
}
