package ai

import (
	"fmt"
	"net/http"

	"github.com/kbesada/flux-code-cli/internal/config"
)

// Registry builds AI clients based on config.
type Registry struct {
	constructors map[string]func(cfg config.Provider, httpClient *http.Client) (Client, error)
}

// NewRegistry creates a registry with default constructors.
func NewRegistry() *Registry {
	return &Registry{
		constructors: map[string]func(cfg config.Provider, httpClient *http.Client) (Client, error){
			"custom": func(p config.Provider, hc *http.Client) (Client, error) {
				return NewStandardClient(StandardClientConfig{
					BaseURL:    p.BaseURL,
					APIKey:     p.APIKey,
					Model:      p.Model,
					Provider:   "custom",
					HTTPClient: hc,
				})
			},
			"openai": func(p config.Provider, hc *http.Client) (Client, error) {
				return NewStandardClient(StandardClientConfig{
					BaseURL:    p.BaseURL,
					APIKey:     p.APIKey,
					Model:      p.Model,
					Provider:   "openai",
					HTTPClient: hc,
				})
			},
			"ollama": func(p config.Provider, hc *http.Client) (Client, error) {
				// Ollama defaults
				baseURL := p.BaseURL
				if baseURL == "" {
					baseURL = "http://localhost:11434/v1"
				}
				return NewStandardClient(StandardClientConfig{
					BaseURL:    baseURL,
					APIKey:     p.APIKey, // Often empty for local Ollama
					Model:      p.Model,
					Provider:   "ollama",
					HTTPClient: hc,
				})
			},
			"openrouter": func(p config.Provider, hc *http.Client) (Client, error) {
				return NewStandardClient(StandardClientConfig{
					BaseURL:    "https://openrouter.ai/api/v1",
					APIKey:     p.APIKey,
					Model:      p.Model,
					Provider:   "openrouter",
					HTTPClient: hc,
				})
			},
		},
	}
}

// Register adds/overrides a constructor for a provider key.
func (r *Registry) Register(name string, ctor func(cfg config.Provider, httpClient *http.Client) (Client, error)) {
	r.constructors[name] = ctor
}

// Build creates a client for the given provider name using config and optional http.Client.
func (r *Registry) Build(providerName string, cfg *config.Config, hc *http.Client) (Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	provCfg, ok := cfg.Providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider %q not found in config", providerName)
	}

	ctor, ok := r.constructors[providerName]
	if !ok {
		// fallback to custom if defined
		if fallback, ok := r.constructors["custom"]; ok {
			return fallback(provCfg, hc)
		}
		return nil, fmt.Errorf("provider %q has no constructor", providerName)
	}

	return ctor(provCfg, hc)
}
