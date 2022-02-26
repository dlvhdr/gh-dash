package listviewport

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/styles"
)

var (
	pagerStyle = lipgloss.NewStyle().
		MarginTop(1).
		Bold(true).
		Foreground(styles.DefaultTheme.FaintText)
)
