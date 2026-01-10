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
		n.renderTitleBlock(),
		n.renderActivity(),
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
			return style.Render("")
		}
		return style.Foreground(n.Ctx.Styles.Colors.OpenIssue).Render("")
	case "Discussion":
		if !isRead {
			style = style.Foreground(n.Ctx.Theme.SecondaryText)
		}
		return style.Render("")
	case "Release":
		if !isRead {
			style = style.Foreground(n.Ctx.Theme.SuccessText)
		}
		return style.Render("")
	case "Commit":
		if !isRead {
			style = style.Foreground(n.Ctx.Theme.SecondaryText)
		}
		return style.Render("")
	case "CheckSuite":
		if !isRead {
			style = style.Foreground(n.Ctx.Theme.WarningText)
		}
		return style.Render(constants.WorkflowIcon)
	case "RepositoryVulnerabilityAlert":
		if !isRead {
			style = style.Foreground(n.Ctx.Theme.ErrorText)
		}
		return style.Render(constants.SecurityIcon)
	default:
		// Generic notification icon for unknown types
		if !isRead {
			style = style.Foreground(n.Ctx.Theme.SecondaryText)
		}
		return style.Render(constants.NotificationIcon)
	}
}

// getStylePrefix extracts ANSI codes from a lipgloss style without the trailing reset.
// This allows styled text to be concatenated without breaking parent background colors.
func getStylePrefix(s lipgloss.Style) string {
	rendered := s.Render("")
	// Strip trailing reset sequence if present
	if len(rendered) >= 4 && rendered[len(rendered)-4:] == "\x1b[0m" {
		return rendered[:len(rendered)-4]
	}
	return rendered
}

// renderTitleBlock returns a 3-line block:
// Line 1: repo/name #number [bookmark icon if bookmarked]
// Line 2: Title (bold for unread)
// Line 3: Activity description
// Note: Truncation is handled dynamically by the table component based on actual column width
// Note: Uses raw ANSI codes without resets to preserve parent background colors
func (n *Notification) renderTitleBlock() string {
	// Line 1: repo #number (secondary color, no ANSI reset to preserve background)
	repoStyle := lipgloss.NewStyle().Foreground(n.Ctx.Theme.SecondaryText)
	if !n.Data.IsUnread() {
		// Dim for read notifications
		repoStyle = lipgloss.NewStyle().Foreground(n.Ctx.Theme.FaintText)
	}
	repoPrefix := getStylePrefix(repoStyle)
	repo := n.Data.GetRepoNameWithOwner()
	number := n.Data.GetNumber()
	line1 := repo
	if number > 0 {
		line1 = fmt.Sprintf("%s #%d", repo, number)
	}
	// Add bookmark icon if bookmarked (using raw ANSI for warning color)
	if data.GetBookmarkStore().IsBookmarked(n.Data.GetId()) {
		bookmarkPrefix := getStylePrefix(lipgloss.NewStyle().Foreground(n.Ctx.Theme.WarningText))
		line1 = line1 + " " + bookmarkPrefix + ""
	}
	line1Rendered := repoPrefix + line1

	// Line 2: Title (bold for unread)
	titleStyle := n.getReadAwareStyle()
	if n.Data.IsUnread() {
		titleStyle = titleStyle.Bold(true)
	}
	titlePrefix := getStylePrefix(titleStyle)
	title := n.Data.GetTitle()
	line2Rendered := titlePrefix + title

	// Line 3: Activity description (no ANSI reset)
	activityPrefix := getStylePrefix(lipgloss.NewStyle().Foreground(n.Ctx.Theme.FaintText))
	line3 := n.Data.ActivityDescription
	if line3 == "" {
		// Fallback to reason-based description
		line3 = n.getReasonDescription()
	}
	line3Rendered := activityPrefix + line3

	return line1Rendered + "\n" + line2Rendered + "\n" + line3Rendered
}

// getReasonDescription returns a fallback description based on notification reason
func (n *Notification) getReasonDescription() string {
	reason := n.Data.GetReason()
	subjectType := n.Data.GetSubjectType()

	switch reason {
	case "review_requested":
		return "Review requested"
	case "subscribed":
		return "Activity on subscribed thread"
	case "mention":
		return "You were mentioned"
	case "author":
		return "Activity on your thread"
	case "comment":
		switch subjectType {
		case "PullRequest":
			return "New comment on pull request"
		case "Issue":
			return "New comment on issue"
		}
		return "New comment"
	case "assign":
		return "You were assigned"
	case "state_change":
		return "State changed"
	case "ci_activity":
		return "CI activity"
	default:
		return ""
	}
}

// renderActivity shows the new comments count with icon
func (n *Notification) renderActivity() string {
	if n.Data.NewCommentsCount <= 0 {
		return ""
	}
	// Use raw ANSI foreground codes without reset to avoid breaking row background
	// White foreground for count, green foreground for icon
	white := "\x1b[97m" // Bright white
	green := "\x1b[32m" // Green
	return white + fmt.Sprintf("+%d ", n.Data.NewCommentsCount) + green + constants.CommentsIcon
}

func (n *Notification) renderUpdatedAt() string {
	timeFormat := n.Ctx.Config.Defaults.DateFormat

	updatedAtOutput := ""
	if timeFormat == "" || timeFormat == "relative" {
		// Use non-breaking space (U+00A0) to prevent wrap
		updatedAtOutput = utils.TimeElapsed(n.Data.GetUpdatedAt()) + "\u00A0ago"
	} else {
		updatedAtOutput = n.Data.GetUpdatedAt().Format(timeFormat)
	}

	return n.getReadAwareStyle().Render(updatedAtOutput)
}
