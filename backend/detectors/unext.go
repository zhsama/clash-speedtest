package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/zhsama/clash-speedtest/unlock"
)

// TestUNext 测试 U-NEXT 解锁情况
func TestUNext(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "U-NEXT",
	}

	req, err := http.NewRequest("GET", "https://www.unext.jp/", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("Accept-Language", "ja-JP,ja;q=0.9,en;q=0.8")
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
		fallthrough
	case strings.Contains(htmlContent, "利用できません"):
		result.Status = "Failed"
		result.Info = "Not Available"
		return result

	case strings.Contains(htmlContent, "blocked"):
		fallthrough
	case strings.Contains(htmlContent, "restricted"):
		fallthrough
	case strings.Contains(htmlContent, "ブロック"):
		result.Status = "Failed"
		result.Info = "Blocked"
		return result

	case strings.Contains(htmlContent, "geo-blocked"):
		fallthrough
	case strings.Contains(htmlContent, "location"):
		fallthrough
	case strings.Contains(htmlContent, "地域制限"):
		result.Status = "Failed"
		result.Info = "Geo-blocked"
		return result
	}

	// 尝试获取地区信息
	if strings.Contains(htmlContent, "market") ||
		strings.Contains(htmlContent, "region") ||
		strings.Contains(htmlContent, "country") ||
		strings.Contains(htmlContent, "地域") {
		result.Status = "Success"
		result.Region = "JP"
		return result
	}

	// 检查是否显示正常内容
	if strings.Contains(htmlContent, "unext") ||
		strings.Contains(htmlContent, "u-next") ||
		strings.Contains(htmlContent, "映画") ||
		strings.Contains(htmlContent, "ドラマ") ||
		strings.Contains(htmlContent, "アニメ") ||
		strings.Contains(htmlContent, "動画") {
		result.Status = "Success"
		result.Region = "JP"
		return result
	}

	result.Status = "Failed"
	result.Info = "Unknown Error"
	return result
}

func init() {
	// 注册 U-NEXT 测试
	unlock.StreamTests = append(unlock.StreamTests, TestUNext)
}