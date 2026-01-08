package notificationrow

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type Notification struct {
	Ctx  *context.ProgramContext
	Data *Data
}

func (n *Notification) ToTableRow() table.Row {
	return table.Row{
		n.renderType(),
		n.renderBookmark(),
		n.renderRepoName(),
		n.renderTitle(),
		n.renderNewComments(),
		n.renderReason(),
		n.renderUpdatedAt(),
	}
}

func (n *Notification) getTextStyle() lipgloss.Style {
	return components.GetIssueTextStyle(n.Ctx)
}

// getReadAwareStyle returns a style that is dimmed/faint for read notifications
func (n *Notification) getReadAwareStyle() lipgloss.Style {
	style := n.getTextStyle()
	if !n.Data.IsUnread() {
		// Dim read notifications
		style = style.Foreground(n.Ctx.Theme.FaintText)
	}
	return style
}

func (n *Notification) renderUnreadIndicator() string {
	if n.Data.IsUnread() {
		return lipgloss.NewStyle().Foreground(n.Ctx.Theme.PrimaryText).Render("")
	}
	return n.getTextStyle().Render(" ")
}

func (n *Notification) renderBookmark() string {
	if data.GetBookmarkStore().IsBookmarked(n.Data.GetId()) {
		return lipgloss.NewStyle().Foreground(n.Ctx.Theme.WarningText).Render("")
	}
	return " "
}

func (n *Notification) renderType() string {
	style := lipgloss.NewStyle()
	isRead := !n.Data.IsUnread()

	// For read notifications, use faint styling for all icons
	if isRead {
		style = style.Foreground(n.Ctx.Theme.FaintText)
	}

	switch n.Data.GetSubjectType() {
	case "PullRequest":
		// Use state-based icons/colors matching prrow.go
		switch n.Data.SubjectState {
		case "MERGED":
			if !isRead {
				style = style.Foreground(n.Ctx.Styles.Colors.MergedPR)
			}
			return style.Render(constants.MergedIcon)
		case "CLOSED":
			if !isRead {
				style = style.Foreground(n.Ctx.Styles.Colors.ClosedPR)
			}
			return style.Render(constants.ClosedIcon)
		default: // OPEN or unknown (not yet fetched)
			if n.Data.IsDraft {
				return style.Foreground(n.Ctx.Theme.FaintText).Render(constants.DraftIcon)
			}
			if !isRead {
				style = style.Foreground(n.Ctx.Styles.Colors.OpenPR)
			}
			return style.Render(constants.OpenIcon)
		}
	case "Issue":
		// Use state-based icons/colors matching issuerow.go
		if n.Data.SubjectState == "CLOSED" || isRead {
			return style.Render("")
		}
		return style.Foreground(n.Ctx.Styles.Colors.OpenIssue).Render("")
	case "Discussion":
		if !isRead {
			style = style.Foreground(n.Ctx.Theme.SecondaryText)
		}
		return style.Render("")
	case "Release":
		if !isRead {
			style = style.Foreground(n.Ctx.Theme.SuccessText)
		}
		return style.Render("")
	case "Commit":
		if !isRead {
			style = style.Foreground(n.Ctx.Theme.SecondaryText)
		}
		return style.Render("")
	default:
		return style.Render("")
	}
}

func (n *Notification) renderRepoName() string {
	return n.getReadAwareStyle().Render(n.Data.Notification.Repository.FullName)
}

func (n *Notification) renderTitle() string {
	style := n.getReadAwareStyle()
	if n.Data.IsUnread() {
		style = style.Bold(true)
	}
	title := n.Data.GetTitle()
	if n.Data.Actor != "" {
		actorStyle := lipgloss.NewStyle().Foreground(n.Ctx.Theme.ActorText)
		title = title + " " + actorStyle.Render("@"+n.Data.Actor)
	}
	return style.Render(title)
}

func (n *Notification) renderNewComments() string {
	if n.Data.NewCommentsCount <= 0 {
		return ""
	}
	style := lipgloss.NewStyle().Foreground(n.Ctx.Theme.SuccessText)
	return style.Render(fmt.Sprintf("+%d", n.Data.NewCommentsCount))
}

func (n *Notification) renderReason() string {
	reason := n.Data.GetReason()
	style := n.Ctx.Styles.Common.FaintTextStyle

	switch reason {
	case "review_requested":
		return style.Render("review")
	case "subscribed":
		return style.Render("subscribed")
	case "mention":
		return style.Foreground(n.Ctx.Theme.WarningText).Render("mention")
	case "author":
		return style.Render("author")
	case "comment":
		return style.Render("comment")
	case "assign":
		return style.Render("assigned")
	case "state_change":
		return style.Render("state")
	case "ci_activity":
		return style.Render("CI")
	default:
		return style.Render(reason)
	}
}

func (n *Notification) renderUpdatedAt() string {
	timeFormat := n.Ctx.Config.Defaults.DateFormat

	updatedAtOutput := ""
	if timeFormat == "" || timeFormat == "relative" {
		updatedAtOutput = utils.TimeElapsed(n.Data.GetUpdatedAt())
	} else {
		updatedAtOutput = n.Data.GetUpdatedAt().Format(timeFormat)
	}

	return n.getReadAwareStyle().Render(updatedAtOutput)
}
