package theme

import (
	"image/color"
	"testing"

	"charm.land/log/v2"
	"github.com/charmbracelet/x/ansi"
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
		require.Equal(
			t,
			color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
			parsed.PrimaryText.Dark,
		)
	})

	t.Run("Should use ANSI color indices", func(t *testing.T) {
		colors := config.ColorThemeConfig{
			Inline: config.ColorTheme{
				Text: config.ColorThemeText{
					Primary: "12",
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
		require.Equal(t, ansi.BasicColor(12), parsed.PrimaryText.Light)
		require.Equal(t, ansi.BasicColor(12), parsed.PrimaryText.Dark)
	})
}
