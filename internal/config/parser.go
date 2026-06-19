package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/maps"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	yamlmarshaller "gopkg.in/yaml.v3"

	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

var hexColorRegex = regexp.MustCompile(`^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$`)

var conf = koanf.Conf{
	Delim:       ".",
	StrictMerge: true,
}

const DashDir = "gh-dash"

const RepoConfigFileName = ".gh-dash.yml"

const ConfigYmlFileName = "config.yml"

// TODO: use this
const ConfigYamlFileName = "config.yaml"

const DEFAULT_XDG_CONFIG_DIRNAME = ".config"

var validate *validator.Validate

/* Stringer implementation for ViewType */
type ViewType string

func (vt ViewType) String() string {
	return string(vt)
}

func (vt ViewType) MarshalJSON() ([]byte, error) {
	return []byte(vt.String()), nil
}

func (a *ViewType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	case "notifications":
		*a = NotificationsView
	case "prs":
		*a = PRsView
	case "issues":
		*a = IssuesView
	case "repo":
		*a = RepoView
	}

	return nil
}

const (
	NotificationsView ViewType = "notifications"
	PRsView           ViewType = "prs"
	IssuesView        ViewType = "issues"
	RepoView          ViewType = "repo"
)

type SectionConfig struct {
	Title   string
	Filters string
	Limit   *int      `yaml:"limit,omitempty"`
	Type    *ViewType `yaml:"type,omitempty"`
}

type PrsSectionConfig struct {
	Title   string
	Filters string
	Limit   *int            `yaml:"limit,omitempty"`
	Layout  PrsLayoutConfig `yaml:"layout,omitempty"`
	Type    *ViewType       `yaml:"type,omitempty"`
}

type IssuesSectionConfig struct {
	Title   string
	Filters string
	Limit   *int               `yaml:"limit,omitempty"`
	Layout  IssuesLayoutConfig `yaml:"layout,omitempty"`
}

type NotificationsSectionConfig struct {
	Title   string
	Filters string
	Limit   *int `yaml:"limit,omitempty"`
}

type PreviewConfig struct {
	Open     bool
	Width    float64 `yaml:"width"              validate:"gt=0"`
	Height   float64 `yaml:"height,omitempty"`
	Position string  `yaml:"position,omitempty"`
}

type NullableBool struct {
	Value *bool
}

func (nb NullableBool) MarshalJSON() ([]byte, error) {
	log.Error("marshalling", "nb", nb)
	if nb.Value != nil {
		return json.Marshal(nb.Value)
	}

	return json.Marshal(false)
}

func (nullBool *NullableBool) UnmarshalJSON(b []byte) error {
	var unmarshalledJson bool
	nb := NullableBool{}
	if nullBool == nil {
		return nil
	}

	err := json.Unmarshal(b, &unmarshalledJson)
	if err != nil {
		return err
	}
	nb.Value = &unmarshalledJson
	*nullBool = nb
	return nil
}

type ColumnConfig struct {
	Width  *int  `yaml:"width,omitempty"  validate:"omitempty,gt=0"`
	Hidden *bool `yaml:"hidden,omitempty"`
}

type PrsLayoutConfig struct {
	UpdatedAt    ColumnConfig `yaml:"updatedAt,omitempty"`
	CreatedAt    ColumnConfig `yaml:"createdAt,omitempty"`
	Repo         ColumnConfig `yaml:"repo,omitempty"`
	Author       ColumnConfig `yaml:"author,omitempty"`
	AuthorIcon   ColumnConfig `yaml:"authorIcon,omitempty"`
	Labels       ColumnConfig `yaml:"labels,omitempty"`
	Assignees    ColumnConfig `yaml:"assignees,omitempty"`
	Title        ColumnConfig `yaml:"title,omitempty"`
	Base         ColumnConfig `yaml:"base,omitempty"`
	ReviewStatus ColumnConfig `yaml:"reviewStatus,omitempty"`
	State        ColumnConfig `yaml:"state,omitempty"`
	Ci           ColumnConfig `yaml:"ci,omitempty"`
	Lines        ColumnConfig `yaml:"lines,omitempty"`
	NumComments  ColumnConfig `yaml:"numComments,omitempty"`
}

