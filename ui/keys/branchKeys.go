package keys

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	log "github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/config"
)

type BranchKeyMap struct {
	Checkout    key.Binding
	FastForward key.Binding
	Push        key.Binding
	Delete      key.Binding
	ViewPRs     key.Binding
}

var BranchKeys = BranchKeyMap{
	Checkout: key.NewBinding(
		key.WithKeys("C", " "),
		key.WithHelp("C/space", "checkout"),
	),
	FastForward: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "fast-forward"),
	),
	Push: key.NewBinding(
		key.WithKeys("P"),
		key.WithHelp("P", "push"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d", "backspace"),
		key.WithHelp("d/backspace", "delete"),
	),
	ViewPRs: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "Switch to PRs"),
	),
}

func BranchFullHelp() []key.Binding {
	return []key.Binding{
		BranchKeys.Checkout,
		BranchKeys.FastForward,
		BranchKeys.Push,
		BranchKeys.Delete,
		BranchKeys.ViewPRs,
	}
}

func rebindBranchKeys(keys []config.Keybinding) error {
	for _, branchKey := range keys {
		if branchKey.Builtin == "" {
			continue
		}

		log.Debug("Rebinding branch key", "builtin", branchKey.Builtin, "key", branchKey.Key)

		var key *key.Binding

		switch branchKey.Builtin {
		case "checkout":
			key = &BranchKeys.Checkout
		case "viewPRs":
			key = &BranchKeys.ViewPRs
		default:
			return fmt.Errorf("unknown built-in branch key: '%s'", branchKey.Builtin)
		}

		key.SetKeys(branchKey.Key)
		key.SetHelp(branchKey.Key, key.Help().Desc)
	}

	return nil
}
