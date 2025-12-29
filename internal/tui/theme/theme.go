package theme

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

type Theme struct {
	MainBackground          lipgloss.AdaptiveColor // config.Theme.Colors.Background.Main
	SelectedBackground      lipgloss.AdaptiveColor // config.Theme.Colors.Background.Selected
	PrimaryBorder           lipgloss.AdaptiveColor // config.Theme.Colors.Border.Primary
	FaintBorder             lipgloss.AdaptiveColor // config.Theme.Colors.Border.Faint
	SecondaryBorder         lipgloss.AdaptiveColor // config.Theme.Colors.Border.Secondary
	FaintText               lipgloss.AdaptiveColor // config.Theme.Colors.Text.Faint
	PrimaryText             lipgloss.AdaptiveColor // config.Theme.Colors.Text.Primary
	SecondaryText           lipgloss.AdaptiveColor // config.Theme.Colors.Text.Secondary
	InvertedText            lipgloss.AdaptiveColor // config.Theme.Colors.Text.Inverted
	SuccessText             lipgloss.AdaptiveColor // config.Theme.Colors.Text.Success
	WarningText             lipgloss.AdaptiveColor // config.Theme.Colors.Text.Warning
	ErrorText               lipgloss.AdaptiveColor // config.Theme.Colors.Text.Error
	NewContributorIconColor lipgloss.AdaptiveColor // config.Theme.Colors.Icon.NewContributor
	ContributorIconColor    lipgloss.AdaptiveColor // config.Theme.Colors.Icon.Contributor
	CollaboratorIconColor   lipgloss.AdaptiveColor // config.Theme.Colors.Icon.Collaborator
	MemberIconColor         lipgloss.AdaptiveColor // config.Theme.Colors.Icon.Member
	OwnerIconColor          lipgloss.AdaptiveColor // config.Theme.Colors.Icon.Owner
	UnknownRoleIconColor    lipgloss.AdaptiveColor // config.Theme.Colors.Icon.UnknownRole
	NewContributorIcon      string                 // config.Theme.Icons.NewContributor
	ContributorIcon         string                 // config.Theme.Icons.Contributor
	CollaboratorIcon        string                 // config.Theme.Icons.Collaborator
	MemberIcon              string                 // config.Theme.Icons.Member
	OwnerIcon               string                 // config.Theme.Icons.Owner
	UnknownRoleIcon         string                 // config.Theme.Icons.UnknownRole
}

var DefaultTheme = &Theme{
	MainBackground:          lipgloss.AdaptiveColor{Light: "", Dark: ""},
	PrimaryBorder:           lipgloss.AdaptiveColor{Light: "013", Dark: "008"},
	SecondaryBorder:         lipgloss.AdaptiveColor{Light: "008", Dark: "007"},
	SelectedBackground:      lipgloss.AdaptiveColor{Light: "006", Dark: "008"},
	FaintBorder:             lipgloss.AdaptiveColor{Light: "254", Dark: "000"},
	PrimaryText:             lipgloss.AdaptiveColor{Light: "000", Dark: "015"},
	SecondaryText:           lipgloss.AdaptiveColor{Light: "244", Dark: "251"},
	FaintText:               lipgloss.AdaptiveColor{Light: "007", Dark: "245"},
	InvertedText:            lipgloss.AdaptiveColor{Light: "015", Dark: "236"},
	SuccessText:             lipgloss.AdaptiveColor{Light: "002", Dark: "002"},
	WarningText:             lipgloss.AdaptiveColor{Light: "003", Dark: "003"},
	ErrorText:               lipgloss.AdaptiveColor{Light: "001", Dark: "001"},
	NewContributorIconColor: lipgloss.AdaptiveColor{Light: "077", Dark: "077"},
	ContributorIconColor:    lipgloss.AdaptiveColor{Light: "075", Dark: "075"},
	CollaboratorIconColor:   lipgloss.AdaptiveColor{Light: "178", Dark: "178"},
	MemberIconColor:         lipgloss.AdaptiveColor{Light: "178", Dark: "178"},
	OwnerIconColor:          lipgloss.AdaptiveColor{Light: "178", Dark: "178"},
	UnknownRoleIconColor:    lipgloss.AdaptiveColor{Light: "178", Dark: "178"},
	NewContributorIcon:      constants.NewContributorIcon,
	ContributorIcon:         constants.ContributorIcon,
	CollaboratorIcon:        constants.CollaboratorIcon,
	MemberIcon:              constants.MemberIcon,
	OwnerIcon:               constants.OwnerIcon,
	UnknownRoleIcon:         constants.UnknownRoleIcon,
}

