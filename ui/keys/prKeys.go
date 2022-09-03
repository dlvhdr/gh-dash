package keys

import "github.com/charmbracelet/bubbles/key"

type PRKeyMap struct {
	Diff     key.Binding
	Checkout key.Binding
	Close    key.Binding
}

var PRKeys = PRKeyMap{
	Diff: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "diff"),
	),
	Checkout: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "checkout"),
	),
	Close: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "close"),
	),
}

func PRFullHelp() []key.Binding {
	return []key.Binding{
		PRKeys.Diff,
		PRKeys.Checkout,
		PRKeys.Close,
	}
}
