package components

import (
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

type Message struct {
	Role      Role
	Content   string
	Timestamp time.Time
}

type Messages struct {
	items    []Message
	renderer *glamour.TermRenderer
	width    int
}

func NewMessages(width int) Messages {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)

	return Messages{
		items:    []Message{},
		renderer: r,
		width:    width,
	}
}

func (m *Messages) Add(role Role, content string) {
	m.items = append(m.items, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

func (m *Messages) Clear() {
	m.items = []Message{}
}

func (m *Messages) Count() int {
	return len(m.items)
}

func (m *Messages) Items() []Message {
	return m.items
}

func (m Messages) Render() string {
	var output strings.Builder

	for _, msg := range m.items {
		switch msg.Role {
		case RoleUser:
			output.WriteString(m.renderUserMessage(msg))
		case RoleAssistant:
			output.WriteString(m.renderAssistantMessage(msg))
		case RoleSystem:
			output.WriteString(m.renderSystemMessage(msg))
		}
		output.WriteString("\n")
	}

	return output.String()
}

func (m Messages) renderUserMessage(msg Message) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		PaddingLeft(2)

	header := headerStyle.Render("You")
	content := contentStyle.Render(msg.Content)

	return header + "\n" + content + "\n"
}

func (m Messages) renderAssistantMessage(msg Message) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D4AA"))

	header := headerStyle.Render("Assistant")

	// Render markdown
	rendered, err := m.renderer.Render(msg.Content)
	if err != nil {
		rendered = msg.Content
	}
	// Trim extra newlines from glamour
	rendered = strings.TrimSpace(rendered)

	return header + "\n" + rendered + "\n"
}

func (m Messages) renderSystemMessage(msg Message) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true).
		PaddingLeft(2)

	return style.Render(msg.Content) + "\n"
}

func (m *Messages) SetWidth(w int) {
	m.width = w
	m.renderer, _ = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(w),
	)
}
