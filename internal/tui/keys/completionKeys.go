package keys

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	log "charm.land/log/v2"
	"github.com/dlvhdr/gh-dash/v4/internal/config"
)

type CmpKeyMap struct {
	NextKey               key.Binding
	PrevKey               key.Binding
	SelectKey             key.Binding
	RefreshSuggestionsKey key.Binding
	ToggleSuggestions     key.Binding
}

var CmpKeys = CmpKeyMap{
	NextKey: key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓/ctrl+n", "next"),
	),
	PrevKey: key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑/ctrl+p", "previous"),
	),
	SelectKey: key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("ctrl+y", "select"),
	),
	RefreshSuggestionsKey: key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("ctrl+f", "refresh"),
	),
	ToggleSuggestions: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "toggle"),
	),
}

func rebindCmpKeys(keys []config.Keybinding) error {
	CustomCmpBindings = []key.Binding{}

	for _, cmpKey := range keys {
		if cmpKey.Builtin == "" {
			// Handle custom commands
			if cmpKey.Command != "" {
				name := cmpKey.Name
				if name == "" {
					name = config.TruncateCommand(cmpKey.Command)
				}

				customBinding := key.NewBinding(
					key.WithKeys(cmpKey.Key),
					key.WithHelp(cmpKey.Key, name),
				)

				CustomCmpBindings = append(CustomCmpBindings, customBinding)
			}
			continue
		}
		log.Debug("Rebinding Completions Key", "builtin", cmpKey.Builtin, "key", cmpKey.Key)

		var key *key.Binding

		switch cmpKey.Builtin {
		case "nextSuggestion":
			key = &CmpKeys.NextKey
		case "previousSuggestion":
			key = &CmpKeys.PrevKey
		case "selectSuggestion":
			key = &CmpKeys.SelectKey
		case "refreshSuggestions":
			key = &CmpKeys.RefreshSuggestionsKey
		case "toggleSuggestions":
			key = &CmpKeys.ToggleSuggestions
		default:
			return fmt.Errorf("unknown builtin: '%s'", cmpKey.Builtin)
		}

		key.SetKeys(cmpKey.Key)

		helpDesc := key.Help().Desc
		if cmpKey.Name != "" {
			helpDesc = cmpKey.Name
		}
		key.SetHelp(cmpKey.Key, helpDesc)
	}

	return nil
}
