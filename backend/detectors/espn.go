package detectors

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/faceair/clash-speedtest/unlock"
)

// TestESPN 测试 ESPN+ 解锁情况
func TestESPN(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "ESPN+",
	}

	// 第一步：获取 token
	tokenData := strings.NewReader(`grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Atoken-exchange&latitude=0&longitude=0&platform=browser&subject_token=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJzdWIiOiJjYWJmMDNkMi0xMmEyLTQ0YjYtODJjOS1lOWJkZGNhMzYwNjkiLCJhdWQiOiJ1cm46YmFtdGVjaDpzZXJ2aWNlOnRva2VuIiwibmJmIjoxNjMyMjMwMTY4LCJpc3MiOiJ1cm46YmFtdGVjaDpzZXJ2aWNlOmRldmljZSIsImV4cCI6MjQ5NjIzMDE2OCwiaWF0IjoxNjMyMjMwMTY4LCJqdGkiOiJhYTI0ZWI5Yi1kNWM4LTQ5ODctYWI4ZS1jMDdhMWVhMDgxNzAifQ.8RQ-44KqmctKgdXdQ7E1DmmWYq0gIZsQw3vRL8RvCtrM_hSEHa-CkTGIFpSLpJw8sMlmTUp5ZGwvhghX-4HXfg&subject_token_type=urn%3Abamtech%3Aparams%3Aoauth%3Atoken-type%3Adevice`)

	tokenReq, err := http.NewRequest("POST", "https://espn.api.edge.bamgrid.com/token", tokenData)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Token Request Error"
		return result
	}

	tokenReq.Header.Set("User-Agent", unlock.UA_Browser)
	tokenReq.Header.Set("Authorization", "Bearer ZXNwbiZicm93c2VyJjEuMC4w.ptUt7QxsteaRruuPmGZFaJByOoqKvDP2a5YkInHrc7c")
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Token Network Error"
		return result
	}
	defer tokenResp.Body.Close()

	// 第二步：注册设备
	deviceData := strings.NewReader(`{"query":"mutation registerDevice($input: RegisterDeviceInput!) {\n            registerDevice(registerDevice: $input) {\n                grant {\n                    grantType\n                    assertion\n                }\n            }\n        }","variables":{"input":{"deviceFamily":"browser","applicationRuntime":"chrome","deviceProfile":"windows","deviceLanguage":"zh-CN","attributes":{"osDeviceIds":[],"manufacturer":"microsoft","model":null,"operatingSystem":"windows","operatingSystemVersion":"10.0","browserName":"chrome","browserVersion":"96.0.4664"}}}}`)

	deviceReq, err := http.NewRequest("POST", "https://espn.api.edge.bamgrid.com/graph/v1/device/graphql", deviceData)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Create Device Request Error"
		return result
	}

	deviceReq.Header.Set("User-Agent", unlock.UA_Browser)
	deviceReq.Header.Set("Authorization", "ZXNwbiZicm93c2VyJjEuMC4w.ptUt7QxsteaRruuPmGZFaJByOoqKvDP2a5YkInHrc7c")
	deviceReq.Header.Set("Content-Type", "application/json")

	deviceResp, err := client.Do(deviceReq)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Device Network Error"
		return result
	}
	defer deviceResp.Body.Close()

	body, err := io.ReadAll(deviceResp.Body)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Read Response Error"
		return result
	}

	var response struct {
		Extensions struct {
			Sdk struct {
				Session struct {
					Location struct {
						CountryCode string `json:"countryCode"`
					}
					InSupportedLocation bool `json:"inSupportedLocation"`
				}
			}
		}
	}

	if err := json.Unmarshal(body, &response); err != nil {
		result.Status = "Failed"
		result.Info = "Parse Response Error"
		return result
	}

	if response.Extensions.Sdk.Session.Location.CountryCode == "US" && response.Extensions.Sdk.Session.InSupportedLocation {
		result.Status = "Success"
		result.Region = "US"
		return result
	}

	result.Status = "Failed"
	result.Info = "Region Restricted"
	return result
}

func init() {
	// 注册 ESPN+ 测试
	unlock.StreamTests = append(unlock.StreamTests, TestESPN)
}
