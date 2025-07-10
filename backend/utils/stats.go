package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"
)

// TestStatistics contains comprehensive statistics for speed test results
type TestStatistics struct {
	TotalProxies      int                    `json:"total_proxies"`
	TestedProxies     int                    `json:"tested_proxies"`
	SuccessfulProxies int                    `json:"successful_proxies"`
	FailedProxies     int                    `json:"failed_proxies"`
	SkippedProxies    int                    `json:"skipped_proxies"`
	
	// 时间统计
	StartTime         time.Time             `json:"start_time"`
	EndTime           time.Time             `json:"end_time"`
	TotalDuration     time.Duration         `json:"total_duration"`
	AverageTestTime   time.Duration         `json:"average_test_time"`
	
	// 延迟统计
	LatencyStats      *LatencyStatistics    `json:"latency_stats"`
	
	// 速度统计
	DownloadStats     *SpeedStatistics      `json:"download_stats"`
	UploadStats       *SpeedStatistics      `json:"upload_stats"`
	
	// 协议分布
	ProtocolStats     map[string]int        `json:"protocol_stats"`
	
	// 地理分布
	CountryStats      map[string]int        `json:"country_stats"`
	
	// 最佳节点
	BestLatencyProxy  *ProxyRanking         `json:"best_latency_proxy"`
	BestDownloadProxy *ProxyRanking         `json:"best_download_proxy"`
	BestUploadProxy   *ProxyRanking         `json:"best_upload_proxy"`
	
	// 错误统计
	ErrorStats        map[string]int        `json:"error_stats"`
}

// LatencyStatistics contains latency-related statistics
type LatencyStatistics struct {
	Mean       float64 `json:"mean_ms"`
	Median     float64 `json:"median_ms"`
	Min        float64 `json:"min_ms"`
	Max        float64 `json:"max_ms"`
	StdDev     float64 `json:"std_dev_ms"`
	P95        float64 `json:"p95_ms"`
	P99        float64 `json:"p99_ms"`
	
	// 抖动统计
	JitterMean float64 `json:"jitter_mean_ms"`
	JitterMax  float64 `json:"jitter_max_ms"`
	
	// 丢包统计
	PacketLossMean float64 `json:"packet_loss_mean_percent"`
	PacketLossMax  float64 `json:"packet_loss_max_percent"`
}

// SpeedStatistics contains speed-related statistics
type SpeedStatistics struct {
	Mean       float64 `json:"mean_mbps"`
	Median     float64 `json:"median_mbps"`
	Min        float64 `json:"min_mbps"`
	Max        float64 `json:"max_mbps"`
	StdDev     float64 `json:"std_dev_mbps"`
	P95        float64 `json:"p95_mbps"`
	P99        float64 `json:"p99_mbps"`
	
	// 分级统计
	Excellent  int `json:"excellent_count"`  // > 100 Mbps
	Good       int `json:"good_count"`       // 50-100 Mbps
	Average    int `json:"average_count"`    // 10-50 Mbps
	Poor       int `json:"poor_count"`       // < 10 Mbps
}

// ProxyRanking represents a ranked proxy result
type ProxyRanking struct {
	ProxyName     string  `json:"proxy_name"`
	ProxyType     string  `json:"proxy_type"`
	Country       string  `json:"country,omitempty"`
	CountryCode   string  `json:"country_code,omitempty"`
	Value         float64 `json:"value"`
	Unit          string  `json:"unit"`
	Latency       float64 `json:"latency_ms,omitempty"`
	DownloadSpeed float64 `json:"download_speed_mbps,omitempty"`
	UploadSpeed   float64 `json:"upload_speed_mbps,omitempty"`
}

// TestResult represents a simplified test result for statistics
type TestResult struct {
	ProxyName     string
	ProxyType     string
	Country       string
	CountryCode   string
	Latency       time.Duration
	Jitter        time.Duration
	PacketLoss    float64
	DownloadSpeed float64  // bytes per second
	UploadSpeed   float64  // bytes per second
	Success       bool
	ErrorType     string
	TestDuration  time.Duration
}

// StatisticsCalculator provides methods to calculate comprehensive statistics
type StatisticsCalculator struct {
	results []TestResult
}

