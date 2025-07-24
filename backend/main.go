package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/zhsama/clash-speedtest/config"
	"github.com/zhsama/clash-speedtest/detectors"
	"github.com/zhsama/clash-speedtest/logger"
	"github.com/zhsama/clash-speedtest/server"
	"github.com/zhsama/clash-speedtest/unlock"
	"github.com/metacubex/mihomo/log"
)

func main() {
	// Parse command line flags
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	appConfig, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger with config
	logConfig := logger.FromAppConfig(appConfig)
	logger.InitLoggerWithConfig(logConfig)

	// Ensure proper cleanup on exit
	defer logger.Cleanup()

	// Enable mihomo logs for debugging
	log.SetLevel(log.DEBUG)

	// Register unlock detectors
	registerUnlockDetectors()

	logger.Logger.Info("Starting Clash SpeedTest API Server",
		slog.String("version", "2.0.0"),
		slog.String("port", fmt.Sprintf("%d", appConfig.Server.Port)),
		slog.String("config_file", configPath),
	)

	// Create and start server
	srv := server.NewServer(appConfig.Server.Port)
	
	// Start the server (this will block until shutdown)
	if err := srv.Start(); err != nil {
		logger.Logger.Error("Server failed to start", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Logger.Info("Server exited")
}

// registerUnlockDetectors 注册所有解锁检测器
func registerUnlockDetectors() {
	logger.Logger.Info("Registering unlock detectors")
	
	// 全球主流平台 (优先级 1)
	unlock.RegisterLegacyDetector("Netflix", 1, detectors.TestNetflix)
	unlock.RegisterLegacyDetector("YouTube", 1, detectors.TestYouTube)
	unlock.RegisterLegacyDetector("Disney+", 1, detectors.TestDisney)
	unlock.RegisterLegacyDetector("IP Check", 1, detectors.TestIPCheck)
	
	// AI 服务 (优先级 2)
	unlock.RegisterLegacyDetector("ChatGPT", 2, detectors.TestOpenAI)
	unlock.RegisterLegacyDetector("Gemini", 2, detectors.TestGemini)
	unlock.RegisterLegacyDetector("Meta AI", 2, detectors.TestMetaAI)
	
	// 音乐和游戏平台 (优先级 2)
	unlock.RegisterLegacyDetector("Spotify", 2, detectors.TestSpotify)
	unlock.RegisterLegacyDetector("Steam", 2, detectors.TestSteam)
	
	// 美国流媒体平台 (优先级 2-3)
	unlock.RegisterLegacyDetector("HBO Max", 2, detectors.TestHBOMax)
	unlock.RegisterLegacyDetector("Hulu", 2, detectors.TestHulu)
	unlock.RegisterLegacyDetector("Prime Video", 2, detectors.TestPrimeVideo)
	unlock.RegisterLegacyDetector("Paramount+", 3, detectors.TestParamount)
	unlock.RegisterLegacyDetector("Discovery+", 3, detectors.TestDiscovery)
	unlock.RegisterLegacyDetector("ESPN+", 3, detectors.TestESPN)
	unlock.RegisterLegacyDetector("Peacock", 3, detectors.TestPeacock)
	unlock.RegisterLegacyDetector("Funimation", 3, detectors.TestFunimation)
	unlock.RegisterLegacyDetector("DAZN", 3, detectors.TestDAZN)
	unlock.RegisterLegacyDetector("Hotstar", 3, detectors.TestHotstar)
	unlock.RegisterLegacyDetector("YouTube CDN", 3, detectors.TestYouTubeCDN)
	
	// 中国大陆平台 (优先级 2-4)
	unlock.RegisterLegacyDetector("Bilibili", 2, detectors.TestBilibiliMainland)
	unlock.RegisterLegacyDetector("Bilibili HKMCTW", 4, detectors.TestBilibiliHKMCTW)
	unlock.RegisterLegacyDetector("Bilibili TW", 4, detectors.TestBilibiliTW)
	
	// 日本平台 (优先级 4)
	unlock.RegisterLegacyDetector("Abema", 4, detectors.TestAbema)
	unlock.RegisterLegacyDetector("GYAO", 4, detectors.TestGYAO)
	unlock.RegisterLegacyDetector("TVer", 4, detectors.TestTVer)
	unlock.RegisterLegacyDetector("U-Next", 4, detectors.TestUNext)
	unlock.RegisterLegacyDetector("DMM", 4, detectors.TestDMM)
	unlock.RegisterLegacyDetector("Telasa", 4, detectors.TestTelasa)
	unlock.RegisterLegacyDetector("Paravi", 4, detectors.TestParavi)
	unlock.RegisterLegacyDetector("Video Market", 4, detectors.TestVideoMarket)
	unlock.RegisterLegacyDetector("Radiko", 4, detectors.TestRadiko)
	
	// 港台地区平台 (优先级 4)
	unlock.RegisterLegacyDetector("TVB", 4, detectors.TestTVB)
	unlock.RegisterLegacyDetector("EncoreTVB", 4, detectors.TestEncoreTVB)
	unlock.RegisterLegacyDetector("Viu", 4, detectors.TestViu)
	unlock.RegisterLegacyDetector("KKTV", 4, detectors.TestKKTV)
	unlock.RegisterLegacyDetector("Line TV", 4, detectors.TestLineTV)
	unlock.RegisterLegacyDetector("Hami Video", 4, detectors.TestHamiVideo)
	unlock.RegisterLegacyDetector("Bahamut", 4, detectors.TestBahamut)
	unlock.RegisterLegacyDetector("Catchplay", 4, detectors.TestCatchplay)
	unlock.RegisterLegacyDetector("HBO Go Asia", 4, detectors.TestHBOGoAsia)
	
	// 其他服务 (优先级 5)
	unlock.RegisterLegacyDetector("Google Play Store", 5, detectors.TestGooglePlayStore)
	
	logger.Logger.Info("Unlock detectors registration completed")
}