package ui

import (
	"time"

	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

const exitPromptTimeout = 2 * time.Second

type clearExitPromptMsg struct{}

type Model struct {
	width          int
	height         int
	ready          bool
	quitting       bool
	lastCtrlC      time.Time
	showExitPrompt bool
}

func NewModel() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			now := time.Now()
			if m.showExitPrompt && now.Sub(m.lastCtrlC) < exitPromptTimeout {
				m.quitting = true
				return m, tea.Quit
			}
			m.lastCtrlC = now
			m.showExitPrompt = true
			return m, tea.Tick(exitPromptTimeout, func(t time.Time) tea.Msg {
				return clearExitPromptMsg{}
			})
		case "esc", "q":
			// Reset exit prompt on other keys
			m.showExitPrompt = false
		}
	case clearExitPromptMsg:
		m.showExitPrompt = false
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}
	if !m.ready {
		return "Initializing..."
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		m.renderMessages(),
		m.renderInput(),
		m.renderStatusBar(),
	)
}

func (m Model) renderHeader() string {
	title := LogoStyle.Render("flux") + "  AI Coding Assistant"
	return HeaderStyle.Width(m.width).Render(title)
}

func (m Model) renderMessages() string {
	// Calculate available height for messages
	// Header: 1 line, Input: 3 lines (with border), Status: 1 line
	msgHeight := m.height - 5
	if msgHeight < 1 {
		msgHeight = 1
	}

	placeholder := "No messages yet. Type something to begin..."
	return MessageAreaStyle.
		Width(m.width).
		Height(msgHeight).
		Render(placeholder)
}

func (m Model) renderInput() string {
	prompt := "> Type your message..."
	return InputStyle.Width(m.width - 2).Render(prompt)
}

func (m Model) renderStatusBar() string {
	var status string
	if m.showExitPrompt {
		status = ExitPromptStyle.Render("Press Ctrl+C again to exit")
	} else {
		status = "Ctrl+C to exit"
	}
	return StatusBarStyle.Width(m.width).Render(status)
}
