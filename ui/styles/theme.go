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
	MainText           lipgloss.AdaptiveColor
	Border             lipgloss.AdaptiveColor
	SecondaryBorder    lipgloss.AdaptiveColor
	WarningText        lipgloss.AdaptiveColor
	SuccessText        lipgloss.AdaptiveColor
	FaintBorder        lipgloss.AdaptiveColor
	FaintText          lipgloss.AdaptiveColor
	SelectedBackground lipgloss.AdaptiveColor
	SecondaryText      lipgloss.AdaptiveColor
	SubleMainText      lipgloss.AdaptiveColor
}

var theme *Theme

var DefaultTheme = func() Theme {
	if theme != nil {
		return *theme
	}

	_shimHex := func(hex config.HexColor) lipgloss.AdaptiveColor {
		return lipgloss.AdaptiveColor{Light: string(hex), Dark: string(hex)}
	}
	_shimAnsi := func(code string) lipgloss.AdaptiveColor {
		return lipgloss.AdaptiveColor{Light: code, Dark: code}
	}

	cfg, _ := config.ParseConfig()

	if cfg.Theme == nil {
		theme = &Theme{
			MainText:           _shimAnsi("255"),
			SubleMainText:      _shimAnsi("254"),
			Border:             _shimAnsi("0"),
			SecondaryBorder:    _shimAnsi("8"),
			WarningText:        _shimAnsi("1"),
			SuccessText:        _shimAnsi("2"),
			FaintBorder:        _shimAnsi("240"),
			FaintText:          _shimAnsi("253"),
			SelectedBackground: _shimAnsi("6"),
			SecondaryText:      _shimAnsi("11"),
		}
	} else {
		theme = &Theme{
			MainText:           _shimHex(cfg.Theme.Colors.Inline.Text.Primary),
			SubleMainText:      _shimHex(cfg.Theme.Colors.Inline.Text.Inverted),
			Border:             _shimHex(cfg.Theme.Colors.Inline.Border.Primary),
			SecondaryBorder:    _shimHex(cfg.Theme.Colors.Inline.Border.Secondary),
			WarningText:        _shimHex(cfg.Theme.Colors.Inline.Text.Warning),
			SuccessText:        _shimHex(cfg.Theme.Colors.Inline.Text.Success),
			FaintBorder:        _shimHex(cfg.Theme.Colors.Inline.Border.Faint),
			FaintText:          _shimHex(cfg.Theme.Colors.Inline.Text.Faint),
			SelectedBackground: _shimHex(cfg.Theme.Colors.Inline.Background.Selected),
			SecondaryText:      _shimHex(cfg.Theme.Colors.Inline.Text.Secondary),
		}
	}

	return *theme
}()
