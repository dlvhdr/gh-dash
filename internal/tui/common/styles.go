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
	MainTextStyle   lipgloss.Style
	FaintTextStyle  lipgloss.Style
	FooterStyle     lipgloss.Style
	ErrorStyle      lipgloss.Style
	SuccessStyle    lipgloss.Style
	DraftGlyph      string
	PersonGlyph     string
	WaitingGlyph    string
	WaitingDotGlyph string
	FailureGlyph    string
	SuccessGlyph    string
	MergedGlyph     string
	CommentGlyph    string
}

func BuildStyles(theme theme.Theme) CommonStyles {
	var s CommonStyles

	s.MainTextStyle = lipgloss.NewStyle().
		Foreground(theme.PrimaryText).
		Bold(true)
	s.FaintTextStyle = lipgloss.NewStyle().
		Foreground(theme.FaintText)
	s.FooterStyle = lipgloss.NewStyle().
		Background(theme.SelectedBackground).
		Height(FooterHeight)

	s.ErrorStyle = lipgloss.NewStyle().Foreground(theme.ErrorText)
	s.SuccessStyle = lipgloss.NewStyle().Foreground(theme.SuccessText)

	s.PersonGlyph = lipgloss.NewStyle().
		Foreground(theme.FaintText).
		Render(constants.PersonIcon)
	s.WaitingGlyph = lipgloss.NewStyle().
		Foreground(theme.WarningText).
		Render(constants.WaitingIcon)
	s.WaitingDotGlyph = lipgloss.NewStyle().
		Foreground(theme.WarningText).
		Render(constants.DotIcon)
	s.FailureGlyph = lipgloss.NewStyle().
		Foreground(theme.ErrorText).
		Render(constants.FailureIcon)
	s.SuccessGlyph = lipgloss.NewStyle().
		Foreground(theme.SuccessText).
		Render(constants.SuccessIcon)
	s.CommentGlyph = lipgloss.NewStyle().
		Foreground(theme.PrimaryText).
		Render(constants.CommentIcon)
	s.DraftGlyph = lipgloss.NewStyle().
		Foreground(theme.PrimaryText).
		Render(constants.DraftIcon)
	s.MergedGlyph = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#A371F7",
			Dark:  "#A371F7",
		}).
		Render(constants.MergedIcon)
	return s
}

// RenderPreviewHeader renders the repo/number line at the top of preview panes
// (e.g., "owner/repo · #123" or "#123 · owner/repo")
func RenderPreviewHeader(theme theme.Theme, width int, text string) string {
	return lipgloss.NewStyle().
		PaddingLeft(1).
		Width(width).
		Background(theme.SelectedBackground).
		Foreground(theme.SecondaryText).
		Render(text)
}

// RenderPreviewTitle renders the title block with background highlight
func RenderPreviewTitle(theme theme.Theme, styles CommonStyles, width int, title string) string {
	return lipgloss.NewStyle().Height(3).Width(width).Background(
		theme.SelectedBackground).PaddingLeft(1).Render(
		lipgloss.PlaceVertical(3, lipgloss.Center, styles.MainTextStyle.
			Background(theme.SelectedBackground).
			Render(title),
		),
	)
}
