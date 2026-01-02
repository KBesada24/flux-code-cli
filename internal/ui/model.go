package ui

import (
	"time"

	"context"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kbesada/flux-code-cli/internal/ai"
	"github.com/kbesada/flux-code-cli/internal/ui/components"
)

const exitPromptTimeout = 2 * time.Second

type clearExitPromptMsg struct{}
type streamChunkMsg string
type streamDoneMsg struct{}
type streamErrMsg struct{ err error }

type Model struct {
	// Components
	input    components.Input
	viewport components.Viewport
	messages components.Messages

	// AI
	aiClient     ai.Client
	streaming    bool
	streamBuf    string
	streamEvents <-chan ai.StreamEvent
	cancelFn     context.CancelFunc
	err          error

	// State
	width          int
	height         int
	ready          bool
	quitting       bool
	lastCtrlC      time.Time
	showExitPrompt bool
}

func NewModel(client ai.Client) Model {
	return Model{
		input:    components.NewInput(),
		messages: components.NewMessages(80),
		aiClient: client,
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
			if m.streaming {
				// Cancel current stream
				if m.cancelFn != nil {
					m.cancelFn()
				}
				m.streaming = false
				return m, nil
			}

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
			if !m.streaming && m.input.Value() != "" {
				return m, m.sendMessage()
			}
			return m, nil
		default:
			// Reset exit prompt on other keys
			m.showExitPrompt = false
		}
	case clearExitPromptMsg:
		m.showExitPrompt = false

	case streamChunkMsg:
		m.streamBuf += string(msg)
		m.updateStreamingMessage()
		return m, m.waitForNextChunk()

	case streamDoneMsg:
		m.finalizeStream()
		if m.cancelFn != nil {
			m.cancelFn()
		}
		return m, nil

	case streamErrMsg:
		m.err = msg.err
		m.streaming = false
		if m.cancelFn != nil {
			m.cancelFn()
		}
		m.messages.Add(components.RoleAssistant,
			"Error: "+msg.err.Error())
		m.viewport.SetContent(m.messages.Render())
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.handleResize()
		m.ready = true
	}

	// Update input component (only if not streaming)
	if !m.streaming {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update viewport component
	var cmd tea.Cmd
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

func (m *Model) sendMessage() tea.Cmd {
	userMsg := m.input.Value()
	m.input.Reset()

	// Add user message
	m.messages.Add(components.RoleUser, userMsg)
	m.viewport.SetContent(m.messages.Render())
	m.viewport.GotoBottom()

	// Start streaming
	m.streaming = true
	m.streamBuf = ""

	// Set up cancellation context synchronously
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFn = cancel

	// Build history in main thread
	history := m.buildMessageHistory()

	// Start stream and store channel
	m.streamEvents = m.aiClient.Stream(ctx, history)

	return m.waitForNextChunk()
}

func (m *Model) waitForNextChunk() tea.Cmd {
	return func() tea.Msg {
		event, ok := <-m.streamEvents
		if !ok {
			return streamDoneMsg{}
		}

		if event.Error != nil {
			return streamErrMsg{err: event.Error}
		}
		if event.Done {
			return streamDoneMsg{}
		}
		if event.Content != "" {
			return streamChunkMsg(event.Content)
		}
		return nil // Continue waiting if empty content but not done
	}
}

func (m *Model) buildMessageHistory() []ai.Message {
	var history []ai.Message

	// Add system prompt if configured
	// history = append(history, ai.Message{Role: "system", Content: systemPrompt})

	for _, msg := range m.messages.Items() {
		history = append(history, ai.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	return history
}

func (m *Model) updateStreamingMessage() {
	// Render messages + current streaming buffer
	content := m.messages.Render()
	content += "\n" + m.renderStreamingIndicator() + m.streamBuf
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m *Model) finalizeStream() {
	m.messages.Add(components.RoleAssistant, m.streamBuf)
	m.streamBuf = ""
	m.streaming = false
	m.viewport.SetContent(m.messages.Render())
	m.viewport.GotoBottom()
}

func (m Model) renderStreamingIndicator() string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D4AA")).
		Render("Assistant") + " (streaming...)\n"
}
