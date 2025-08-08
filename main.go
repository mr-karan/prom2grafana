package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/mr-karan/prom2grafana/internal/ai"
	"github.com/mr-karan/prom2grafana/internal/config"
	"github.com/mr-karan/prom2grafana/internal/handlers"
	"github.com/mr-karan/prom2grafana/internal/server"
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

func main() {
	// Handle version flag
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.Parse()

	if showVersion {
		fmt.Printf("prom2grafana %s (commit: %s, built at: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Load configuration
	cfg := config.New()
	cfg.SetupLogging()

	slog.Info("Starting prom2grafana", "version", version, "port", cfg.Port())

	// Initialize AI client
	aiClient, err := ai.New(cfg, systemPrompt)
	if err != nil {
		log.Fatalf("Failed to initialize AI client: %v", err)
	}

	// Initialize handlers
	convertHandler := handlers.NewConvertHandler(aiClient)

	// Initialize and start server
	srv := server.New(cfg.Port(), content, convertHandler)
	if err := srv.Start(); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
