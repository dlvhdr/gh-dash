package ui

import "github.com/charmbracelet/lipgloss"

var (
	headerHeight       = 6
	footerHeight       = 2
	singleRuneWidth    = 4
	mainContentPadding = 0
	cellPadding        = cellStyle.GetPaddingLeft() + cellStyle.GetPaddingRight()

	emptyCellWidth     = lipgloss.Width(cellStyle.Render(""))
	reviewCellWidth    = emptyCellWidth
	mergeableCellWidth = emptyCellWidth
	ciCellWidth        = lipgloss.Width(cellStyle.Render("CI"))
	linesCellWidth     = lipgloss.Width(cellStyle.Render("+12345 / -12345"))
	prAuthorCellWidth  = 15 + cellPadding
	prRepoCellWidth    = 20 + cellPadding
	updatedAtCellWidth = lipgloss.Width("Updated At") + cellPadding
	usedWidth          = emptyCellWidth + reviewCellWidth + mergeableCellWidth +
		ciCellWidth + linesCellWidth + prAuthorCellWidth + prRepoCellWidth + updatedAtCellWidth

	indigo       = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	subtleIndigo = lipgloss.AdaptiveColor{Light: "#5A57B5", Dark: "#242347"}
	selected     = lipgloss.AdaptiveColor{Light: subtleIndigo.Light, Dark: subtleIndigo.Dark}
	border       = lipgloss.AdaptiveColor{Light: indigo.Light, Dark: indigo.Dark}
	subtleBorder = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special      = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(indigo.Dark))

	tab = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(highlight).
		Faint(true).
		Padding(0, 1)

	activeTab = tab.
			Copy().
			Faint(false).
			Bold(true).
			Border(activeTabBorder, true)

	tabGap = tab.Copy().
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	emptyStateStyle = lipgloss.NewStyle().
			Faint(true).
			PaddingLeft(2).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(border).
			BorderBottom(true)

	pullRequestStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(subtleBorder).
				BorderBottom(true)

	selectedPullRequestStyle = lipgloss.NewStyle().
					Background(lipgloss.Color(subtleIndigo.Dark)).
					Foreground(lipgloss.Color(subtleIndigo.Light)).
					BorderForeground(subtleBorder).
					Inherit(pullRequestStyle)

	cellStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			MaxHeight(1)

	selectedCellStyle = cellStyle.Copy().Background(selected)

	singleRuneCellStyle = cellStyle.Copy().PaddingRight(1).Width(singleRuneWidth)

	selectedRuneCellStyle = singleRuneCellStyle.Copy().Background(selected)

	selectionPointerStyle = lipgloss.NewStyle()
)

func makeCellStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return selectedCellStyle.Copy()
	}

	return cellStyle.Copy()
}

func makeRuneCellStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return selectedRuneCellStyle.Copy()
	}

	return singleRuneCellStyle.Copy()
}
