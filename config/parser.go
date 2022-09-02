package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

const DashDir = "gh-dash"
const ConfigFileName = "config.yml"
const DEFAULT_XDG_CONFIG_DIRNAME = ".config"

var validate *validator.Validate

type ViewType string

const (
	PRsView    ViewType = "prs"
	IssuesView ViewType = "issues"
)

type SectionConfig struct {
	Title   string
	Filters string
	Limit   *int `yaml:"limit,omitempty"`
}

type PreviewConfig struct {
	Open  bool
	Width int
}

type Defaults struct {
	Preview     PreviewConfig `yaml:"preview"`
	PrsLimit    int           `yaml:"prsLimit"`
	IssuesLimit int           `yaml:"issuesLimit"`
	View        ViewType      `yaml:"view"`
}

type Keybinding struct {
	Key     string `yaml:"key"`
	Command string `yaml:"command"`
}

type Keybindings struct {
	Prs []Keybinding `yaml:"prs"`
}

type Pager struct {
	Diff string `yaml:"diff"`
}

type HexColor string

type ColorThemeText struct {
	Primary   HexColor `yaml:"primary" validate:"hexcolor"`
	Secondary HexColor `yaml:"secondary" validate:"hexcolor"`
	Inverted  HexColor `yaml:"inverted" validate:"hexcolor"`
	Faint     HexColor `yaml:"faint" validate:"hexcolor"`
	Warning   HexColor `yaml:"warning" validate:"hexcolor"`
	Success   HexColor `yaml:"success" validate:"hexcolor"`
}

type ColorThemeBorder struct {
	Primary   HexColor `yaml:"primary" validate:"hexcolor"`
	Secondary HexColor `yaml:"secondary" validate:"hexcolor"`
	Faint     HexColor `yaml:"faint" validate:"hexcolor"`
}

type ColorThemeBackground struct {
	Selected HexColor `yaml:"selected" validate:"hexcolor"`
}

type ColorTheme struct {
	Text       ColorThemeText       `yaml:"text" validate:"required,dive"`
	Background ColorThemeBackground `yaml:"background" validate:"required,dive"`
	Border     ColorThemeBorder     `yaml:"border" validate:"required,dive"`
}

type ColorThemeConfig struct {
	Inline ColorTheme `yaml:",inline" validate:"dive"`
}

type ThemeConfig struct {
	Colors ColorThemeConfig `yaml:"colors,omitempty" validate:"dive"`
}

type Config struct {
	PRSections     []SectionConfig   `yaml:"prSections"`
	IssuesSections []SectionConfig   `yaml:"issuesSections"`
	Defaults       Defaults          `yaml:"defaults"`
	Keybindings    Keybindings       `yaml:"keybindings"`
	RepoPaths      map[string]string `yaml:"repoPaths"`
	Theme          *ThemeConfig      `yaml:"theme,omitempty" validate:"omitempty,dive"`
	Pager          Pager             `yaml:"pager"`
}

type configError struct {
	configDir string
	parser    ConfigParser
	err       error
}

type ConfigParser struct{}

func (parser ConfigParser) getDefaultConfig() Config {
	return Config{
		Defaults: Defaults{
			Preview: PreviewConfig{
				Open:  true,
				Width: 50,
			},
			PrsLimit:    20,
			IssuesLimit: 20,
			View:        PRsView,
		},
		PRSections: []SectionConfig{
			{
				Title:   "My Pull Requests",
				Filters: "is:open author:@me",
			},
			{
				Title:   "Needs My Review",
				Filters: "is:open review-requested:@me",
			},
			{
				Title:   "Involved",
				Filters: "is:open involves:@me -author:@me",
			},
		},
		IssuesSections: []SectionConfig{
			{
				Title:   "My Issues",
				Filters: "is:open author:@me",
			},
			{
				Title:   "Assigned",
				Filters: "is:open assignee:@me",
			},
			{
				Title:   "Involved",
				Filters: "is:open involves:@me -author:@me",
			},
		},
		Keybindings: Keybindings{
			Prs: []Keybinding{},
		},
		RepoPaths: map[string]string{},
	}
}