type IssuesLayoutConfig struct {
	UpdatedAt   ColumnConfig `yaml:"updatedAt,omitempty"`
	CreatedAt   ColumnConfig `yaml:"createdAt,omitempty"`
	State       ColumnConfig `yaml:"state,omitempty"`
	Repo        ColumnConfig `yaml:"repo,omitempty"`
	Title       ColumnConfig `yaml:"title,omitempty"`
	Creator     ColumnConfig `yaml:"creator,omitempty"`
	CreatorIcon ColumnConfig `yaml:"creatorIcon,omitempty"`
	Assignees   ColumnConfig `yaml:"assignees,omitempty"`
	Comments    ColumnConfig `yaml:"comments,omitempty"`
	Reactions   ColumnConfig `yaml:"reactions,omitempty"`
}

type LayoutConfig struct {
	Prs    PrsLayoutConfig    `yaml:"prs,omitempty"`
	Issues IssuesLayoutConfig `yaml:"issues,omitempty"`
}

type Defaults struct {
	Preview                PreviewConfig `yaml:"preview"`
	PrsLimit               int           `yaml:"prsLimit"`
	PrApproveComment       string        `yaml:"prApproveComment,omitempty"`
	IssuesLimit            int           `yaml:"issuesLimit"`
	NotificationsLimit     int           `yaml:"notificationsLimit"`
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
	Command string `yaml:"command,omitempty"`
	Builtin string `yaml:"builtin,omitempty"`
	Name    string `yaml:"name,omitempty"`
}

func (kb Keybinding) NewBinding(previous *key.Binding) key.Binding {
	helpDesc := ""
	if previous != nil {
		helpDesc = previous.Help().Desc
	}

	if kb.Name != "" {
		helpDesc = kb.Name
	}

	return key.NewBinding(
		key.WithKeys(kb.Key),
		key.WithHelp(kb.Key, helpDesc),
	)
}

type Keybindings struct {
	Universal     []Keybinding `yaml:"universal,omitempty"`
	Issues        []Keybinding `yaml:"issues,omitempty"`
	Prs           []Keybinding `yaml:"prs,omitempty"`
	Branches      []Keybinding `yaml:"branches,omitempty"`
	Notifications []Keybinding `yaml:"notifications,omitempty"`
	Cmp           []Keybinding `yaml:"completions,omitempty"`
}

type Pager struct {
	Diff string `yaml:"diff"`
}

type Color string

func (c Color) String() string {
	return string(c)
}

func (c Color) IsZero() bool {
	return c.String() == ""
}

type ColorThemeIcon struct {
	NewContributor Color `yaml:"newcontributor" validate:"omitempty,color"`
	Contributor    Color `yaml:"contributor"    validate:"omitempty,color"`
	Collaborator   Color `yaml:"collaborator"   validate:"omitempty,color"`
	Member         Color `yaml:"member"         validate:"omitempty,color"`
	Owner          Color `yaml:"owner"          validate:"omitempty,color"`
	UnknownRole    Color `yaml:"unknownrole"    validate:"omitempty,color"`
}

type ColorThemeText struct {
	Primary   Color `yaml:"primary,omitzero,omitempty" validate:"omitzero,omitempty,color"`
	Secondary Color `yaml:"secondary"                  validate:"omitempty,color"`
	Inverted  Color `yaml:"inverted"                   validate:"omitempty,color"`
	Faint     Color `yaml:"faint"                      validate:"omitempty,color"`
	Warning   Color `yaml:"warning"                    validate:"omitempty,color"`
	Success   Color `yaml:"success"                    validate:"omitempty,color"`
	Error     Color `yaml:"error"                      validate:"omitempty,color"`
	Actor     Color `yaml:"actor"                      validate:"omitempty,color"`
}

