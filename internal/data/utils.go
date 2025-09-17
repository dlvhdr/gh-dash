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

func IsStatusWaiting(status string) bool {
	return status == "PENDING" ||
		status == "QUEUED" ||
		status == "IN_PROGRESS" ||
		status == "WAITING"
}

func IsConclusionASkip(conclusion string) bool {
	return conclusion == "SKIPPED"
}

func IsConclusionAFailure(conclusion string) bool {
	return conclusion == "FAILURE" || conclusion == "TIMED_OUT" || conclusion == "STARTUP_FAILURE"
}

func IsConclusionASuccess(conclusion string) bool {
	return conclusion == "SUCCESS"
}

func GetAuthorRoleIcon(role string, theme theme.Theme) string {
	// https://docs.github.com/en/graphql/reference/enums#commentauthorassociation
	switch role {
	case "FIRST_TIMER", "FIRST_TIME_CONTRIBUTOR", "NONE":
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(theme.NewContributorIconColor)).Render(theme.NewContributorIcon)
	case "COLLABORATOR":
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(theme.CollaboratorIconColor)).Render(theme.CollaboratorIcon)
	case "CONTRIBUTOR":
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(theme.ContributorIconColor)).Render(theme.ContributorIcon)
	case "MEMBER":
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(theme.MemberIconColor)).Render(theme.MemberIcon)
	case "OWNER":
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(theme.OwnerIconColor)).Render(theme.OwnerIcon)
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(theme.UnknownRoleIconColor)).Render(theme.UnknownRoleIcon)
	}
}
