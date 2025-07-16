package detectors

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/faceair/clash-speedtest/unlock"
	"github.com/metacubex/mihomo/constant"
)

// BilibiliDetector Bilibili检测器
type BilibiliDetector struct {
	*unlock.BaseDetector
}

// NewBilibiliDetector 创建Bilibili检测器
func NewBilibiliDetector() *BilibiliDetector {
	return &BilibiliDetector{
		BaseDetector: unlock.NewBaseDetector("Bilibili", 2), // 中优先级
	}
}

// Detect 检测Bilibili解锁状态
func (d *BilibiliDetector) Detect(ctx context.Context, proxy constant.Proxy) *unlock.UnlockResult {
	d.LogDetectionStart(proxy)

	client := unlock.CreateHTTPClient(ctx, proxy)

	// 方法1: 检测台湾专属内容
	result1 := d.checkTaiwanContent(ctx, client)
	if result1.Status == unlock.StatusUnlocked && result1.Region == "TW" {
		result1.CheckedAt = time.Now()
		d.LogDetectionResult(proxy, result1)
		return result1
	}

	// 方法2: 检测港澳台内容
	result2 := d.checkHKMOTWContent(ctx, client)
	if result2.Status == unlock.StatusUnlocked {
		result2.CheckedAt = time.Now()
		d.LogDetectionResult(proxy, result2)
		return result2
	}

	// 方法3: 检测大陆内容访问
	result3 := d.checkMainlandContent(ctx, client)

	result3.CheckedAt = time.Now()
	d.LogDetectionResult(proxy, result3)
	return result3
}

// checkTaiwanContent 检测台湾专属内容
func (d *BilibiliDetector) checkTaiwanContent(ctx context.Context, client *http.Client) *unlock.UnlockResult {
	// 访问台湾专属动画
	resp, err := unlock.MakeRequest(ctx, client, "GET", "https://www.bilibili.com/bangumi/play/ss21542", nil)
	if err != nil {
		return d.CreateErrorResult("Failed to connect to Bilibili", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return d.CreateErrorResult("Failed to read Bilibili response", err)
	}

	bodyStr := string(body)

	if strings.Contains(bodyStr, "地区限制") ||
		strings.Contains(bodyStr, "区域限制") ||
		strings.Contains(bodyStr, "版权方要求") {
		return d.CreateResult(unlock.StatusLocked, "", "Taiwan exclusive content blocked")
	}

	if resp.StatusCode == 200 &&
		!strings.Contains(bodyStr, "error") &&
		strings.Contains(bodyStr, "bangumi") {
		return d.CreateResult(unlock.StatusUnlocked, "TW", "Taiwan exclusive content accessible")
	}

	return d.CreateResult(unlock.StatusFailed, "", "Unable to determine Taiwan content status")
}

// checkHKMOTWContent 检测港澳台内容
func (d *BilibiliDetector) checkHKMOTWContent(ctx context.Context, client *http.Client) *unlock.UnlockResult {
	// 访问港澳台内容
	resp, err := unlock.MakeRequest(ctx, client, "GET", "https://www.bilibili.com/bangumi/play/ss28341", nil)
	if err != nil {
		return d.CreateErrorResult("Failed to connect to Bilibili", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return d.CreateErrorResult("Failed to read Bilibili response", err)
	}

	bodyStr := string(body)

	if strings.Contains(bodyStr, "地区限制") ||
		strings.Contains(bodyStr, "区域限制") ||
		strings.Contains(bodyStr, "版权方要求") {
		return d.CreateResult(unlock.StatusLocked, "", "HKMOTW content blocked")
	}

	if resp.StatusCode == 200 &&
		!strings.Contains(bodyStr, "error") &&
		strings.Contains(bodyStr, "bangumi") {
		return d.CreateResult(unlock.StatusUnlocked, "HKMOTW", "HKMOTW content accessible")
	}

	return d.CreateResult(unlock.StatusFailed, "", "Unable to determine HKMOTW content status")
}

// checkMainlandContent 检测大陆内容访问
func (d *BilibiliDetector) checkMainlandContent(ctx context.Context, client *http.Client) *unlock.UnlockResult {
	// 访问主站首页
	resp, err := unlock.MakeRequest(ctx, client, "GET", "https://www.bilibili.com", nil)
	if err != nil {
		return d.CreateErrorResult("Failed to connect to Bilibili", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return d.CreateResult(unlock.StatusUnlocked, "CN", "Bilibili mainland accessible")
	}

	return d.CreateResult(unlock.StatusLocked, "", "Bilibili not accessible")
}

// init 函数用于自动注册检测器
func init() {
	unlock.Register(NewBilibiliDetector())
}