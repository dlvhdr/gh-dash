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
	Repos   []string
}

const PrsDir = "prs"
const SectionsFileName = "sections.yml"

type configError struct {
	configDir string
	err       error
}

func (e configError) Error() string {
	return fmt.Sprintf(
		`Couldn't find a sections.yml configuration file.
Create one under: %s

Example of a sections.yml file:
  - title: My Pull Requests
    repos:
      - dlvhdr/gh-prs
    filters: author:@me
  - title: Needs My Review
    repos:
      - dlvhdr/gh-prs
    filters: assignee:@me
  - title: Subscribed
    repos:
      - cli/cli
      - charmbracelet/glamour
      - charmbracelet/lipgloss
    filters: -author:@me

For more info, go to https://github.com/dlvhdr/gh-prs
press q to exit.

Original error: %v`,
		path.Join(e.configDir, PrsDir, SectionsFileName),
		e.err,
	)
}

func ParseSectionsConfig() ([]SectionConfig, error) {
	var sections []SectionConfig
	configDir := os.Getenv("XDG_CONFIG_HOME")
	var err error
	if configDir == "" {
		configDir, err = os.UserConfigDir()
		if err != nil {
			return sections, configError{configDir: configDir, err: err}
		}
	}

	data, err := os.ReadFile(filepath.Join(configDir, PrsDir, SectionsFileName))
	if err != nil {
		return sections, configError{configDir: configDir, err: err}
	}

	err = yaml.Unmarshal([]byte(data), &sections)
	if err != nil {
		return sections, fmt.Errorf("Failed parsing sections.yml: %w", err)
	}

	return sections, nil
}
