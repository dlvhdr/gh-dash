package common

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

// defaultLabelColor is used when no color is provided (e.g., GitLab labels)
const defaultLabelColor = "6e7681"

func RenderLabels(sidebarWidth int, labels []data.Label, pillStyle lipgloss.Style) string {
	width := sidebarWidth

	renderedRows := []string{}

	rowContentsWidth := 0
	currentRowLabels := []string{}

	for _, l := range labels {
		// Use default color if no color is provided
		colorHex := l.Color
		if colorHex == "" {
			colorHex = defaultLabelColor
		}
		// Remove # prefix if present (GitLab includes it, GitHub doesn't)
		if len(colorHex) > 0 && colorHex[0] == '#' {
			colorHex = colorHex[1:]
		}
		c := lipgloss.Color("#" + colorHex)
		currentLabel := pillStyle.
			BorderForeground(c).
			Background(c).
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
