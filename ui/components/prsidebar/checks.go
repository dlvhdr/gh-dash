package prsidebar

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/v4/data"
)

func (sidebar *Model) renderChecks() string {
	title := sidebar.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Render(" Checks")

	commits := sidebar.pr.Data.Commits.Nodes
	if len(commits) == 0 {
		return ""
	}

	var checks []string
	for _, review := range sidebar.pr.Data.LatestReviews.Nodes {
		checks = append(checks, sidebar.renderReviewHeader(review))
	}

	lastCommit := commits[0]
	for _, node := range lastCommit.Commit.StatusCheckRollup.Contexts.Nodes {
		if node.Typename == "CheckRun" {
			checkRun := node.CheckRun
			renderedStatus := sidebar.renderCheckRunConclusion(checkRun)
			name := renderCheckRunName(checkRun)
			checks = append(checks, lipgloss.JoinHorizontal(lipgloss.Top, renderedStatus, " ", name))
		} else if node.Typename == "StatusContext" {
			statusContext := node.StatusContext
			status := sidebar.renderStatusContextConclusion(statusContext)
			checks = append(checks, lipgloss.JoinHorizontal(lipgloss.Top, status, " ", renderStatusContextName(statusContext)))
		}
	}

	if len(checks) == 0 {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			lipgloss.NewStyle().
				Italic(true).
				PaddingLeft(2).
				Width(sidebar.getIndentedContentWidth()).
				Render("No checks to display..."),
		)
	}

	renderedChecks := lipgloss.JoinVertical(lipgloss.Left, checks...)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().PaddingLeft(2).Width(sidebar.getIndentedContentWidth()).Render(renderedChecks),
	)
}

func (m *Model) renderCheckRunConclusion(checkRun data.CheckRun) string {
	conclusionStr := string(checkRun.Conclusion)
	if data.IsStatusWaiting(string(checkRun.Status)) {
		return m.ctx.Styles.Common.WaitingGlyph
	}

	if data.IsConclusionAFailure(conclusionStr) {
		return m.ctx.Styles.Common.FailureGlyph
	}

	return m.ctx.Styles.Common.SuccessGlyph
}

func (m *Model) renderStatusContextConclusion(statusContext data.StatusContext) string {
	conclusionStr := string(statusContext.State)
	if data.IsStatusWaiting(conclusionStr) {
		return m.ctx.Styles.Common.WaitingGlyph
	}

	if data.IsConclusionAFailure(conclusionStr) {
		return m.ctx.Styles.Common.FailureGlyph
	}

	return m.ctx.Styles.Common.SuccessGlyph
}

func renderCheckRunName(checkRun data.CheckRun) string {
	var parts []string
	creator := strings.TrimSpace(string(checkRun.CheckSuite.Creator.Login))
	if creator != "" {
		parts = append(parts, creator)
	}

	workflow := strings.TrimSpace(string(checkRun.CheckSuite.WorkflowRun.Workflow.Name))
	if workflow != "" {
		parts = append(parts, workflow)
	}

	name := strings.TrimSpace(string(checkRun.Name))
	if name != "" {
		parts = append(parts, name)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		strings.Join(parts, "/"),
	)
}

func renderStatusContextName(statusContext data.StatusContext) string {
	var parts []string
	creator := strings.TrimSpace(string(statusContext.Creator.Login))
	if creator != "" {
		parts = append(parts, creator)
	}

	context := strings.TrimSpace(string(statusContext.Context))
	if context != "" && context != "/" {
		parts = append(parts, context)
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		strings.Join(parts, "/"),
	)
}
