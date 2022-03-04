package prsidebar

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/styles"
)

var (
	borderWidth    = 1
	pagerHeight    = 2
	contentPadding = 2

	openPR   = lipgloss.AdaptiveColor{Light: "#42A0FA", Dark: "#42A0FA"}
	closedPR = lipgloss.AdaptiveColor{Light: "#C38080", Dark: "#C38080"}
	mergedPR = lipgloss.AdaptiveColor{Light: "#A371F7", Dark: "#A371F7"}

	pillStyle = styles.MainTextStyle.Copy().
			Foreground(styles.DefaultTheme.SubleMainText).
			PaddingLeft(1).
			PaddingRight(1)

	sideBarStyle = lipgloss.NewStyle().
			Padding(0, contentPadding).
			BorderLeft(true).
			BorderStyle(lipgloss.Border{
			Top:         "",
			Bottom:      "",
			Left:        "│",
			Right:       "",
			TopLeft:     "",
			TopRight:    "",
			BottomRight: "",
			BottomLeft:  "",
		}).
		BorderForeground(styles.DefaultTheme.Border)

	pagerStyle = lipgloss.NewStyle().
			Height(pagerHeight).
			PaddingTop(1).
			Bold(true).
			Foreground(styles.DefaultTheme.FaintText)
)
