package table

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/styles"
)

var (
	cellStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			MaxHeight(1)

	titleCellStyle = cellStyle.Copy().
			Bold(true).
			Foreground(styles.DefaultTheme.MainText)

	singleRuneTitleCellStyle = titleCellStyle.Copy().Width(styles.SingleRuneWidth)

	headerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(styles.DefaultTheme.SecondaryBorder).
			BorderBottom(true)
)
