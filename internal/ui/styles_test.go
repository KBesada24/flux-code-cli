package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorsDefined(t *testing.T) {
	colors := map[string]lipgloss.Color{
		"PrimaryColor":   PrimaryColor,
		"SecondaryColor": SecondaryColor,
		"ErrorColor":     ErrorColor,
		"WarningColor":   WarningColor,
		"MutedColor":     MutedColor,
		"TextColor":      TextColor,
		"BgColor":        BgColor,
	}

	for name, color := range colors {
		if string(color) == "" {
			t.Errorf("%s should not be empty", name)
		}
	}
}

func TestStylesRender(t *testing.T) {
	styles := map[string]lipgloss.Style{
		"HeaderStyle":      HeaderStyle,
		"MessageAreaStyle": MessageAreaStyle,
		"InputStyle":       InputStyle,
		"StatusBarStyle":   StatusBarStyle,
		"LogoStyle":        LogoStyle,
	}

	for name, style := range styles {
		// This should not panic
		result := style.Render("test")
		if result == "" {
			t.Errorf("%s.Render should not return empty string", name)
		}
	}
}
