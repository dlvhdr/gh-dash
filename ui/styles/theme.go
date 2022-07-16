package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/config"
)

var (
	SingleRuneWidth    = 4
	MainContentPadding = 1
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
}

var theme *Theme

var DefaultTheme = func() Theme {
	if theme != nil {
		return *theme
	}

	_shimHex := func(hex config.HexColor) lipgloss.AdaptiveColor {
		return lipgloss.AdaptiveColor{Light: string(hex), Dark: string(hex)}
	}

	cfg, _ := config.ParseConfig()

	if cfg.Theme == nil {
		theme = &Theme{
			SelectedBackground: lipgloss.AdaptiveColor{Light: "006", Dark: "006"},
			PrimaryBorder:      lipgloss.AdaptiveColor{Light: "000", Dark: "015"},
			FaintBorder:        lipgloss.AdaptiveColor{Light: "007", Dark: "008"},
			SecondaryBorder:    lipgloss.AdaptiveColor{Light: "008", Dark: "007"},
			FaintText:          lipgloss.AdaptiveColor{Light: "243", Dark: "249"},
			PrimaryText:        lipgloss.AdaptiveColor{Light: "000", Dark: "015"},
			SecondaryText:      lipgloss.AdaptiveColor{Light: "237", Dark: "255"},
			InvertedText:       lipgloss.AdaptiveColor{Light: "015", Dark: "015"},
			SuccessText:        lipgloss.AdaptiveColor{Light: "002", Dark: "002"},
			WarningText:        lipgloss.AdaptiveColor{Light: "001", Dark: "001"},
		}
	} else {
		theme = &Theme{
			SelectedBackground: _shimHex(cfg.Theme.Colors.Inline.Background.Selected),
			PrimaryBorder:      _shimHex(cfg.Theme.Colors.Inline.Border.Primary),
			FaintBorder:        _shimHex(cfg.Theme.Colors.Inline.Border.Faint),
			SecondaryBorder:    _shimHex(cfg.Theme.Colors.Inline.Border.Secondary),
			FaintText:          _shimHex(cfg.Theme.Colors.Inline.Text.Faint),
			PrimaryText:        _shimHex(cfg.Theme.Colors.Inline.Text.Primary),
			SecondaryText:      _shimHex(cfg.Theme.Colors.Inline.Text.Secondary),
			InvertedText:       _shimHex(cfg.Theme.Colors.Inline.Text.Inverted),
			SuccessText:        _shimHex(cfg.Theme.Colors.Inline.Text.Success),
			WarningText:        _shimHex(cfg.Theme.Colors.Inline.Text.Warning),
		}
	}

	return *theme
}()
