package data

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

type RowData interface {
	GetRepoNameWithOwner() string
	GetTitle() string
	GetNumber() int
	GetUrl() string
	GetUpdatedAt() time.Time
}

func GetAuthorRoleIcon(role string, theme theme.Theme) string {
	// https://docs.github.com/en/graphql/reference/enums#commentauthorassociation
	switch role {
	case "FIRST_TIMER", "FIRST_TIME_CONTRIBUTOR", "NONE":
		return stripAnsiReset(lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(
			theme.NewContributorIconColor)).Render(theme.NewContributorIcon))
	case "COLLABORATOR":
		return stripAnsiReset(lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(
			theme.CollaboratorIconColor)).Render(theme.CollaboratorIcon))
	case "CONTRIBUTOR":
		return stripAnsiReset(lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(
			theme.ContributorIconColor)).Render(theme.ContributorIcon))
	case "MEMBER":
		return stripAnsiReset(lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(theme.MemberIconColor)).Render(theme.MemberIcon))
	case "OWNER":
		return stripAnsiReset(lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(theme.OwnerIconColor)).Render(theme.OwnerIcon))
	default:
		return stripAnsiReset(lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(
			theme.UnknownRoleIconColor)).Render(theme.UnknownRoleIcon))
	}
}

func stripAnsiReset(value string) string {
	return strings.ReplaceAll(value, "\x1b[0m", "")
}
