package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	colorPrimary   = lipgloss.Color("#007aff")
	colorSuccess   = lipgloss.Color("#34c759")
	colorWarning   = lipgloss.Color("#ff9500")
	colorError     = lipgloss.Color("#ff3b30")
	colorMuted     = lipgloss.Color("#8e8e93")
	colorBackground = lipgloss.Color("#1c1c1e")
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)

	successStyle = lipgloss.NewStyle().Foreground(colorSuccess)
	warningStyle = lipgloss.NewStyle().Foreground(colorWarning)
	errorStyle   = lipgloss.NewStyle().Foreground(colorError)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(colorPrimary).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)
)

func renderTitle(title, subtitle string) string {
	s := titleStyle.Render("EnvKit") + " " + titleStyle.Copy().Foreground(lipgloss.Color("#ffffff")).Render(title)
	if subtitle != "" {
		s += "\n" + subtitleStyle.Render(subtitle)
	}
	return s
}

func renderHelp(keys string) string {
	return helpStyle.Render(keys)
}
