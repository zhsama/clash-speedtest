package detectors

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/zhsama/clash-speedtest/unlock"
)

// TestOpenAI 测试 ChatGPT/OpenAI 解锁情况
func TestOpenAI(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "ChatGPT",
	}

	req, err := http.NewRequest("GET", "https://api.openai.com/compliance/cookie_requirements", nil)
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

	bodyStr := string(body)

	// 检查是否被阻止
	if strings.Contains(bodyStr, "unsupported_country") ||
		strings.Contains(bodyStr, "vpn") {
		result.Status = "Failed"
		result.Info = "Blocked"
		return result
	}

	// 尝试解析JSON响应获取地区信息
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err == nil {
		if country, ok := apiResponse["country"].(string); ok {
			result.Status = "Success"
			result.Region = strings.ToUpper(country)
			return result
		}
	}

	// 如果响应成功但没有地区信息
	if resp.StatusCode == 200 {
		result.Status = "Success"
		result.Region = "Available"
		return result
	}

	result.Status = "Failed"
	result.Info = "Unknown Error"
	return result
}

func init() {
	// 注册 ChatGPT 测试
	unlock.StreamTests = append(unlock.StreamTests, TestOpenAI)
}