package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const StateFileName = "state.yml"
const DEFAULT_XDG_STATE_DIRNAME = ".local/state"

// State holds runtime state that should persist between sessions
type State struct {
	PreviewWidth int `yaml:"previewWidth,omitempty"`
}

// GetStatePath returns the path to the state file
// State files are stored in $XDG_STATE_HOME (defaults to ~/.local/state)
func GetStatePath() (string, error) {
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		stateDir = filepath.Join(homeDir, DEFAULT_XDG_STATE_DIRNAME)
	}
	return filepath.Join(stateDir, DashDir, StateFileName), nil
}

// LoadState loads the state from the state file
func LoadState() (State, error) {
	state := State{}

	statePath, err := GetStatePath()
	if err != nil {
		return state, err
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil // No state file yet, return empty state
		}
		return state, err
	}

	err = yaml.Unmarshal(data, &state)
	return state, err
}

// SaveState saves the state to the state file
func SaveState(state State) error {
	statePath, err := GetStatePath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0644)
}

// SavePreviewWidth saves just the preview width to state
func SavePreviewWidth(width int) error {
	state, _ := LoadState() // Ignore error, start fresh if needed
	state.PreviewWidth = width
	return SaveState(state)
}
