package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/zhsama/clash-speedtest/unlock"
)

// TestHBOGoAsia 测试 HBO Go Asia 解锁情况
func TestHBOGoAsia(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "HBO Go Asia",
	}

	req, err := http.NewRequest("GET", "https://api2.hbogoasia.com/v1/geog?lang=undefined&version=0&bundleId=www.hbogoasia.com", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("Origin", "https://www.hbogoasia.com")

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

	htmlContent := string(body)

	// 检查地区代码
	for _, region := range []string{"PH", "HK", "SG", "TW", "TH", "ID", "MY"} {
		if strings.Contains(htmlContent, `"country":"`+region+`"`) {
			result.Status = "Success"
			result.Region = region
			return result
		}
	}

	if strings.Contains(htmlContent, "UnauthorizedLocation") {
		result.Status = "Failed"
		result.Info = "Region Not Available"
	} else {
		result.Status = "Failed"
		result.Info = "Unknown Error"
	}

	return result
}

func init() {
	// 注册 HBO Go Asia 测试
	unlock.StreamTests = append(unlock.StreamTests, TestHBOGoAsia)
}
