package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server ServerConfig `yaml:"server"`
	Logger LoggerConfig `yaml:"logger"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// LoggerConfig contains logging configuration
type LoggerConfig struct {
	Level         string `yaml:"level"`
	OutputToFile  bool   `yaml:"output_to_file"`
	LogDir        string `yaml:"log_dir"`
	LogFileName   string `yaml:"log_file_name"`
	MaxSize       int64  `yaml:"max_size"`         // Maximum log file size in bytes
	MaxFiles      int    `yaml:"max_files"`        // Maximum number of log files to keep
	RotateOnStart bool   `yaml:"rotate_on_start"`  // Whether to rotate log on startup
	EnableConsole bool   `yaml:"enable_console"`   // Whether to also output to console
	Format        string `yaml:"format"`           // Log format: "text" or "json"
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Logger: LoggerConfig{
			Level:         "INFO",
			OutputToFile:  true,
			LogDir:        "logs",
			LogFileName:   "clash-speedtest.log",
			MaxSize:       10 * 1024 * 1024, // 10MB
			MaxFiles:      5,
			RotateOnStart: true,
			EnableConsole: true,
			Format:        "text",
		},
	}
}

// LoadConfig loads configuration from file with fallback to defaults
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file
		if err := config.SaveToFile(configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config file: %w", err)
		}
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables
	config.overrideWithEnv()

	return config, nil
}

// SaveToFile saves the configuration to a file
func (c *Config) SaveToFile(configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// overrideWithEnv overrides configuration with environment variables
func (c *Config) overrideWithEnv() {
	// Server configuration
	if port := os.Getenv("PORT"); port != "" {
		if p, err := parseInt(port); err == nil {
			c.Server.Port = p
		}
	}
	if host := os.Getenv("HOST"); host != "" {
		c.Server.Host = host
	}

	// Logger configuration
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		c.Logger.Level = strings.ToUpper(level)
	}
	if logDir := os.Getenv("LOG_DIR"); logDir != "" {
		c.Logger.LogDir = logDir
	}
	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		c.Logger.LogFileName = logFile
	}
	if toFile := os.Getenv("LOG_TO_FILE"); toFile != "" {
		c.Logger.OutputToFile = parseBool(toFile)
	}
	if toConsole := os.Getenv("LOG_TO_CONSOLE"); toConsole != "" {
		c.Logger.EnableConsole = parseBool(toConsole)
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		c.Logger.Format = strings.ToLower(format)
	}
}

// GetSlogLevel converts string log level to slog.Level
func (c *LoggerConfig) GetSlogLevel() slog.Level {
	switch strings.ToUpper(c.Level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Helper functions
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func parseBool(s string) bool {
	switch strings.ToLower(s) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}