// NewStatisticsCalculator creates a new statistics calculator
func NewStatisticsCalculator() *StatisticsCalculator {
	return &StatisticsCalculator{
		results: make([]TestResult, 0),
	}
}

// AddResult adds a test result to the calculator
func (sc *StatisticsCalculator) AddResult(result TestResult) {
	sc.results = append(sc.results, result)
}

// Calculate computes comprehensive statistics from all added results
func (sc *StatisticsCalculator) Calculate() *TestStatistics {
	if len(sc.results) == 0 {
		return &TestStatistics{}
	}

	stats := &TestStatistics{
		TotalProxies:  len(sc.results),
		ProtocolStats: make(map[string]int),
		CountryStats:  make(map[string]int),
		ErrorStats:    make(map[string]int),
	}

	// 分离成功和失败的结果
	var successResults []TestResult
	var latencies []float64
	var downloadSpeeds []float64
	var uploadSpeeds []float64
	var jitters []float64
	var packetLosses []float64
	var testDurations []time.Duration

	for _, result := range sc.results {
		// 统计协议分布
		stats.ProtocolStats[result.ProxyType]++
		
		// 统计国家分布
		if result.Country != "" {
			stats.CountryStats[result.Country]++
		}
		
		if result.Success {
			stats.SuccessfulProxies++
			successResults = append(successResults, result)
			
			if result.Latency > 0 {
				latencies = append(latencies, float64(result.Latency.Milliseconds()))
				jitters = append(jitters, float64(result.Jitter.Milliseconds()))
				packetLosses = append(packetLosses, result.PacketLoss)
			}
			
			if result.DownloadSpeed > 0 {
				downloadSpeeds = append(downloadSpeeds, result.DownloadSpeed/(1024*1024)) // Convert to Mbps
			}
			
			if result.UploadSpeed > 0 {
				uploadSpeeds = append(uploadSpeeds, result.UploadSpeed/(1024*1024)) // Convert to Mbps
			}
			
			if result.TestDuration > 0 {
				testDurations = append(testDurations, result.TestDuration)
			}
		} else {
			stats.FailedProxies++
			if result.ErrorType != "" {
				stats.ErrorStats[result.ErrorType]++
			}
		}
	}

	stats.TestedProxies = stats.SuccessfulProxies + stats.FailedProxies

	// 计算延迟统计
	if len(latencies) > 0 {
		stats.LatencyStats = calculateLatencyStats(latencies, jitters, packetLosses)
	}

	// 计算下载速度统计
	if len(downloadSpeeds) > 0 {
		stats.DownloadStats = calculateSpeedStats(downloadSpeeds)
	}

	// 计算上传速度统计
	if len(uploadSpeeds) > 0 {
		stats.UploadStats = calculateSpeedStats(uploadSpeeds)
	}

	// 计算时间统计
	if len(testDurations) > 0 {
		var totalDuration time.Duration
		for _, d := range testDurations {
			totalDuration += d
		}
		stats.AverageTestTime = totalDuration / time.Duration(len(testDurations))
	}

	// 找出最佳节点
	stats.BestLatencyProxy = findBestProxy(successResults, "latency")
	stats.BestDownloadProxy = findBestProxy(successResults, "download")
	stats.BestUploadProxy = findBestProxy(successResults, "upload")

	return stats
}

// calculateLatencyStats calculates latency-related statistics
func calculateLatencyStats(latencies, jitters, packetLosses []float64) *LatencyStatistics {
	sort.Float64s(latencies)
	sort.Float64s(jitters)
	sort.Float64s(packetLosses)

	stats := &LatencyStatistics{
		Mean:   calculateMean(latencies),
		Median: calculateMedian(latencies),
		Min:    latencies[0],
		Max:    latencies[len(latencies)-1],
		StdDev: calculateStdDev(latencies),
		P95:    calculatePercentile(latencies, 95),
		P99:    calculatePercentile(latencies, 99),
	}

	if len(jitters) > 0 {
		stats.JitterMean = calculateMean(jitters)
		stats.JitterMax = jitters[len(jitters)-1]
	}

	if len(packetLosses) > 0 {
		stats.PacketLossMean = calculateMean(packetLosses)
		stats.PacketLossMax = packetLosses[len(packetLosses)-1]
	}

	return stats
}

