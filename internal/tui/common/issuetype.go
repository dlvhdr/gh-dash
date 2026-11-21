package common

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

// Foreground colors for each IssueTypeColor value, from the palette GitHub uses for issue types in its light and
// dark themes.
var issueTypeColors = map[string]compat.AdaptiveColor{
	"GRAY":   {Light: lipgloss.Color("#59636E"), Dark: lipgloss.Color("#9198A1")},
	"BLUE":   {Light: lipgloss.Color("#0969DA"), Dark: lipgloss.Color("#4493F8")},
	"GREEN":  {Light: lipgloss.Color("#1A7F37"), Dark: lipgloss.Color("#3FB950")},
	"YELLOW": {Light: lipgloss.Color("#9A6700"), Dark: lipgloss.Color("#D29922")},
	"ORANGE": {Light: lipgloss.Color("#BC4C00"), Dark: lipgloss.Color("#DB6D28")},
	"RED":    {Light: lipgloss.Color("#D1242F"), Dark: lipgloss.Color("#F85149")},
	"PINK":   {Light: lipgloss.Color("#BF3989"), Dark: lipgloss.Color("#DB61A2")},
	"PURPLE": {Light: lipgloss.Color("#8250DF"), Dark: lipgloss.Color("#AB7DF8")},
}

// IssueTypeColor maps an IssueTypeColor value to a terminal color. An unknown value gets the GRAY color, so a color
// GitHub adds later still renders.
func IssueTypeColor(name string) compat.AdaptiveColor {
	if c, ok := issueTypeColors[name]; ok {
		return c
	}
	return issueTypeColors["GRAY"]
}

// RenderIssueTypePill renders an issue type the way a label is rendered: a pill in a darkened shade of the type's
// color, with lighter text.
func RenderIssueTypePill(issueType data.IssueType, pillStyle lipgloss.Style) string {
	c := lipgloss.Darken(IssueTypeColor(issueType.Color).Dark, 0.3)
	return pillStyle.
		BorderForeground(c).
		Foreground(lipgloss.Lighten(c, 0.5)).
		Background(c).
		Render(issueType.Name)
}
