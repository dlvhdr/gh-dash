package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/config"
)

type Theme struct {
	SelectedBackground lipgloss.AdaptiveColor // config.Theme.Colors.Background.Selected
	PrimaryBorder      lipgloss.AdaptiveColor // config.Theme.Colors.Border.Primary
	FaintBorder        lipgloss.AdaptiveColor // config.Theme.Colors.Border.Faint
	SecondaryBorder    lipgloss.AdaptiveColor // config.Theme.Colors.Border.Secondary
	FaintText          lipgloss.AdaptiveColor // config.Theme.Colors.Text.Faint
	PrimaryText        lipgloss.AdaptiveColor // config.Theme.Colors.Text.Primary
	SecondaryText      lipgloss.AdaptiveColor // config.Theme.Colors.Text.Secondary
	InvertedText       lipgloss.AdaptiveColor // config.Theme.Colors.Text.Inverted
	SuccessText        lipgloss.AdaptiveColor // config.Theme.Colors.Text.Success
	WarningText        lipgloss.AdaptiveColor // config.Theme.Colors.Text.Warning
	ErrorText          lipgloss.AdaptiveColor // config.Theme.Colors.Text.Error
}

var DefaultTheme = &Theme{
	PrimaryBorder:      lipgloss.AdaptiveColor{Light: "013", Dark: "008"},
	SecondaryBorder:    lipgloss.AdaptiveColor{Light: "008", Dark: "007"},
	SelectedBackground: lipgloss.AdaptiveColor{Light: "006", Dark: "008"},
	FaintBorder:        lipgloss.AdaptiveColor{Light: "254", Dark: "000"},
	PrimaryText:        lipgloss.AdaptiveColor{Light: "000", Dark: "015"},
	SecondaryText:      lipgloss.AdaptiveColor{Light: "244", Dark: "251"},
	FaintText:          lipgloss.AdaptiveColor{Light: "007", Dark: "245"},
	InvertedText:       lipgloss.AdaptiveColor{Light: "015", Dark: "236"},
	SuccessText:        lipgloss.AdaptiveColor{Light: "002", Dark: "002"},
	WarningText:        lipgloss.AdaptiveColor{Light: "003", Dark: "003"},
	ErrorText:          lipgloss.AdaptiveColor{Light: "001", Dark: "001"},
}

func ParseTheme(cfg *config.Config) Theme {
	_shimHex := func(hex config.HexColor, fallback lipgloss.AdaptiveColor) lipgloss.AdaptiveColor {
		if hex == "" {
			return fallback
		}
		return lipgloss.AdaptiveColor{Light: string(hex), Dark: string(hex)}
	}

	if cfg.Theme.Colors != nil {
		DefaultTheme = &Theme{
			SelectedBackground: _shimHex(
				cfg.Theme.Colors.Inline.Background.Selected,
				DefaultTheme.SelectedBackground,
			),
			PrimaryBorder: _shimHex(
				cfg.Theme.Colors.Inline.Border.Primary,
				DefaultTheme.PrimaryBorder,
			),
			FaintBorder: _shimHex(
				cfg.Theme.Colors.Inline.Border.Faint,
				DefaultTheme.FaintBorder,
			),
			SecondaryBorder: _shimHex(
				cfg.Theme.Colors.Inline.Border.Secondary,
				DefaultTheme.SecondaryBorder,
			),
			FaintText: _shimHex(
				cfg.Theme.Colors.Inline.Text.Faint,
				DefaultTheme.FaintText,
			),
			PrimaryText: _shimHex(
				cfg.Theme.Colors.Inline.Text.Primary,
				DefaultTheme.PrimaryText,
			),
			SecondaryText: _shimHex(
				cfg.Theme.Colors.Inline.Text.Secondary,
				DefaultTheme.SecondaryText,
			),
			InvertedText: _shimHex(
				cfg.Theme.Colors.Inline.Text.Inverted,
				DefaultTheme.InvertedText,
			),
			SuccessText: _shimHex(
				cfg.Theme.Colors.Inline.Text.Success,
				DefaultTheme.SuccessText,
			),
			WarningText: _shimHex(
				cfg.Theme.Colors.Inline.Text.Warning,
				DefaultTheme.WarningText,
			),
			ErrorText: _shimHex(
				cfg.Theme.Colors.Inline.Text.Error,
				DefaultTheme.ErrorText,
			),
		}
	}

	log.Debug("Parsing theme", "config", cfg.Theme.Colors, "theme", DefaultTheme)

	return *DefaultTheme
}
