package detectors

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestEncoreTVB 测试 encoreTVB 解锁情况
func TestEncoreTVB(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "encoreTVB",
	}

	req, err := http.NewRequest("GET", "https://edge.api.brightcove.com/playback/v1/accounts/5324042807001/videos/6005570109001", nil)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Request Error"
		return result
	}

	req.Header.Set("User-Agent", unlock.UA_Browser)
	req.Header.Set("Accept", "application/json;pk=BCpkADawqM2Gpjj8SlY2mj4FgJJMfUpxTNtHWXOItY1PvamzxGstJbsgc-zFOHkCVcKeeOhPUd9MNHEGJoVy1By1Hrlh9rOXArC5M5MTcChJGU6maC8qhQ4Y8W-QYtvi8Nq34bUb9IOvoKBLeNF4D9Avskfe9rtMoEjj6ImXu_i4oIhYS0dx7x1AgHvtAaZFFhq3LBGtR-ZcsSqxNzVg-4PRUI9zcytQkk_YJXndNSfhVdmYmnxkgx1XXisGv1FG5GOmEK4jZ_Ih0riX5icFnHrgniADr4bA2G7TYh4OeGBrYLyFN_BDOvq3nFGrXVWrTLhaYyjxOr4rZqJPKK2ybmMsq466Ke1ZtE-wNQ")
	req.Header.Set("Origin", "https://www.encoretvb.com")

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
		ErrorSubcode string `json:"error_subcode"`
		AccountId    string `json:"account_id"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Response Error"
		return result
	}

	if response.ErrorSubcode == "CLIENT_GEO" {
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	}

	if response.AccountId != "0" {
		result.Status = "Success"
		result.Region = "HKG"
		return result
	}

	result.Status = "Failed"
	result.Info = "Unknown Error"
	return result
}

func init() {
	// 注册 encoreTVB 测试
	unlock.StreamTests = append(unlock.StreamTests, TestEncoreTVB)
}
