package notificationrow

import (
	"fmt"
	"strings"

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

// getReadAwareStyle returns the text style for notifications.
// Read/unread status is indicated by the blue dot, not text color.
func (n *Notification) getReadAwareStyle() lipgloss.Style {
	return n.getTextStyle()
}

func (n *Notification) renderType() string {
	style := lipgloss.NewStyle()
	isUnread := n.Data.IsUnread()

	// Icons are always colored - only the blue dot indicates unread status
	var icon string
	switch n.Data.GetSubjectType() {
	case "PullRequest":
		// Use state-based icons/colors matching prrow.go
		switch n.Data.SubjectState {
		case "MERGED":
			style = style.Foreground(n.Ctx.Styles.Colors.MergedPR)
			icon = constants.MergedIcon
		case "CLOSED":
			style = style.Foreground(n.Ctx.Styles.Colors.ClosedPR)
			icon = constants.ClosedIcon
		default: // OPEN or unknown (not yet fetched)
			if n.Data.IsDraft {
				style = style.Foreground(n.Ctx.Theme.FaintText)
				icon = constants.DraftIcon
			} else {
				style = style.Foreground(n.Ctx.Styles.Colors.OpenPR)
				icon = constants.OpenIcon
			}
		}
	case "Issue":
		// Use state-based icons/colors matching issuerow.go
		if n.Data.SubjectState == "CLOSED" {
			icon = ""
			style = style.Foreground(n.Ctx.Styles.Colors.ClosedPR)
		} else {
			icon = ""
			style = style.Foreground(n.Ctx.Styles.Colors.OpenIssue)
		}
	case "Discussion":
		icon = ""
		style = style.Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"})
	case "Release":
		icon = ""
		style = style.Foreground(lipgloss.AdaptiveColor{Light: "#0969da", Dark: "#58a6ff"})
	case "Commit":
		icon = ""
		style = style.Foreground(n.Ctx.Theme.SecondaryText)
	case "CheckSuite":
		// Parse title to determine workflow status (similar to gitify approach)
		title := strings.ToLower(n.Data.GetTitle())
		switch {
		case strings.Contains(title, "failed"):
			icon = constants.FailureIcon
			style = style.Foreground(n.Ctx.Theme.ErrorText)
		case strings.Contains(title, "succeeded"):
			icon = constants.SuccessIcon
			style = style.Foreground(n.Ctx.Theme.SuccessText)
		case strings.Contains(title, "cancelled"), strings.Contains(title, "canceled"):
			icon = constants.WorkflowRunIcon
			style = style.Foreground(n.Ctx.Theme.FaintText)
		case strings.Contains(title, "skipped"):
			icon = constants.WorkflowRunIcon
			style = style.Foreground(n.Ctx.Theme.FaintText)
		default:
			icon = constants.WorkflowRunIcon
			style = style.Foreground(n.Ctx.Theme.WarningText)
		}
	case "RepositoryVulnerabilityAlert":
		icon = constants.SecurityIcon
		style = style.Foreground(n.Ctx.Theme.ErrorText)
	default:
		// Generic notification icon for unknown types
		icon = constants.NotificationIcon
		style = style.Foreground(n.Ctx.Theme.SecondaryText)
	}

	// Use raw ANSI codes without reset to preserve parent background colors
	iconPrefix := utils.GetStylePrefix(style)

	// Always render 3 lines to match the Title column (repo, title, activity)
	// This ensures proper background color when row is selected
	// Line 1: icon
	// Line 2: blue dot (unread) or empty
	// Line 3: empty
	if isUnread {
		dotPrefix := utils.GetStylePrefix(lipgloss.NewStyle().Foreground(lipgloss.Color("33")))
		return iconPrefix + icon + "\n" + dotPrefix + constants.DotIcon + "\n"
	}
	return iconPrefix + icon + "\n\n"
}

// renderTitleBlock returns a 3-line block:
// Line 1: repo/name #number [bookmark icon if bookmarked]
// Line 2: Title (bold for unread)
// Line 3: Activity description
// Note: Truncation is handled dynamically by the table component based on actual column width
// Note: Uses raw ANSI codes without resets to preserve parent background colors
func (n *Notification) renderTitleBlock() string {
	// Line 1: repo #number (secondary color, no ANSI reset to preserve background)
	// Read/unread status is indicated by the blue dot, not text color
	repoStyle := lipgloss.NewStyle().Foreground(n.Ctx.Theme.SecondaryText)
	repoPrefix := utils.GetStylePrefix(repoStyle)
	repo := n.Data.GetRepoNameWithOwner()
	number := n.Data.GetNumber()
	line1 := repo
	if number > 0 {
		line1 = fmt.Sprintf("%s #%d", repo, number)
	}
	// Add bookmark icon if bookmarked (using raw ANSI for warning color)
	if data.GetBookmarkStore().IsBookmarked(n.Data.GetId()) {
		bookmarkPrefix := utils.GetStylePrefix(lipgloss.NewStyle().Foreground(n.Ctx.Theme.WarningText))
		line1 = line1 + " " + bookmarkPrefix + ""
	}
	line1Rendered := repoPrefix + line1

	// Line 2: Title (bold for unread)
	titleStyle := n.getReadAwareStyle()
	if n.Data.IsUnread() {
		titleStyle = titleStyle.Bold(true)
	}
	titlePrefix := utils.GetStylePrefix(titleStyle)
	title := n.Data.GetTitle()
	line2Rendered := titlePrefix + title

	// Line 3: Activity description (no ANSI reset)
	activityPrefix := utils.GetStylePrefix(lipgloss.NewStyle().Foreground(n.Ctx.Theme.FaintText))
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
// Returns 3 lines to match Title column for proper background highlighting
func (n *Notification) renderActivity() string {
	if n.Data.NewCommentsCount <= 0 {
		return "\n\n"
	}
	// Use raw ANSI foreground codes without reset to avoid breaking row background
	// White foreground for count, green foreground for icon
	white := "\x1b[97m" // Bright white
	green := "\x1b[32m" // Green
	return white + fmt.Sprintf("+%d ", n.Data.NewCommentsCount) + green + constants.CommentsIcon + "\n\n"
}

// renderUpdatedAt returns the time since last update
// Returns 3 lines to match Title column for proper background highlighting
func (n *Notification) renderUpdatedAt() string {
	timeFormat := n.Ctx.Config.Defaults.DateFormat

	updatedAtOutput := ""
	if timeFormat == "" || timeFormat == "relative" {
		// Use non-breaking space (U+00A0) to prevent wrap
		updatedAtOutput = utils.TimeElapsed(n.Data.GetUpdatedAt()) + "\u00A0ago"
	} else {
		updatedAtOutput = n.Data.GetUpdatedAt().Format(timeFormat)
	}

	// Use raw ANSI codes without reset to preserve parent background colors
	stylePrefix := utils.GetStylePrefix(n.getReadAwareStyle())
	return stylePrefix + updatedAtOutput + "\n\n"
}
