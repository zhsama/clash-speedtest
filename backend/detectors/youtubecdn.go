package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/faceair/clash-speedtest/unlock"
)

// IATA机场代码和城市对应关系
var (
	IATACODE = map[string]string{
		"TPE": "Taipei",
		"HKG": "Hong Kong",
		"NRT": "Tokyo",
		"KIX": "Osaka",
		"ICN": "Seoul",
		"BKK": "Bangkok",
		"SIN": "Singapore",
		"KUL": "Kuala Lumpur",
		"LAX": "Los Angeles",
		"SJC": "San Jose",
		"SEA": "Seattle",
		"LHR": "London",
		"FRA": "Frankfurt",
		"AMS": "Amsterdam",
		"CDG": "Paris",
	}
)

// TestYouTubeCDN 测试 YouTube CDN 位置
func TestYouTubeCDN(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "YouTube CDN",
	}

	req, err := http.NewRequest("GET", "https://redirector.googlevideo.com/report_mapping", nil)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Response Error"
		return result
	}

	content := string(body)
	if content == "" {
		result.Status = "Failed"
		result.Info = "Empty Response"
		return result
	}

	// 提取IATA代码
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		result.Status = "Failed"
		result.Info = "Parse Response Error"
		return result
	}

	// 查找第一行包含 "=>" 的内容
	var firstLine string
	for _, line := range lines {
		if strings.Contains(line, "=>") {
			firstLine = line
			break
		}
	}

	if firstLine == "" {
		result.Status = "Failed"
		result.Info = "Location Not Found"
		return result
	}

	// 提取IATA代码
	parts := strings.Fields(firstLine)
	if len(parts) < 3 {
		result.Status = "Failed"
		result.Info = "Parse IATA Code Error"
		return result
	}

	// 提取ISP和IATA代码
	serverInfo := strings.Split(parts[2], "-")
	if len(serverInfo) < 2 {
		result.Status = "Failed"
		result.Info = "Parse Server Info Error"
		return result
	}

	isp := strings.ToUpper(serverInfo[0])
	iataCode := strings.ToUpper(serverInfo[1][:3])

	// 检查是否为IDC路由器
	isIDC := strings.Contains(content, "router")

	// 查找IATA代码对应的位置
	location, exists := IATACODE[iataCode]
	if !exists {
		result.Status = "Failed"
		result.Info = "IATA: " + iataCode + " Not Found"
		return result
	}

	result.Status = "Success"
	if isIDC {
		result.Region = location
	} else {
		result.Region = location
		result.Info = isp
	}

	return result
}

func init() {
	// 注册 YouTube CDN 测试
	unlock.StreamTests = append(unlock.StreamTests, TestYouTubeCDN)
}
