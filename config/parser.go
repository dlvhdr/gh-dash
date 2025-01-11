package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"

	"github.com/dlvhdr/gh-dash/v4/utils"
)

const DashDir = "gh-dash"

const ConfigYmlFileName = "config.yml"

const ConfigYamlFileName = "config.yaml"

const DEFAULT_XDG_CONFIG_DIRNAME = ".config"

var validate *validator.Validate

type ViewType string

const (
	PRsView    ViewType = "prs"
	IssuesView ViewType = "issues"
	RepoView   ViewType = "repo"
)

type SectionConfig struct {
	Title   string
	Filters string
	Limit   *int `yaml:"limit,omitempty"`
	Type    *ViewType
}

type PrsSectionConfig struct {
	Title   string
	Filters string
	Limit   *int            `yaml:"limit,omitempty"`
	Layout  PrsLayoutConfig `yaml:"layout,omitempty"`
	Type    *ViewType
}

type IssuesSectionConfig struct {
	Title   string
	Filters string
	Limit   *int               `yaml:"limit,omitempty"`
	Layout  IssuesLayoutConfig `yaml:"layout,omitempty"`
}

type PreviewConfig struct {
	Open  bool
	Width int
}

type ColumnConfig struct {
	Width  *int  `yaml:"width,omitempty"  validate:"omitempty,gt=0"`
	Hidden *bool `yaml:"hidden,omitempty"`
}

type PrsLayoutConfig struct {
	UpdatedAt    ColumnConfig `yaml:"updatedAt,omitempty"`
	Repo         ColumnConfig `yaml:"repo,omitempty"`
	Author       ColumnConfig `yaml:"author,omitempty"`
	Assignees    ColumnConfig `yaml:"assignees,omitempty"`
	Title        ColumnConfig `yaml:"title,omitempty"`
	Base         ColumnConfig `yaml:"base,omitempty"`
	ReviewStatus ColumnConfig `yaml:"reviewStatus,omitempty"`
	State        ColumnConfig `yaml:"state,omitempty"`
	Ci           ColumnConfig `yaml:"ci,omitempty"`
	Lines        ColumnConfig `yaml:"lines,omitempty"`
}

type IssuesLayoutConfig struct {
	UpdatedAt ColumnConfig `yaml:"updatedAt,omitempty"`
	State     ColumnConfig `yaml:"state,omitempty"`
	Repo      ColumnConfig `yaml:"repo,omitempty"`
	Title     ColumnConfig `yaml:"title,omitempty"`
	Creator   ColumnConfig `yaml:"creator,omitempty"`
	Assignees ColumnConfig `yaml:"assignees,omitempty"`
	Comments  ColumnConfig `yaml:"comments,omitempty"`
	Reactions ColumnConfig `yaml:"reactions,omitempty"`
}

type LayoutConfig struct {
	Prs    PrsLayoutConfig    `yaml:"prs,omitempty"`
	Issues IssuesLayoutConfig `yaml:"issues,omitempty"`
}

type Defaults struct {
	Preview                PreviewConfig `yaml:"preview"`
	PrsLimit               int           `yaml:"prsLimit"`
	IssuesLimit            int           `yaml:"issuesLimit"`
	View                   ViewType      `yaml:"view"`
	Layout                 LayoutConfig  `yaml:"layout,omitempty"`
	RefetchIntervalMinutes int           `yaml:"refetchIntervalMinutes,omitempty"`
	DateFormat             string        `yaml:"dateFormat,omitempty"`
}

type RepoConfig struct {
	BranchesRefetchIntervalSeconds int `yaml:"branchesRefetchIntervalSeconds,omitempty"`
	PrsRefetchIntervalSeconds      int `yaml:"prsRefetchIntervalSeconds,omitempty"`
}

type Keybinding struct {
	Key     string `yaml:"key"`
	Command string `yaml:"command"`
	Builtin string `yaml:"builtin"`
}

func (kb Keybinding) NewBinding(previous *key.Binding) key.Binding {
	helpDesc := ""
	if previous != nil {
		helpDesc = previous.Help().Desc
	}

	return key.NewBinding(
		key.WithKeys(kb.Key),
		key.WithHelp(kb.Key, helpDesc),
	)
}

