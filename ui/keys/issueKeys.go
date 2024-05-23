package keys

import "github.com/charmbracelet/bubbles/key"

type IssueKeyMap struct {
	Assign   key.Binding
	Unassign key.Binding
	Comment  key.Binding
	Close    key.Binding
	Reopen   key.Binding
	ViewPRs  key.Binding
}

var IssueKeys = IssueKeyMap{
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
	Close: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "close"),
	),
	Reopen: key.NewBinding(
		key.WithKeys("X"),
		key.WithHelp("X", "reopen"),
	),
	ViewPRs: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "switch to PRs"),
	),
}

func IssueFullHelp() []key.Binding {
	return []key.Binding{
		IssueKeys.Assign,
		IssueKeys.Unassign,
		IssueKeys.Comment,
		IssueKeys.Close,
		IssueKeys.Reopen,
		IssueKeys.ViewPRs,
	}
}
