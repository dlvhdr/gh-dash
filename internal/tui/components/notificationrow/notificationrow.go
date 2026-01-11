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
		icon = ""
		if n.Data.SubjectState == "CLOSED" {
			style = style.Foreground(n.Ctx.Styles.Colors.ClosedPR)
		} else {
			style = style.Foreground(n.Ctx.Styles.Colors.OpenIssue)
		}
	case "Discussion":
		icon = ""
		style = style.Foreground(n.Ctx.Theme.SecondaryText)
	case "Release":
		icon = ""
		style = style.Foreground(n.Ctx.Theme.SuccessText)
	case "Commit":
		icon = ""
		style = style.Foreground(n.Ctx.Theme.SecondaryText)
	case "CheckSuite":
		icon = constants.WorkflowIcon
		style = style.Foreground(n.Ctx.Theme.WarningText)
	case "RepositoryVulnerabilityAlert":
		icon = constants.SecurityIcon
		style = style.Foreground(n.Ctx.Theme.ErrorText)
	default:
		// Generic notification icon for unknown types
		icon = constants.NotificationIcon
		style = style.Foreground(n.Ctx.Theme.SecondaryText)
	}

	// Add blue dot below icon for unread notifications (like GitHub's UI)
	if isUnread {
		dotStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
		return style.Render(icon) + "\n" + dotStyle.Render(constants.DotIcon)
	}
	return style.Render(icon)
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
	// Read/unread status is indicated by the blue dot, not text color
	repoStyle := lipgloss.NewStyle().Foreground(n.Ctx.Theme.SecondaryText)
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
