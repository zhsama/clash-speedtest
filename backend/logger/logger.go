package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhsama/clash-speedtest/config"
)

var Logger *slog.Logger

// LogConfig contains configuration for logging
type LogConfig struct {
	Level         slog.Level
	OutputToFile  bool
	LogDir        string
	LogFileName   string
	MaxSize       int64 // Maximum log file size in bytes (default: 10MB)
	MaxFiles      int   // Maximum number of log files to keep (default: 5)
	RotateOnStart bool  // Whether to rotate log on startup
	EnableConsole bool  // Whether to also output to console
	Format        string // Log format: "text" or "json"
}

// DefaultLogConfig returns the default logging configuration
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:         slog.LevelInfo,
		OutputToFile:  true,
		LogDir:        "logs",
		LogFileName:   "clash-speedtest.log",
		MaxSize:       10 * 1024 * 1024, // 10MB
		MaxFiles:      5,
		RotateOnStart: true,
		EnableConsole: true,
		Format:        "text",
	}
}

// FromAppConfig converts application config to LogConfig
func FromAppConfig(appConfig *config.Config) *LogConfig {
	return &LogConfig{
		Level:         appConfig.Logger.GetSlogLevel(),
		OutputToFile:  appConfig.Logger.OutputToFile,
		LogDir:        appConfig.Logger.LogDir,
		LogFileName:   appConfig.Logger.LogFileName,
		MaxSize:       appConfig.Logger.MaxSize,
		MaxFiles:      appConfig.Logger.MaxFiles,
		RotateOnStart: appConfig.Logger.RotateOnStart,
		EnableConsole: appConfig.Logger.EnableConsole,
		Format:        appConfig.Logger.Format,
	}
}

var currentConfig *LogConfig
var logFile *os.File

func init() {
	InitLogger()
}

// InitLogger initializes the global logger with default configuration
func InitLogger() {
	config := DefaultLogConfig()
	InitLoggerWithConfig(config)
}

// InitLoggerWithConfig initializes the global logger with custom configuration
func InitLoggerWithConfig(config *LogConfig) {
	currentConfig = config

	// Override log level from environment variable
	if envLevel := getLogLevel(); envLevel != slog.LevelInfo {
		config.Level = envLevel
	}

	opts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.Level == slog.LevelDebug,
	}

	var writers []io.Writer

	// Add console output if enabled
	if config.EnableConsole {
		writers = append(writers, os.Stdout)
	}

	// Add file output if enabled
	if config.OutputToFile {
		fileWriter, err := setupFileLogging(config)
		if err != nil {
			// Fallback to console only if file setup fails
			fmt.Fprintf(os.Stderr, "Failed to setup file logging: %v\n", err)
			writers = []io.Writer{os.Stdout}
		} else {
			writers = append(writers, fileWriter)
		}
	}

	// Create multi-writer if we have multiple outputs
	var writer io.Writer
	if len(writers) == 1 {
		writer = writers[0]
	} else if len(writers) > 1 {
		writer = io.MultiWriter(writers...)
	} else {
		writer = os.Stdout // Fallback
	}

	// Create handler based on the format configuration
	var handler slog.Handler
	switch strings.ToLower(config.Format) {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	case "text":
		handler = slog.NewTextHandler(writer, opts)
	default:
		// Auto-detect format based on output type
		if config.OutputToFile && !config.EnableConsole {
			// File only - use JSON format for better structure
			handler = slog.NewJSONHandler(writer, opts)
		} else {
			// Console or mixed - use text format for readability
			handler = slog.NewTextHandler(writer, opts)
		}
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)

	Logger.Info("Logger initialized",
		slog.String("level", config.Level.String()),
		slog.Bool("file_output", config.OutputToFile),
		slog.Bool("console_output", config.EnableConsole),
		slog.String("log_dir", config.LogDir),
	)
}

