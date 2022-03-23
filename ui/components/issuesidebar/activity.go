package issuesidebar

import (
	"regexp"
	"sort"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/data"
	"github.com/dlvhdr/gh-prs/ui/markdown"
	"github.com/dlvhdr/gh-prs/ui/styles"
	"github.com/dlvhdr/gh-prs/utils"
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
		renderedComment, err := renderComment(comment, markdownRenderer)
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

	return lipgloss.JoinVertical(lipgloss.Left, renderTitle(), bodyStyle.Render(body))
}

func renderTitle() string {
	return styles.MainTextStyle.Copy().
		MarginBottom(1).
		Underline(true).
		Render(" Comments")
}

func renderEmptyState() string {
	return lipgloss.NewStyle().Italic(true).Render("No comments...")
}

func renderComment(comment data.Comment, markdownRenderer glamour.TermRenderer) (string, error) {
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		styles.MainTextStyle.Copy().Render(comment.Author.Login),
		" ",
		lipgloss.NewStyle().Foreground(styles.DefaultTheme.FaintText).Render(utils.TimeElapsed(comment.UpdatedAt)),
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
