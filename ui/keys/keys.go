package keys

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	log "github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/config"
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
	RefreshAll    key.Binding
	PageDown      key.Binding
	PageUp        key.Binding
	NextSection   key.Binding
	PrevSection   key.Binding
	Search        key.Binding
	CopyUrl       key.Binding
	CopyNumber    key.Binding
	Help          key.Binding
	Quit          key.Binding
}

func CreateKeyMapForView(viewType config.ViewType) help.KeyMap {
	Keys.viewType = viewType
	return Keys
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	var additionalKeys []key.Binding
	if k.viewType == config.PRsView {
		additionalKeys = PRFullHelp()
	} else if k.viewType == config.RepoView {
		additionalKeys = BranchFullHelp()
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
	return []key.Binding{
		k.Up,
		k.Down,
		k.PrevSection,
		k.NextSection,
		k.FirstLine,
		k.LastLine,
		k.PageDown,
		k.PageUp,
	}
}

func (k KeyMap) AppKeys() []key.Binding {
	return []key.Binding{
		k.Refresh,
		k.RefreshAll,
		k.TogglePreview,
		k.OpenGithub,
		k.CopyNumber,
		k.CopyUrl,
		k.Search,
	}
}

func (k KeyMap) QuitAndHelpKeys() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

var Keys = &KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	FirstLine: key.NewBinding(
		key.WithKeys("g", "home"),
		key.WithHelp("g/home", "first item"),
	),
	LastLine: key.NewBinding(
		key.WithKeys("G", "end"),
		key.WithHelp("G/end", "last item"),
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
	RefreshAll: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "refresh all"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("Ctrl+d", "preview page down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("Ctrl+u", "preview page up"),
	),
	NextSection: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("󰁔/l", "next section"),
	),
	PrevSection: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("󰁍/h", "previous section"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	CopyNumber: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "copy number"),
	),
	CopyUrl: key.NewBinding(
		key.WithKeys("Y"),
		key.WithHelp("Y", "copy url"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// Rebind will update our saved keybindings from configuration values.
func Rebind(universal, issueKeys, prKeys, branchKeys []config.Keybinding) error {
	err := rebindUniversal(universal)
	if err != nil {
		return err
	}

	err = rebindPRKeys(prKeys)
	if err != nil {
		return err
	}

	err = rebindBranchKeys(branchKeys)
	if err != nil {
		return err
	}

	return rebindIssueKeys(issueKeys)
}

func rebindUniversal(universal []config.Keybinding) error {
	log.Debug("Rebinding universal keys", "keys", universal)
	for _, kb := range universal {
		if kb.Builtin == "" {
			continue
		}

		log.Debug("Rebinding universal key", "builtin", kb.Builtin, "key", kb.Key)

		var key *key.Binding

		switch kb.Builtin {
		case "up":
			key = &Keys.Up
		case "down":
			key = &Keys.Down
		case "firstLine":
			key = &Keys.FirstLine
		case "lastLine":
			key = &Keys.LastLine
		case "togglePreview":
			key = &Keys.TogglePreview
		case "openGithub":
			key = &Keys.OpenGithub
		case "refresh":
			key = &Keys.Refresh
		case "refreshAll":
			key = &Keys.RefreshAll
		case "pageDown":
			key = &Keys.PageDown
		case "pageUp":
			key = &Keys.PageUp
		case "nextSection":
			key = &Keys.NextSection
		case "prevSection":
			key = &Keys.PrevSection
		case "search":
			key = &Keys.Search
		case "copyurl":
			key = &Keys.CopyUrl
		case "copyNumber":
			key = &Keys.CopyNumber
		case "help":
			key = &Keys.Help
		case "quit":
			key = &Keys.Quit
		default:
			return fmt.Errorf("unknown built-in universal key: '%s'", kb.Builtin)
		}

		key.SetKeys(kb.Key)
		key.SetHelp(kb.Key, key.Help().Desc)
	}

	return nil
}