func ParseTheme(cfg *config.Config) Theme {
	_shimHex := func(hex config.HexColor, fallback lipgloss.AdaptiveColor) lipgloss.AdaptiveColor {
		if hex == "" {
			return fallback
		}
		return lipgloss.AdaptiveColor{Light: string(hex), Dark: string(hex)}
	}
	_shimIcon := func(icon string, fallback string) string {
		if icon != "" {
			return icon
		}
		return fallback
	}

	if cfg.Theme.Colors != nil {
		DefaultTheme.MainBackground = _shimHex(
			cfg.Theme.Colors.Inline.Background.Main,
			DefaultTheme.MainBackground,
		)
		DefaultTheme.SelectedBackground = _shimHex(
			cfg.Theme.Colors.Inline.Background.Selected,
			DefaultTheme.SelectedBackground,
		)
		DefaultTheme.PrimaryBorder = _shimHex(
			cfg.Theme.Colors.Inline.Border.Primary,
			DefaultTheme.PrimaryBorder,
		)
		DefaultTheme.FaintBorder = _shimHex(
			cfg.Theme.Colors.Inline.Border.Faint,
			DefaultTheme.FaintBorder,
		)
		DefaultTheme.SecondaryBorder = _shimHex(
			cfg.Theme.Colors.Inline.Border.Secondary,
			DefaultTheme.SecondaryBorder,
		)
		DefaultTheme.FaintText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Faint,
			DefaultTheme.FaintText,
		)
		DefaultTheme.PrimaryText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Primary,
			DefaultTheme.PrimaryText,
		)
		DefaultTheme.SecondaryText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Secondary,
			DefaultTheme.SecondaryText,
		)
		DefaultTheme.InvertedText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Inverted,
			DefaultTheme.InvertedText,
		)
		DefaultTheme.SuccessText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Success,
			DefaultTheme.SuccessText,
		)
		DefaultTheme.WarningText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Warning,
			DefaultTheme.WarningText,
		)
		DefaultTheme.ErrorText = _shimHex(
			cfg.Theme.Colors.Inline.Text.Error,
			DefaultTheme.ErrorText,
		)
		DefaultTheme.NewContributorIconColor = _shimHex(
			cfg.Theme.Colors.Inline.Icon.NewContributor,
			DefaultTheme.NewContributorIconColor,
		)
		DefaultTheme.ContributorIconColor = _shimHex(
			cfg.Theme.Colors.Inline.Icon.Contributor,
			DefaultTheme.ContributorIconColor,
		)
		DefaultTheme.CollaboratorIconColor = _shimHex(
			cfg.Theme.Colors.Inline.Icon.Collaborator,
			DefaultTheme.CollaboratorIconColor,
		)
		DefaultTheme.MemberIconColor = _shimHex(
			cfg.Theme.Colors.Inline.Icon.Member,
			DefaultTheme.MemberIconColor,
		)
		DefaultTheme.OwnerIconColor = _shimHex(
			cfg.Theme.Colors.Inline.Icon.Owner,
			DefaultTheme.OwnerIconColor,
		)
	}

	if cfg.ShowAuthorIcons && cfg.Theme.Icons != nil {
		DefaultTheme.NewContributorIcon = _shimIcon(
			cfg.Theme.Icons.Inline.NewContributor,
			DefaultTheme.NewContributorIcon,
		)
		DefaultTheme.ContributorIcon = _shimIcon(
			cfg.Theme.Icons.Inline.Contributor,
			DefaultTheme.ContributorIcon,
		)
		DefaultTheme.CollaboratorIcon = _shimIcon(
			cfg.Theme.Icons.Inline.Collaborator,
			DefaultTheme.CollaboratorIcon,
		)
		DefaultTheme.MemberIcon = _shimIcon(
			cfg.Theme.Icons.Inline.Member,
			DefaultTheme.MemberIcon,
		)
		DefaultTheme.OwnerIcon = _shimIcon(
			cfg.Theme.Icons.Inline.Owner,
			DefaultTheme.OwnerIcon,
		)
		DefaultTheme.UnknownRoleIcon = _shimIcon(
			cfg.Theme.Icons.Inline.UnknownRole,
			DefaultTheme.UnknownRoleIcon,
		)
	}

	return *DefaultTheme
}

func HasDarkBackground(theme Theme) bool {
	bg := strings.TrimSpace(theme.MainBackground.Light)
	if bg == "" {
		bg = strings.TrimSpace(theme.MainBackground.Dark)
	}
	if bg == "" {
		return lipgloss.HasDarkBackground()
	}
	isDark, ok := isDarkHexColor(bg)
	if !ok {
		return lipgloss.HasDarkBackground()
	}
	return isDark
}

func isDarkHexColor(value string) (bool, bool) {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimPrefix(trimmed, "#")
	if len(trimmed) == 3 {
		trimmed = fmt.Sprintf("%c%c%c%c%c%c", trimmed[0], trimmed[0], trimmed[1], trimmed[1], trimmed[2], trimmed[2])
	}
	if len(trimmed) != 6 {
		return false, false
	}
	r, err := strconv.ParseUint(trimmed[0:2], 16, 8)
	if err != nil {
		return false, false
	}
	g, err := strconv.ParseUint(trimmed[2:4], 16, 8)
	if err != nil {
		return false, false
	}
	b, err := strconv.ParseUint(trimmed[4:6], 16, 8)
	if err != nil {
		return false, false
	}
	luminance := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
	return luminance < 128, true
}
