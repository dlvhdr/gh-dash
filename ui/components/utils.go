package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/ui/styles"
)

func GetIssueTextStyle(state string) lipgloss.Style {
	if state == "OPEN" {
		return lipgloss.NewStyle().Foreground(styles.DefaultTheme.PrimaryText)
	}
	return lipgloss.NewStyle().Foreground(styles.DefaultTheme.FaintText)
}

func RenderIssueTitle(state string, title string, number int) string {
	prNumber := fmt.Sprintf("#%d", number)
	var prNumberFg lipgloss.AdaptiveColor
	if state != "OPEN" {
		prNumberFg = styles.DefaultTheme.FaintText
	} else {
		prNumberFg = styles.DefaultTheme.SecondaryText
	}
	prNumber = lipgloss.NewStyle().Foreground(prNumberFg).Render(prNumber)

	rTitle := GetIssueTextStyle(state).Render(title)

	// TODO: hack - see issue https://github.com/charmbracelet/lipgloss/issues/144
	// Provide ability to prevent insertion of Reset sequence #144
	prNumber = strings.Replace(prNumber, "\x1b[0m", "", -1)

	res := fmt.Sprintf("%s %s", prNumber, rTitle)
	return res
}
