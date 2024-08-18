package issuesidebar

import (
	"regexp"
	"sort"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/markdown"
	"github.com/dlvhdr/gh-dash/v4/utils"
)

type RenderedActivity struct {
	UpdatedAt      time.Time
	RenderedString string
}

func (m *Model) renderActivity() string {
	width := m.getIndentedContentWidth() - 2
	markdownRenderer := markdown.GetMarkdownRenderer(width)

	var activity []RenderedActivity
	for _, comment := range m.issue.Data.Comments.Nodes {
		renderedComment, err := m.renderComment(comment, markdownRenderer)
		if err != nil {
			continue
		}
		activity = append(activity, RenderedActivity{
			UpdatedAt:      comment.UpdatedAt,
			RenderedString: renderedComment,
		})
	}

	sort.Slice(activity, func(i, j int) bool {
		return activity[i].UpdatedAt.Before(activity[j].UpdatedAt)
	})

	body := ""
	bodyStyle := lipgloss.NewStyle().PaddingLeft(2)
	if len(activity) == 0 {
		body = renderEmptyState()
	} else {
		var renderedActivities []string
		for _, activity := range activity {
			renderedActivities = append(renderedActivities, activity.RenderedString)
		}
		body = lipgloss.JoinVertical(lipgloss.Left, renderedActivities...)
	}

	return lipgloss.JoinVertical(lipgloss.Left, m.renderActivitiesTitle(), bodyStyle.Render(body))
}

func (m Model) renderActivitiesTitle() string {
	return m.ctx.Styles.Common.MainTextStyle.
		MarginBottom(1).
		Underline(true).
		Render(" Comments")
}

func renderEmptyState() string {
	return lipgloss.NewStyle().Italic(true).Render("No comments...")
}

func (m *Model) renderComment(comment data.IssueComment, markdownRenderer glamour.TermRenderer) (string, error) {
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		m.ctx.Styles.Common.MainTextStyle.Render(comment.Author.Login),
		" ",
		lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(utils.TimeElapsed(comment.UpdatedAt)),
	)

	regex := regexp.MustCompile(`((\n)+|^)([^\r\n]*\|[^\r\n]*(\n)?)+`)
	body := regex.ReplaceAllString(comment.Body, "")
	body, err := markdownRenderer.Render(body)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
	), err
}
