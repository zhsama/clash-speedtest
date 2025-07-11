package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/faceair/clash-speedtest/logger"
	"github.com/faceair/clash-speedtest/speedtester"
)

// generateTaskID generates a unique task ID
func generateTaskID() string {
	return fmt.Sprintf("task-%d-%s", time.Now().Unix(), time.Now().Format("150405"))
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Debug("Health check requested")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// sendError sends an error response
func sendError(w http.ResponseWriter, message string) {
	logger.Logger.Error("Sending error response", slog.String("error", message))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(TestResponse{
		Success: false,
		Error:   message,
	})
}

// sendSuccess sends a success response
func sendSuccess(w http.ResponseWriter, results []*speedtester.Result) {
	logger.Logger.Info("Sending successful response", slog.Int("result_count", len(results)))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TestResponse{
		Success: true,
		Results: results,
	})
}
