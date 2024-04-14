package keys

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/dlvhdr/gh-dash/config"
)

type PRKeyMap struct {
	Assign      key.Binding
	Unassign    key.Binding
	Comment     key.Binding
	Diff        key.Binding
	Checkout    key.Binding
	Close       key.Binding
	Ready       key.Binding
	Reopen      key.Binding
	Merge       key.Binding
	WatchChecks key.Binding
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
		key.WithKeys("W"),
		key.WithHelp("W", "ready for review"),
	),
	Merge: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "merge"),
	),
	WatchChecks: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "Watch checks"),
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
		PRKeys.WatchChecks,
	}
}

func rebindPRKeys(keys []config.Keybinding) error {
	for _, prKey := range keys {
		var k key.Binding

		switch prKey.Builtin {
		case "assign":
			k = PRKeys.Assign
		case "unassign":
			k = PRKeys.Unassign
		case "comment":
			k = PRKeys.Comment
		case "diff":
			k = PRKeys.Diff
		case "checkout":
			k = PRKeys.Checkout
		case "close":
			k = PRKeys.Close
		case "ready":
			k = PRKeys.Ready
		case "reopen":
			k = PRKeys.Reopen
		case "merge":
			k = PRKeys.Merge
		case "watchchecks":
			k = PRKeys.WatchChecks
		default:
			// TODO: return an error here
			return nil
		}

		// Not really sure if this is the best idea but I am not
		// sure how else we are meant to define alt keys.
		if len(prKey.Keys) > 0 {
			k.SetKeys(prKey.Keys...)
		} else {
			k.SetKeys(prKey.Key)
		}
	}

	return nil
}
