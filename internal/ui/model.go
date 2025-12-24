package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kbesada/flux-code-cli/internal/ui/components"
)

const exitPromptTimeout = 2 * time.Second

type clearExitPromptMsg struct{}

type Model struct {
	// Components
	input    components.Input
	viewport components.Viewport
	messages components.Messages

	// State
	width          int
	height         int
	ready          bool
	quitting       bool
	lastCtrlC      time.Time
	showExitPrompt bool
}

func NewModel() Model {
	return Model{
		input:    components.NewInput(),
		messages: components.NewMessages(80),
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

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
		case "enter":
			// Send message if input has content
			if value := m.input.Value(); value != "" {
				m.messages.Add(components.RoleUser, value)
				m.input.Reset()
				m.viewport.SetContent(m.messages.Render())
				m.viewport.GotoBottom()
				// Reset exit prompt on activity
				m.showExitPrompt = false
			}
			return m, nil
		default:
			// Reset exit prompt on other keys
			m.showExitPrompt = false
		}
	case clearExitPromptMsg:
		m.showExitPrompt = false
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.handleResize()
		m.ready = true
	}

	// Update input component
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	// Update viewport component
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
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
		m.viewport.View(),
		m.input.View(),
		m.renderStatusBar(),
	)
}

func (m *Model) handleResize() {
	headerHeight := 1
	statusHeight := 1
	inputHeight := 5

	viewportHeight := m.height - headerHeight - statusHeight - inputHeight
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	m.viewport.SetSize(m.width, viewportHeight)
	m.input.SetWidth(m.width - 4)
	m.messages.SetWidth(m.width - 4)
	m.viewport.SetContent(m.messages.Render())
}

func (m Model) renderHeader() string {
	title := LogoStyle.Render("flux") + "  AI Coding Assistant"
	return HeaderStyle.Width(m.width).Render(title)
}

func (m Model) renderStatusBar() string {
	var status string
	if m.showExitPrompt {
		status = ExitPromptStyle.Render("Press Ctrl+C again to exit")
	} else {
		status = "Ctrl+C to exit • Enter to send"
	}
	return StatusBarStyle.Width(m.width).Render(status)
}
