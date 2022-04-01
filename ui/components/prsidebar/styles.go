package prsidebar

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/ui/styles"
)

var (
	openPR   = lipgloss.AdaptiveColor{Light: "#42A0FA", Dark: "#42A0FA"}
	closedPR = lipgloss.AdaptiveColor{Light: "#C38080", Dark: "#C38080"}
	mergedPR = lipgloss.AdaptiveColor{Light: "#A371F7", Dark: "#A371F7"}

	pillStyle = styles.MainTextStyle.Copy().
			Foreground(styles.DefaultTheme.SubleMainText).
			PaddingLeft(1).
			PaddingRight(1)
)
