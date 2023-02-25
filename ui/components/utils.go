package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/ui/context"
)

func GetIssueTextStyle(
	ctx *context.ProgramContext,
	state string,
) lipgloss.Style {
	if state == "OPEN" {
		return lipgloss.NewStyle().Foreground(ctx.Theme.PrimaryText)
	}
	return lipgloss.NewStyle().Foreground(ctx.Theme.FaintText)
}

func RenderIssueTitle(
	ctx *context.ProgramContext,
	state string,
	title string,
	number int,
) string {
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

func GetPrevCyclicItem(currItem, totalItems int) int {
	prevItem := (currItem - 1) % totalItems
	if prevItem < 0 {
		prevItem += totalItems
	}

	return prevItem
}

func GetNextCyclicItem(currItem, totalItems int) int {
	return (currItem + 1) % totalItems
}
