package tabs

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/styles"
)

var (
	tab = lipgloss.NewStyle().
		Faint(true).
		Bold(true).
		Padding(0, 2)

	activeTab = tab.
			Copy().
			Foreground(styles.DefaultTheme.MainText).
			Faint(false)

	tabGap = tab.Copy().
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	tabsRow = lipgloss.NewStyle().
		PaddingTop(1).
		PaddingBottom(0).
		BorderBottom(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderBottomForeground(styles.DefaultTheme.Border)
)
