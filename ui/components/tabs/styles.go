package tabs

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/ui/styles"
)

var (
	tabsBorderHeight   = 1
	tabsContentHeight  = 2
	TabsHeight         = tabsBorderHeight + tabsContentHeight
	viewSwitcherMargin = 1

	tab = lipgloss.NewStyle().
		Faint(true).
		Padding(0, 2)

	activeTab = tab.
			Copy().
			Faint(false).
			Bold(true).
			Background(styles.DefaultTheme.SelectedBackground).
			Foreground(styles.DefaultTheme.PrimaryText)

	tabSeparator = lipgloss.NewStyle().
			Foreground(styles.DefaultTheme.SecondaryBorder)

	tabsRow = lipgloss.NewStyle().
		Height(tabsContentHeight).
		PaddingTop(1).
		PaddingBottom(0).
		BorderBottom(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderBottomForeground(styles.DefaultTheme.PrimaryBorder)

	viewSwitcher = lipgloss.NewStyle()

	activeView = lipgloss.NewStyle().
			Foreground(styles.DefaultTheme.PrimaryText).
			Bold(true).
			Background(styles.DefaultTheme.SelectedBackground)

	viewsSeparator = lipgloss.NewStyle().
			BorderForeground(styles.DefaultTheme.PrimaryBorder).
			BorderStyle(lipgloss.NormalBorder()).
			BorderRight(true)

	inactiveView = lipgloss.NewStyle().
			Background(styles.DefaultTheme.FaintBorder).
			Foreground(styles.DefaultTheme.SecondaryText)
)
