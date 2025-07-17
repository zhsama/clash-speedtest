package detectors

import (
	"io"
	"net/http"
	"strings"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestGYAO 测试 GYAO 解锁情况
func TestGYAO(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "GYAO",
	}

	req, err := http.NewRequest("GET", "https://gyao.yahoo.co.jp/apis/playback/graphql?appId=dj00aiZpPUNJeDh2cU1RazU3UCZzPWNvbnN1bWVyc2VjcmV0Jng9NTk-&query=%20query%20Playback(%24videoId%3A%20ID!%2C%20%24logicaAgent%3A%20LogicaAgent!%2C%20%24clientSpaceId%3A%20String!%2C%20%24os%3A%20Os!%2C%20%24device%3A%20Device!)%20%7B%20content(%20parameter%3A%20%7B%20contentId%3A%20%24videoId%20logicaAgent%3A%20%24logicaAgent%20clientSpaceId%3A%20%24clientSpaceId%20os%3A%20%24os%20device%3A%20%24device%20view%3A%20WEB%20%7D%20)%20%7B%20tracking%20%7B%20streamLog%20vrLog%20stLog%20%7D%20inStreamAd%20%7B%20forcePlayback%20source%20%7B%20__typename%20...%20on%20YjAds%20%7B%20ads%20%7B%20location%20time%20adRequests%20%7B%20__typename%20...%20on%20YjAdOnePfWeb%20%7B%20adDs%20placementCategoryId%20%7D%20...%20on%20YjAdOnePfProgrammaticWeb%20%7B%20adDs%20%7D%20...%20on%20YjAdAmobee%20%7B%20url%20%7D%20...%20on%20YjAdGam%20%7B%20url%20%7D%20%7D%20%7D%20%7D%20...%20on%20Vmap%20%7B%20url%20%7D%20...%20on%20CatchupVmap%20%7B%20url%20siteId%20%7D%20%7D%20%7D%20video%20%7B%20id%20title%20delivery%20%7B%20id%20drm%20%7D%20duration%20images%20%7B%20url%20width%20height%20%7D%20cpId%20playableAge%20maxPixel%20embeddingPermission%20playableAgents%20gyaoUrl%20%7D%20%7D%20%7D%20&variables=%7B%22videoId%22%3A%225fb4e68c-aef7-4f63-88e9-8cfeb35e9065%22%2C%22logicaAgent%22%3A%22PC_WEB%22%2C%22clientSpaceId%22%3A%221183050133%22%2C%22os%22%3A%22UNKNOWN%22%2C%22device%22%3A%22PC%22%7D", nil)
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

	if strings.Contains(string(body), "not in japan") {
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	}

	result.Status = "Success"
	result.Region = "JPN"
	return result
}

func init() {
	// 注册 GYAO 测试
	unlock.StreamTests = append(unlock.StreamTests, TestGYAO)
}
