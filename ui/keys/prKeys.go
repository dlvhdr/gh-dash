package keys

import "github.com/charmbracelet/bubbles/key"

type PRKeyMap struct {
	Diff     key.Binding
}

var PRKeys = PRKeyMap{
	Diff: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "diff"),
	),
}

func PRFullHelp() []key.Binding {
	return []key.Binding{
		PRKeys.Diff,
	}
}
