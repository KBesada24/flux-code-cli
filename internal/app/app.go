package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kbesada/flux-code-cli/internal/ai"
	"github.com/kbesada/flux-code-cli/internal/config"
	"github.com/kbesada/flux-code-cli/internal/ui"
)

func Run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get provider config
	provider, ok := cfg.Providers[cfg.Provider]
	if !ok {
		// Fallback to default if configured provider not found
		cfg.Provider = "ollama"
		var exists bool
		provider, exists = cfg.Providers["ollama"]
		if !exists {
			// If even fallback is missing (rare), set reasonable defaults
			provider = config.Provider{
				BaseURL: "http://localhost:11434/v1",
				Model:   "codellama",
			}
		}
	}

	// Create AI client
	client, err := ai.NewClient(ai.ProviderConfig{
		Name:    cfg.Provider,
		APIKey:  provider.APIKey,
		BaseURL: provider.BaseURL,
		Model:   provider.Model,
	})
	if err != nil {
		return fmt.Errorf("failed to create AI client: %w", err)
	}

	model := ui.NewModel(client)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