// setupFileLogging sets up file-based logging with rotation
func setupFileLogging(config *LogConfig) (io.Writer, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(config.LogDir, config.LogFileName)

	// Rotate existing log file if needed
	if config.RotateOnStart {
		if err := rotateLogFile(logPath, config.MaxFiles); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
		}
	}

	// Open log file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logFile = file

	// Check if file needs rotation based on size
	if stat, err := file.Stat(); err == nil && stat.Size() > config.MaxSize {
		file.Close()
		if err := rotateLogFile(logPath, config.MaxFiles); err != nil {
			return nil, fmt.Errorf("failed to rotate large log file: %w", err)
		}
		// Reopen after rotation
		file, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to reopen log file after rotation: %w", err)
		}
		logFile = file
	}

	return file, nil
}

// rotateLogFile rotates the log file by renaming it with a timestamp
func rotateLogFile(logPath string, maxFiles int) error {
	// Check if log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return nil // Nothing to rotate
	}

	// Generate timestamp for the rotated file
	timestamp := time.Now().Format("20060102_150405")
	dir := filepath.Dir(logPath)
	baseName := strings.TrimSuffix(filepath.Base(logPath), filepath.Ext(logPath))
	ext := filepath.Ext(logPath)

	rotatedPath := filepath.Join(dir, fmt.Sprintf("%s_%s%s", baseName, timestamp, ext))

	// Rename current log file
	if err := os.Rename(logPath, rotatedPath); err != nil {
		return fmt.Errorf("failed to rename log file: %w", err)
	}

	// Clean up old log files if we exceed maxFiles
	go cleanupOldLogFiles(dir, baseName, ext, maxFiles)

	return nil
}

// cleanupOldLogFiles removes old log files to maintain the maximum count
func cleanupOldLogFiles(dir, baseName, ext string, maxFiles int) {
	pattern := filepath.Join(dir, fmt.Sprintf("%s_*%s", baseName, ext))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	// If we have more files than allowed, remove the oldest ones
	if len(matches) > maxFiles {
		// Sort files by modification time (oldest first)
		// For simplicity, we'll just remove excess files
		// In a production system, you'd want to sort by timestamp
		for i := 0; i < len(matches)-maxFiles; i++ {
			os.Remove(matches[i])
		}
	}
}

// getLogLevel returns the log level based on environment variable
func getLogLevel() slog.Level {
	levelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch levelStr {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// SetLevel allows changing the log level at runtime
func SetLevel(level slog.Level) {
	if currentConfig != nil {
		currentConfig.Level = level
		InitLoggerWithConfig(currentConfig)
	}
}

// RotateLogNow forces immediate log rotation
func RotateLogNow() error {
	if currentConfig != nil && currentConfig.OutputToFile && logFile != nil {
		logPath := filepath.Join(currentConfig.LogDir, currentConfig.LogFileName)

		// Close current file
		logFile.Close()

		// Rotate
		if err := rotateLogFile(logPath, currentConfig.MaxFiles); err != nil {
			return err
		}

		// Reinitialize logging
		InitLoggerWithConfig(currentConfig)
		return nil
	}
	return fmt.Errorf("file logging not enabled or not initialized")
}

// Cleanup closes the log file properly
func Cleanup() {
	if logFile != nil {
		Logger.Info("Closing log file")
		logFile.Close()
		logFile = nil
	}
}

// LogHTTPRequest logs HTTP request information
func LogHTTPRequest(method, path, remoteAddr string, statusCode int, duration string) {
	Logger.Info("HTTP Request",
		slog.String("method", method),
		slog.String("path", path),
		slog.String("remote_addr", remoteAddr),
		slog.Int("status_code", statusCode),
		slog.String("duration", duration),
	)
}

// LogError logs error with context
func LogError(msg string, err error, attrs ...slog.Attr) {
	args := make([]any, 0, len(attrs)*2+2)
	args = append(args, slog.String("error", err.Error()))
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value)
	}
	Logger.Error(msg, args...)
}

// LogSpeedTest logs speed test related information
func LogSpeedTest(event string, attrs ...slog.Attr) {
	args := make([]any, 0, len(attrs)*2+2)
	args = append(args, slog.String("component", "speedtest"))
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value)
	}
	Logger.Info(event, args...)
}
