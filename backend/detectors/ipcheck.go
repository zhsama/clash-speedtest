package detectors

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/zhsama/clash-speedtest/unlock"
	"github.com/andybalholm/brotli"
)

// 随机UA列表
var (
	pcUserAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2.1 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/120.0.2210.133 Safari/537.36",
	}

	mobileUserAgents = []string{
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 14; SM-S918B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.144 Mobile Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/120.0.6099.119 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.144 Mobile Safari/537.36",
		"Mozilla/5.0 (iPad; CPU OS 17_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 13; M2102J20SG) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.144 Mobile Safari/537.36",
	}

	languages = []string{
		"zh-CN,zh;q=0.9,en;q=0.8",
		"en-US,en;q=0.9",
		"zh-TW,zh;q=0.9,en;q=0.8",
		"ja-JP,ja;q=0.9,en;q=0.8",
		"ko-KR,ko;q=0.9,en;q=0.8",
		"en-GB,en;q=0.9",
	}
)

// 生成随机请求头
func generateRandomHeaders(isMobile bool) http.Header {
	headers := make(http.Header)

	// 随机选择UA
	var ua string
	if isMobile {
		ua = mobileUserAgents[rand.Intn(len(mobileUserAgents))]
	} else {
		ua = pcUserAgents[rand.Intn(len(pcUserAgents))]
	}

	// 随机选择语言
	lang := languages[rand.Intn(len(languages))]

	// 基础请求头
	headers.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	headers.Set("accept-encoding", "gzip, br")
	headers.Set("accept-language", lang)
	headers.Set("cache-control", "no-cache")
	headers.Set("pragma", "no-cache")
	headers.Set("sec-ch-ua-mobile", fmt.Sprintf("?%d", map[bool]int{true: 1, false: 0}[isMobile]))
	headers.Set("sec-ch-ua-platform", map[bool]string{true: `"Android"`, false: `"Windows"`}[isMobile])
	headers.Set("sec-fetch-dest", "document")
	headers.Set("sec-fetch-mode", "navigate")
	headers.Set("sec-fetch-site", "cross-site")
	headers.Set("sec-fetch-user", "?1")
	headers.Set("upgrade-insecure-requests", "1")
	headers.Set("user-agent", ua)

	// Cloudflare 相关请求头
	headers.Set("cf-device-type", map[bool]string{true: "mobile", false: "desktop"}[isMobile])
	headers.Set("cf-visitor", `{"scheme":"https"}`)
	headers.Set("x-forwarded-proto", "https")
	headers.Set("x-requested-with", "XMLHttpRequest")
	headers.Set("dnt", "1")

	// 随机生成 Client Hints
	headers.Set("sec-ch-ua", generateSecChUA(ua))
	headers.Set("sec-ch-ua-full-version-list", generateSecChUAFullVersionList(ua))

	return headers
}

// 生成 sec-ch-ua 头
func generateSecChUA(ua string) string {
	if strings.Contains(ua, "Chrome") {
		version := extractVersion(ua, "Chrome")
		return fmt.Sprintf(`"Google Chrome";v="%s", "Not=A?Brand";v="8", "Chromium";v="%s"`, version, version)
	} else if strings.Contains(ua, "Firefox") {
		version := extractVersion(ua, "Firefox")
		return fmt.Sprintf(`"Firefox";v="%s"`, version)
	} else if strings.Contains(ua, "Safari") && !strings.Contains(ua, "Chrome") {
		version := extractVersion(ua, "Version")
		return fmt.Sprintf(`"Safari";v="%s"`, version)
	}
	return `"Not=A?Brand";v="8"`
}

// 生成 sec-ch-ua-full-version-list 头
func generateSecChUAFullVersionList(ua string) string {
	if strings.Contains(ua, "Chrome") {
		version := extractVersion(ua, "Chrome")
		return fmt.Sprintf(`"Google Chrome";v="%s.0.0.0", "Not=A?Brand";v="8.0.0.0", "Chromium";v="%s.0.0.0"`, version, version)
	}
	return ""
}

// 从 UA 中提取版本号
func extractVersion(ua string, browser string) string {
	re := regexp.MustCompile(browser + `/(\d+)`)
	matches := re.FindStringSubmatch(ua)
	if len(matches) > 1 {
		return matches[1]
	}
	return "0"
}