// calculateSpeedStats calculates speed-related statistics
func calculateSpeedStats(speeds []float64) *SpeedStatistics {
	sort.Float64s(speeds)

	stats := &SpeedStatistics{
		Mean:   calculateMean(speeds),
		Median: calculateMedian(speeds),
		Min:    speeds[0],
		Max:    speeds[len(speeds)-1],
		StdDev: calculateStdDev(speeds),
		P95:    calculatePercentile(speeds, 95),
		P99:    calculatePercentile(speeds, 99),
	}

	// 分级统计
	for _, speed := range speeds {
		switch {
		case speed > 100:
			stats.Excellent++
		case speed >= 50:
			stats.Good++
		case speed >= 10:
			stats.Average++
		default:
			stats.Poor++
		}
	}

	return stats
}

// findBestProxy finds the best proxy based on the specified metric
func findBestProxy(results []TestResult, metric string) *ProxyRanking {
	if len(results) == 0 {
		return nil
	}

	var bestResult *TestResult
	var bestValue float64
	var unit string

	for i, result := range results {
		var value float64
		switch metric {
		case "latency":
			if result.Latency <= 0 {
				continue
			}
			value = float64(result.Latency.Milliseconds())
			unit = "ms"
			if bestResult == nil || value < bestValue {
				bestResult = &results[i]
				bestValue = value
			}
		case "download":
			if result.DownloadSpeed <= 0 {
				continue
			}
			value = result.DownloadSpeed / (1024 * 1024) // Convert to Mbps
			unit = "Mbps"
			if bestResult == nil || value > bestValue {
				bestResult = &results[i]
				bestValue = value
			}
		case "upload":
			if result.UploadSpeed <= 0 {
				continue
			}
			value = result.UploadSpeed / (1024 * 1024) // Convert to Mbps
			unit = "Mbps"
			if bestResult == nil || value > bestValue {
				bestResult = &results[i]
				bestValue = value
			}
		}
	}

	if bestResult == nil {
		return nil
	}

	return &ProxyRanking{
		ProxyName:     bestResult.ProxyName,
		ProxyType:     bestResult.ProxyType,
		Country:       bestResult.Country,
		CountryCode:   bestResult.CountryCode,
		Value:         bestValue,
		Unit:          unit,
		Latency:       float64(bestResult.Latency.Milliseconds()),
		DownloadSpeed: bestResult.DownloadSpeed / (1024 * 1024),
		UploadSpeed:   bestResult.UploadSpeed / (1024 * 1024),
	}
}

// Mathematical helper functions

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	n := len(values)
	if n%2 == 0 {
		return (values[n/2-1] + values[n/2]) / 2
	}
	return values[n/2]
}

func calculateStdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := calculateMean(values)
	var sumSquares float64
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	variance := sumSquares / float64(len(values))
	return math.Sqrt(variance)
}

func calculatePercentile(values []float64, percentile int) float64 {
	if len(values) == 0 {
		return 0
	}
	if percentile <= 0 {
		return values[0]
	}
	if percentile >= 100 {
		return values[len(values)-1]
	}
	
	index := float64(percentile) / 100.0 * float64(len(values)-1)
	if index == float64(int(index)) {
		return values[int(index)]
	}
	
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	weight := index - float64(lower)
	
	return values[lower]*(1-weight) + values[upper]*weight
}

// FormatDuration formats a duration to a human-readable string
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// ToJSON converts statistics to JSON
func (stats *TestStatistics) ToJSON() (string, error) {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetSummary returns a brief summary of the statistics
func (stats *TestStatistics) GetSummary() string {
	successRate := float64(stats.SuccessfulProxies) / float64(stats.TotalProxies) * 100
	
	summary := fmt.Sprintf("Test Summary: %d/%d proxies tested successfully (%.1f%%)",
		stats.SuccessfulProxies, stats.TotalProxies, successRate)
	
	if stats.LatencyStats != nil {
		summary += fmt.Sprintf(", Avg Latency: %.1fms", stats.LatencyStats.Mean)
	}
	
	if stats.DownloadStats != nil {
		summary += fmt.Sprintf(", Avg Download: %.1f Mbps", stats.DownloadStats.Mean)
	}
	
	if stats.UploadStats != nil {
		summary += fmt.Sprintf(", Avg Upload: %.1f Mbps", stats.UploadStats.Mean)
	}
	
	return summary
}