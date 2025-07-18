package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/faceair/clash-speedtest/config"
	"github.com/faceair/clash-speedtest/detectors"
	"github.com/faceair/clash-speedtest/logger"
	"github.com/faceair/clash-speedtest/server"
	"github.com/faceair/clash-speedtest/unlock"
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
	
	// 注册主要流媒体平台检测器
	unlock.RegisterLegacyDetector("Netflix", 1, detectors.TestNetflix)
	unlock.RegisterLegacyDetector("YouTube", 1, detectors.TestYouTube)
	unlock.RegisterLegacyDetector("Disney+", 1, detectors.TestDisney)
	unlock.RegisterLegacyDetector("ChatGPT", 2, detectors.TestOpenAI)
	unlock.RegisterLegacyDetector("Spotify", 2, detectors.TestSpotify)
	unlock.RegisterLegacyDetector("Bilibili", 2, detectors.TestBilibiliMainland)
	unlock.RegisterLegacyDetector("HBO Max", 2, detectors.TestHBOMax)
	unlock.RegisterLegacyDetector("Hulu", 2, detectors.TestHulu)
	unlock.RegisterLegacyDetector("Prime Video", 2, detectors.TestPrimeVideo)
	unlock.RegisterLegacyDetector("DAZN", 3, detectors.TestDAZN)
	unlock.RegisterLegacyDetector("Paramount+", 3, detectors.TestParamount)
	unlock.RegisterLegacyDetector("Discovery+", 3, detectors.TestDiscovery)
	unlock.RegisterLegacyDetector("ESPN+", 3, detectors.TestESPN)
	unlock.RegisterLegacyDetector("Peacock", 3, detectors.TestPeacock)
	unlock.RegisterLegacyDetector("Funimation", 3, detectors.TestFunimation)
	unlock.RegisterLegacyDetector("Hotstar", 3, detectors.TestHotstar)
	
	logger.Logger.Info("Unlock detectors registration completed")
}