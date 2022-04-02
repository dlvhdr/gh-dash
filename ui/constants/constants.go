package constants

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/ui/styles"
)

type KeyMap struct {
	Up            key.Binding
	Down          key.Binding
	FirstItem     key.Binding
	LastItem      key.Binding
	TogglePreview key.Binding
	OpenGithub    key.Binding
	Refresh       key.Binding
	PageDown      key.Binding
	PageUp        key.Binding
	NextSection   key.Binding
	PrevSection   key.Binding
	Help          key.Binding
	Quit          key.Binding
}

type Dimensions struct {
	Width  int
	Height int
}

var (
	WaitingGlyph = lipgloss.NewStyle().Foreground(styles.DefaultTheme.FaintText).Render("")
	FailureGlyph = lipgloss.NewStyle().Foreground(styles.DefaultTheme.WarningText).Render("")
	SuccessGlyph = lipgloss.NewStyle().Foreground(styles.DefaultTheme.SuccessText).Render("")

	Keys = KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		FirstItem: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g/home", "first item"),
		),
		LastItem: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G/end", "last item"),
		),
		PrevSection: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("/h", "previous section"),
		),
		NextSection: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("/l", "next section"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("Ctrl+u", "preview page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("Ctrl+d", "preview page down"),
		),
		TogglePreview: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "open in Preview"),
		),
		OpenGithub: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open in GitHub"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
)