type Keybindings struct {
	Universal []Keybinding `yaml:"universal"`
	Issues    []Keybinding `yaml:"issues"`
	Prs       []Keybinding `yaml:"prs"`
	Branches  []Keybinding `yaml:"branches"`
}

type Pager struct {
	Diff string `yaml:"diff"`
}

type HexColor string

type ColorThemeText struct {
	Primary   HexColor `yaml:"primary"   validate:"omitempty,hexcolor"`
	Secondary HexColor `yaml:"secondary" validate:"omitempty,hexcolor"`
	Inverted  HexColor `yaml:"inverted"  validate:"omitempty,hexcolor"`
	Faint     HexColor `yaml:"faint"     validate:"omitempty,hexcolor"`
	Warning   HexColor `yaml:"warning"   validate:"omitempty,hexcolor"`
	Success   HexColor `yaml:"success"   validate:"omitempty,hexcolor"`
	Error     HexColor `yaml:"error"     validate:"omitempty,hexcolor"`
}

type ColorThemeBorder struct {
	Primary   HexColor `yaml:"primary"   validate:"omitempty,hexcolor"`
	Secondary HexColor `yaml:"secondary" validate:"omitempty,hexcolor"`
	Faint     HexColor `yaml:"faint"     validate:"omitempty,hexcolor"`
}

type ColorThemeBackground struct {
	Selected HexColor `yaml:"selected" validate:"omitempty,hexcolor"`
}

type ColorTheme struct {
	Text       ColorThemeText       `yaml:"text"       validate:"required"`
	Background ColorThemeBackground `yaml:"background" validate:"required"`
	Border     ColorThemeBorder     `yaml:"border"     validate:"required"`
}

type ColorThemeConfig struct {
	Inline ColorTheme `yaml:",inline"`
}

type TableUIThemeConfig struct {
	ShowSeparator bool `yaml:"showSeparator" default:"true"`
	Compact       bool `yaml:"compact" default:"false"`
}

type UIThemeConfig struct {
	SectionsShowCount bool               `yaml:"sectionsShowCount" default:"true"`
	Table             TableUIThemeConfig `yaml:"table"`
}

type ThemeConfig struct {
	Ui     UIThemeConfig     `yaml:"ui,omitempty"     validate:"omitempty"`
	Colors *ColorThemeConfig `yaml:"colors,omitempty" validate:"omitempty"`
}

type Config struct {
	PRSections     []PrsSectionConfig    `yaml:"prSections"`
	IssuesSections []IssuesSectionConfig `yaml:"issuesSections"`
	Repo           RepoConfig            `yaml:"repo"`
	Defaults       Defaults              `yaml:"defaults"`
	Keybindings    Keybindings           `yaml:"keybindings"`
	RepoPaths      map[string]string     `yaml:"repoPaths"`
	Theme          *ThemeConfig          `yaml:"theme,omitempty" validate:"omitempty"`
	Pager          Pager                 `yaml:"pager"`
	ConfirmQuit    bool                  `yaml:"confirmQuit"`
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
			PrsLimit:               20,
			IssuesLimit:            20,
			View:                   PRsView,
			RefetchIntervalMinutes: 30,
			Layout: LayoutConfig{
				Prs: PrsLayoutConfig{
					UpdatedAt: ColumnConfig{
						Width: utils.IntPtr(lipgloss.Width("2mo  ")),
					},
					Repo: ColumnConfig{
						Width: utils.IntPtr(20),
					},
					Author: ColumnConfig{
						Width: utils.IntPtr(15),
					},
					Assignees: ColumnConfig{
						Width:  utils.IntPtr(20),
						Hidden: utils.BoolPtr(true),
					},
					Base: ColumnConfig{
						Width:  utils.IntPtr(15),
						Hidden: utils.BoolPtr(true),
					},
					Lines: ColumnConfig{
						Width: utils.IntPtr(lipgloss.Width(" +31.4k -31.6k ")),
					},
				},
				Issues: IssuesLayoutConfig{
					UpdatedAt: ColumnConfig{
						Width: utils.IntPtr(lipgloss.Width("2mo  ")),
					},
					Repo: ColumnConfig{
						Width: utils.IntPtr(15),
					},
					Creator: ColumnConfig{
						Width: utils.IntPtr(10),
					},
					Assignees: ColumnConfig{
						Width:  utils.IntPtr(20),
						Hidden: utils.BoolPtr(true),
					},
				},
			},
		},
		Repo: RepoConfig{
			BranchesRefetchIntervalSeconds: 30,
			PrsRefetchIntervalSeconds:      60,
		},
		PRSections: []PrsSectionConfig{
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
		IssuesSections: []IssuesSectionConfig{
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
			Universal: []Keybinding{},
			Issues:    []Keybinding{},
			Prs:       []Keybinding{},
		},
		RepoPaths: map[string]string{},
		Theme: &ThemeConfig{
			Ui: UIThemeConfig{
				SectionsShowCount: true,
				Table: TableUIThemeConfig{
					ShowSeparator: true,
					Compact:       false,
				},
			},
		},
		ConfirmQuit: false,
	}
}

