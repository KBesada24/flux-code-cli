package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Provider != "ollama" {
		t.Errorf("Expected default provider 'ollama', got '%s'", cfg.Provider)
	}

	if cfg.UI.Theme != "dark" {
		t.Errorf("Expected default theme 'dark', got '%s'", cfg.UI.Theme)
	}

	if cfg.UI.WordWrap != 80 {
		t.Errorf("Expected default word_wrap 80, got %d", cfg.UI.WordWrap)
	}

	if !cfg.UI.ShowTokens {
		t.Error("Expected show_tokens to be true by default")
	}

	if !cfg.UI.SyntaxHighlighting {
		t.Error("Expected syntax_highlighting to be true by default")
	}
}

func TestGet(t *testing.T) {
	_, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	cfg := Get()
	if cfg == nil {
		t.Error("Get() should return config after Load()")
	}
}

func TestEnvExpansion(t *testing.T) {
	// Set up test environment variable
	os.Setenv("TEST_API_KEY", "test-key-123")
	defer os.Unsetenv("TEST_API_KEY")

	// Manually test expansion
	expanded := os.ExpandEnv("${TEST_API_KEY}")
	if expanded != "test-key-123" {
		t.Errorf("Expected 'test-key-123', got '%s'", expanded)
	}
}
