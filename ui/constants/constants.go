package constants

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/styles"
)

var (
	WaitingGlyph = lipgloss.NewStyle().Foreground(styles.DefaultTheme.FaintText).Render("")
	FailureGlyph = lipgloss.NewStyle().Foreground(styles.DefaultTheme.WarningText).Render("")
	SuccessGlyph = lipgloss.NewStyle().Foreground(styles.DefaultTheme.SuccessText).Render("")
)
