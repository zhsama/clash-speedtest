package unlock

import (
	"time"
)

// DefaultUnlockConfig 返回默认的解锁检测配置
func DefaultUnlockConfig() *UnlockTestConfig {
	return &UnlockTestConfig{
		Enabled:       true,
		Platforms:     []string{"Netflix", "YouTube", "Disney+", "ChatGPT", "Spotify", "Bilibili"},
		Concurrent:    5,
		Timeout:       10,
		RetryOnError:  true,
		IncludeIPInfo: true,
	}
}

// ValidateConfig 验证配置
func ValidateConfig(config *UnlockTestConfig) error {
	if config == nil {
		return nil
	}

	// 验证并发数
	if config.Concurrent <= 0 {
		config.Concurrent = 5
	}
	if config.Concurrent > 20 {
		config.Concurrent = 20
	}

	// 验证超时时间
	if config.Timeout <= 0 {
		config.Timeout = 10
	}
	if config.Timeout > 60 {
		config.Timeout = 60
	}

	// 验证平台列表
	if len(config.Platforms) == 0 {
		config.Platforms = []string{"Netflix", "YouTube", "Disney+", "ChatGPT", "Spotify", "Bilibili"}
	}

	return nil
}

// PlatformSettings 平台设置
type PlatformSettings struct {
	Name        string
	DisplayName string
	Category    PlatformCategory
	Priority    int
	Enabled     bool
	Description string
	Timeout     time.Duration
	RetryCount  int
}

// GetDefaultPlatformSettings 获取默认平台设置
func GetDefaultPlatformSettings() map[string]PlatformSettings {
	return map[string]PlatformSettings{
		"Netflix": {
			Name:        "Netflix",
			DisplayName: "Netflix",
			Category:    CategoryStreaming,
			Priority:    1,
			Enabled:     true,
			Description: "Netflix 流媒体服务解锁检测",
			Timeout:     10 * time.Second,
			RetryCount:  2,
		},
		"YouTube": {
			Name:        "YouTube",
			DisplayName: "YouTube Premium",
			Category:    CategoryStreaming,
			Priority:    1,
			Enabled:     true,
			Description: "YouTube Premium 服务解锁检测",
			Timeout:     10 * time.Second,
			RetryCount:  2,
		},
		"Disney+": {
			Name:        "Disney+",
			DisplayName: "Disney Plus",
			Category:    CategoryStreaming,
			Priority:    2,
			Enabled:     true,
			Description: "Disney+ 流媒体服务解锁检测",
			Timeout:     10 * time.Second,
			RetryCount:  2,
		},
		"ChatGPT": {
			Name:        "ChatGPT",
			DisplayName: "OpenAI ChatGPT",
			Category:    CategorySocial,
			Priority:    2,
			Enabled:     true,
			Description: "OpenAI ChatGPT 服务解锁检测",
			Timeout:     10 * time.Second,
			RetryCount:  2,
		},
		"Spotify": {
			Name:        "Spotify",
			DisplayName: "Spotify",
			Category:    CategoryStreaming,
			Priority:    2,
			Enabled:     true,
			Description: "Spotify 音乐流媒体服务解锁检测",
			Timeout:     10 * time.Second,
			RetryCount:  2,
		},
		"Bilibili": {
			Name:        "Bilibili",
			DisplayName: "哔哩哔哩",
			Category:    CategoryRegional,
			Priority:    3,
			Enabled:     true,
			Description: "哔哩哔哩视频服务解锁检测",
			Timeout:     10 * time.Second,
			RetryCount:  2,
		},
	}
}

// GetPlatformsByCategory 按分类获取平台
func GetPlatformsByCategory(category PlatformCategory) []string {
	settings := GetDefaultPlatformSettings()
	var platforms []string
	
	for name, setting := range settings {
		if setting.Category == category && setting.Enabled {
			platforms = append(platforms, name)
		}
	}
	
	return platforms
}

// GetPlatformsByPriority 按优先级获取平台
func GetPlatformsByPriority(priority int) []string {
	settings := GetDefaultPlatformSettings()
	var platforms []string
	
	for name, setting := range settings {
		if setting.Priority == priority && setting.Enabled {
			platforms = append(platforms, name)
		}
	}
	
	return platforms
}

// IsPlatformSupported 检查平台是否支持
func IsPlatformSupported(platform string) bool {
	settings := GetDefaultPlatformSettings()
	setting, exists := settings[platform]
	return exists && setting.Enabled
}