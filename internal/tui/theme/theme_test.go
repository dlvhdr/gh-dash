package theme

import (
	"testing"

	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

func TestTheme(t *testing.T) {
	t.Run("Should use the configured colors", func(t *testing.T) {
		colors := config.ColorThemeConfig{
			Inline: config.ColorTheme{
				Text: config.ColorThemeText{
					Primary:   "#FF0000",
					Secondary: "",
					Inverted:  "",
					Faint:     "",
					Warning:   "",
					Success:   "",
					Error:     "",
				},
			},
		}
		thm := config.ThemeConfig{
			Colors: &colors,
		}
		cfg := config.Config{
			Theme: &thm,
		}

		parsed := ParseTheme(&cfg)
		require.Equal(t, "#FF0000", parsed.PrimaryText.Dark)
	})
}
