package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type SectionConfig struct {
	Title   string
	Filters string
}

const PrsDir = "prs"
const ConfigFileName = "config.yml"

type configError struct {
	configDir string
	err       error
}

func (e configError) Error() string {
	return fmt.Sprintf(
		`Couldn't find a config.yml configuration file.
Create one under: %s

Example of a config.yml file:
  - title: My Pull Requests
    filters: author:@me
  - title: Needs My Review
    filters: review-requested:@me
	- title: Subscribed
		filters: -author:@me repo:cli/cli repo:charmbracelet/glamour repo:charmbracelet/lipgloss

For more info, go to https://github.com/dlvhdr/gh-prs
press q to exit.

Original error: %v`,
		path.Join(e.configDir, PrsDir, ConfigFileName),
		e.err,
	)
}

func ParseSectionsConfig() ([]SectionConfig, error) {
	var err error
	var sections []SectionConfig
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir, err = os.UserConfigDir()
		if err != nil {
			return sections, configError{configDir: configDir, err: err}
		}
	}

	data, err := os.ReadFile(filepath.Join(configDir, PrsDir, ConfigFileName))
	if err != nil {
		return sections, configError{configDir: configDir, err: err}
	}

	err = yaml.Unmarshal([]byte(data), &sections)
	if err != nil {
		return sections, fmt.Errorf("failed parsing config.yml: %w", err)
	}

	return sections, nil
}
