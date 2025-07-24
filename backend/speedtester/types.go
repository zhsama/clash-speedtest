package speedtester

import (
	"time"

	"github.com/zhsama/clash-speedtest/unlock"
	"github.com/metacubex/mihomo/constant"
)

// Config speed test configuration
type Config struct {
	ConfigPaths      string
	FilterRegex      string
	IncludeNodes     []string
	ExcludeNodes     []string
	ProtocolFilter   []string
	ServerURL        string
	DownloadSize     int
	UploadSize       int
	Timeout          time.Duration
	Concurrent       int
	MaxLatency       time.Duration
	MinDownloadSpeed float64
	MinUploadSpeed   float64
	FastMode         bool
	RenameNodes      bool
	TestMode         string
	UnlockConfig     *unlock.UnlockTestConfig
}

// SpeedTester speed tester
type SpeedTester struct {
	config         *Config
	unlockDetector *unlock.Detector
}

// CProxy proxy configuration
type CProxy struct {
	constant.Proxy
	Config map[string]any
}
