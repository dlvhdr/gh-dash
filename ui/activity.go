package ui

import (
	"sort"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/data"
	"github.com/dlvhdr/gh-prs/ui/markdown"
	"github.com/dlvhdr/gh-prs/utils"
)

type RenderedActivity struct {
	UpdatedAt      time.Time
	RenderedString string
}

func (sidebar *Sidebar) renderActivity() string {
	width := sidebar.model.getSidebarWidth() - 8
	markdownRenderer := markdown.GetMarkdownRenderer(width)

	var activity []RenderedActivity
	for _, comment := range sidebar.pr.Data.Comments.Nodes {
		renderedComment, err := renderComment(comment, markdownRenderer)
		if err != nil {
			continue
		}
		activity = append(activity, RenderedActivity{
			UpdatedAt:      comment.UpdatedAt,
			RenderedString: renderedComment,
		})
	}

	for _, review := range sidebar.pr.Data.LatestReviews.Nodes {
		renderedReview, err := renderReview(review, markdownRenderer)
		if err != nil {
			continue
		}
		activity = append(activity, RenderedActivity{
			UpdatedAt:      review.UpdatedAt,
			RenderedString: renderedReview,
		})
	}

	sort.Slice(activity, func(i, j int) bool {
		return activity[i].UpdatedAt.After(activity[j].UpdatedAt)
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
	return mainTextStyle.Copy().
		MarginBottom(1).
		Underline(true).
		Render(" Comments")
}

func renderEmptyState() string {
	return lipgloss.NewStyle().Italic(true).Render("No comments...")
}

func renderComment(comment data.Comment, markdownRenderer glamour.TermRenderer) (string, error) {
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		mainTextStyle.Copy().Render(comment.Author.Login),
		" ",
		lipgloss.NewStyle().Foreground(faintText).Render(utils.TimeElapsed(comment.UpdatedAt)),
	)
	body, err := markdownRenderer.Render(comment.Body)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
	), err
}

func renderReview(review data.Review, markdownRenderer glamour.TermRenderer) (string, error) {
	header := renderReviewHeader(review)
	body, err := markdownRenderer.Render(review.Body)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
	), err
}

func renderReviewHeader(review data.Review) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		renderReviewDecision(review.State),
		" ",
		mainTextStyle.Copy().Render(review.Author.Login),
		" ",
		lipgloss.NewStyle().Foreground(faintText).Render("reviewed "+utils.TimeElapsed(review.UpdatedAt)),
	)
}

func renderReviewDecision(decision string) string {
	switch decision {
	case "PENDING":
		return waitingGlyph
	case "COMMENTED":
		return lipgloss.NewStyle().Foreground(faintText).Render("")
	case "APPROVED":
		return successGlyph
	case "CHANGES_REQUESTED":
		return failureGlyph
	}

	return ""
}
