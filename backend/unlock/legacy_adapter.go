package unlock

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/faceair/clash-speedtest/logger"
	"github.com/metacubex/mihomo/constant"
)

// LegacyDetectorAdapter 适配器，将旧的函数式检测器适配到新的 UnlockDetector 接口
type LegacyDetectorAdapter struct {
	*BaseDetector
	detectorFunc func(*http.Client) *StreamResult
}

// NewLegacyDetectorAdapter 创建适配器实例
func NewLegacyDetectorAdapter(platformName string, priority int, detectorFunc func(*http.Client) *StreamResult) *LegacyDetectorAdapter {
	return &LegacyDetectorAdapter{
		BaseDetector: NewBaseDetector(platformName, priority),
		detectorFunc: detectorFunc,
	}
}

// Detect 实现 UnlockDetector 接口的 Detect 方法
func (l *LegacyDetectorAdapter) Detect(ctx context.Context, proxy constant.Proxy) *UnlockResult {
	l.LogDetectionStart(proxy)

	// 创建通过代理的 HTTP 客户端
	client := createHTTPClient(ctx, proxy)

	// 调用旧的检测函数
	streamResult := l.detectorFunc(client)

	// 转换结果格式
	result := l.convertStreamResultToUnlockResult(streamResult)

	l.LogDetectionResult(proxy, result)
	return result
}

// convertStreamResultToUnlockResult 将 StreamResult 转换为 UnlockResult
func (l *LegacyDetectorAdapter) convertStreamResultToUnlockResult(streamResult *StreamResult) *UnlockResult {
	var status UnlockStatus
	var message string

	switch streamResult.Status {
	case "Success":
		status = StatusUnlocked
		if streamResult.Info != "" {
			message = streamResult.Info
		} else {
			message = "Successfully unlocked"
		}
	case "Failed":
		status = StatusLocked
		if streamResult.Info != "" {
			message = streamResult.Info
		} else {
			message = "Not available in this region"
		}
	default:
		status = StatusError
		message = "Unknown status: " + streamResult.Status
	}

	return &UnlockResult{
		Platform: streamResult.Platform,
		Status:   status,
		Region:   streamResult.Region,
		Message:  message,
	}
}

// RegisterLegacyDetector 注册单个旧版检测器
func RegisterLegacyDetector(platformName string, priority int, detectorFunc func(*http.Client) *StreamResult) {
	if detectorFunc != nil {
		adapter := NewLegacyDetectorAdapter(platformName, priority, detectorFunc)
		Register(adapter)
		
		logger.Logger.Debug("Registered platform detector",
			slog.String("platform", platformName),
			slog.Int("priority", priority),
		)
	} else {
		logger.Logger.Warn("Detector function not found",
			slog.String("platform", platformName),
		)
	}
}