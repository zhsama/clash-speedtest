package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestDMM 测试 DMM 解锁情况
func TestDMM(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "DMM",
	}

	req, err := http.NewRequest("GET", "https://api-public.dmm.com/v1/region", nil)
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

	response := string(body)

	switch {
	case strings.Contains(response, `"country":"JPN"`):
		result.Status = "Success"
		result.Region = "JP"
		return result
	case strings.Contains(response, "IP_COUNTRY"):
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	}

	result.Status = "Failed"
	result.Info = "Not Available"
	return result
}

func init() {
	// 注册 DMM 测试
	unlock.StreamTests = append(unlock.StreamTests, TestDMM)
}
