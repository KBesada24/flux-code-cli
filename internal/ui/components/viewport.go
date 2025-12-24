package components

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type Viewport struct {
	viewport viewport.Model
	ready    bool
}

func NewViewport(width, height int) Viewport {
	vp := viewport.New(width, height)
	vp.YPosition = 0

	return Viewport{
		viewport: vp,
		ready:    true,
	}
}

func (v Viewport) Update(msg tea.Msg) (Viewport, tea.Cmd) {
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

func (v Viewport) View() string {
	return v.viewport.View()
}

func (v *Viewport) SetContent(content string) {
	v.viewport.SetContent(content)
}

func (v *Viewport) SetSize(width, height int) {
	v.viewport.Width = width
	v.viewport.Height = height
}

func (v *Viewport) GotoBottom() {
	v.viewport.GotoBottom()
}

func (v Viewport) ScrollPercent() float64 {
	return v.viewport.ScrollPercent()
}

func (v Viewport) Ready() bool {
	return v.ready
}
