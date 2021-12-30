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

const DefaultConfigContents = `- title: My Pull Requests
  filters: is:open author:@me
- title: Needs My Review
  filters: is:open review-requested:@me
- title: Subscribed
  filters: is:open -author:@me repo:cli/cli repo:dlvhdr/gh-prs`

func (e configError) Error() string {
	return fmt.Sprintf(
		`Couldn't find a config.yml configuration file.
Create one under: %s

Example of a config.yml file:
%s

For more info, go to https://github.com/dlvhdr/gh-prs
press q to exit.

Original error: %v`,
		path.Join(e.configDir, PrsDir, ConfigFileName),
		DefaultConfigContents,
		e.err,
	)
}

func writeDefaultConfigContents(newConfigFile *os.File) error {
	_, err := newConfigFile.WriteString(DefaultConfigContents)

	if err != nil {
		return err
	}

	return nil
}

func createConfigFileIfMissing(configFilePath string) error {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		newConfigFile, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			return err
		}

		defer newConfigFile.Close()
		return writeDefaultConfigContents(newConfigFile)
	}

	return nil
}

func getConfigFileOrCreateIfMissing() (*string, error) {
	var err error
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir, err = os.UserConfigDir()
		if err != nil {
			return nil, configError{configDir: configDir, err: err}
		}
	}

	prsConfigDir := filepath.Join(configDir, PrsDir)
	err = os.MkdirAll(prsConfigDir, os.ModePerm)
	if err != nil {
		return nil, configError{configDir: configDir, err: err}
	}

	configFilePath := filepath.Join(prsConfigDir, ConfigFileName)
	err = createConfigFileIfMissing(configFilePath)
	if err != nil {
		return nil, configError{configDir: configDir, err: err}
	}

	return &configFilePath, nil
}

type parsingError struct {
	err error
}

func (e parsingError) Error() string {
	return fmt.Sprintf("failed parsing config.yml: %v", e.err)
}

func readConfigFileSections(path string) ([]SectionConfig, error) {
	var sections []SectionConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return sections, configError{configDir: path, err: err}
	}

	err = yaml.Unmarshal([]byte(data), &sections)
	if err != nil {
		return sections, parsingError{err: err}
	}

	return sections, nil
}

func ParseConfig() ([]SectionConfig, error) {
	var err error
	configFilePath, err := getConfigFileOrCreateIfMissing()
	if err != nil {
		return nil, parsingError{err: err}
	}

	sections, err := readConfigFileSections(*configFilePath)
	if err != nil {
		return nil, parsingError{err: err}
	}

	return sections, nil
}
