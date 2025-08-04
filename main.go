package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// Build variables set by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

//go:embed index.html static/*
var content embed.FS

//go:embed grafana_dashboard_prompt.md
var grafanaDashboardPrompt string

// Global koanf instance
var k = koanf.New(".")

type ConvertRequest struct {
	Metrics string `json:"metrics"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// DashboardResponse represents the structured output from AI
type DashboardResponse struct {
	GrafanaDashboard string `json:"grafana_dashboard" jsonschema:"description=Complete Grafana dashboard JSON as a string"`
	PrometheusAlerts string `json:"prometheus_alerts" jsonschema:"description=Prometheus alerts in YAML format"`
}

var systemPrompt = `You are an expert Site-Reliability Engineer who specializes in Grafana dashboards and Prometheus-based alerting.

` + grafanaDashboardPrompt + `

The user will paste a block of Prometheus metric samples.

You must generate a JSON response with two fields:

1. "grafana_dashboard" - A complete Grafana dashboard JSON as a STRING (not an object). The dashboard must be a valid JSON string that can be imported into Grafana based on the comprehensive documentation above.

2. "prometheus_alerts" - Prometheus alerts in YAML format as a STRING. Include:
   • Meaningful alert rules based on the metrics
   • Annotations with summary and description
   • Labels with severity: warning or critical
   • Use appropriate thresholds based on metric types

IMPORTANT: The grafana_dashboard field must contain the ENTIRE dashboard JSON as a string, not just a number or placeholder.`

const (
	// Maximum request body size (1MB)
	maxRequestBodySize = 1 << 20 // 1MB
	// API timeout duration
	apiTimeout = 30 * time.Second
)

func main() {
	// Handle version flag
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.Parse()

	if showVersion {
		fmt.Printf("prom2grafana %s (commit: %s, built at: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Load environment variables first
	if err := k.Load(env.Provider("", env.Opt{}), nil); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	// Setup structured logging based on LOG_LEVEL env var
	logLevel := slog.LevelInfo
	if k.String("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Get port with default
	port := k.String("PORT")
	if port == "" {
		port = "8080"
	}
	slog.Info("Starting prom2grafana", "version", version, "port", port)

	// Create a sub filesystem for static files
	staticFS, err := fs.Sub(content, "static")
	if err != nil {
		log.Fatal(err)
	}

	// Serve static files from embedded filesystem
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Serve index.html from embedded filesystem
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := content.ReadFile("index.html")
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(data); err != nil {
			slog.Error("Failed to write response", "error", err)
		}
	})

	// API endpoint
	http.HandleFunc("/convert", handleConvert)

	slog.Info("Server starting", "port", port, "url", fmt.Sprintf("http://localhost:%s", port))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
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

	var req ConvertRequest
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

	// Get API configuration
	apiKey := k.String("OPENAI_API_KEY")
	if apiKey == "" {
		slog.Error("OpenAI API key not configured")
		respondWithError(w, http.StatusInternalServerError, "OpenAI API key not configured")
		return
	}

	// Get configurable API URL and model
	apiURL := k.String("OPENAI_API_URL")
	if apiURL == "" {
		apiURL = "https://openrouter.ai/api/v1" // Default to OpenRouter
	}

	model := k.String("OPENAI_MODEL")
	if model == "" {
		model = "google/gemini-2.5-flash" // Default model
	}

	// Create client with custom config
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = apiURL
	client := openai.NewClientWithConfig(config)
	slog.Debug("API client created", "base_url", config.BaseURL, "model", model)

	// Generate JSON schema for structured output
	var dashboardResp DashboardResponse
	schema, err := jsonschema.GenerateSchemaForType(dashboardResp)
	if err != nil {
		slog.Error("Failed to generate JSON schema", "error", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to prepare request")
		return
	}

	// Prepare the request with structured output
	chatReq := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: req.Metrics,
			},
		},
		Temperature: 0.1,
		MaxTokens:   65536, // 64k tokens for larger dashboards
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "dashboard_generator",
				Schema: schema,
				Strict: false, // Allow some flexibility
			},
		},
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), apiTimeout)
	defer cancel()

	// Make the API call
	slog.Info("Calling OpenRouter API", "model", chatReq.Model, "max_tokens", chatReq.MaxTokens)
	apiStart := time.Now()
	resp, err := client.CreateChatCompletion(ctx, chatReq)
	apiDuration := time.Since(apiStart)
	slog.Info("API call completed", "duration", apiDuration)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			slog.Error("API request timeout", "error", err)
			respondWithError(w, http.StatusGatewayTimeout, "Request timeout - please try again")
			return
		}
		slog.Error("OpenRouter API error", "error", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to generate dashboard")
		return
	}

	if len(resp.Choices) == 0 {
		slog.Error("No choices in API response")
		respondWithError(w, http.StatusInternalServerError, "No response from AI")
		return
	}

	// Parse the structured response
	aiResponse := resp.Choices[0].Message.Content
	slog.Debug("AI response received", "response_length", len(aiResponse))

	// Unmarshal the structured response
	err = schema.Unmarshal(aiResponse, &dashboardResp)
	if err != nil {
		slog.Error("Failed to unmarshal structured response", "error", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to parse AI response")
		return
	}

	// Validate that we got a proper dashboard
	if len(dashboardResp.GrafanaDashboard) == 0 {
		slog.Error("Empty Grafana dashboard in response")
		respondWithError(w, http.StatusInternalServerError, "Invalid dashboard response")
		return
	}

	slog.Info("Successfully generated dashboard", "dashboard_size", len(dashboardResp.GrafanaDashboard), "alerts_size", len(dashboardResp.PrometheusAlerts))

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
	if err := json.NewEncoder(w).Encode(ErrorResponse{Error: message}); err != nil {
		slog.Error("Failed to encode error response", "error", err)
	}
}
