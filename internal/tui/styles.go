package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	PrimaryColor   = lipgloss.Color("#7D56F4")
	SecondaryColor = lipgloss.Color("#04B575")
	SuccessColor   = lipgloss.Color("#23D18B")
	ErrorColor     = lipgloss.Color("#ED567A")
	BgColor        = lipgloss.Color("#1A1B26")

	// Styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(PrimaryColor).
			Padding(0, 1).
			Bold(true)

	TitleStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			MarginLeft(2)

	StatusStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Italic(true).
			MarginLeft(2)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1).
			MarginTop(1)

	LogPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3C3C5C")).
			Padding(0, 1).
			MarginTop(0)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(PrimaryColor).
			PaddingLeft(1)

	NormalStyle = lipgloss.NewStyle().
			PaddingLeft(1)

	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)