func (parser ConfigParser) getDefaultConfigYamlContents() string {
	defaultConfig := parser.getDefaultConfig()
	yaml, _ := yaml.Marshal(defaultConfig)

	return string(yaml)
}

func (e configError) Error() string {
	return fmt.Sprintf(
		`Couldn't find a config.yml or a config.yaml configuration file.
Create one under: %s

Example of a config.yml file:
%s

For more info, go to https://github.com/dlvhdr/gh-dash
press q to exit.

Original error: %v`,
		path.Join(e.configDir, DashDir, ConfigYmlFileName),
		string(e.parser.getDefaultConfigYamlContents()),
		e.err,
	)
}

func (parser ConfigParser) writeDefaultConfigContents(
	newConfigFile *os.File,
) error {
	_, err := newConfigFile.WriteString(parser.getDefaultConfigYamlContents())

	if err != nil {
		return err
	}

	return nil
}

func (parser ConfigParser) createConfigFileIfMissing(
	configFilePath string,
) error {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		newConfigFile, err := os.OpenFile(
			configFilePath,
			os.O_RDWR|os.O_CREATE|os.O_EXCL,
			0666,
		)
		if err != nil {
			return err
		}

		defer newConfigFile.Close()
		return parser.writeDefaultConfigContents(newConfigFile)
	}

	return nil
}

func (parser ConfigParser) getDefaultConfigFileOrCreateIfMissing(repoPath string) (string, error) {
	var configFilePath string
	ghDashConfig := os.Getenv("GH_DASH_CONFIG")

	// First try GH_DASH_CONFIG
	if ghDashConfig != "" {
		configFilePath = ghDashConfig
		// Then try to see if we're currently in a git repo
	} else if repoPath != "" {
		basename := repoPath + "/." + DashDir
		repoConfigYml := basename + ".yml"
		repoConfigYaml := basename + ".yaml"
		if _, err := os.Stat(repoConfigYml); err == nil {
			configFilePath = repoConfigYml
		} else if _, err := os.Stat(repoConfigYaml); err == nil {
			configFilePath = repoConfigYaml
		}
		if configFilePath != "" {
			return configFilePath, nil
		}
	}

	// Then fallback to global config
	if configFilePath == "" {
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			configDir = filepath.Join(homeDir, DEFAULT_XDG_CONFIG_DIRNAME)
		}

		dashConfigDir := filepath.Join(configDir, DashDir)
		configFilePath = filepath.Join(dashConfigDir, ConfigYmlFileName)
	}

	// Ensure directory exists before attempting to create file
	configDir := filepath.Dir(configFilePath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err = os.MkdirAll(configDir, os.ModePerm); err != nil {
			return "", configError{
				parser:    parser,
				configDir: configDir,
				err:       err,
			}
		}
	}

	if err := parser.createConfigFileIfMissing(configFilePath); err != nil {
		return "", configError{parser: parser, configDir: configDir, err: err}
	}

	return configFilePath, nil
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
	repoFF := IsFeatureEnabled(FF_REPO_VIEW)
	if config.Defaults.View == RepoView && !repoFF {
		config.Defaults.View = PRsView
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

func ParseConfig(path string, repoPath string) (Config, error) {
	parser := initParser()

	var config Config
	var err error
	var configFilePath string

	if path == "" {
		configFilePath, err = parser.getDefaultConfigFileOrCreateIfMissing(repoPath)
		if err != nil {
			return config, parsingError{err: err}
		}
	} else {
		configFilePath = path
	}

	config, err = parser.readConfigFile(configFilePath)
	if err != nil {
		return config, parsingError{err: err}
	}

	return config, nil
}
