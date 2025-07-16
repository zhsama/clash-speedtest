package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/faceair/clash-speedtest/utils/stats"
	"gopkg.in/yaml.v3"
)

// ExportFormat represents different export formats
type ExportFormat string

const (
	FormatJSON  ExportFormat = "json"
	FormatCSV   ExportFormat = "csv"
	FormatYAML  ExportFormat = "yaml"
	FormatClash ExportFormat = "clash" // Clash configuration format
)

// ExportOptions contains options for exporting results
type ExportOptions struct {
	Format          ExportFormat `json:"format"`
	OutputPath      string       `json:"output_path"`
	IncludeFailures bool         `json:"include_failures"`
	SortBy          string       `json:"sort_by"`           // "latency", "download", "upload", "name"
	TopN            int          `json:"top_n"`             // Export only top N results (0 = all)
	MinLatency      int          `json:"min_latency_ms"`    // Filter by minimum latency
	MaxLatency      int          `json:"max_latency_ms"`    // Filter by maximum latency
	MinDownload     float64      `json:"min_download_mbps"` // Filter by minimum download speed
	MinUpload       float64      `json:"min_upload_mbps"`   // Filter by minimum upload speed
}

// ExportableResult represents a result that can be exported
type ExportableResult struct {
	ProxyName     string    `json:"proxy_name" csv:"Proxy Name"`
	ProxyType     string    `json:"proxy_type" csv:"Proxy Type"`
	ProxyServer   string    `json:"proxy_server" csv:"Server"`
	ProxyPort     int       `json:"proxy_port" csv:"Port"`
	Country       string    `json:"country" csv:"Country"`
	CountryCode   string    `json:"country_code" csv:"Country Code"`
	City          string    `json:"city" csv:"City"`
	ISP           string    `json:"isp" csv:"ISP"`
	Latency       int64     `json:"latency_ms" csv:"Latency (ms)"`
	Jitter        int64     `json:"jitter_ms" csv:"Jitter (ms)"`
	PacketLoss    float64   `json:"packet_loss_percent" csv:"Packet Loss (%)"`
	DownloadSpeed float64   `json:"download_speed_mbps" csv:"Download (Mbps)"`
	UploadSpeed   float64   `json:"upload_speed_mbps" csv:"Upload (Mbps)"`
	TestTime      time.Time `json:"test_time" csv:"Test Time"`
	Status        string    `json:"status" csv:"Status"`
	ErrorMessage  string    `json:"error_message,omitempty" csv:"Error Message"`

	// Original proxy configuration for Clash export
	ProxyConfig map[string]any `json:"proxy_config,omitempty" csv:"-"`
}

// ClashConfig represents a Clash configuration file
type ClashConfig struct {
	Port               int              `yaml:"port"`
	SocksPort          int              `yaml:"socks-port"`
	AllowLan           bool             `yaml:"allow-lan"`
	Mode               string           `yaml:"mode"`
	LogLevel           string           `yaml:"log-level"`
	ExternalController string           `yaml:"external-controller"`
	Proxies            []map[string]any `yaml:"proxies"`
	ProxyGroups        []map[string]any `yaml:"proxy-groups"`
	Rules              []string         `yaml:"rules"`
}

// Exporter handles exporting test results in various formats
type Exporter struct {
	results []ExportableResult
	stats   *stats.TestStatistics
}

// NewExporter creates a new exporter
func NewExporter() *Exporter {
	return &Exporter{
		results: make([]ExportableResult, 0),
	}
}

// AddResult adds a result to be exported
func (e *Exporter) AddResult(result ExportableResult) {
	e.results = append(e.results, result)
}

// SetStatistics sets the test statistics
func (e *Exporter) SetStatistics(stats *stats.TestStatistics) {
	e.stats = stats
}

