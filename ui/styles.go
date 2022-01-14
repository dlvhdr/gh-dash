package ui

import "github.com/charmbracelet/lipgloss"

var (
	headerHeight       = 6
	footerHeight       = 2
	prRowHeight        = 2
	singleRuneWidth    = 4
	mainContentPadding = 1
	pagerHeight        = 2
	cellPadding        = cellStyle.GetPaddingLeft() + cellStyle.GetPaddingRight()

	reviewCellWidth    = singleRuneWidth
	mergeableCellWidth = singleRuneWidth
	ciCellWidth        = lipgloss.Width(cellStyle.Render("CI"))
	linesCellWidth     = lipgloss.Width(cellStyle.Render("123450 / -123450"))
	prAuthorCellWidth  = 15
	prRepoCellWidth    = 20
	updatedAtCellWidth = lipgloss.Width(cellStyle.Render(" Updated"))
	usedWidth          = reviewCellWidth + mergeableCellWidth +
		ciCellWidth + linesCellWidth + prAuthorCellWidth + prRepoCellWidth + updatedAtCellWidth

	indigo             = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#383B5B"}
	subtleIndigo       = lipgloss.AdaptiveColor{Light: "#5A57B5", Dark: "#242347"}
	selectedBackground = lipgloss.AdaptiveColor{Light: subtleIndigo.Light, Dark: subtleIndigo.Dark}
	border             = lipgloss.AdaptiveColor{Light: indigo.Light, Dark: indigo.Dark}
	secondaryBorder    = lipgloss.AdaptiveColor{Light: indigo.Light, Dark: "#39386b"}
	faintBorder        = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#28283b"}
	mainText           = lipgloss.AdaptiveColor{Light: "#242347", Dark: "#E2E1ED"}
	secondaryText      = lipgloss.AdaptiveColor{Light: indigo.Light, Dark: "#666CA6"}
	faintText          = lipgloss.AdaptiveColor{Light: indigo.Light, Dark: "#3E4057"}
	warningText        = lipgloss.AdaptiveColor{Light: "#F23D5C", Dark: "#F23D5C"}
	successText        = lipgloss.AdaptiveColor{Light: "#3DF294", Dark: "#3DF294"}
	openPR             = lipgloss.AdaptiveColor{Light: "#42A0FA", Dark: "#42A0FA"}
	closedPR           = lipgloss.AdaptiveColor{Light: "#C38080", Dark: "#C38080"}
	mergedPR           = lipgloss.AdaptiveColor{Light: "#A371F7", Dark: "#A371F7"}

	tab = lipgloss.NewStyle().
		Faint(true).
		Bold(true).
		Padding(0, 2)

	activeTab = tab.
			Copy().
			Foreground(mainText).
			Faint(false)

	tabGap = tab.Copy().
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	tabsRow = lipgloss.NewStyle().
		PaddingTop(1).
		PaddingBottom(0).
		BorderBottom(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderBottomForeground(border)

	emptyStateStyle = lipgloss.NewStyle().
			Faint(true).
			PaddingLeft(1).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(secondaryBorder).
			BorderBottom(true)

	pullRequestStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(faintBorder).
				BorderBottom(true)

	cellStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			MaxHeight(1)

	titleCellStyle = cellStyle.Copy().
			Bold(true).
			Foreground(mainText)

	selectedCellStyle = cellStyle.Copy().Background(selectedBackground)

	singleRuneTitleCellStyle = titleCellStyle.Copy().Width(singleRuneWidth)

	singleRuneCellStyle = cellStyle.Copy().Width(singleRuneWidth)

	selectedSingleRuneCellStyle = singleRuneCellStyle.Copy().Background(selectedBackground)

	spinnerStyle = lipgloss.NewStyle().PaddingLeft(2)

	helpStyle = lipgloss.NewStyle().
			Height(footerHeight).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(border)

	sideBarStyle = lipgloss.NewStyle().
			Padding(0, 2).
			BorderLeft(true).
			BorderStyle(lipgloss.Border{
			Top:         "",
			Bottom:      "",
			Left:        "│",
			Right:       "",
			TopLeft:     "",
			TopRight:    "",
			BottomRight: "",
			BottomLeft:  "",
		}).
		BorderForeground(border)

	mainTextStyle = lipgloss.NewStyle().
			Foreground(mainText).
			Bold(true)

	pillStyle = mainTextStyle.Copy().
			Foreground(subtleIndigo).
			PaddingLeft(1).
			PaddingRight(1)

	pagerStyle = lipgloss.NewStyle().
			MarginTop(1).
			Bold(true).
			Foreground(faintText)
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
