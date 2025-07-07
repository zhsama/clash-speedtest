package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/faceair/clash-speedtest/speedtester"
	"github.com/metacubex/mihomo/log"
)

type TestRequest struct {
	ConfigPaths      string  `json:"configPaths"`
	FilterRegex      string  `json:"filterRegex"`
	ServerURL        string  `json:"serverUrl"`
	DownloadSize     int     `json:"downloadSize"`
	UploadSize       int     `json:"uploadSize"`
	Timeout          int     `json:"timeout"`
	Concurrent       int     `json:"concurrent"`
	MaxLatency       int     `json:"maxLatency"`
	MinDownloadSpeed float64 `json:"minDownloadSpeed"`
	MinUploadSpeed   float64 `json:"minUploadSpeed"`
	StashCompatible  bool    `json:"stashCompatible"`
	RenameNodes      bool    `json:"renameNodes"`
}

type TestResponse struct {
	Success bool                    `json:"success"`
	Error   string                  `json:"error,omitempty"`
	Results []*speedtester.Result   `json:"results,omitempty"`
}

var port = flag.String("port", "8090", "API server port")

func main() {
	flag.Parse()
	log.SetLevel(log.SILENT)

	http.HandleFunc("/api/test", handleTest)
	http.HandleFunc("/api/health", handleHealth)

	// Enable CORS
	handler := corsMiddleware(http.DefaultServeMux)

	server := &http.Server{
		Addr:    ":" + *port,
		Handler: handler,
	}

	// 设置优雅关闭
	go func() {
		fmt.Printf("API Server running on port %s\n", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalln("Server forced to shutdown:", err)
	}
	fmt.Println("Server exited")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body: "+err.Error())
		return
	}

	// Set defaults
	if req.FilterRegex == "" {
		req.FilterRegex = ".+"
	}
	if req.ServerURL == "" {
		req.ServerURL = "https://speed.cloudflare.com"
	}
	if req.DownloadSize == 0 {
		req.DownloadSize = 50 * 1024 * 1024
	}
	if req.UploadSize == 0 {
		req.UploadSize = 20 * 1024 * 1024
	}
	if req.Timeout == 0 {
		req.Timeout = 5
	}
	if req.Concurrent == 0 {
		req.Concurrent = 4
	}
	if req.MaxLatency == 0 {
		req.MaxLatency = 800
	}
	if req.MinDownloadSpeed == 0 {
		req.MinDownloadSpeed = 5
	}
	if req.MinUploadSpeed == 0 {
		req.MinUploadSpeed = 2
	}

	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths:      req.ConfigPaths,
		FilterRegex:      req.FilterRegex,
		ServerURL:        req.ServerURL,
		DownloadSize:     req.DownloadSize,
		UploadSize:       req.UploadSize,
		Timeout:          time.Duration(req.Timeout) * time.Second,
		Concurrent:       req.Concurrent,
		MaxLatency:       time.Duration(req.MaxLatency) * time.Millisecond,
		MinDownloadSpeed: req.MinDownloadSpeed * 1024 * 1024,
		MinUploadSpeed:   req.MinUploadSpeed * 1024 * 1024,
	})

	allProxies, err := speedTester.LoadProxies(req.StashCompatible)
	if err != nil {
		sendError(w, "Failed to load proxies: "+err.Error())
		return
	}

	results := make([]*speedtester.Result, 0)
	speedTester.TestProxies(allProxies, func(result *speedtester.Result) {
		results = append(results, result)
	})

	// Filter results
	filteredResults := make([]*speedtester.Result, 0)
	for _, result := range results {
		if req.MaxLatency > 0 && result.Latency > time.Duration(req.MaxLatency)*time.Millisecond {
			continue
		}
		if req.MinDownloadSpeed > 0 && result.DownloadSpeed < req.MinDownloadSpeed*1024*1024 {
			continue
		}
		if req.MinUploadSpeed > 0 && result.UploadSpeed < req.MinUploadSpeed*1024*1024 {
			continue
		}
		filteredResults = append(filteredResults, result)
	}

	sendSuccess(w, filteredResults)
}

func sendError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(TestResponse{
		Success: false,
		Error:   message,
	})
}

func sendSuccess(w http.ResponseWriter, results []*speedtester.Result) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TestResponse{
		Success: true,
		Results: results,
	})
}