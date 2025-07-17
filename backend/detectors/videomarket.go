package detectors

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestVideoMarket 测试 VideoMarket 解锁情况
func TestVideoMarket(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "VideoMarket",
	}

	// 第一步：获取 access token
	tokenData := strings.NewReader(`grant_type=client_credentials&client_id=1eolxdrti3t58m2f2k8yi0kli105743b6f8c8295&client_secret=lco0nndn3l9tcbjdfdwlswmee105743b739cfb5a`)
	tokenReq, err := http.NewRequest("POST", "https://api-p.videomarket.jp/v2/authorize/access_token", tokenData)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Token Request Error"
		return result
	}

	tokenReq.Header.Set("User-Agent", unlock.UA_Browser)
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Token Network Error"
		return result
	}
	defer tokenResp.Body.Close()

	tokenBody, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Token Response Error"
		return result
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.Unmarshal(tokenBody, &tokenResponse); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Token Response Error"
		return result
	}

	if tokenResponse.AccessToken == "" {
		result.Status = "Failed"
		result.Info = "No Access Token"
		return result
	}

	// 第二步：获取 play key
	playData := strings.NewReader(`fullStoryId=118008001&playChromeCastFlag=false&loginFlag=0`)
	playReq, err := http.NewRequest("POST", "https://api-p.videomarket.jp/v2/api/play/keyissue", playData)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Play Request Error"
		return result
	}

	playReq.Header.Set("User-Agent", unlock.UA_Browser)
	playReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	playReq.Header.Set("X-Authorization", tokenResponse.AccessToken)

	playResp, err := client.Do(playReq)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Play Network Error"
		return result
	}
	defer playResp.Body.Close()

	playBody, err := io.ReadAll(playResp.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Play Response Error"
		return result
	}

	var playResponse struct {
		PlayKey string `json:"PlayKey"`
	}

	if err := json.Unmarshal(playBody, &playResponse); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Play Response Error"
		return result
	}

	// 第三步：验证 play key
	authReq, err := http.NewRequest("GET", "https://api-p.videomarket.jp/v2/api/play/keyauth?playKey="+playResponse.PlayKey+"&deviceType=3&bitRate=0&loginFlag=0&connType=", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Auth Request Error"
		return result
	}

	authReq.Header.Set("User-Agent", unlock.UA_Browser)
	authReq.Header.Set("X-Authorization", tokenResponse.AccessToken)

	authResp, err := client.Do(authReq)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Auth Network Error"
		return result
	}
	defer authResp.Body.Close()

	switch authResp.StatusCode {
	case 200, 408:
		result.Status = "Success"
		result.Region = "JPN"
		return result
	case 403:
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	default:
		result.Status = "Failed"
		result.Info = "Unknown Response"
		return result
	}
}

func init() {
	// 注册 VideoMarket 测试
	unlock.StreamTests = append(unlock.StreamTests, TestVideoMarket)
}
