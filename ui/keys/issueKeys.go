package keys

import "github.com/charmbracelet/bubbles/key"

type IssueKeyMap struct {
	Comment key.Binding
	Close   key.Binding
	Reopen  key.Binding
}

var IssueKeys = IssueKeyMap{
	Comment: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "comment"),
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

func IssueFullHelp() []key.Binding {
	return []key.Binding{
		IssueKeys.Comment,
		IssueKeys.Close,
		IssueKeys.Reopen,
	}
}
