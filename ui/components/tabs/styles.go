package tabs

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/styles"
)

var (
	tab = lipgloss.NewStyle().
		Faint(true).
		Padding(0, 2)

	activeTab = tab.
			Copy().
			Faint(false).
			Bold(true).
			Background(styles.DefaultTheme.SelectedBackground).
			Foreground(styles.DefaultTheme.MainText)

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
