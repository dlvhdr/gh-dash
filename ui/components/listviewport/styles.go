package listviewport

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/styles"
)

var (
	pagerHeight = 2

	pagerStyle = lipgloss.NewStyle().
			Height(pagerHeight).
			MaxHeight(pagerHeight).
			PaddingTop(1).
			Bold(true).
			Foreground(styles.DefaultTheme.FaintText)
)
