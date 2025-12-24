package ui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	PrimaryColor   = lipgloss.Color("#7D56F4") // Purple
	SecondaryColor = lipgloss.Color("#00D4AA") // Teal
	ErrorColor     = lipgloss.Color("#FF6B6B") // Red
	WarningColor   = lipgloss.Color("#FFB86C") // Orange
	MutedColor     = lipgloss.Color("#626262") // Gray
	TextColor      = lipgloss.Color("#FAFAFA") // White
	BgColor        = lipgloss.Color("#1E1E1E") // Dark
)

// Component styles
var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			Padding(0, 1)

	MessageAreaStyle = lipgloss.NewStyle().
				Padding(1, 2)

	InputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(0, 1)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Padding(0, 1)

	ExitPromptStyle = lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true)

	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor)
)
