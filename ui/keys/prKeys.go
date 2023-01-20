package keys

import "github.com/charmbracelet/bubbles/key"

type PRKeyMap struct {
	Assign   key.Binding
	Unassign key.Binding
	Comment  key.Binding
	Diff     key.Binding
	Checkout key.Binding
	Close    key.Binding
	Ready    key.Binding
	Reopen   key.Binding
	Merge    key.Binding
}

var PRKeys = PRKeyMap{
	Assign: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "assign"),
	),
	Unassign: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "unassign"),
	),
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
	Ready: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "ready for review"),
	),
	Merge: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "merge"),
	),
}

func PRFullHelp() []key.Binding {
	return []key.Binding{
		PRKeys.Assign,
		PRKeys.Unassign,
		PRKeys.Comment,
		PRKeys.Diff,
		PRKeys.Checkout,
		PRKeys.Close,
		PRKeys.Ready,
		PRKeys.Reopen,
		PRKeys.Merge,
	}
}
