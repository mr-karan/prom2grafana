package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/mr-karan/prom2grafana/internal/ai"
	"github.com/mr-karan/prom2grafana/internal/models"
)

const (
	// Maximum request body size (1MB)
	maxRequestBodySize = 1 << 20 // 1MB
	// API timeout duration
	apiTimeout = 30 * time.Second
)

// ConvertHandler handles the /convert endpoint
type ConvertHandler struct {
	aiClient *ai.Client
}

// NewConvertHandler creates a new convert handler
func NewConvertHandler(aiClient *ai.Client) *ConvertHandler {
	return &ConvertHandler{
		aiClient: aiClient,
	}
}

// Handle processes the convert request
func (h *ConvertHandler) Handle(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	slog.Info("Handling convert request", "method", r.Method, "path", r.URL.Path)

	if r.Method != "POST" {
		slog.Debug("Invalid method", "method", r.Method)
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var req models.ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if err.Error() == "http: request body too large" {
			slog.Error("Request body too large", "error", err)
			respondWithError(w, http.StatusRequestEntityTooLarge, "Request body too large (max 1MB)")
			return
		}
		slog.Error("Failed to decode request body", "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Metrics) == "" {
		slog.Debug("Empty metrics provided")
		respondWithError(w, http.StatusBadRequest, "Metrics cannot be empty")
		return
	}

	slog.Info("Request validated", "metrics_length", len(req.Metrics))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), apiTimeout)
	defer cancel()

	// Generate dashboard using AI client
	dashboardResp, err := h.aiClient.GenerateDashboard(ctx, req.Metrics)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			slog.Error("API request timeout", "error", err)
			respondWithError(w, http.StatusGatewayTimeout, "Request timeout - please try again")
			return
		}
		slog.Error("AI generation failed", "error", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to generate dashboard")
		return
	}

	slog.Info("Successfully generated dashboard", 
		"dashboard_size", len(dashboardResp.GrafanaDashboard), 
		"alerts_size", len(dashboardResp.PrometheusAlerts))

	// Send the result back
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(dashboardResp); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}

	totalDuration := time.Since(start)
	slog.Info("Request completed successfully", "total_duration", totalDuration)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	slog.Debug("Sending error response", "code", code, "message", message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(models.ErrorResponse{Error: message}); err != nil {
		slog.Error("Failed to encode error response", "error", err)
	}
}