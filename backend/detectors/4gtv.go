package detectors

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/zhsama/clash-speedtest/unlock"
)

// Test4GTV 测试 4GTV 解锁情况
func Test4GTV(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "4GTV",
	}

	// 构建请求体
	data := strings.NewReader(`value=D33jXJ0JVFkBqV%2BZSi1mhPltbejAbPYbDnyI9hmfqjKaQwRQdj7ZKZRAdb16%2FRUrE8vGXLFfNKBLKJv%2BfDSiD%2BZJlUa5Msps2P4IWuTrUP1%2BCnS255YfRadf%2BKLUhIPj`)

	req, err := http.NewRequest("POST", "https://api2.4gtv.tv/Vod/GetVodUrl3", data)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		Success bool `json:"success"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Response Error"
		return result
	}

	if response.Success {
		result.Status = "Success"
		result.Region = "TWN"
		return result
	}

	result.Status = "Failed"
	result.Info = "Region Restricted"
	return result
}

func init() {
	// 注册 4GTV 测试
	unlock.StreamTests = append(unlock.StreamTests, Test4GTV)
}