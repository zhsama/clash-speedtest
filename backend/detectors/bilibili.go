package detectors

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestBilibiliMainland 测试哔哩哔哩大陆限定解锁情况
func TestBilibiliMainland(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Bilibili China Mainland Only",
	}

	// 生成随机session
	session := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d", time.Now().UnixNano()))))

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.bilibili.com/pgc/player/web/playurl?avid=82846771&qn=0&type=&otype=json&ep_id=307247&fourk=1&fnver=0&fnval=16&session=%s&module=bangumi", session), nil)
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

	var response struct {
		Code int `json:"code"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Response Error"
		return result
	}

	switch response.Code {
	case 0:
		result.Status = "Success"
		result.Region = "CHN"
		return result
	case -10403:
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	default:
		result.Status = "Failed"
		result.Info = fmt.Sprintf("Error Code: %d", response.Code)
		return result
	}
}

// TestBilibiliHKMCTW 测试哔哩哔哩港澳台限定解锁情况
func TestBilibiliHKMCTW(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Bilibili HongKong/Macau/Taiwan",
	}

	// 生成随机session
	session := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d", time.Now().UnixNano()))))

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.bilibili.com/pgc/player/web/playurl?avid=18281381&cid=29892777&qn=0&type=&otype=json&ep_id=183799&fourk=1&fnver=0&fnval=16&session=%s&module=bangumi", session), nil)
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

	var response struct {
		Code int `json:"code"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Response Error"
		return result
	}

	switch response.Code {
	case 0:
		result.Status = "Success"
		result.Region = "HKG/MAC/TWN"
		return result
	case -10403:
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	default:
		result.Status = "Failed"
		result.Info = fmt.Sprintf("Error Code: %d", response.Code)
		return result
	}
}

// TestBilibiliTW 测试哔哩哔哩台湾限定解锁情况
func TestBilibiliTW(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "Bilibili Taiwan Only",
	}

	// 生成随机session
	session := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d", time.Now().UnixNano()))))

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.bilibili.com/pgc/player/web/playurl?avid=50762638&cid=100279344&qn=0&type=&otype=json&ep_id=268176&fourk=1&fnver=0&fnval=16&session=%s&module=bangumi", session), nil)
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

	var response struct {
		Code int `json:"code"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Response Error"
		return result
	}

	switch response.Code {
	case 0:
		result.Status = "Success"
		result.Region = "TWN"
		return result
	case -10403:
		result.Status = "Failed"
		result.Info = "Region Restricted"
		return result
	default:
		result.Status = "Failed"
		result.Info = fmt.Sprintf("Error Code: %d", response.Code)
		return result
	}
}

func init() {
	// 注册 Bilibili 测试
	unlock.StreamTests = append(unlock.StreamTests, TestBilibiliMainland, TestBilibiliHKMCTW, TestBilibiliTW)
}