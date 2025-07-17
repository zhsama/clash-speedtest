package detectors

import (
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestSteam 测试 Steam 商店货币区域
func TestSteam(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Steam",
	}

	req, err := http.NewRequest("GET", "https://store.steampowered.com/app/761830", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

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

	// 尝试多种方式匹配货币信息
	patterns := []string{
		`"priceCurrency":"([^"]+)"`,
		`data-price-final[^>]+>([A-Z]{2,3})\s`,
		`\$([A-Z]{2,3})\s+\d+\.\d+`,
		`¥\s*\d+`,    // 日元
		`₩\s*\d+`,    // 韩元
		`NT\$\s*\d+`, // 新台币
		`HK\$\s*\d+`, // 港币
		`S\$\s*\d+`,  // 新加坡元
		`A\$\s*\d+`,  // 澳元
		`₹\s*\d+`,    // 印度卢比
		`€\s*\d+`,    // 欧元
		`£\s*\d+`,    // 英镑
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(htmlContent); len(matches) > 0 {
			result.Status = "Success"
			switch {
			case strings.Contains(matches[0], "¥"):
				result.Region = "JPY"
			case strings.Contains(matches[0], "₩"):
				result.Region = "KRW"
			case strings.Contains(matches[0], "NT$"):
				result.Region = "TWD"
			case strings.Contains(matches[0], "HK$"):
				result.Region = "HKD"
			case strings.Contains(matches[0], "S$"):
				result.Region = "SGD"
			case strings.Contains(matches[0], "A$"):
				result.Region = "AUD"
			case strings.Contains(matches[0], "₹"):
				result.Region = "INR"
			case strings.Contains(matches[0], "€"):
				result.Region = "EUR"
			case strings.Contains(matches[0], "£"):
				result.Region = "GBP"
			default:
				if len(matches) > 1 {
					result.Region = matches[1]
				} else {
					result.Region = matches[0]
				}
			}
			return result
		}
	}

	// 检查是否被重定向到年龄验证页面
	if strings.Contains(htmlContent, "agecheck") || strings.Contains(htmlContent, "age_check") {
		result.Status = "Failed"
		result.Info = "Age Check Required"
		return result
	}

	// 检查是否在维护
	if strings.Contains(htmlContent, "maintenance") {
		result.Status = "Failed"
		result.Info = "Store Maintenance"
		return result
	}

	result.Status = "Failed"
	result.Info = "Currency Not Found"
	return result
}

func init() {
	// 注册 Steam 测试
	unlock.StreamTests = append(unlock.StreamTests, TestSteam)
}