package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kbesada/flux-code-cli/internal/ui"
)

func Run() error {
	model := ui.NewModel()
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
