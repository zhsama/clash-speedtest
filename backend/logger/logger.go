package logger

import (
	"log/slog"
	"os"
	"strings"
)

var Logger *slog.Logger

func init() {
	InitLogger()
}

// InitLogger initializes the global logger with structured logging
func InitLogger() {
	level := getLogLevel()
	
	opts := &slog.HandlerOptions{
		Level: level,
		AddSource: level == slog.LevelDebug,
	}
	
	// Create a JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, opts)
	Logger = slog.New(handler)
	
	// Set as default logger
	slog.SetDefault(Logger)
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
	opts := &slog.HandlerOptions{
		Level: level,
		AddSource: level == slog.LevelDebug,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
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