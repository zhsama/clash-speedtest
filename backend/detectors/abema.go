package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/zhsama/clash-speedtest/unlock"
)

// TestAbema 测试 Abema TV 解锁情况
func TestAbema(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Abema",
	}

	req, err := http.NewRequest("GET", "https://api.abema.io/v1/ip/check?device=android", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("Accept", "application/json")

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

	if strings.Contains(string(body), `"country":"JP"`) {
		result.Status = "Success"
		result.Region = "JP"
		return result
	}

	result.Status = "Failed"
	result.Info = "Not Available"
	return result
}

func init() {
	// 注册 Abema TV 测试
	unlock.StreamTests = append(unlock.StreamTests, TestAbema)
}