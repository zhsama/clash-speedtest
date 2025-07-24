package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/zhsama/clash-speedtest/unlock"
)

// TestBahamut 测试 Bahamut 动画疯解锁情况
func TestBahamut(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Bahamut",
	}

	req, err := http.NewRequest("GET", "https://ani.gamer.com.tw/ajax/token.php?adID=89422&sn=14667", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://ani.gamer.com.tw")
	req.Header.Set("Referer", "https://ani.gamer.com.tw/animeVideo.php")

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
	case strings.Contains(response, "error code: 1011"):
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	case strings.Contains(response, "error code: 1015"):
		result.Status = "Failed"
		result.Info = "IP Blocked"
		return result
	case strings.Contains(response, "error code:"):
		result.Status = "Failed"
		result.Info = "Error"
		return result
	case strings.Contains(response, "animeSn"):
		result.Status = "Success"
		result.Region = "TW"
		return result
	}

	result.Status = "Failed"
	result.Info = "Unknown Error"
	return result
}

func init() {
	// 注册 Bahamut 动画疯测试
	unlock.StreamTests = append(unlock.StreamTests, TestBahamut)
}
