package components

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type Input struct {
	textarea textarea.Model
	focused  bool
}

func NewInput() Input {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.Focus()
	ta.CharLimit = 4000
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	return Input{
		textarea: ta,
		focused:  true,
	}
}

func (i Input) Update(msg tea.Msg) (Input, tea.Cmd) {
	var cmd tea.Cmd
	i.textarea, cmd = i.textarea.Update(msg)
	return i, cmd
}

func (i Input) View() string {
	return i.textarea.View()
}

func (i Input) Value() string {
	return i.textarea.Value()
}

func (i *Input) Reset() {
	i.textarea.Reset()
}

func (i *Input) SetWidth(w int) {
	i.textarea.SetWidth(w)
}

func (i *Input) Focus() tea.Cmd {
	return i.textarea.Focus()
}

func (i *Input) Blur() {
	i.textarea.Blur()
}

func (i Input) Focused() bool {
	return i.focused
}
