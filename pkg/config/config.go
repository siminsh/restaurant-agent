package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AnthropicAPIKey    string
	Port               string
	ClaudeModel        string
	MaxAgentIterations int
}

func Load() (*Config, error) {
	// Load .env file if present (ignore error if missing)
	godotenv.Load()

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-sonnet-4-5-20250929"
	}

	maxIter := 15
	if v := os.Getenv("MAX_AGENT_ITERATIONS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_AGENT_ITERATIONS: %w", err)
		}
		if n <= 0 {
			return nil, fmt.Errorf("MAX_AGENT_ITERATIONS must be positive, got %d", n)
		}
		maxIter = n
	}

	return &Config{
		AnthropicAPIKey:    apiKey,
		Port:               port,
		ClaudeModel:        model,
		MaxAgentIterations: maxIter,
	}, nil
}
