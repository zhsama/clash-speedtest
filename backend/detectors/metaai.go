package detectors

import (
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/zhsama/clash-speedtest/unlock"
)

// TestMetaAI 测试 Meta AI 区域限制
func TestMetaAI(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Meta AI",
	}

	req, err := http.NewRequest("GET", "https://www.meta.ai/", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	// 设置请求头
	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("Accept", "*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="120", "Not_A Brand";v="24", "Google Chrome";v="120"`)
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

	content := string(body)

	// 检查是否被阻止
	isBlocked := strings.Contains(content, "AbraGeoBlockedErrorRoot")
	isOK := strings.Contains(content, "AbraHomeRootConversationQuery")

	if !isBlocked && !isOK {
		result.Status = "Failed"
		result.Info = "Page Error"
		return result
	}

	if isBlocked {
		result.Status = "Failed"
		result.Info = "Not Available"
		return result
	}

	if isOK {
		// 提取区域代码
		re := regexp.MustCompile(`"code"\s*:\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(content); len(matches) > 1 {
			parts := strings.Split(matches[1], "_")
			if len(parts) > 1 {
				result.Status = "Success"
				result.Region = parts[1]
				return result
			}
		}
		result.Status = "Success"
		result.Region = "Available"
		return result
	}

	result.Status = "Failed"
	result.Info = "Unknown Error"
	return result
}

func init() {
	// 注册 Meta AI 测试
	unlock.StreamTests = append(unlock.StreamTests, TestMetaAI)
}
