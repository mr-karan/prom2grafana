package ai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/mr-karan/prom2grafana/internal/config"
	m "github.com/mr-karan/prom2grafana/internal/models"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// Client handles AI interactions
type Client struct {
	client       *openai.Client
	config       *config.Config
	systemPrompt string
}

// New creates a new AI client
func New(cfg *config.Config, systemPrompt string) (*Client, error) {
	apiKey := cfg.OpenAIAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	// Create client with custom config
	openaiConfig := openai.DefaultConfig(apiKey)
	openaiConfig.BaseURL = cfg.OpenAIAPIURL()
	client := openai.NewClientWithConfig(openaiConfig)

	return &Client{
		client:       client,
		config:       cfg,
		systemPrompt: systemPrompt,
	}, nil
}

// getModelsToTry returns the list of models to try in order
func (c *Client) getModelsToTry() []string {
	// Check for multiple models configuration
	modelsStr := c.config.OpenAIModels()
	if modelsStr != "" {
		models := strings.Split(modelsStr, ",")
		for i, model := range models {
			models[i] = strings.TrimSpace(model)
		}
		return models
	}

	// Fallback to single model configuration (backward compatibility)
	return []string{c.config.OpenAIModel()}
}

// GenerateDashboard generates a dashboard using AI with model fallback
func (c *Client) GenerateDashboard(ctx context.Context, metrics string) (*m.DashboardResponse, error) {
	models := c.getModelsToTry()
	slog.Info("Models configured for fallback", "models", models)

	return c.tryModelsWithFallback(ctx, models, metrics)
}

// tryModelsWithFallback attempts to generate dashboard using multiple models with fallback
func (c *Client) tryModelsWithFallback(ctx context.Context, models []string, metrics string) (*m.DashboardResponse, error) {
	// Generate JSON schema for structured output
	var dashboardResp m.DashboardResponse
	schema, err := jsonschema.GenerateSchemaForType(dashboardResp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JSON schema: %w", err)
	}

	var lastErr error
	var attemptedModels []string

	for _, model := range models {
		attemptedModels = append(attemptedModels, model)

		// Prepare the request with structured output
		chatReq := openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: c.systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: metrics,
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

		// Make the API call
		slog.Info("Trying model", "model", model, "max_tokens", chatReq.MaxTokens)
		apiStart := time.Now()
		resp, err := c.client.CreateChatCompletion(ctx, chatReq)
		apiDuration := time.Since(apiStart)

		if err != nil {
			slog.Warn("Model failed", "model", model, "error", err, "duration", apiDuration)
			lastErr = err
			continue
		}

		if len(resp.Choices) == 0 {
			slog.Warn("No choices in response", "model", model)
			lastErr = fmt.Errorf("no response from model %s", model)
			continue
		}

		// Parse the structured response
		aiResponse := resp.Choices[0].Message.Content
		slog.Debug("AI response received", "model", model, "response_length", len(aiResponse))

		// Unmarshal the structured response
		err = schema.Unmarshal(aiResponse, &dashboardResp)
		if err != nil {
			slog.Warn("Failed to parse response", "model", model, "error", err)
			lastErr = fmt.Errorf("failed to parse response from %s: %w", model, err)
			continue
		}

		// Validate that we got a proper dashboard
		if len(dashboardResp.GrafanaDashboard) == 0 {
			slog.Warn("Empty dashboard response", "model", model)
			lastErr = fmt.Errorf("empty dashboard from model %s", model)
			continue
		}

		// Success!
		slog.Info("Model succeeded", "model", model, "duration", apiDuration)
		return &dashboardResp, nil
	}

	// All models failed
	return nil, fmt.Errorf("all models failed (tried: %s), last error: %w",
		strings.Join(attemptedModels, ", "), lastErr)
}
