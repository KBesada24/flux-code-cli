package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kbesada/flux-code-cli/internal/config"
	"github.com/kbesada/flux-code-cli/internal/ui"
)

func Run() error {
	// Load configuration (errors are non-fatal, uses defaults)
	_, _ = config.Load()

	model := ui.NewModel()
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
