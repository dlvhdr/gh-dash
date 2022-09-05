package keys

import "github.com/charmbracelet/bubbles/key"

type PRKeyMap struct {
	Comment  key.Binding
	Diff     key.Binding
	Checkout key.Binding
	Close    key.Binding
	Reopen   key.Binding
}

var PRKeys = PRKeyMap{
	Comment: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "comment"),
	),
	Diff: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "diff"),
	),
	Checkout: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "checkout"),
	),
	Close: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "close"),
	),
	Reopen: key.NewBinding(
		key.WithKeys("X"),
		key.WithHelp("X", "reopen"),
	),
}

func PRFullHelp() []key.Binding {
	return []key.Binding{
		PRKeys.Comment,
		PRKeys.Diff,
		PRKeys.Checkout,
		PRKeys.Close,
		PRKeys.Reopen,
	}
}
