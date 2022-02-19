package pr

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/styles"
)

var (
	successText                 = lipgloss.AdaptiveColor{Light: "#3DF294", Dark: "#3DF294"}
	openPR                      = lipgloss.AdaptiveColor{Light: "#42A0FA", Dark: "#42A0FA"}
	closedPR                    = lipgloss.AdaptiveColor{Light: "#C38080", Dark: "#C38080"}
	mergedPR                    = lipgloss.AdaptiveColor{Light: "#A371F7", Dark: "#A371F7"}
	singleRuneCellStyle         = cellStyle.Copy()
	selectedSingleRuneCellStyle = singleRuneCellStyle.Copy().Background(styles.DefaultTheme.SelectedBackground)
	cellStyle                   = lipgloss.NewStyle()
	selectedCellStyle           = cellStyle.Copy().Background(styles.DefaultTheme.SelectedBackground)
)

func makeCellStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return selectedCellStyle.Copy()
	}

	return cellStyle.Copy()
}

func makeRuneCellStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return selectedSingleRuneCellStyle.Copy()
	}

	return singleRuneCellStyle.Copy()
}
