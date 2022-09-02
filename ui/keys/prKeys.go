package keys

import "github.com/charmbracelet/bubbles/key"

type PRKeyMap struct {
	Diff     key.Binding
	Checkout key.Binding
	Close    key.Binding
	Comment  key.Binding
	Edit     key.Binding
	Merge    key.Binding
	Ready    key.Binding
	Reopen   key.Binding
	Review   key.Binding
	Help     key.Binding
	Quit     key.Binding
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
	Comment: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "comment"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Merge: key.NewBinding(
		key.WithKeys("M"),
		key.WithHelp("M", "merge"),
	),
	Ready: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "ready for review"),
	),
	Reopen: key.NewBinding(
		key.WithKeys("O"),
		key.WithHelp("O", "reopen"),
	),
	Review: key.NewBinding(
		key.WithKeys("W"),
		key.WithHelp("W", "review"),
	),
}

func PRFullHelp() []key.Binding {
	return []key.Binding{
		PRKeys.Diff,
		PRKeys.Checkout,
		PRKeys.Close,
		PRKeys.Comment,
		PRKeys.Edit,
		PRKeys.Merge,
		PRKeys.Ready,
		PRKeys.Reopen,
		PRKeys.Review,
	}
}
