package detectors

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestTelasa 测试 Telasa 解锁情况
func TestTelasa(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Telasa",
	}

	req, err := http.NewRequest("GET", "https://api-videopass-anon.kddi-video.com/v1/playback/system_status", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("X-Device-ID", "d36f8e6b-e344-4f5e-9a55-90aeb3403799")

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
		Status struct {
			Type    string `json:"type"`
			Subtype string `json:"subtype"`
		} `json:"status"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Response Error"
		return result
	}

	if response.Status.Subtype == "IPLocationNotAllowed" {
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	}

	if response.Status.Type != "" {
		result.Status = "Success"
		result.Region = "JPN"
		return result
	}

	result.Status = "Failed"
	result.Info = "Unknown Response"
	return result
}

func init() {
	// 注册 Telasa 测试
	unlock.StreamTests = append(unlock.StreamTests, TestTelasa)
}
