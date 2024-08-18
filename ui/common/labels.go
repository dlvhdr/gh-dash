package common

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/v4/data"
)

func RenderLabels(sidebarWidth int, labels []data.Label, pillStyle lipgloss.Style) string {
	width := sidebarWidth

	renderedRows := []string{}

	rowContentsWidth := 0
	currentRowLabels := []string{}

	for _, l := range labels {
		currentLabel := pillStyle.
			Background(lipgloss.Color("#" + l.Color)).
			Render(l.Name)

		currentLabelWidth := lipgloss.Width(currentLabel)

		if rowContentsWidth+currentLabelWidth <= width {
			currentRowLabels = append(
				currentRowLabels,
				currentLabel,
			)
			rowContentsWidth += currentLabelWidth
		} else {
			currentRowLabels = append(currentRowLabels, "\n")
			renderedRows = append(renderedRows, lipgloss.JoinHorizontal(lipgloss.Top, currentRowLabels...))

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
