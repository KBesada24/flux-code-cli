package config

import (
	"os"

	"github.com/spf13/viper"
)

var cfg *Config

func Load() (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("provider", "ollama")
	v.SetDefault("ui.theme", "dark")
	v.SetDefault("ui.word_wrap", 80)
	v.SetDefault("ui.show_tokens", true)
	v.SetDefault("ui.syntax_highlighting", true)
	v.SetDefault("system.system_prompt", "You are a helpful AI coding assistant.")

	// Config paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("$HOME/.config/flux")
	v.AddConfigPath(".")

	// Environment variables
	v.SetEnvPrefix("FLUX")
	v.AutomaticEnv()

	// Read config
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Unmarshal
	cfg = &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// Expand environment variables in API keys
	for name, provider := range cfg.Providers {
		provider.APIKey = os.ExpandEnv(provider.APIKey)
		cfg.Providers[name] = provider
	}

	return cfg, nil
}

func Get() *Config {
	return cfg
}
