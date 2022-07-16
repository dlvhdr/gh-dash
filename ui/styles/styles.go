package styles

import "github.com/charmbracelet/lipgloss"

var (
	MainTextStyle = lipgloss.NewStyle().
		Foreground(DefaultTheme.PrimaryText).
		Bold(true)
)
