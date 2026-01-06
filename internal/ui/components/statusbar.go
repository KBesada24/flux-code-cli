package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kbesada/flux-code-cli/internal/git"
)

type StatusBar struct {
	width     int
	gitStatus string
	model     string
	provider  string
}

func NewStatusBar() StatusBar {
	return StatusBar{}
}

func (s *StatusBar) Update() {
	// Update git status
	if repo, err := git.Open(""); err == nil {
		if status, err := repo.GetStatus(); err == nil {
			s.gitStatus = status.FormatForStatusBar()
		}
	}
}

func (s StatusBar) View() string {
	leftStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	gitStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D4AA")).
		Bold(true)

	modelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4"))

	left := leftStyle.Render("Ctrl+C quit • Enter send • /help commands")

	var right string
	if s.gitStatus != "" {
		right = gitStyle.Render(" "+s.gitStatus) + " │ "
	}
	right += modelStyle.Render(s.model)

	// Calculate padding
	padding := s.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if padding < 0 {
		padding = 0
	}

	return left + strings.Repeat(" ", padding) + right
}

func (s *StatusBar) SetWidth(w int) {
	s.width = w
}

func (s *StatusBar) SetModel(provider, model string) {
	s.provider = provider
	s.model = fmt.Sprintf("%s/%s", provider, model)
}
