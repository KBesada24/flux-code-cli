package ai

import (
	"context"
)

// Client is the interface for AI providers
type Client interface {
	// Stream sends a prompt and returns a channel of streaming events
	Stream(ctx context.Context, messages []Message) <-chan StreamEvent

	// Complete sends a prompt and returns the full response
	Complete(ctx context.Context, messages []Message) (string, error)

	// GetModel returns the current model name
	GetModel() string

	// SetModel changes the active model
	SetModel(model string)
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	Name    string
	APIKey  string
	BaseURL string
	Model   string
}

// NewClient creates a new AI client based on provider config
func NewClient(cfg ProviderConfig) (Client, error) {
	// All supported providers use OpenAI-compatible API
	return NewOpenAIClient(cfg)
}
