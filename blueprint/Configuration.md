# Configuration System

This document explains the configuration system in gh-dash, including how configuration is loaded, validated, merged, and used throughout the application.

## Table of Contents
1. [Configuration Structure](#configuration-structure)
2. [Configuration Loading](#configuration-loading)
3. [Configuration Merging](#configuration-merging)
4. [Validation](#validation)
5. [Default Values](#default-values)
6. [User Customization](#user-customization)

---

## Configuration Structure

### Main Configuration Object

```go
// internal/config/parser.go

type Config struct {
    // Section definitions
    PRSections     []PrsSectionConfig      // PR filtering sections
    IssuesSections []IssuesSectionConfig   // Issue filtering sections

    // Repository settings
    Repo RepoConfig

    // Default settings
    Defaults Defaults

    // Key bindings
    Keybindings Keybindings

    // Repository path aliases
    RepoPaths map[string]string

    // Theming
    Theme *ThemeConfig

    // Diff viewer
    Pager Pager

    // Behavior
    ConfirmQuit            bool
    ShowAuthorIcons        bool
    SmartFilteringAtLaunch bool
}
```

### PR Sections

```go
type PrsSectionConfig struct {
    Title   string              // Section display title
    Filters string              // GitHub search filters
    Limit   *int                // Max results (optional)
    Layout  PrsLayoutConfig     // Column layout overrides
    Type    *ViewType           // View type override
}

type PrsLayoutConfig struct {
    UpdatedAt    ColumnConfig
    CreatedAt    ColumnConfig
    Repo         ColumnConfig
    Author       ColumnConfig
    AuthorIcon   ColumnConfig
    Assignees    ColumnConfig
    Title        ColumnConfig
    Base         ColumnConfig
    ReviewStatus ColumnConfig
    State        ColumnConfig
    Ci           ColumnConfig
    Lines        ColumnConfig
}

type ColumnConfig struct {
    Width  *int   // Column width in characters (optional)
    Hidden *bool  // Whether to hide column (optional)
}
```

### Issues Sections

```go
type IssuesSectionConfig struct {
    Title   string
    Filters string
    Limit   *int
    Layout  IssuesLayoutConfig
}

type IssuesLayoutConfig struct {
    UpdatedAt   ColumnConfig
    CreatedAt   ColumnConfig
    State       ColumnConfig
    Repo        ColumnConfig
    Title       ColumnConfig
    Creator     ColumnConfig
    CreatorIcon ColumnConfig
    Assignees   ColumnConfig
    Comments    ColumnConfig
    Reactions   ColumnConfig
}
```

### Defaults

```go
type Defaults struct {
    Preview PreviewConfig  // Sidebar settings
    PrsLimit               int
    PrApproveComment       string
    IssuesLimit            int
    View                   ViewType        // Default view (prs/issues/repo)
    Layout                 LayoutConfig    // Global layout settings
    RefetchIntervalMinutes int
    DateFormat             string
}

type PreviewConfig struct {
    Open  bool  // Show sidebar by default
    Width int   // Sidebar width
}
```

### Keybindings

```go
type Keybindings struct {
    Universal []Keybinding  // Available everywhere
    Issues    []Keybinding  // Issue view only
    Prs       []Keybinding  // PR view only
    Branches  []Keybinding  // Repo/branch view only
}

type Keybinding struct {
    Key     string  // Key combination (e.g., "a", "ctrl+p")
    Command string  // Shell command to run (optional)
    Builtin string  // Built-in action (optional)
    Name    string  // Display name (optional)
}
```

### Theme Configuration

```go
type ThemeConfig struct {
    Ui     UIThemeConfig      // UI preferences
    Colors *ColorThemeConfig  // Color overrides
    Icons  *IconThemeConfig   // Icon overrides
}

type ColorThemeConfig struct {
    Inline ColorTheme
}

type ColorTheme struct {
    Icon       ColorThemeIcon
    Text       ColorThemeText
    Background ColorThemeBackground
    Border     ColorThemeBorder
}

type ColorThemeText struct {
    Primary   HexColor  // Main text color
    Secondary HexColor  // Secondary text color
    Inverted  HexColor  // Inverted text color
    Faint     HexColor  // Faint text color
    Warning   HexColor  // Warning color
    Success   HexColor  // Success color
    Error     HexColor  // Error color
}
```

---

## Configuration Loading

### Configuration File Hierarchy

```
Priority (highest to lowest):
1. --config flag          → Explicit config file
2. .gh-dash.yml (repo)    → Repository-local config
3. $GH_DASH_CONFIG        → Environment variable
4. ~/.config/gh-dash/config.yml  → Global config (created if missing)
```

### Loading Process

```go
// internal/config/parser.go

func ParseConfig(location Location) (Config, error) {
    parser := initParser()

    // 1. Get global config path (create if missing)
    globalCfgPath, err := parser.getGlobalConfigPathOrCreateIfMissing()
    if err != nil {
        return Config{}, err
    }

    // 2. Check for user-provided config
    userProvidedCfgPath := parser.getProvidedConfigPath(location)

    // 3. Merge configs if both exist
    if userProvidedCfgPath != "" {
        return parser.mergeConfigs(globalCfgPath, userProvidedCfgPath)
    }

    // 4. Otherwise, load global config only
    err = parser.loadGlobalConfig(globalCfgPath)
    if err != nil {
        return Config{}, err
    }

    return parser.unmarshalConfigWithDefaults()
}
```

### Configuration Detection

```go
func (parser ConfigParser) getProvidedConfigPath(location Location) string {
    // 1. Check --config flag
    if location.ConfigFlag != "" {
        return location.ConfigFlag
    }

    // 2. Check $GH_DASH_CONFIG env var
    if cfg := os.Getenv("GH_DASH_CONFIG"); cfg != "" {
        return cfg
    }

    // 3. Check for repo-local config
    if location.RepoPath != "" {
        repoConfigYml := location.RepoPath + "/.gh-dash.yml"
        repoConfigYaml := location.RepoPath + "/.gh-dash.yaml"

        if _, err := os.Stat(repoConfigYml); err == nil {
            return repoConfigYml
        }
        if _, err := os.Stat(repoConfigYaml); err == nil {
            return repoConfigYaml
        }
    }

    return ""
}
```

---

## Configuration Merging

### Merge Strategy

```go
func (parser ConfigParser) mergeConfigs(
    globalCfgPath,
    userProvidedCfgPath string,
) (Config, error) {
    // 1. Load global config
    parser.k.Load(file.Provider(globalCfgPath), yaml.Parser())

    // 2. Load user config with custom merge function
    parser.k.Load(
        file.Provider(userProvidedCfgPath),
        yaml.Parser(),
        koanf.WithMergeFunc(func(overrides, dest map[string]any) error {
            // Custom merge logic for keybindings
            universalKeybinds := mergeKeybindings(overrides, dest, "universal")
            prsKeybinds := mergeKeybindings(overrides, dest, "prs")
            issuesKeybinds := mergeKeybindings(overrides, dest, "issues")

            // Deep merge most fields
            maps.Merge(overrides, dest)

            // Override keybindings with merged versions
            dest["keybindings"].(map[string]any)["universal"] = universalKeybinds
            dest["keybindings"].(map[string]any)["prs"] = prsKeybinds
            dest["keybindings"].(map[string]any)["issues"] = issuesKeybinds

            // Sections completely replace (not merge)
            dest["prSections"] = overrides["prSections"]
            dest["issuesSections"] = overrides["issuesSections"]

            return nil
        }),
    )

    return parser.unmarshalConfigWithDefaults()
}
```

### Keybinding Merge Logic

```go
func mergeKeybindings(src, dest map[string]any, typ string) []map[string]string {
    // Create map of keybindings by key
    keybindsMap := make(map[string]map[string]string)

    // Add source keybindings (user config)
    for _, keybind := range src["keybindings"].(map[string]any)[typ].([]any) {
        keybind := keybind.(map[string]any)
        keybindsMap[keybind["key"].(string)] = castToStringMap(keybind)
    }

    // Add dest keybindings (global config) if key not in source
    for _, keybind := range dest["keybindings"].(map[string]any)[typ].([]any) {
        keybind := keybind.(map[string]any)
        key := keybind["key"].(string)

        if _, exists := keybindsMap[key]; !exists {
            keybindsMap[key] = castToStringMap(keybind)
        }
    }

    // Convert back to slice
    merged := make([]map[string]string, 0, len(keybindsMap))
    for _, keybind := range keybindsMap {
        merged = append(merged, keybind)
    }

    return merged
}
```

### Merge Behavior

| Field | Merge Strategy |
|-------|---------------|
| Sections (prs/issues) | **Replace** - User config completely replaces global |
| Keybindings | **Union** - Merge by key, user overrides global |
| Theme | **Deep merge** - User values override specific colors |
| Defaults | **Deep merge** - User values override defaults |
| Other fields | **Deep merge** - User values take precedence |

---

## Validation

### Validation Rules

```go
// Using go-playground/validator tags

type Config struct {
    PRSections []PrsSectionConfig `yaml:"prSections"`
    Theme      *ThemeConfig       `yaml:"theme" validate:"omitempty"`
}

type ColumnConfig struct {
    Width  *int  `yaml:"width" validate:"omitempty,gt=0"`  // Must be > 0 if present
    Hidden *bool `yaml:"hidden"`
}

type ColorThemeText struct {
    Primary HexColor `yaml:"primary" validate:"omitempty,hexcolor"`
}
```

### Validation Process

```go
var validate *validator.Validate

func (parser ConfigParser) unmarshalConfigWithDefaults() (Config, error) {
    cfg := parser.getDefaultConfig()

    // Unmarshal config
    err := parser.k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "yaml"})
    if err != nil {
        return Config{}, err
    }

    // Validate
    err = validate.Struct(cfg)
    return cfg, err
}

// Example validation errors:
// - Width must be greater than 0
// - Invalid hex color format
// - Required field missing
```

### Custom Validators

```go
func init() {
    validate = validator.New()

    // Use yaml tag for field names
    validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
        name := strings.Split(fld.Tag.Get("yaml"), ",")[0]
        if name == "-" {
            return ""
        }
        return name
    })

    // Custom validators could be added here
    // validate.RegisterValidation("customRule", customValidatorFunc)
}
```

---

## Default Values

### Default Configuration

```go
func (parser ConfigParser) getDefaultConfig() Config {
    return Config{
        Defaults: Defaults{
            Preview: PreviewConfig{
                Open:  true,
                Width: 50,
            },
            PrsLimit:               20,
            PrApproveComment:       "LGTM",
            IssuesLimit:            20,
            View:                   PRsView,
            RefetchIntervalMinutes: 30,
            DateFormat:             "",
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
                    // ... more defaults
                },
            },
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
        Theme: &ThemeConfig{
            Ui: UIThemeConfig{
                SectionsShowCount: true,
                Table: TableUIThemeConfig{
                    ShowSeparator: true,
                    Compact:       false,
                },
            },
        },
        ConfirmQuit:            false,
        ShowAuthorIcons:        true,
        SmartFilteringAtLaunch: true,
    }
}
```

### Applying Defaults

```go
// Defaults are applied first, then user config overrides
func (parser ConfigParser) unmarshalConfigWithDefaults() (Config, error) {
    // Start with defaults
    cfg := parser.getDefaultConfig()

    // Unmarshal user config over defaults
    err := parser.k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "yaml"})

    return cfg, err
}
```

---

## User Customization

### Example Configuration File

```yaml
# ~/.config/gh-dash/config.yml

prSections:
  - title: "My PRs"
    filters: "is:open author:@me"
    limit: 20
    layout:
      author:
        hidden: true
      repo:
        width: 25

  - title: "Team PRs"
    filters: "is:open team:myorg/myteam"
    limit: 50

issuesSections:
  - title: "Assigned to Me"
    filters: "is:open assignee:@me"
  - title: "High Priority"
    filters: "is:open label:priority-high"

defaults:
  view: prs
  preview:
    open: true
    width: 60
  prsLimit: 30
  issuesLimit: 30
  refetchIntervalMinutes: 10
  layout:
    prs:
      updatedAt:
        width: 8
      lines:
        hidden: true

keybindings:
  universal:
    - key: "q"
      builtin: "quit"
    - key: "r"
      builtin: "refresh"

  prs:
    - key: "a"
      builtin: "approve"
    - key: "m"
      builtin: "merge"
    - key: "o"
      builtin: "open"
    - key: "x"
      command: "echo 'Custom command!'"
      name: "Custom Action"

theme:
  colors:
    text:
      primary: "#E0E0E0"
      secondary: "#808080"
      success: "#00FF00"
      error: "#FF0000"
    border:
      primary: "#404040"
    background:
      selected: "#2d2d2d"

  icons:
    contributor: "👤"
    owner: "👑"

  ui:
    sectionsShowCount: true
    table:
      showSeparator: true
      compact: false

pager:
  diff: "delta"

confirmQuit: true
showAuthorIcons: true
smartFilteringAtLaunch: true
```

### Repository-Local Override

```yaml
# /path/to/myproject/.gh-dash.yml

# Override global config for this repo
prSections:
  - title: "Project PRs"
    filters: "is:open repo:owner/myproject"
    limit: 100

defaults:
  view: prs
```

### Configuration via Environment

```bash
# Set config path
export GH_DASH_CONFIG=/path/to/custom/config.yml

# Feature flags
export FF_REPO_VIEW=1
export FF_MOCK_DATA=1

gh-dash
```

---

## Configuration Best Practices

### 1. Start with Defaults

```yaml
# Minimal config - uses all defaults
prSections:
  - title: "My PRs"
    filters: "is:open author:@me"
```

### 2. Override Selectively

```yaml
# Override only what you need
defaults:
  prsLimit: 50  # Override default of 20

# Other defaults remain unchanged
```

### 3. Use Repository-Local Configs

```yaml
# .gh-dash.yml in project root
# Project-specific sections
prSections:
  - title: "Backend PRs"
    filters: "is:open label:backend"
  - title: "Frontend PRs"
    filters: "is:open label:frontend"
```

### 4. Template Variables in Filters

```yaml
prSections:
  - title: "Recent PRs"
    # Use template functions
    filters: 'is:open updated:>{{ now | dateModify "-168h" | date "2006-01-02" }}'
```

### 5. Custom Keybindings

```yaml
keybindings:
  prs:
    # Run custom script
    - key: "ctrl+t"
      command: "gh pr checks"
      name: "Run Checks"

    # Override default
    - key: "a"
      builtin: "approve"
```

---

## Summary

The configuration system in gh-dash:

1. **Hierarchical**: Global → Local → Flag
2. **Mergeable**: Configs merge intelligently
3. **Validated**: Type-safe with validation
4. **Defaulted**: Works out of the box
5. **Flexible**: Override any setting
6. **Type-safe**: Compile-time checks
7. **Documented**: Self-documenting via examples

Key features:
- **YAML-based**: Human-readable, editable
- **Validated**: go-playground/validator
- **Merged**: Koanf library
- **Hierarchical**: Multiple config sources
- **Type-safe**: Strong typing

When configuring:
- Start with defaults
- Override selectively
- Use repo-local configs for project-specific settings
- Validate with `--debug` to see config loading
- Reference example config for all options
