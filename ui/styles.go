package ui

import "github.com/charmbracelet/lipgloss"

type ColorPair struct {
	Dark  string
	Light string
}

func NewColorPair(dark, light string) ColorPair {
	return ColorPair{dark, light}
}

var (
	Indigo       = NewColorPair("#7571F9", "#5A56E0")
	SubtleIndigo = NewColorPair("#514DC1", "#7D79F6")
	Cream        = NewColorPair("#FFFDF5", "#FFFDF5")
	YellowGreen  = NewColorPair("#ECFD65", "#04B575")
	Fuschia      = NewColorPair("#EE6FF8", "#EE6FF8")
	Green        = NewColorPair("#04B575", "#04B575")
	Red          = NewColorPair("#ED567A", "#FF4672")
	FaintRed     = NewColorPair("#C74665", "#FF6F91")
	SpinnerColor = NewColorPair("#747373", "#8E8E8E")
	NoColor      = NewColorPair("", "")

	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

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

	singleRuneWidth = 4

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(Indigo.Dark))

	Tab = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(highlight).
		Faint(true).
		Padding(0, 1)

	ActiveTab = Tab.Copy().Faint(false).Bold(true).Border(activeTabBorder, true)

	TabGap = Tab.Copy().
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	EmptyStateStyle = lipgloss.NewStyle().
			Faint(true).
			PaddingLeft(2).
			MarginBottom(1)

	PullRequestStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("235")).
				BorderBottom(true)

	SelectedPullRequestStyle = lipgloss.NewStyle().
					Background(lipgloss.Color(NoColor.Light)).
					Foreground(lipgloss.Color(SubtleIndigo.Light)).
					Inherit(PullRequestStyle)

	CellStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			MaxHeight(1)

	SingleRuneCellStyle = CellStyle.Copy().MarginRight(1).Width(singleRuneWidth)
)
