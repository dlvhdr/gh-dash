package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/styles"
)

var (
	footerHeight = 3

	helpTextStyle = lipgloss.NewStyle().Foreground(styles.DefaultTheme.SecondaryText)
	helpStyle     = lipgloss.NewStyle().
			Height(footerHeight - 1).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(styles.DefaultTheme.Border)
)
