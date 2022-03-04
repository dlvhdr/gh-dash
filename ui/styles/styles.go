package styles

import "github.com/charmbracelet/lipgloss"

var (
	MainTextStyle = lipgloss.NewStyle().
		Foreground(DefaultTheme.MainText).
		Bold(true)
)
