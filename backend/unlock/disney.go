package unlock

import (
	"io"
	"strings"
	"time"

	"github.com/metacubex/mihomo/constant"
)

// DisneyDetector Disney+检测器
type DisneyDetector struct {
	*BaseDetector
}

// NewDisneyDetector 创建Disney+检测器
func NewDisneyDetector() *DisneyDetector {
	return &DisneyDetector{
		BaseDetector: NewBaseDetector("Disney+", 1), // 高优先级
	}
}

// Detect 检测Disney+解锁状态
func (d *DisneyDetector) Detect(proxy constant.Proxy, timeout time.Duration) *UnlockResult {
	d.logDetectionStart(proxy)
	
	client := createHTTPClient(proxy, timeout)
	
	// 访问Disney+主页
	resp, err := makeRequest(client, "GET", "https://www.disneyplus.com", nil)
	if err != nil {
		result := d.createErrorResult("Failed to connect to Disney+", err)
		d.logDetectionResult(proxy, result)
		return result
	}
	defer resp.Body.Close()
	
	// 检查重定向URL
	finalURL := resp.Request.URL.String()
	
	var result *UnlockResult
	if strings.Contains(finalURL, "/unavailable") || 
	   strings.Contains(finalURL, "/blocked") ||
	   strings.Contains(finalURL, "/unsupported") {
		result = d.createResult(StatusLocked, "", "Disney+ not available in this region")
	} else {
		// 读取响应内容进行进一步检查
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			result = d.createErrorResult("Failed to read Disney+ response", err)
		} else {
			bodyStr := string(body)
			if strings.Contains(bodyStr, "not available") || 
			   strings.Contains(bodyStr, "access denied") {
				result = d.createResult(StatusLocked, "", "Disney+ blocked in this region")
			} else if strings.Contains(bodyStr, "sign up") || 
					  strings.Contains(bodyStr, "subscribe") ||
					  strings.Contains(bodyStr, "bundle") {
				// 提取地区信息
				region := d.extractDisneyRegion(finalURL, bodyStr)
				result = d.createResult(StatusUnlocked, region, "Disney+ available")
			} else {
				result = d.createResult(StatusFailed, "", "Unable to determine Disney+ status")
			}
		}
	}
	
	d.logDetectionResult(proxy, result)
	return result
}

// extractDisneyRegion 从Disney+响应中提取地区信息
func (d *DisneyDetector) extractDisneyRegion(url, body string) string {
	// 从URL中提取地区
	if strings.Contains(url, "disneyplus.com") {
		return "US"
	}
	if strings.Contains(url, ".jp/") {
		return "JP"
	}
	if strings.Contains(url, ".co.uk/") {
		return "GB"
	}
	if strings.Contains(url, ".ca/") {
		return "CA"
	}
	if strings.Contains(url, ".com.au/") {
		return "AU"
	}
	
	// 从内容中提取地区信息
	if strings.Contains(body, `"market":"US"`) {
		return "US"
	}
	if strings.Contains(body, `"market":"JP"`) {
		return "JP"
	}
	if strings.Contains(body, `"market":"GB"`) {
		return "GB"
	}
	
	return ""
}