type ColorThemeBorder struct {
	Primary   Color `yaml:"primary"   validate:"omitempty,color"`
	Secondary Color `yaml:"secondary" validate:"omitempty,color"`
	Faint     Color `yaml:"faint"     validate:"omitempty,color"`
}

type ColorThemeBackground struct {
	Selected Color `yaml:"selected" validate:"omitempty,color"`
}

type ColorTheme struct {
	Icon       ColorThemeIcon       `yaml:"icon,omitempty"       validate:"required,omitempty"`
	Text       ColorThemeText       `yaml:"text,omitempty"       validate:"required,omitempty"`
	Background ColorThemeBackground `yaml:"background,omitempty" validate:"required,omitempty"`
	Border     ColorThemeBorder     `yaml:"border,omitempty"     validate:"required,omitempty"`
}

type ColorThemeConfig struct {
	Inline ColorTheme `yaml:",inline,squash"`
}

type IconTheme struct {
	NewContributor string `yaml:"newcontributor,omitempty"`
	Contributor    string `yaml:"contributor,omitempty"`
	Collaborator   string `yaml:"collaborator,omitempty"`
	Member         string `yaml:"member,omitempty"`
	Owner          string `yaml:"owner,omitempty"`
	UnknownRole    string `yaml:"unknownrole,omitempty"`
}

type IconThemeConfig struct {
	Inline IconTheme `yaml:",inline"`
}

type TableUIThemeConfig struct {
	ShowSeparator bool `yaml:"showSeparator" default:"true"`
	Compact       bool `yaml:"compact"       default:"false"`
}

type UIThemeConfig struct {
	SectionsShowCount bool               `yaml:"sectionsShowCount" default:"true"`
	Table             TableUIThemeConfig `yaml:"table"`
}

type ThemeConfig struct {
	Ui     UIThemeConfig     `yaml:"ui,omitempty"     validate:"omitempty"`
	Colors *ColorThemeConfig `yaml:"colors,omitempty" validate:"omitempty"`
	Icons  *IconThemeConfig  `yaml:"icons,omitempty"  validate:"omitempty"`
}

type Config struct {
	Include                  []string                     `yaml:"include,omitempty"`
	PRSections               []PrsSectionConfig           `yaml:"prSections"`
	IssuesSections           []IssuesSectionConfig        `yaml:"issuesSections"`
	NotificationsSections    []NotificationsSectionConfig `yaml:"notificationsSections"`
	Repo                     RepoConfig                   `yaml:"repo,omitempty"`
	Defaults                 Defaults                     `yaml:"defaults"`
	Keybindings              Keybindings                  `yaml:"keybindings"`
	RepoPaths                map[string]string            `yaml:"repoPaths"`
	Theme                    *ThemeConfig                 `yaml:"theme,omitempty"           validate:"omitempty"`
	Pager                    Pager                        `yaml:"pager"`
	ConfirmQuit              bool                         `yaml:"confirmQuit"`
	ShowAuthorIcons          bool                         `yaml:"showAuthorIcons,omitempty"`
	SmartFilteringAtLaunch   bool                         `yaml:"smartFilteringAtLaunch"                         default:"true"`
	IncludeReadNotifications bool                         `yaml:"includeReadNotifications"                       default:"true"`
}

type configError struct {
	configDir string
	parser    ConfigParser
	err       error
}

type ConfigParser struct {
	k *koanf.Koanf
}

