package styles

import "github.com/charmbracelet/lipgloss"

var indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#383B5B"}

type Theme struct {
	MainText lipgloss.AdaptiveColor
	Border   lipgloss.AdaptiveColor
}

var DefaultTheme = Theme{
	MainText: lipgloss.AdaptiveColor{Light: "#242347", Dark: "#E2E1ED"},
	Border:   lipgloss.AdaptiveColor{Light: indigo.Light, Dark: indigo.Dark},
}
