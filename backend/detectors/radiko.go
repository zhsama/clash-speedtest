package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/zhsama/clash-speedtest/unlock"
)

// TestRadiko 测试 Radiko 解锁情况
func TestRadiko(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Radiko",
	}

	req, err := http.NewRequest("GET", "https://radiko.jp/area?_=1625406539531", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

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
	if strings.Contains(response, `classs="OUT"`) || strings.Contains(response, "OUT") {
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	}

	if strings.Contains(response, "JAPAN") {
		result.Status = "Success"
		result.Region = "JPN"
		return result
	}

	result.Status = "Failed"
	result.Info = "Unknown Response"
	return result
}

func init() {
	// 注册 Radiko 测试
	unlock.StreamTests = append(unlock.StreamTests, TestRadiko)
}
