package detectors

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/zhsama/clash-speedtest/unlock"
)

// TestCatchplay 测试 Catchplay+ 解锁情况
func TestCatchplay(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Catchplay+",
	}

	req, err := http.NewRequest("GET", "https://sunapi.catchplay.com/geo", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("Authorization", "Basic NTQ3MzM0NDgtYTU3Yi00MjU2LWE4MTEtMzdlYzNkNjJmM2E0Ok90QzR3elJRR2hLQ01sSDc2VEoy")

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

	var response struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Response Error"
		return result
	}

	if response.Code == "100016" {
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	}

	result.Status = "Success"
	result.Region = response.Code
	return result
}

func init() {
	// 注册 Catchplay+ 测试
	unlock.StreamTests = append(unlock.StreamTests, TestCatchplay)
}