func (parser ConfigParser) getDefaultConfig() Config {
	return Config{
		Defaults: Defaults{
			Preview: PreviewConfig{
				Open:     true,
				Width:    0.45,
				Height:   0.60,
				Position: "auto",
			},
			PrsLimit:               20,
			PrApproveComment:       "LGTM",
			IssuesLimit:            20,
			NotificationsLimit:     20,
			View:                   PRsView,
			RefetchIntervalMinutes: 30,
			Layout: LayoutConfig{
				Prs: PrsLayoutConfig{
					UpdatedAt: ColumnConfig{
						Width: utils.IntPtr(lipgloss.Width("2mo  ")),
					},
					CreatedAt: ColumnConfig{
						Width: utils.IntPtr(lipgloss.Width("2mo  ")),
					},
					Repo: ColumnConfig{
						Width: utils.IntPtr(20),
					},
					Author: ColumnConfig{
						Width: utils.IntPtr(15),
					},
					AuthorIcon: ColumnConfig{
						Hidden: utils.BoolPtr(false),
					},
					Labels: ColumnConfig{
						Width:  utils.IntPtr(22),
						Hidden: utils.BoolPtr(true),
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
					CreatedAt: ColumnConfig{
						Width: utils.IntPtr(lipgloss.Width("2mo  ")),
					},
					Repo: ColumnConfig{
						Width: utils.IntPtr(15),
					},
					Creator: ColumnConfig{
						Width: utils.IntPtr(10),
					},
					CreatorIcon: ColumnConfig{
						Hidden: utils.BoolPtr(false),
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
		NotificationsSections: []NotificationsSectionConfig{
			{
				Title:   "All",
				Filters: "",
			},
			{
				Title:   "Created",
				Filters: "reason:author",
			},
			{
				Title:   "Participating",
				Filters: "reason:participating",
			},
			{
				Title:   "Mentioned",
				Filters: "reason:mention",
			},
			{
				Title:   "Review Requested",
				Filters: "reason:review-requested",
			},
			{
				Title:   "Assigned",
				Filters: "reason:assign",
			},
			{
				Title:   "Subscribed",
				Filters: "reason:subscribed",
			},
			{
				Title:   "Team Mentioned",
				Filters: "reason:team-mention",
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
		ConfirmQuit:              false,
		ShowAuthorIcons:          true,
		SmartFilteringAtLaunch:   true,
		IncludeReadNotifications: true,
	}
}

func (parser ConfigParser) getDefaultConfigYamlContents() (string, error) {
	defaultConfig := parser.getDefaultConfig()
	log.Debug("loading default config yaml contents")

	b, err := yamlmarshaller.Marshal(defaultConfig)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (e configError) Error() string {
	content, err := e.parser.getDefaultConfigYamlContents()
	if err != nil {
		return fmt.Sprintf(
			"encountered error while trying to generate default config yaml contents: %v",
			err,
		)
	}
	return fmt.Sprintf(
		`Couldn't find a config.yml or a config.yaml configuration file.
Create one under: %s

Example of a config.yml file:
%s

For more info, go to https://github.com/dlvhdr/gh-dash
press q to exit.

Original error: %v`,
		path.Join(e.configDir, DashDir, ConfigYmlFileName),
		content,
		e.err,
	)
}

func (parser ConfigParser) writeDefaultConfigContents(
	newConfigFile *os.File,
) error {
	content, err := parser.getDefaultConfigYamlContents()
	if err != nil {
		return err
	}
	_, err = newConfigFile.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func (parser ConfigParser) createConfigFileIfMissing(
	configFilePath string,
) error {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		log.Info("default config doesn't exist - writing", "path", configFilePath, "err", err)

		newConfigFile, err := os.OpenFile(
			configFilePath,
			os.O_RDWR|os.O_CREATE|os.O_EXCL,
			0o666,
		)
		if err != nil {
			return err
		}

		defer newConfigFile.Close()
		return parser.writeDefaultConfigContents(newConfigFile)
	}

	return nil
}

func (parser ConfigParser) getGlobalConfigPathOrCreateIfMissing() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, DEFAULT_XDG_CONFIG_DIRNAME)
	}

	configFilePath := filepath.Join(configDir, DashDir, ConfigYmlFileName)
	log.Debug("using global config path", "path", configFilePath)

	// Ensure directory exists before attempting to create file
	configDir = filepath.Dir(configFilePath)
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

func (parser ConfigParser) getProvidedConfigPath(location Location) string {
	var userProvidedCfgPath string
	// First try the provided --config flag
	if location.ConfigFlag != "" {
		userProvidedCfgPath = location.ConfigFlag
	} else if cfg := os.Getenv("GH_DASH_CONFIG"); cfg != "" {
		// then try the GH_DASH_CONFIG env var
		userProvidedCfgPath = cfg
	} else if location.RepoPath != "" {
		// Then try to see if we're currently in a git repo
		basename := location.RepoPath + "/." + DashDir
		repoConfigYml := basename + ".yml"
		repoConfigYaml := basename + ".yaml"
		if _, err := os.Stat(repoConfigYml); err == nil {
			userProvidedCfgPath = repoConfigYml
		} else if _, err := os.Stat(repoConfigYaml); err == nil {
			userProvidedCfgPath = repoConfigYaml
		}
	}

	return userProvidedCfgPath
}

// keybindingTypes are the keybinding groups that are unioned across config
// layers rather than being replaced wholesale.
var keybindingTypes = []string{"universal", "prs", "issues", "completions"}

// sectionTypes are replaced wholesale by any layer that defines them.
var sectionTypes = []string{"prSections", "issuesSections", "notificationsSections"}

func mergeOption() koanf.Option {
	return koanf.WithMergeFunc(func(overrides, dest map[string]any) error {
		// Snapshot the layer's sections before maps.Merge so the section guard
		// below can tell which sections this layer actually declared.
		overridesCopy := maps.Copy(overrides)

		unioned := make(map[string][]map[string]string, len(keybindingTypes))
		for _, typ := range keybindingTypes {
			unioned[typ] = mergeKeybindings(overrides, dest, typ)
		}

		maps.Merge(overrides, dest)

		keybindings, ok := dest["keybindings"].(map[string]any)
		if !ok {
			keybindings = map[string]any{}
			dest["keybindings"] = keybindings
		}
		for typ, binds := range unioned {
			keybindings[typ] = binds
		}
		for _, key := range sectionTypes {
			if v, ok := overridesCopy[key]; ok {
				dest[key] = v
			}
		}

		return nil
	})
}

func (parser ConfigParser) getIncludes(cfgPath string) ([]string, error) {
	k := koanf.NewWithConf(conf)
	if err := k.Load(file.Provider(cfgPath), yaml.Parser()); err != nil {
		return nil, err
	}
	return k.Strings("include"), nil
}

func resolveIncludePath(includingCfgPath, includePath string) string {
	if strings.HasPrefix(includePath, "~") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			return strings.Replace(includePath, "~", homeDir, 1)
		}
	}
	if filepath.IsAbs(includePath) {
		return includePath
	}
	return filepath.Join(filepath.Dir(includingCfgPath), includePath)
}

func (parser ConfigParser) loadConfig(cfgPath string) error {
	return parser.loadConfigWithIncludes(cfgPath, map[string]bool{})
}

// seen holds absolute paths already loaded, so an include cycle can't loop forever.
func (parser ConfigParser) loadConfigWithIncludes(cfgPath string, seen map[string]bool) error {
	abs, err := filepath.Abs(cfgPath)
	if err != nil {
		abs = cfgPath
	}
	if seen[abs] {
		log.Warn("Skipping already-included config to avoid a cycle", "path", cfgPath)
		return nil
	}
	seen[abs] = true

	includes, err := parser.getIncludes(cfgPath)
	if err != nil {
		return parsingError{err: err, path: cfgPath}
	}
	// Load includes first so cfgPath's own values take precedence over them.
	for _, include := range includes {
		includePath := resolveIncludePath(cfgPath, include)
		if err := parser.loadConfigWithIncludes(includePath, seen); err != nil {
			return err
		}
	}

	if err := parser.k.Load(file.Provider(cfgPath), yaml.Parser(), mergeOption()); err != nil {
		return parsingError{err: err, path: cfgPath}
	}
	log.Info("Loaded config", "path", cfgPath)
	return nil
}

func (parser ConfigParser) mergeConfigs(globalCfgPath, userProvidedCfgPath string) (Config, error) {
	if err := parser.loadConfig(globalCfgPath); err != nil {
		return Config{}, err
	}
	if err := parser.loadConfig(userProvidedCfgPath); err != nil {
		return Config{}, err
	}
	return parser.unmarshalConfigWithDefaults()
}

func mergeKeybindings(overrides, dest map[string]any, typ string) []map[string]string {
	byKey := make(map[string]map[string]string)
	// dest first, then overrides, so a key bound in both resolves to overrides.
	for _, layer := range []map[string]any{dest, overrides} {
		for _, keybind := range keybindingsByType(layer, typ) {
			if key, ok := keybind["key"]; ok {
				byKey[key] = keybind
			}
		}
	}

	merged := make([]map[string]string, 0, len(byKey))
	for _, keybind := range byKey {
		merged = append(merged, keybind)
	}
	return merged
}

func keybindingsByType(layer map[string]any, typ string) []map[string]string {
	keybindings, ok := layer["keybindings"].(map[string]any)
	if !ok {
		return nil
	}

	// Raw YAML parses to []any, but a previous merge pass writes the type back as
	// []map[string]string, so both shapes can show up once 3+ layers are merged.
	switch list := keybindings[typ].(type) {
	case []map[string]string:
		return list
	case []any:
		out := make([]map[string]string, 0, len(list))
		for _, item := range list {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			casted := make(map[string]string)
			for key, val := range m {
				if s, ok := val.(string); ok {
					casted[key] = s
				}
			}
			out = append(out, casted)
		}
		return out
	default:
		return nil
	}
}

type parsingError struct {
	path string
	err  error
}

func (e parsingError) Error() string {
	return fmt.Sprintf("failed parsing config at path %s with error %v", e.path, e.err)
}

func validateColor(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if hexColorRegex.MatchString(s) {
		return true
	}
	n, err := strconv.Atoi(s)
	return err == nil && n >= 0 && n <= 255
}

func initParser() ConfigParser {
	validate = validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("yaml"), ",")
		if name == "-" {
			return ""
		}
		return name
	})

	validate.RegisterValidation("color", validateColor)

	return ConfigParser{
		k: koanf.NewWithConf(conf),
	}
}

type Location struct {
	RepoPath         string // path if inside a git repo
	ConfigFlag       string // Config passed with explicit --config flag
	SkipGlobalConfig bool   // Skip loading global config (for testing)
}

func ParseConfig(location Location) (Config, error) {
	parser := initParser()

	var config Config
	var err error

	userProvidedCfgPath := parser.getProvidedConfigPath(location)

	// For testing: skip global config and load only the provided config
	if location.SkipGlobalConfig && userProvidedCfgPath != "" {
		if err := parser.loadConfig(userProvidedCfgPath); err != nil {
			return Config{}, err
		}
		return parser.unmarshalConfigWithDefaults()
	}

	globalCfgPath, err := parser.getGlobalConfigPathOrCreateIfMissing()
	if err != nil {
		return config, parsingError{path: globalCfgPath, err: err}
	}

	if userProvidedCfgPath != "" {
		mergedCfg, err := parser.mergeConfigs(globalCfgPath, userProvidedCfgPath)
		if err != nil {
			return Config{}, err
		}
		return mergedCfg, nil
	}

	if err = parser.loadConfig(globalCfgPath); err != nil {
		log.Error("failed loading global config", "err", err)
		return Config{}, err
	}

	return parser.unmarshalConfigWithDefaults()
}

func (parser ConfigParser) unmarshalConfigWithDefaults() (Config, error) {
	cfg := parser.getDefaultConfig()
	err := parser.k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "yaml"})
	if err != nil {
		return Config{}, err
	}

	repoFF := IsFeatureEnabled(FF_REPO_VIEW)
	if cfg.Defaults.View == RepoView && !repoFF {
		cfg.Defaults.View = PRsView
	}

	err = validate.Struct(cfg)
	return cfg, err
}
