package common

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

var (
	HeaderHeight       = 2
	SearchHeight       = 3
	FooterHeight       = 1
	ExpandedHelpHeight = 17
	InputBoxHeight     = 8
	SingleRuneWidth    = 4
	MainContentPadding = 1
	TabsBorderHeight   = 1
	TabsContentHeight  = 2
	TabsHeight         = TabsBorderHeight + TabsContentHeight
	ViewSwitcherMargin = 1
	TableHeaderHeight  = 2
)

type CommonStyles struct {
	MainTextStyle  lipgloss.Style
	FaintTextStyle lipgloss.Style
	FooterStyle    lipgloss.Style
	ErrorStyle     lipgloss.Style
	DraftGlyph     string
	PersonGlyph    string
	WaitingGlyph   string
	FailureGlyph   string
	SuccessGlyph   string
	MergedGlyph    string
	CommentGlyph   string
}

func BuildStyles(theme theme.Theme) CommonStyles {
	var s CommonStyles

	s.MainTextStyle = lipgloss.NewStyle().
		Foreground(theme.PrimaryText).
		Bold(true).
		Background(theme.MainBackground)
	s.FaintTextStyle = lipgloss.NewStyle().
		Foreground(theme.FaintText).
		Background(theme.MainBackground)
	s.FooterStyle = lipgloss.NewStyle().
		Background(theme.SelectedBackground).
		Height(FooterHeight)

	s.ErrorStyle = s.FooterStyle.
		Foreground(theme.ErrorText).
		MaxHeight(FooterHeight)

	s.PersonGlyph = lipgloss.NewStyle().
		Foreground(theme.FaintText).
		Background(theme.MainBackground).
		Render(constants.PersonIcon)
	s.WaitingGlyph = lipgloss.NewStyle().
		Foreground(theme.WarningText).
		Background(theme.MainBackground).
		Render(constants.WaitingIcon)
	s.FailureGlyph = lipgloss.NewStyle().
		Foreground(theme.ErrorText).
		Background(theme.MainBackground).
		Render(constants.FailureIcon)
	s.SuccessGlyph = lipgloss.NewStyle().
		Foreground(theme.SuccessText).
		Background(theme.MainBackground).
		Render(constants.SuccessIcon)
	s.CommentGlyph = lipgloss.NewStyle().
		Foreground(theme.PrimaryText).
		Background(theme.MainBackground).
		Render(constants.CommentIcon)
	s.DraftGlyph = lipgloss.NewStyle().
		Foreground(theme.PrimaryText).
		Background(theme.MainBackground).
		Render(constants.DraftIcon)
	s.MergedGlyph = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#A371F7",
			Dark:  "#A371F7",
		}).
		Background(theme.MainBackground).
		Render(constants.MergedIcon)
	return s
}
