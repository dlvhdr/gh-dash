package keys

import "github.com/charmbracelet/bubbles/key"

type IssueKeyMap struct {
	Comment key.Binding
}

var IssueKeys = IssueKeyMap{
	Comment: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "comment"),
	),
}

func IssueFullHelp() []key.Binding {
	return []key.Binding{
		IssueKeys.Comment,
	}
}
