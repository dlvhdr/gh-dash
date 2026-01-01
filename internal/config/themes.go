package config

import (
	"embed"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed themes/*.yml
var embeddedThemes embed.FS

// ThemeFileText represents text colors in a theme file (without omitzero)
type ThemeFileText struct {
	Primary   string `yaml:"primary"`
	Secondary string `yaml:"secondary"`
	Inverted  string `yaml:"inverted"`
	Faint     string `yaml:"faint"`
	Warning   string `yaml:"warning"`
	Success   string `yaml:"success"`
	Error     string `yaml:"error"`
}

// ThemeFileBackground represents background colors in a theme file
type ThemeFileBackground struct {
	Main     string `yaml:"main"`
	Selected string `yaml:"selected"`
}

// ThemeFileBorder represents border colors in a theme file
type ThemeFileBorder struct {
	Primary   string `yaml:"primary"`
	Secondary string `yaml:"secondary"`
	Faint     string `yaml:"faint"`
}

// ThemeFileColors represents the colors section in a theme file
type ThemeFileColors struct {
	Text       ThemeFileText       `yaml:"text"`
	Background ThemeFileBackground `yaml:"background"`
	Border     ThemeFileBorder     `yaml:"border"`
}

// ThemeFile represents a theme definition file
type ThemeFile struct {
	Name   string          `yaml:"name"`
	Colors ThemeFileColors `yaml:"colors"`
}

// AvailableTheme represents a theme that can be selected
type AvailableTheme struct {
	ID       string // filename without extension
	Name     string // display name
	Path     string // path to theme file (empty for embedded)
	Embedded bool   // true if this is a built-in theme
}

// LoadAvailableThemes returns all available themes (embedded + user themes)
func LoadAvailableThemes() ([]AvailableTheme, error) {
	themes := make([]AvailableTheme, 0)

	// Load embedded themes
	entries, err := embeddedThemes.ReadDir("themes")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		data, err := embeddedThemes.ReadFile("themes/" + entry.Name())
		if err != nil {
			continue
		}

		var tf ThemeFile
		if err := yaml.Unmarshal(data, &tf); err != nil {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".yml")
		themes = append(themes, AvailableTheme{
			ID:       id,
			Name:     tf.Name,
			Embedded: true,
		})
	}

	// Load user themes from config directory
	userThemesDir, err := getUserThemesPath()
	if err == nil {
		if entries, err := os.ReadDir(userThemesDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
					continue
				}

				themePath := filepath.Join(userThemesDir, entry.Name())
				data, err := os.ReadFile(themePath)
				if err != nil {
					continue
				}

				var tf ThemeFile
				if err := yaml.Unmarshal(data, &tf); err != nil {
					continue
				}

				id := strings.TrimSuffix(entry.Name(), ".yml")
				// Check if this overrides an embedded theme
				overrides := false
				for i, t := range themes {
					if t.ID == id {
						themes[i].Path = themePath
						themes[i].Embedded = false
						themes[i].Name = tf.Name
						overrides = true
						break
					}
				}

				if !overrides {
					themes = append(themes, AvailableTheme{
						ID:   id,
						Name: tf.Name,
						Path: themePath,
					})
				}
			}
		}
	}

	// Sort themes alphabetically by name
	sort.Slice(themes, func(i, j int) bool {
		// Keep "default" first
		if themes[i].ID == "default" {
			return true
		}
		if themes[j].ID == "default" {
			return false
		}
		return themes[i].Name < themes[j].Name
	})

	return themes, nil
}

// LoadTheme loads a theme by ID and returns the color configuration
func LoadTheme(id string) (*ThemeFile, error) {
	// Try user themes first
	userThemesDir, err := getUserThemesPath()
	if err == nil {
		themePath := filepath.Join(userThemesDir, id+".yml")
		if data, err := os.ReadFile(themePath); err == nil {
			var tf ThemeFile
			if err := yaml.Unmarshal(data, &tf); err == nil {
				return &tf, nil
			}
		}
	}

	// Fall back to embedded themes
	data, err := embeddedThemes.ReadFile("themes/" + id + ".yml")
	if err != nil {
		return nil, err
	}

	var tf ThemeFile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		return nil, err
	}

	return &tf, nil
}

// ApplyThemeToConfig applies a theme file to the config
func ApplyThemeToConfig(cfg *Config, tf *ThemeFile) {
	if cfg.Theme == nil {
		cfg.Theme = &ThemeConfig{}
	}
	if cfg.Theme.Colors == nil {
		cfg.Theme.Colors = &ColorThemeConfig{}
	}

	// Convert ThemeFile types to config types
	cfg.Theme.Colors.Inline.Text = ColorThemeText{
		Primary:   HexColor(tf.Colors.Text.Primary),
		Secondary: HexColor(tf.Colors.Text.Secondary),
		Inverted:  HexColor(tf.Colors.Text.Inverted),
		Faint:     HexColor(tf.Colors.Text.Faint),
		Warning:   HexColor(tf.Colors.Text.Warning),
		Success:   HexColor(tf.Colors.Text.Success),
		Error:     HexColor(tf.Colors.Text.Error),
	}
	cfg.Theme.Colors.Inline.Background = ColorThemeBackground{
		Main:     HexColor(tf.Colors.Background.Main),
		Selected: HexColor(tf.Colors.Background.Selected),
	}
	cfg.Theme.Colors.Inline.Border = ColorThemeBorder{
		Primary:   HexColor(tf.Colors.Border.Primary),
		Secondary: HexColor(tf.Colors.Border.Secondary),
		Faint:     HexColor(tf.Colors.Border.Faint),
	}
}

func getUserThemesPath() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, DEFAULT_XDG_CONFIG_DIRNAME)
	}
	return filepath.Join(configDir, DashDir, "themes"), nil
}

// GetCurrentTheme gets the currently saved theme ID from state
func GetCurrentTheme() string {
	state, err := LoadState()
	if err != nil || state.Theme == "" {
		return "default"
	}
	return state.Theme
}

// SaveCurrentTheme saves the current theme ID to state
func SaveCurrentTheme(themeID string) error {
	state, _ := LoadState()
	state.Theme = themeID
	return SaveState(state)
}
