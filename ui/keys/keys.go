package keys

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/dlvhdr/gh-dash/config"
)

type KeyMap struct {
	viewType      config.ViewType
	Up            key.Binding
	Down          key.Binding
	FirstLine     key.Binding
	LastLine      key.Binding
	TogglePreview key.Binding
	OpenGithub    key.Binding
	Refresh       key.Binding
	PageDown      key.Binding
	PageUp        key.Binding
	NextSection   key.Binding
	PrevSection   key.Binding
	SwitchView    key.Binding
	Search        key.Binding
	Help          key.Binding
	Quit          key.Binding
}

func GetKeyMap(viewType config.ViewType) help.KeyMap {
	Keys.viewType = viewType
	return Keys
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	var additionalKeys []key.Binding
	if k.viewType == config.PRsView {
		additionalKeys = PRFullHelp()
	} else {
		additionalKeys = IssueFullHelp()
	}

	return [][]key.Binding{
		k.NavigationKeys(),
		k.AppKeys(),
		additionalKeys,
		k.QuitAndHelpKeys(),
	}
}

func (k KeyMap) NavigationKeys() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.PrevSection, k.NextSection, k.FirstLine, k.LastLine, k.PageDown, k.PageUp}
}

func (k KeyMap) AppKeys() []key.Binding {
	return []key.Binding{k.Refresh, k.SwitchView, k.TogglePreview, k.OpenGithub, k.Search}
}

func (k KeyMap) QuitAndHelpKeys() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

var Keys = KeyMap{
	Up:            key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
	Down:          key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
	FirstLine:     key.NewBinding(key.WithKeys("g", "home"), key.WithHelp("g/home", "first item")),
	LastLine:      key.NewBinding(key.WithKeys("G", "end"), key.WithHelp("G/end", "last item")),
	TogglePreview: key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "open in Preview")),
	OpenGithub:    key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open in GitHub")),
	Refresh:       key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	PageDown:      key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("Ctrl+d", "preview page down")),
	PageUp:        key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("Ctrl+u", "preview page up")),
	NextSection:   key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("/l", "next section")),
	PrevSection:   key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("/h", "previous section")),
	SwitchView:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "switch view")),
	Search:        key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	Help:          key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
	Quit:          key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
}
