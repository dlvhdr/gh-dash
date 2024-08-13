package components

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/ui/context"
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

func GetIssueTextStyle(
	ctx *context.ProgramContext,
) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ctx.Theme.PrimaryText)
}

func RenderIssueTitle(
	ctx *context.ProgramContext,
	state string,
	title string,
	number int,
) string {
	prNumber := ""
	if ctx.Config.Theme.Ui.Table.Compact {
		prNumber = fmt.Sprintf("#%d ", number)
		var prNumberFg lipgloss.AdaptiveColor
		if state != "OPEN" {
			prNumberFg = ctx.Theme.FaintText
		} else {
			prNumberFg = ctx.Theme.SecondaryText
		}
		prNumber = lipgloss.NewStyle().Foreground(prNumberFg).Render(prNumber)
		// TODO: hack - see issue https://github.com/charmbracelet/lipgloss/issues/144
		// Provide ability to prevent insertion of Reset sequence #144
		prNumber = strings.Replace(prNumber, "\x1b[0m", "", -1)

	}

	rTitle := GetIssueTextStyle(ctx).Render(title)

	res := fmt.Sprintf("%s%s", prNumber, rTitle)
	return res
}
