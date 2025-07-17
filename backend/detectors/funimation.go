package detectors

import (
	"net/http"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestFunimation 测试 Funimation 解锁情况
func TestFunimation(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Funimation",
	}

	req, err := http.NewRequest("GET", "https://www.funimation.com", nil)
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

	if resp.StatusCode == 403 {
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	}

	// 检查 region cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "region" {
			result.Status = "Success"
			result.Region = cookie.Value
			return result
		}
	}

	result.Status = "Failed"
	result.Info = "Region Not Found"
	return result
}

func init() {
	// 注册 Funimation 测试
	unlock.StreamTests = append(unlock.StreamTests, TestFunimation)
}
