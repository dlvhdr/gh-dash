package prsidebar

import (
	"fmt"
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

	var activities []RenderedActivity
	var comments []comment

	for _, review := range m.pr.Data.ReviewThreads.Nodes {
		path := review.Path
		line := review.Line
		for _, c := range review.Comments.Nodes {
			comments = append(comments, comment{
				Author:    c.Author.Login,
				Body:      c.Body,
				UpdatedAt: c.UpdatedAt,
				Path:      &path,
				Line:      &line,
			})
		}
	}

	for _, c := range m.pr.Data.Comments.Nodes {
		comments = append(comments, comment{
			Author:    c.Author.Login,
			Body:      c.Body,
			UpdatedAt: c.UpdatedAt,
		})
	}

	for _, comment := range comments {
		renderedComment, err := m.renderComment(comment, markdownRenderer)
		if err != nil {
			continue
		}
		activities = append(activities, RenderedActivity{
			UpdatedAt:      comment.UpdatedAt,
			RenderedString: renderedComment,
		})
	}

	for _, review := range m.pr.Data.LatestReviews.Nodes {
		renderedReview, err := m.renderReview(review, markdownRenderer)
		if err != nil {
			continue
		}
		activities = append(activities, RenderedActivity{
			UpdatedAt:      review.UpdatedAt,
			RenderedString: renderedReview,
		})
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].UpdatedAt.Before(activities[j].UpdatedAt)
	})

	body := ""
	bodyStyle := lipgloss.NewStyle().PaddingLeft(2)
	if len(activities) == 0 {
		body = renderEmptyState()
	} else {
		var renderedActivities []string
		for _, activity := range activities {
			renderedActivities = append(renderedActivities, activity.RenderedString)
		}
		body = lipgloss.JoinVertical(lipgloss.Left, renderedActivities...)
	}

	return lipgloss.JoinVertical(lipgloss.Left, m.renderActivityTitle(), bodyStyle.Render(body))
}

func (m *Model) renderActivityTitle() string {
	return m.ctx.Styles.Common.MainTextStyle.
		MarginBottom(1).
		Underline(true).
		Render(" Comments")
}

func renderEmptyState() string {
	return lipgloss.NewStyle().Italic(true).Render("No comments...")
}

type comment struct {
	Author    string
	UpdatedAt time.Time
	Body      string
	Path      *string
	Line      *int
}

func (m *Model) renderComment(comment comment, markdownRenderer glamour.TermRenderer) (string, error) {
	width := m.getIndentedContentWidth()
	authorAndTime := lipgloss.NewStyle().
		Width(width).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.ctx.Theme.FaintBorder).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			m.ctx.Styles.Common.MainTextStyle.Render(comment.Author),
			" ",
			lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(utils.TimeElapsed(comment.UpdatedAt)),
		))

	var header string
	if comment.Path != nil && comment.Line != nil {
		filePath := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Width(width).Render(
			fmt.Sprintf(
				"%s#l%d",
				*comment.Path,
				*comment.Line,
			),
		)
		header = lipgloss.JoinVertical(lipgloss.Left, authorAndTime, filePath, "")
	} else {
		header = authorAndTime
	}

	regex := regexp.MustCompile(`((\n)+|^)([^\r\n]*\|[^\r\n]*(\n)?)+`)
	body := regex.ReplaceAllString(comment.Body, "")
	body, err := markdownRenderer.Render(body)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
	), err
}

func (m *Model) renderReview(review data.Review, markdownRenderer glamour.TermRenderer) (string, error) {
	header := m.renderReviewHeader(review)
	body, err := markdownRenderer.Render(review.Body)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
	), err
}

func (m *Model) renderReviewHeader(review data.Review) string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderReviewDecision(review.State),
		" ",
		m.ctx.Styles.Common.MainTextStyle.Render(review.Author.Login),
		" ",
		lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render("reviewed "+utils.TimeElapsed(review.UpdatedAt)),
	)
}

func (m *Model) renderReviewDecision(decision string) string {
	switch decision {
	case "PENDING":
		return m.ctx.Styles.Common.WaitingGlyph
	case "COMMENTED":
		return lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render("󰈈")
	case "APPROVED":
		return m.ctx.Styles.Common.SuccessGlyph
	case "CHANGES_REQUESTED":
		return m.ctx.Styles.Common.FailureGlyph
	}

	return ""
}
