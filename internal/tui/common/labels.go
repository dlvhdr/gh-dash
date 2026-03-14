package common

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

func renderLabelPill(label data.Label, pillStyle lipgloss.Style, suffix string) string {
	c := lipgloss.Color("#" + label.Color)
	return pillStyle.
		BorderForeground(c).
		Background(c).
		Render(label.Name) + suffix
}

func RenderLabels(sidebarWidth int, labels []data.Label, pillStyle lipgloss.Style) string {
	width := sidebarWidth

	renderedRows := []string{}

	rowContentsWidth := 0
	currentRowLabels := []string{}

	for _, l := range labels {
		currentLabel := renderLabelPill(l, pillStyle, "")

		currentLabelWidth := lipgloss.Width(currentLabel)

		if rowContentsWidth+currentLabelWidth <= width {
			currentRowLabels = append(
				currentRowLabels,
				currentLabel,
			)
			rowContentsWidth += currentLabelWidth
		} else {
			currentRowLabels = append(currentRowLabels, "\n")
			renderedRows = append(
				renderedRows,
				lipgloss.JoinHorizontal(lipgloss.Top, currentRowLabels...),
			)

			currentRowLabels = []string{currentLabel}
			rowContentsWidth = currentLabelWidth
		}

		// +1 for the space between labels
		currentRowLabels = append(currentRowLabels, " ")
		rowContentsWidth += 1
	}

	renderedRows = append(renderedRows, lipgloss.JoinHorizontal(lipgloss.Top, currentRowLabels...))

	return lipgloss.JoinVertical(lipgloss.Left, renderedRows...)
}

func RenderLabelsWithLimit(
	sidebarWidth int,
	maxRows int,
	labels []data.Label,
	pillStyle lipgloss.Style,
) string {
	return RenderLabelsWithLimitAndRowStyle(
		sidebarWidth,
		maxRows,
		labels,
		pillStyle,
		lipgloss.NewStyle(),
	)
}

func RenderLabelsWithLimitAndRowStyle(
	sidebarWidth int,
	maxRows int,
	labels []data.Label,
	pillStyle lipgloss.Style,
	rowStyle lipgloss.Style,
) string {
	if sidebarWidth <= 0 || len(labels) == 0 {
		return ""
	}

	rowStylePrefix := stylePrefix(rowStyle)
	renderedRows := make([]string, 0, maxRows)
	rowContentsWidth := 0
	currentRowLabels := []string{}
	remainingLabels := 0
	rowCount := 0

	for idx, label := range labels {
		currentLabel := renderLabelPill(label, pillStyle, rowStylePrefix)
		currentLabelWidth := lipgloss.Width(currentLabel)

		if rowContentsWidth > 0 && rowContentsWidth+1+currentLabelWidth > sidebarWidth {
			rowCount++
			if maxRows > 0 && rowCount >= maxRows {
				remainingLabels = len(labels) - idx
				break
			}

			renderedRows = append(renderedRows, strings.Join(currentRowLabels, " "))
			currentRowLabels = []string{}
			rowContentsWidth = 0
		}

		if rowContentsWidth == 0 && currentLabelWidth > sidebarWidth {
			rowCount++
			currentRowLabels = append(currentRowLabels, truncateStyled(currentLabel, sidebarWidth, rowStylePrefix))
			renderedRows = append(renderedRows, strings.Join(currentRowLabels, " "))
			currentRowLabels = []string{}
			rowContentsWidth = 0
			if maxRows > 0 && rowCount >= maxRows {
				remainingLabels = len(labels) - idx - 1
				break
			}
			continue
		}

		if rowContentsWidth > 0 {
			rowContentsWidth++
		}
		currentRowLabels = append(currentRowLabels, currentLabel)
		rowContentsWidth += currentLabelWidth
	}

	if remainingLabels > 0 && maxRows > 0 && rowCount >= maxRows && len(currentRowLabels) == 0 {
		return strings.Join(renderedRows, "\n")
	}

	if remainingLabels > 0 {
		overflowLabel := pillStyle.Render("+"+strconv.Itoa(remainingLabels)) + rowStylePrefix
		overflowWidth := lipgloss.Width(overflowLabel)
		for {
			if rowContentsWidth == 0 {
				if overflowWidth > sidebarWidth {
					currentRowLabels = []string{truncateStyled(overflowLabel, sidebarWidth, rowStylePrefix)}
				} else {
					currentRowLabels = []string{overflowLabel}
				}
				break
			}

			if rowContentsWidth+1+overflowWidth <= sidebarWidth {
				currentRowLabels = append(currentRowLabels, overflowLabel)
				break
			}

			remainingLabels++
			lastItemIdx := len(currentRowLabels) - 1
			rowContentsWidth -= lipgloss.Width(currentRowLabels[lastItemIdx])
			currentRowLabels = currentRowLabels[:lastItemIdx]
			if len(currentRowLabels) > 0 {
				rowContentsWidth--
			} else {
				rowContentsWidth = 0
			}

			overflowLabel = pillStyle.Render("+"+strconv.Itoa(remainingLabels)) + rowStylePrefix
			overflowWidth = lipgloss.Width(overflowLabel)
		}
	}

	if len(currentRowLabels) > 0 {
		renderedRows = append(renderedRows, strings.Join(currentRowLabels, " "))
	}

	return strings.Join(renderedRows, "\n")
}

func stylePrefix(style lipgloss.Style) string {
	rendered := style.Render("x")
	if rendered == "" {
		return ""
	}

	prefix, _, found := strings.Cut(rendered, "x")
	if !found {
		return ""
	}

	return prefix
}

func truncateStyled(value string, width int, stylePrefix string) string {
	return ansi.Truncate(value, width, constants.Ellipsis) + stylePrefix
}
