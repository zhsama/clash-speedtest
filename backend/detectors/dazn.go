package detectors

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestDAZN 测试 DAZN 解锁情况
func TestDAZN(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "DAZN",
	}

	req, err := http.NewRequest("GET", "https://startup.core.indazn.com/misl/v5/Startup", nil)
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

	// 解析 JSON 响应
	var data struct {
		Region struct {
			IsAllowed   bool   `json:"isAllowed"`
			CountryCode string `json:"countryCode"`
			CountryName string `json:"country"`
		} `json:"region"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Error"
		return result
	}

	if data.Region.IsAllowed {
		result.Status = "Success"
		result.Region = data.Region.CountryCode
		return result
	}

	result.Status = "Failed"
	result.Info = "Not Available"
	return result
}

func init() {
	// 注册 DAZN 测试
	unlock.StreamTests = append(unlock.StreamTests, TestDAZN)
}
