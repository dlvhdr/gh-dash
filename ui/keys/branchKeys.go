package keys

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	log "github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
)

type BranchKeyMap struct {
	Checkout    key.Binding
	New         key.Binding
	CreatePr    key.Binding
	FastForward key.Binding
	Push        key.Binding
	ForcePush   key.Binding
	Delete      key.Binding
	UpdatePr    key.Binding
	ViewPRs     key.Binding
}

var BranchKeys = BranchKeyMap{
	Checkout: key.NewBinding(
		key.WithKeys("C", " "),
		key.WithHelp("C/space", "checkout"),
	),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new"),
	),
	CreatePr: key.NewBinding(
		key.WithKeys("O"),
		key.WithHelp("O", "create PR"),
	),
	FastForward: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "fast-forward"),
	),
	Push: key.NewBinding(
		key.WithKeys("P"),
		key.WithHelp("P", "push"),
	),
	ForcePush: key.NewBinding(
		key.WithKeys("F"),
		key.WithHelp("F", "force-push"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d", "backspace"),
		key.WithHelp("d/backspace", "delete"),
	),
	UpdatePr: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "update PR"),
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
		BranchKeys.ForcePush,
		BranchKeys.New,
		BranchKeys.CreatePr,
		BranchKeys.Delete,
		BranchKeys.UpdatePr,
		BranchKeys.ViewPRs,
	}
}

func rebindBranchKeys(keys []config.Keybinding) error {
	CustomBranchBindings = []key.Binding{}

	for _, branchKey := range keys {
		if branchKey.Builtin == "" {
			// Handle custom commands
			if branchKey.Command != "" {
				name := branchKey.Name
				if branchKey.Name == "" {
					name = config.TruncateCommand(branchKey.Command)
				}

				customBinding := key.NewBinding(
					key.WithKeys(branchKey.Key),
					key.WithHelp(branchKey.Key, name),
				)

				CustomBranchBindings = append(CustomBranchBindings, customBinding)
			}
			continue
		}

		log.Debug("Rebinding branch key", "builtin", branchKey.Builtin, "key", branchKey.Key)

		var key *key.Binding

		switch branchKey.Builtin {
		case "new":
			key = &BranchKeys.New
		case "createPr":
			key = &BranchKeys.CreatePr
		case "delete":
			key = &BranchKeys.Delete
		case "push":
			key = &BranchKeys.Push
		case "forcePush":
			key = &BranchKeys.ForcePush
		case "fastForward":
			key = &BranchKeys.FastForward
		case "checkout":
			key = &BranchKeys.Checkout
		case "viewPRs":
			key = &BranchKeys.ViewPRs
		case "updatePr":
			key = &BranchKeys.UpdatePr
		default:
			return fmt.Errorf("unknown built-in branch key: '%s'", branchKey.Builtin)
		}

		key.SetKeys(branchKey.Key)

		helpDesc := key.Help().Desc
		if branchKey.Name != "" {
			helpDesc = branchKey.Name
		}
		key.SetHelp(branchKey.Key, helpDesc)
	}

	return nil
}