func (parser ConfigParser) getDefaultConfigYamlContents() string {
	defaultConfig := parser.getDefaultConfig()
	yaml, _ := yaml.Marshal(defaultConfig)

	return string(yaml)
}

func (e configError) Error() string {
	return fmt.Sprintf(
		`Couldn't find a config.yml configuration file.
Create one under: %s

Example of a config.yml file:
%s

For more info, go to https://github.com/dlvhdr/gh-dash
press q to exit.

Original error: %v`,
		path.Join(e.configDir, DashDir, ConfigFileName),
		string(e.parser.getDefaultConfigYamlContents()),
		e.err,
	)
}

func (parser ConfigParser) writeDefaultConfigContents(newConfigFile *os.File) error {
	_, err := newConfigFile.WriteString(parser.getDefaultConfigYamlContents())

	if err != nil {
		return err
	}

	return nil
}

func (parser ConfigParser) createConfigFileIfMissing(configFilePath string) error {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		newConfigFile, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			return err
		}

		defer newConfigFile.Close()
		return parser.writeDefaultConfigContents(newConfigFile)
	}

	return nil
}

func (parser ConfigParser) getExistingConfigFile() (*string, error) {
	var err error
	var dashConfigFile string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	xdgConfigDir := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigDir == "" {
		xdgConfigDir = filepath.Join(homeDir, DEFAULT_XDG_CONFIG_DIRNAME)
	}

	dashConfigFile = filepath.Join(xdgConfigDir, DashDir, ConfigFileName)
	if _, err := os.Stat(dashConfigFile); err == nil {
		return &dashConfigFile, nil
	}

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	dashConfigFile = filepath.Join(userConfigDir, DashDir, ConfigFileName)
	if _, err := os.Stat(dashConfigFile); err == nil {
		return &dashConfigFile, nil
	}

	return nil, nil
}

func (parser ConfigParser) getConfigFileOrCreateIfMissing() (*string, error) {
	var err error

	existingConfigFile, err := parser.getExistingConfigFile()
	if err != nil {
		return nil, err
	}
	if existingConfigFile != nil {
		return existingConfigFile, nil
	}

	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		configDir = filepath.Join(homeDir, DEFAULT_XDG_CONFIG_DIRNAME)
	}

	dashConfigDir := filepath.Join(configDir, DashDir)
	err = os.MkdirAll(dashConfigDir, os.ModePerm)
	if err != nil {
		return nil, configError{parser: parser, configDir: configDir, err: err}
	}

	configFilePath := filepath.Join(dashConfigDir, ConfigFileName)
	err = parser.createConfigFileIfMissing(configFilePath)
	if err != nil {
		return nil, configError{parser: parser, configDir: configDir, err: err}
	}

	return &configFilePath, nil
}

type parsingError struct {
	err error
}

func (e parsingError) Error() string {
	return fmt.Sprintf("failed parsing config.yml: %v", e.err)
}

func (parser ConfigParser) readConfigFile(path string) (Config, error) {
	config := parser.getDefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return config, configError{parser: parser, configDir: path, err: err}
	}

	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		return config, err
	}

	err = validate.Struct(config)
	return config, err
}

func initParser() ConfigParser {
	validate = validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.Split(fld.Tag.Get("yaml"), ",")[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return ConfigParser{}
}

func ParseConfig() (Config, error) {
	parser := initParser()

	var config Config
	var err error

	configFilePath, err := parser.getConfigFileOrCreateIfMissing()
	if err != nil {
		return config, parsingError{err: err}
	}

	config, err = parser.readConfigFile(*configFilePath)
	if err != nil {
		return config, parsingError{err: err}
	}

	return config, nil
}
