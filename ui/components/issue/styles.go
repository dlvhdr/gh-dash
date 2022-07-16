package issue

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/ui/styles"
)

var (
	OpenIssue   = lipgloss.AdaptiveColor{Light: "#42A0FA", Dark: "#42A0FA"}
	ClosedIssue = lipgloss.AdaptiveColor{Light: "#C38080", Dark: "#C38080"}

	titleText = lipgloss.NewStyle().Foreground(styles.DefaultTheme.PrimaryText)
)
