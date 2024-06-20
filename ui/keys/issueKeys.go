package keys

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	log "github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/config"
)

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

func rebindIssueKeys(keys []config.Keybinding) error {
	for _, issueKey := range keys {
		if issueKey.Builtin == "" {
			continue
		}

		log.Debug("Rebinding issue key", "builtin", issueKey.Builtin, "key", issueKey.Key)

		var key *key.Binding

		switch issueKey.Builtin {
		case "assign":
			key = &IssueKeys.Assign
		case "unassign":
			key = &IssueKeys.Unassign
		case "comment":
			key = &IssueKeys.Comment
		case "close":
			key = &IssueKeys.Close
		case "reopen":
			key = &IssueKeys.Reopen
		case "viewPrs":
			key = &IssueKeys.ViewPRs
		default:
			return fmt.Errorf("unknown built-in issue key: '%s'", issueKey.Builtin)
		}

		key.SetKeys(issueKey.Key)
		key.SetHelp(issueKey.Key, key.Help().Desc)
	}

	return nil
}
