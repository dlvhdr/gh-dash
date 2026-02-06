package notificationview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type Model struct {
	ctx   *context.ProgramContext
	row   *notificationrow.Data
	width int

	// Cached notification subject data for sidebar display
	subjectPR    *prrow.Data
	subjectIssue *data.IssueData
	subjectId    string // ID of the notification whose subject is cached
}

func NewModel(ctx *context.ProgramContext) Model {
	return Model{
		ctx: ctx,
	}
}

func (m *Model) SetRow(row *notificationrow.Data) {
	m.row = row
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) SetSubjectPR(pr *prrow.Data, notificationId string) {
	m.subjectPR = pr
	m.subjectIssue = nil
	m.subjectId = notificationId
}

func (m *Model) SetSubjectIssue(issue *data.IssueData, notificationId string) {
	m.subjectIssue = issue
	m.subjectPR = nil
	m.subjectId = notificationId
}

func (m *Model) GetSubjectPR() *prrow.Data {
	return m.subjectPR
}

func (m *Model) GetSubjectIssue() *data.IssueData {
	return m.subjectIssue
}

func (m *Model) GetSubjectId() string {
	return m.subjectId
}

func (m *Model) ClearSubject() {
	m.subjectPR = nil
	m.subjectIssue = nil
	m.subjectId = ""
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}

func (m Model) View() string {
	if m.row == nil {
		return ""
	}

	s := strings.Builder{}
	notification := m.row.Notification

	// Title - using common preview styling
	titleBlock := common.RenderPreviewTitle(m.ctx.Theme, m.ctx.Styles.Common, m.width, notification.Subject.Title)

	labelStyle := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.FaintText).
		Width(16)

	valueStyle := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.SecondaryText)

	faintValueStyle := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.FaintText)

	sectionStyle := lipgloss.NewStyle().
		PaddingBottom(1)

	s.WriteString(titleBlock)
	s.WriteString("\n\n")

	// Type with icon
	typeIcon := getTypeIcon(notification.Subject.Type)
	typeRow := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Type"),
		valueStyle.Render(fmt.Sprintf("%s %s", typeIcon, notification.Subject.Type)),
	)
	s.WriteString(sectionStyle.Render(typeRow))
	s.WriteString("\n")

	repoRow := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Repository"),
		valueStyle.Render(notification.Repository.FullName),
	)
	s.WriteString(sectionStyle.Render(repoRow))
	s.WriteString("\n")

	visibility := "Public"
	if notification.Repository.Private {
		visibility = "Private"
	}
	visibilityRow := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Visibility"),
		valueStyle.Render(visibility),
	)
	s.WriteString(sectionStyle.Render(visibilityRow))
	s.WriteString("\n")

	reasonRow := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Reason"),
		valueStyle.Render(formatReason(notification.Reason)),
	)
	s.WriteString(sectionStyle.Render(reasonRow))
	s.WriteString("\n")

	status := "Read"
	if notification.Unread {
		status = "Unread"
	}
	statusRow := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Status"),
		valueStyle.Render(status),
	)
	s.WriteString(sectionStyle.Render(statusRow))
	s.WriteString("\n")

	// Updated at
	updatedRow := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Updated"),
		valueStyle.Render(notification.UpdatedAt.Local().Format("Jan 2, 2006 3:04 PM")),
	)
	s.WriteString(sectionStyle.Render(updatedRow))
	s.WriteString("\n")

	// Last read at
	lastReadValue := "Never"
	if notification.LastReadAt != nil {
		lastReadValue = notification.LastReadAt.Local().Format("Jan 2, 2006 3:04 PM")
	}
	lastReadRow := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Last Read"),
		valueStyle.Render(lastReadValue),
	)
	s.WriteString(sectionStyle.Render(lastReadRow))
	s.WriteString("\n")

	hasComment := "No"
	if notification.Subject.LatestCommentUrl != "" {
		hasComment = "Yes"
	}
	commentRow := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Has Comment"),
		valueStyle.Render(hasComment),
	)
	s.WriteString(sectionStyle.Render(commentRow))
	s.WriteString("\n")

	idRow := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Notification ID"),
		faintValueStyle.Render(notification.Id),
	)
	s.WriteString(sectionStyle.Render(idRow))
	s.WriteString("\n")

	if notification.Subject.Url != "" {
		urlRow := lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render("API URL"),
			faintValueStyle.Render(notification.Subject.Url),
		)
		s.WriteString(sectionStyle.Render(urlRow))
	}

	return s.String()
}

func getTypeIcon(subjectType string) string {
	switch subjectType {
	case "PullRequest":
		return ""
	case "Issue":
		return ""
	case "Discussion":
		return ""
	case "Release":
		return ""
	case "Commit":
		return ""
	case "CheckSuite":
		return ""
	default:
		return ""
	}
}

func formatReason(reason string) string {
	switch reason {
	case "subscribed":
		return "Subscribed"
	case "review_requested":
		return "Review requested"
	case "author":
		return "Author"
	case "comment":
		return "Comment"
	case "mention":
		return "Mentioned"
	case "team_mention":
		return "Team mentioned"
	case "state_change":
		return "State changed"
	case "assign":
		return "Assigned"
	case "ci_activity":
		return "CI activity"
	case "approval_requested":
		return "Approval requested"
	default:
		return reason
	}
}
