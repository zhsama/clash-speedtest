package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/faceair/clash-speedtest/config"
	"github.com/faceair/clash-speedtest/logger"
	"github.com/faceair/clash-speedtest/server"
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