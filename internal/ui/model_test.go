package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	m := NewModel()

	if m.ready {
		t.Error("NewModel should not be ready initially")
	}
	if m.quitting {
		t.Error("NewModel should not be quitting initially")
	}
	if m.width != 0 {
		t.Errorf("NewModel width should be 0, got %d", m.width)
	}
	if m.height != 0 {
		t.Errorf("NewModel height should be 0, got %d", m.height)
	}
}

func TestModelInit(t *testing.T) {
	m := NewModel()
	cmd := m.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestModelUpdateQuitKeys(t *testing.T) {
	// Test that single Ctrl+C shows exit prompt
	t.Run("single_ctrl+c_shows_prompt", func(t *testing.T) {
		m := NewModel()
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}

		newModel, cmd := m.Update(msg)
		model := newModel.(Model)

		if model.quitting {
			t.Error("Single Ctrl+C should not quit")
		}
		if !model.showExitPrompt {
			t.Error("Single Ctrl+C should show exit prompt")
		}
		if cmd == nil {
			t.Error("Should return a tick command for timeout")
		}
	})

	// Test that double Ctrl+C quits
	t.Run("double_ctrl+c_quits", func(t *testing.T) {
		m := NewModel()
		m.showExitPrompt = true
		m.lastCtrlC = time.Now()

		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		newModel, cmd := m.Update(msg)
		model := newModel.(Model)

		if !model.quitting {
			t.Error("Double Ctrl+C should quit")
		}
		if cmd == nil {
			t.Error("Double Ctrl+C should return tea.Quit command")
		}
	})

	// Test that esc/q reset exit prompt
	t.Run("esc_resets_prompt", func(t *testing.T) {
		m := NewModel()
		m.showExitPrompt = true

		msg := tea.KeyMsg{Type: tea.KeyEsc}
		newModel, _ := m.Update(msg)
		model := newModel.(Model)

		if model.showExitPrompt {
			t.Error("Esc should reset exit prompt")
		}
	})
}

func TestModelUpdateWindowResize(t *testing.T) {
	m := NewModel()
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}

	newModel, _ := m.Update(msg)
	model := newModel.(Model)

	if model.width != 100 {
		t.Errorf("Expected width 100, got %d", model.width)
	}
	if model.height != 50 {
		t.Errorf("Expected height 50, got %d", model.height)
	}
	if !model.ready {
		t.Error("Model should be ready after WindowSizeMsg")
	}
}

func TestModelViewNotReady(t *testing.T) {
	m := NewModel()
	view := m.View()

	if view != "Initializing..." {
		t.Errorf("Expected 'Initializing...', got '%s'", view)
	}
}

func TestModelViewQuitting(t *testing.T) {
	m := NewModel()
	m.quitting = true
	view := m.View()

	if view != "Goodbye!\n" {
		t.Errorf("Expected 'Goodbye!\\n', got '%s'", view)
	}
}

func TestModelViewReady(t *testing.T) {
	m := NewModel()
	m.ready = true
	m.width = 80
	m.height = 24

	view := m.View()

	// Check that view contains expected elements
	if !strings.Contains(view, "flux") {
		t.Error("View should contain 'flux' logo")
	}
	if !strings.Contains(view, "Ctrl+C") {
		t.Error("View should contain Ctrl+C instructions")
	}
}
