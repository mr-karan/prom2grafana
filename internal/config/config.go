package config

import (
	"log"
	"log/slog"
	"os"

	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

// Config holds all application configuration
type Config struct {
	k *koanf.Koanf
}

// New creates a new configuration instance
func New() *Config {
	k := koanf.New(".")
	
	// Load environment variables
	if err := k.Load(env.Provider("", env.Opt{}), nil); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	return &Config{k: k}
}

// Port returns the server port
func (c *Config) Port() string {
	port := c.k.String("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

// LogLevel returns the configured log level
func (c *Config) LogLevel() slog.Level {
	logLevel := slog.LevelInfo
	if c.k.String("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	return logLevel
}

// OpenAIAPIKey returns the OpenAI/OpenRouter API key
func (c *Config) OpenAIAPIKey() string {
	return c.k.String("OPENAI_API_KEY")
}

// OpenAIAPIURL returns the API URL
func (c *Config) OpenAIAPIURL() string {
	apiURL := c.k.String("OPENAI_API_URL")
	if apiURL == "" {
		apiURL = "https://openrouter.ai/api/v1"
	}
	return apiURL
}

// OpenAIModel returns the single model configuration
func (c *Config) OpenAIModel() string {
	model := c.k.String("OPENAI_MODEL")
	if model == "" {
		model = "google/gemini-2.5-flash"
	}
	return model
}

// OpenAIModels returns the multiple models configuration
func (c *Config) OpenAIModels() string {
	return c.k.String("OPENAI_MODELS")
}

// SetupLogging configures structured logging
func (c *Config) SetupLogging() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: c.LogLevel(),
	}))
	slog.SetDefault(logger)
}