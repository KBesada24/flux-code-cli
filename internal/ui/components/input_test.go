package components

import (
	"testing"
)

func TestNewInput(t *testing.T) {
	input := NewInput()

	if input.Value() != "" {
		t.Error("New input should have empty value")
	}

	if !input.Focused() {
		t.Error("New input should be focused")
	}
}

func TestInputReset(t *testing.T) {
	input := NewInput()

	// Simulate typing by directly setting (can't easily simulate keystrokes)
	// Just verify Reset doesn't panic
	input.Reset()

	if input.Value() != "" {
		t.Error("After reset, value should be empty")
	}
}

func TestInputSetWidth(t *testing.T) {
	input := NewInput()

	// Should not panic
	input.SetWidth(100)
}

func TestInputBlur(t *testing.T) {
	input := NewInput()

	// Should not panic
	input.Blur()
}
