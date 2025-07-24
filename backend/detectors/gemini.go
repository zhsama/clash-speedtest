package detectors

import (
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/zhsama/clash-speedtest/unlock"
)

// TestGemini 测试 Google Gemini 区域限制
func TestGemini(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Google Gemini",
	}

	req, err := http.NewRequest("GET", "https://gemini.google.com", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", unlock.UA_Browser)

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

	content := string(body)

	// 检查是否可用
	hasAccess := strings.Contains(content, "45631641,null,true")

	// 提取区域代码
	re := regexp.MustCompile(`,2,1,200,"([A-Z]{3})"`)
	matches := re.FindStringSubmatch(content)

	if hasAccess {
		result.Status = "Success"
		if len(matches) > 1 {
			result.Region = matches[1]
		} else {
			result.Region = "Available"
		}
		return result
	}

	result.Status = "Failed"
	result.Info = "Not Available"
	return result
}

func init() {
	// 注册 Google Gemini 测试
	unlock.StreamTests = append(unlock.StreamTests, TestGemini)
}
