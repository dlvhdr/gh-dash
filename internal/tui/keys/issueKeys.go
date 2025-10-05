package keys

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	log "github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
)

type IssueKeyMap struct {
	Label                key.Binding
	Assign               key.Binding
	Unassign             key.Binding
	Comment              key.Binding
	Close                key.Binding
	Reopen               key.Binding
	ToggleSmartFiltering key.Binding
	ViewPRs              key.Binding
}

var IssueKeys = IssueKeyMap{
	Label: key.NewBinding(
		key.WithKeys("L"),
		key.WithHelp("L", "label"),
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
	ToggleSmartFiltering: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "toggle smart filtering"),
	),
	ViewPRs: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "switch to PRs"),
	),
}

func IssueFullHelp() []key.Binding {
	return []key.Binding{
		IssueKeys.Label,
		IssueKeys.Assign,
		IssueKeys.Unassign,
		IssueKeys.Comment,
		IssueKeys.Close,
		IssueKeys.Reopen,
		IssueKeys.ToggleSmartFiltering,
		IssueKeys.ViewPRs,
	}
}

func rebindIssueKeys(keys []config.Keybinding) error {
	CustomIssueBindings = []key.Binding{}

	for _, issueKey := range keys {
		if issueKey.Builtin == "" {
			// Handle custom commands
			if issueKey.Command != "" {
				name := issueKey.Name
				if issueKey.Name == "" {
					name = config.TruncateCommand(issueKey.Command)
				}

				customBinding := key.NewBinding(
					key.WithKeys(issueKey.Key),
					key.WithHelp(issueKey.Key, name),
				)

				CustomIssueBindings = append(CustomIssueBindings, customBinding)
			}
			continue
		}

		log.Debug("Rebinding issue key", "builtin", issueKey.Builtin, "key", issueKey.Key)

		var key *key.Binding

		switch issueKey.Builtin {
		case "label":
			key = &IssueKeys.Label
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

		helpDesc := key.Help().Desc
		if issueKey.Name != "" {
			helpDesc = issueKey.Name
		}
		key.SetHelp(issueKey.Key, helpDesc)
	}

	return nil
}
