package models

// ConvertRequest represents the request to convert metrics
type ConvertRequest struct {
	Metrics string `json:"metrics"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// DashboardResponse represents the structured output from AI
type DashboardResponse struct {
	GrafanaDashboard string `json:"grafana_dashboard" jsonschema:"description=Complete Grafana dashboard JSON as a string"`
	PrometheusAlerts string `json:"prometheus_alerts" jsonschema:"description=Prometheus alerts in YAML format"`
}