// Export exports the results in the specified format
func (e *Exporter) Export(options ExportOptions) error {
	// Filter and sort results
	filteredResults := e.filterResults(options)
	sortedResults := e.sortResults(filteredResults, options.SortBy)

	// Limit to top N if specified
	if options.TopN > 0 && len(sortedResults) > options.TopN {
		sortedResults = sortedResults[:options.TopN]
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(options.OutputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	switch options.Format {
	case FormatJSON:
		return e.exportJSON(sortedResults, options.OutputPath)
	case FormatCSV:
		return e.exportCSV(sortedResults, options.OutputPath)
	case FormatYAML:
		return e.exportYAML(sortedResults, options.OutputPath)
	case FormatClash:
		return e.exportClash(sortedResults, options.OutputPath)
	default:
		return fmt.Errorf("unsupported export format: %s", options.Format)
	}
}

// filterResults filters results based on the specified criteria
func (e *Exporter) filterResults(options ExportOptions) []ExportableResult {
	var filtered []ExportableResult

	for _, result := range e.results {
		// Skip failures if not included
		if !options.IncludeFailures && result.Status != "success" {
			continue
		}

		// Apply filters
		if options.MinLatency > 0 && result.Latency < int64(options.MinLatency) {
			continue
		}
		if options.MaxLatency > 0 && result.Latency > int64(options.MaxLatency) {
			continue
		}
		if options.MinDownload > 0 && result.DownloadSpeed < options.MinDownload {
			continue
		}
		if options.MinUpload > 0 && result.UploadSpeed < options.MinUpload {
			continue
		}

		filtered = append(filtered, result)
	}

	return filtered
}

// sortResults sorts results based on the specified field
func (e *Exporter) sortResults(results []ExportableResult, sortBy string) []ExportableResult {
	// Implementation would include sorting logic
	// For now, return as-is
	return results
}

// exportJSON exports results to JSON format
func (e *Exporter) exportJSON(results []ExportableResult, outputPath string) error {
	data := struct {
		Metadata struct {
			ExportTime   time.Time                `json:"export_time"`
			TotalResults int                      `json:"total_results"`
			Statistics   *stats.TestStatistics `json:"statistics,omitempty"`
		} `json:"metadata"`
		Results []ExportableResult `json:"results"`
	}{
		Results: results,
	}

	data.Metadata.ExportTime = time.Now()
	data.Metadata.TotalResults = len(results)
	data.Metadata.Statistics = e.stats

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// exportCSV exports results to CSV format
func (e *Exporter) exportCSV(results []ExportableResult, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Proxy Name", "Proxy Type", "Server", "Port", "Country", "Country Code",
		"City", "ISP", "Latency (ms)", "Jitter (ms)", "Packet Loss (%)",
		"Download (Mbps)", "Upload (Mbps)", "Test Time", "Status", "Error Message",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, result := range results {
		row := []string{
			result.ProxyName,
			result.ProxyType,
			result.ProxyServer,
			fmt.Sprintf("%d", result.ProxyPort),
			result.Country,
			result.CountryCode,
			result.City,
			result.ISP,
			fmt.Sprintf("%d", result.Latency),
			fmt.Sprintf("%d", result.Jitter),
			fmt.Sprintf("%.2f", result.PacketLoss),
			fmt.Sprintf("%.2f", result.DownloadSpeed),
			fmt.Sprintf("%.2f", result.UploadSpeed),
			result.TestTime.Format("2006-01-02 15:04:05"),
			result.Status,
			result.ErrorMessage,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// exportYAML exports results to YAML format
func (e *Exporter) exportYAML(results []ExportableResult, outputPath string) error {
	data := struct {
		Metadata struct {
			ExportTime   time.Time                `yaml:"export_time"`
			TotalResults int                      `yaml:"total_results"`
			Statistics   *stats.TestStatistics `yaml:"statistics,omitempty"`
		} `yaml:"metadata"`
		Results []ExportableResult `yaml:"results"`
	}{
		Results: results,
	}

	data.Metadata.ExportTime = time.Now()
	data.Metadata.TotalResults = len(results)
	data.Metadata.Statistics = e.stats

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create YAML file: %w", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()
	return encoder.Encode(data)
}

// exportClash exports results to Clash configuration format
func (e *Exporter) exportClash(results []ExportableResult, outputPath string) error {
	config := ClashConfig{
		Port:               7890,
		SocksPort:          7891,
		AllowLan:           false,
		Mode:               "Rule",
		LogLevel:           "info",
		ExternalController: "127.0.0.1:9090",
		Proxies:            make([]map[string]any, 0),
		ProxyGroups: []map[string]any{
			{
				"name":    "🚀 节点选择",
				"type":    "select",
				"proxies": []string{"♻️ 自动选择", "🎯 全球直连"},
			},
			{
				"name":     "♻️ 自动选择",
				"type":     "url-test",
				"proxies":  []string{},
				"url":      "http://www.gstatic.com/generate_204",
				"interval": 300,
			},
			{
				"name":    "🎯 全球直连",
				"type":    "select",
				"proxies": []string{"DIRECT"},
			},
		},
		Rules: []string{
			"DOMAIN-SUFFIX,local,DIRECT",
			"IP-CIDR,127.0.0.0/8,DIRECT",
			"IP-CIDR,172.16.0.0/12,DIRECT",
			"IP-CIDR,192.168.0.0/16,DIRECT",
			"IP-CIDR,10.0.0.0/8,DIRECT",
			"IP-CIDR,17.0.0.0/8,DIRECT",
			"IP-CIDR,100.64.0.0/10,DIRECT",
			"GEOIP,CN,DIRECT",
			"MATCH,🚀 节点选择",
		},
	}

	// Add successful proxies to config
	var proxyNames []string
	for _, result := range results {
		if result.Status == "success" && result.ProxyConfig != nil {
			// Add speed information to proxy name
			enhancedName := fmt.Sprintf("%s | ⬇️%.1fM ⬆️%.1fM ⏱️%dms",
				result.ProxyName,
				result.DownloadSpeed,
				result.UploadSpeed,
				result.Latency,
			)

			// Create a copy of the proxy config and update the name
			proxyConfig := make(map[string]any)
			for k, v := range result.ProxyConfig {
				proxyConfig[k] = v
			}
			proxyConfig["name"] = enhancedName

			config.Proxies = append(config.Proxies, proxyConfig)
			proxyNames = append(proxyNames, enhancedName)
		}
	}

	// Update auto-select group with proxy names
	if len(proxyNames) > 0 {
		config.ProxyGroups[1]["proxies"] = proxyNames
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create Clash config file: %w", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()
	return encoder.Encode(config)
}

// GenerateFilename generates a filename with timestamp
func GenerateFilename(prefix string, format ExportFormat) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s.%s", prefix, timestamp, string(format))
}

// GetSupportedFormats returns all supported export formats
func GetSupportedFormats() []ExportFormat {
	return []ExportFormat{FormatJSON, FormatCSV, FormatYAML, FormatClash}
}

// ValidateExportOptions validates export options
func ValidateExportOptions(options ExportOptions) error {
	if options.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}

	supportedFormats := GetSupportedFormats()
	formatSupported := false
	for _, format := range supportedFormats {
		if options.Format == format {
			formatSupported = true
			break
		}
	}
	if !formatSupported {
		return fmt.Errorf("unsupported format: %s, supported formats: %s",
			options.Format, strings.Join(formatStrings(supportedFormats), ", "))
	}

	if options.TopN < 0 {
		return fmt.Errorf("top_n must be non-negative")
	}

	return nil
}

// Helper function to convert format slice to string slice
func formatStrings(formats []ExportFormat) []string {
	result := make([]string, len(formats))
	for i, format := range formats {
		result[i] = string(format)
	}
	return result
}
