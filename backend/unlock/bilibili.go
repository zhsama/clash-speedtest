package unlock

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/metacubex/mihomo/constant"
)

// BilibiliDetector Bilibili检测器
type BilibiliDetector struct {
	*BaseDetector
}

// NewBilibiliDetector 创建Bilibili检测器
func NewBilibiliDetector() *BilibiliDetector {
	return &BilibiliDetector{
		BaseDetector: NewBaseDetector("Bilibili", 2), // 中优先级
	}
}

// Detect 检测Bilibili解锁状态
func (d *BilibiliDetector) Detect(proxy constant.Proxy, timeout time.Duration) *UnlockResult {
	d.logDetectionStart(proxy)
	
	client := createHTTPClient(proxy, timeout)
	
	// 方法1: 检测台湾专属内容
	result1 := d.checkTaiwanContent(client)
	if result1.Status == StatusUnlocked && result1.Region == "TW" {
		d.logDetectionResult(proxy, result1)
		return result1
	}
	
	// 方法2: 检测港澳台内容
	result2 := d.checkHKMOTWContent(client)
	if result2.Status == StatusUnlocked {
		d.logDetectionResult(proxy, result2)
		return result2
	}
	
	// 方法3: 检测大陆内容访问
	result3 := d.checkMainlandContent(client)
	
	d.logDetectionResult(proxy, result3)
	return result3
}

// checkTaiwanContent 检测台湾专属内容
func (d *BilibiliDetector) checkTaiwanContent(client *http.Client) *UnlockResult {
	// 访问台湾专属动画
	resp, err := makeRequest(client, "GET", "https://www.bilibili.com/bangumi/play/ss21542", nil)
	if err != nil {
		return d.createErrorResult("Failed to connect to Bilibili", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return d.createErrorResult("Failed to read Bilibili response", err)
	}
	
	bodyStr := string(body)
	
	if strings.Contains(bodyStr, "地区限制") || 
	   strings.Contains(bodyStr, "区域限制") ||
	   strings.Contains(bodyStr, "版权方要求") {
		return d.createResult(StatusLocked, "", "Taiwan exclusive content blocked")
	}
	
	if resp.StatusCode == 200 && 
	   !strings.Contains(bodyStr, "error") &&
	   strings.Contains(bodyStr, "bangumi") {
		return d.createResult(StatusUnlocked, "TW", "Taiwan exclusive content accessible")
	}
	
	return d.createResult(StatusFailed, "", "Unable to determine Taiwan content status")
}

// checkHKMOTWContent 检测港澳台内容
func (d *BilibiliDetector) checkHKMOTWContent(client *http.Client) *UnlockResult {
	// 访问港澳台内容
	resp, err := makeRequest(client, "GET", "https://www.bilibili.com/bangumi/play/ss28341", nil)
	if err != nil {
		return d.createErrorResult("Failed to connect to Bilibili", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return d.createErrorResult("Failed to read Bilibili response", err)
	}
	
	bodyStr := string(body)
	
	if strings.Contains(bodyStr, "地区限制") || 
	   strings.Contains(bodyStr, "区域限制") ||
	   strings.Contains(bodyStr, "版权方要求") {
		return d.createResult(StatusLocked, "", "HKMOTW content blocked")
	}
	
	if resp.StatusCode == 200 && 
	   !strings.Contains(bodyStr, "error") &&
	   strings.Contains(bodyStr, "bangumi") {
		return d.createResult(StatusUnlocked, "HKMOTW", "HKMOTW content accessible")
	}
	
	return d.createResult(StatusFailed, "", "Unable to determine HKMOTW content status")
}

// checkMainlandContent 检测大陆内容访问
func (d *BilibiliDetector) checkMainlandContent(client *http.Client) *UnlockResult {
	// 访问主站首页
	resp, err := makeRequest(client, "GET", "https://www.bilibili.com", nil)
	if err != nil {
		return d.createErrorResult("Failed to connect to Bilibili", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		return d.createResult(StatusUnlocked, "CN", "Bilibili mainland accessible")
	}
	
	return d.createResult(StatusLocked, "", "Bilibili not accessible")
}