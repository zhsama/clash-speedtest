package detectors

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestHamiVideo 测试 HamiVideo 解锁情况
func TestHamiVideo(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "HamiVideo",
	}

	req, err := http.NewRequest("GET", "https://hamivideo.hinet.net/api/play.do?id=OTT_VOD_0000249064&freeProduct=1", nil)
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

	var response struct {
		Code string `json:"code"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Response Error"
		return result
	}

	if response.Code == "06001-107" {
		result.Status = "Success"
		result.Region = "TWN"
		return result
	}

	result.Status = "Failed"
	result.Info = "Region Restricted"
	return result
}

func init() {
	// 注册 HamiVideo 测试
	unlock.StreamTests = append(unlock.StreamTests, TestHamiVideo)
}