// 带重试的请求
func doRequestWithRetry(client *http.Client, req *http.Request, maxRetries int, debugMode bool) (*http.Response, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// 重试前等待随机时间
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

			// 重新生成请求头
			isMobile := rand.Float32() < 0.3
			req.Header = generateRandomHeaders(isMobile)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if debugMode {
				fmt.Printf("请求失败 (尝试 %d/%d): %v\n", i+1, maxRetries, err)
			}
			continue
		}

		// 检查是否遇到 Cloudflare 验证
		if resp.StatusCode == 403 || resp.StatusCode == 503 {
			body, _ := readCompressedBody(resp)
			if strings.Contains(string(body), "cloudflare") || strings.Contains(string(body), "cf-") {
				resp.Body.Close()
				if debugMode {
					fmt.Printf("遇到 Cloudflare 验证 (尝试 %d/%d)\n", i+1, maxRetries)
				}
				continue
			}
		}

		return resp, nil
	}

	return nil, fmt.Errorf("达到最大重试次数 (%d): %v", maxRetries, lastErr)
}

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorOrange = "\033[38;5;208m"
	colorWhite  = "\033[37m"
	colorReset  = "\033[0m"
)

type IPRiskResponse struct {
	Risk interface{} `json:"risk"`
}

type GeoResponse struct {
	Country string `json:"country"`
	IP      string `json:"ip"`
}

func readCompressedBody(resp *http.Response) ([]byte, error) {
	var reader io.ReadCloser
	var err error

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
	case "br":
		reader = io.NopCloser(brotli.NewReader(resp.Body))
	default:
		reader = resp.Body
	}

	return io.ReadAll(reader)
}

// GetLocation 获取地理位置信息
func GetLocation(client *http.Client, debugMode bool) (string, error) {
	req, err := http.NewRequest("GET", "https://64.ipcheck.ing/geo", nil)
	if err != nil {
		if debugMode {
			fmt.Printf("创建请求失败: %v\n", err)
		}
		return "N/A", err
	}

	// 使用随机请求头
	isMobile := rand.Float32() < 0.3
	req.Header = generateRandomHeaders(isMobile)

	if debugMode {
		fmt.Println("发送请求头:")
		for k, v := range req.Header {
			fmt.Printf("%s: %v\n", k, v)
		}
	}

	resp, err := doRequestWithRetry(client, req, 3, debugMode)
	if err != nil {
		if debugMode {
			fmt.Printf("请求失败: %v\n", err)
		}
		return "N/A", err
	}
	defer resp.Body.Close()

	body, err := readCompressedBody(resp)
	if err != nil {
		if debugMode {
			fmt.Printf("读取响应失败: %v\n", err)
		}
		return "N/A", err
	}

	if debugMode {
		fmt.Printf("请求 URL: %s\n", req.URL)
		fmt.Printf("响应状态码: %d\n", resp.StatusCode)
		fmt.Printf("响应头: %v\n", resp.Header)
		fmt.Printf("地理位置 API 响应: %s\n", string(body))
	}

	var geoResp GeoResponse
	if err := json.Unmarshal(body, &geoResp); err != nil {
		if debugMode {
			fmt.Printf("JSON 解析错误: %v\n", err)
		}
		return "N/A", err
	}

	if geoResp.Country != "" {
		if debugMode {
			fmt.Printf("成功获取到国家信息: %s\n", geoResp.Country)
		}
		return geoResp.Country, nil
	}
	if debugMode {
		fmt.Println("响应中没有国家信息")
	}
	return "N/A", fmt.Errorf("no country information in response")
}

