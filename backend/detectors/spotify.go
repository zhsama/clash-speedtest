// detectors/spotify.go
package detectors

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/zhsama/clash-speedtest/logger"
	"github.com/zhsama/clash-speedtest/unlock"
)

// TestSpotify 测试 Spotify 解锁情况
func TestSpotify(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Spotify",
	}

	req, err := http.NewRequest("GET", "https://open.spotify.com/", nil)
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

	// 若被重定向到 /unavailable 视为未解锁
	finalURL := resp.Request.URL.String()
	if strings.Contains(finalURL, "unavailable") {
		result.Status = "Failed"
		result.Info = "Not Available"
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Response Error"
		return result
	}
	htmlContent := string(body)

	logger.Logger.Debug("Spotify Response", slog.String("content", htmlContent))

	// 从 HTML 提取 market 国家码
	market, err := extractMarket(htmlContent)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Region Blocked or Unknown"
		return result
	}

	if market != "" {
		result.Status = "Success"
		result.Region = market
	} else {
		result.Status = "Success"
		result.Region = "Available"
	}
	return result
}

// extractMarket 解析首页 HTML，返回 Spotify 判定的国家码
func extractMarket(html string) (string, error) {
	// (?s) 让 . 跨行匹配；只抓 id="appServerConfig" 这一个脚本
	re := regexp.MustCompile(`(?s)<script[^>]+id="appServerConfig"[^>]*>([^<]+)</script>`)
	m := re.FindStringSubmatch(html)
	if len(m) < 2 {
		return "", fmt.Errorf("appServerConfig not found")
	}
	raw := strings.TrimSpace(m[1])

	// 尝试标准 Base64，再尝试 URL-safe Base64
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		data, err = base64.URLEncoding.DecodeString(raw)
		if err != nil {
			return "", fmt.Errorf("base64 decode: %w", err)
		}
	}

	// 解 JSON，只关心 market 字段
	var cfg struct {
		Market string `json:"market"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("json unmarshal: %w", err)
	}
	return strings.ToUpper(cfg.Market), nil
}

func init() {
	// 注册 Spotify 测试
	unlock.StreamTests = append(unlock.StreamTests, TestSpotify)
}
