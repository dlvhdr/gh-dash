package data

import (
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
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(
			theme.NewContributorIconColor)).Background(theme.MainBackground).Render(theme.NewContributorIcon)
	case "COLLABORATOR":
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(
			theme.CollaboratorIconColor)).Background(theme.MainBackground).Render(theme.CollaboratorIcon)
	case "CONTRIBUTOR":
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(
			theme.ContributorIconColor)).Background(theme.MainBackground).Render(theme.ContributorIcon)
	case "MEMBER":
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(theme.MemberIconColor)).Background(theme.MainBackground).Render(theme.MemberIcon)
	case "OWNER":
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(theme.OwnerIconColor)).Background(theme.MainBackground).Render(theme.OwnerIcon)
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(
			theme.UnknownRoleIconColor)).Background(theme.MainBackground).Render(theme.UnknownRoleIcon)
	}
}