// GetLocationWithRisk 获取地理位置和IP纯净度信息
func GetLocationWithRisk(client *http.Client, debugMode bool, enableRisk bool) (string, error) {
	if debugMode {
		fmt.Println("开始获取地理位置信息...")
	}

	// 设置总体超时
	client.Timeout = 5 * time.Second

	// 先获取地理位置
	city, err := GetLocation(client, debugMode)
	if err != nil || city == "N/A" {
		if debugMode {
			fmt.Printf("获取地理位置失败: %v\n", err)
		}
		return "N/A", err
	}

	// 如果不启用风险检测，直接返回地理位置
	if !enableRisk {
		return city, nil
	}

	// 创建一个新的客户端用于风险值请求
	riskClient := &http.Client{
		Timeout:   5 * time.Second,
		Transport: client.Transport,
	}

	// 获取IP
	req, err := http.NewRequest("GET", "https://64.ipcheck.ing/geo", nil)
	if err != nil {
		return city, nil // 出错时返回地理位置
	}

	// 使用随机请求头
	isMobile := rand.Float32() < 0.3
	req.Header = generateRandomHeaders(isMobile)

	// 使用重试机制获取IP
	resp, err := doRequestWithRetry(riskClient, req, 3, debugMode)
	if err != nil {
		return city, nil
	}
	defer resp.Body.Close()

	body, err := readCompressedBody(resp)
	if err != nil {
		return city, nil
	}

	var geoResp struct {
		IP string `json:"ip"`
	}
	if err := json.Unmarshal(body, &geoResp); err != nil {
		return city, nil
	}

	// 获取 IP 风险值
	riskReq, err := http.NewRequest("GET", fmt.Sprintf("https://ipcheck.ing/api/ipchecking?ip=%s&lang=zh-CN", geoResp.IP), nil)
	if err != nil {
		return city, nil
	}

	// 设置相同请求头
	for key, values := range req.Header {
		riskReq.Header[key] = values
	}
	// 添加必要的额外头
	riskReq.Header.Set("Referer", "https://ipcheck.ing/")
	riskReq.Header.Set("Origin", "https://ipcheck.ing/")

	// 使用重试机制获取风险值
	riskResp, err := doRequestWithRetry(riskClient, riskReq, 3, debugMode)
	if err != nil {
		return city, nil
	}
	defer riskResp.Body.Close()

	riskBody, err := readCompressedBody(riskResp)
	if err != nil {
		return city, nil
	}

	if debugMode {
		fmt.Printf("风险值响应: %s\n", string(riskBody))
	}

	var riskData struct {
		ProxyDetect struct {
			Proxy    string      `json:"proxy"`
			Risk     interface{} `json:"risk"`
			Type     string      `json:"type"`
			Operator string      `json:"operator"`
			Protocol string      `json:"protocol"`
		} `json:"proxyDetect"`
	}
	if err := json.Unmarshal(riskBody, &riskData); err != nil {
		if debugMode {
			fmt.Printf("解析风险值响应失败: %v\n", err)
		}
		return city, nil
	}

	if debugMode {
		fmt.Printf("解析后的风险值数据: %+v\n", riskData.ProxyDetect)
	}

	// 根据风险值返回不同结果
	var riskLevel string
	if debugMode {
		fmt.Printf("风险值类型: %T, 值: %v\n", riskData.ProxyDetect.Risk, riskData.ProxyDetect.Risk)
	}

	switch v := riskData.ProxyDetect.Risk.(type) {
	case float64:
		if v == 0 {
			riskLevel = fmt.Sprintf("[%.0f 纯净]", v)
		} else if v < 66 {
			riskLevel = fmt.Sprintf("[%.0f 一般]", v)
		} else {
			riskLevel = fmt.Sprintf("[%.0f 较差]", v)
		}
	case json.Number:
		f, _ := v.Float64()
		if f == 0 {
			riskLevel = fmt.Sprintf("[%.0f 纯净]", f)
		} else if f < 66 {
			riskLevel = fmt.Sprintf("[%.0f 一般]", f)
		} else {
			riskLevel = fmt.Sprintf("[%.0f 较差]", f)
		}
	case nil:
		riskLevel = fmt.Sprintf("[100 非常差]")
	default:
		if str, ok := v.(string); ok && str == "" {
			riskLevel = fmt.Sprintf("[100 非常差]")
		} else {
			riskLevel = fmt.Sprintf("[%v 未知]", v)
		}
	}

	// 如果有风险值，返回带风险值的结果
	if riskData.ProxyDetect.Risk != nil {
		return fmt.Sprintf("%s %s", city, riskLevel), nil
	}

	// 否则只返回地理位置
	return city, nil
}

// TestIPCheck 测试 IP 地理位置检查
func TestIPCheck(client *http.Client) *unlock.StreamResult {
	result := &unlock.StreamResult{
		Platform: "IP Location Check",
	}

	location, err := GetLocation(client, false)
	if err != nil {
		result.Status = "Failed"
		result.Info = "Unable to get location"
		return result
	}

	result.Status = "Success"
	result.Region = location
	return result
}

func init() {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())
	
	// 注册 IP 检查测试
	unlock.StreamTests = append(unlock.StreamTests, TestIPCheck)
}