package components

import (
	"fmt"
	"strings"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/ui/context"
)

func FormatNumber(num int) string {
	if num >= 1000000 {
		million := float64(num) / 1000000.0
		return strconv.FormatFloat(million, 'f', 1, 64) + "M"
	} else if num >= 1000 {
		kilo := float64(num) / 1000.0
		return strconv.FormatFloat(kilo, 'f', 1, 64) + "k"
	}

	return strconv.Itoa(num)
}

func GetIssueTextStyle(ctx *context.ProgramContext, state string) lipgloss.Style {
	if state == "OPEN" {
		return lipgloss.NewStyle().Foreground(ctx.Theme.PrimaryText)
	}
	return lipgloss.NewStyle().Foreground(ctx.Theme.FaintText)
}

func RenderIssueTitle(ctx *context.ProgramContext, state string, title string, number int) string {
	prNumber := fmt.Sprintf("#%d", number)
	var prNumberFg lipgloss.AdaptiveColor
	if state != "OPEN" {
		prNumberFg = ctx.Theme.FaintText
	} else {
		prNumberFg = ctx.Theme.SecondaryText
	}
	prNumber = lipgloss.NewStyle().Foreground(prNumberFg).Render(prNumber)

	rTitle := GetIssueTextStyle(ctx, state).Render(title)

	// TODO: hack - see issue https://github.com/charmbracelet/lipgloss/issues/144
	// Provide ability to prevent insertion of Reset sequence #144
	prNumber = strings.Replace(prNumber, "\x1b[0m", "", -1)

	res := fmt.Sprintf("%s %s", prNumber, rTitle)
	return res
}
