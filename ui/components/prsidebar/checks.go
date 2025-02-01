package prsidebar

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/data"
)

func (sidebar *Model) renderChecks() string {
	title := sidebar.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Render(" Checks")
	w := sidebar.getIndentedContentWidth()
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(sidebar.ctx.Theme.ErrorText).Width(w)
	review := sidebar.viewCheckCategory("Review required", "Code owner review required", false)

	checksBar := sidebar.viewChecksBar()
	checksBottom := lipgloss.JoinVertical(lipgloss.Left, "1 failing, 2 skipped, 15 successful", checksBar)
	checks := sidebar.viewCheckCategory("Some checks were not successful", checksBottom, false)

	merge := sidebar.viewCheckCategory("Merging is blocked", "Waiting on code owner review", true)

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		box.Render(lipgloss.JoinVertical(lipgloss.Left, review, checks, merge)),
	)

	// commits := sidebar.pr.Data.Commits.Nodes
	// if len(commits) == 0 {
	// 	return ""
	// }
	//
	// var checks []string
	// for _, review := range sidebar.pr.Data.LatestReviews.Nodes {
	// 	checks = append(checks, sidebar.renderReviewHeader(review))
	// }
	//
	// lastCommit := commits[0]
	// for _, node := range lastCommit.Commit.StatusCheckRollup.Contexts.Nodes {
	// 	if node.Typename == "CheckRun" {
	// 		checkRun := node.CheckRun
	// 		renderedStatus := sidebar.renderCheckRunConclusion(checkRun)
	// 		name := renderCheckRunName(checkRun)
	// 		checks = append(checks, lipgloss.JoinHorizontal(lipgloss.Top, renderedStatus, " ", name))
	// 	} else if node.Typename == "StatusContext" {
	// 		statusContext := node.StatusContext
	// 		status := sidebar.renderStatusContextConclusion(statusContext)
	// 		checks = append(checks, lipgloss.JoinHorizontal(lipgloss.Top, status, " ", renderStatusContextName(statusContext)))
	// 	}
	// }
	//
	// if len(checks) == 0 {
	// 	return lipgloss.JoinVertical(
	// 		lipgloss.Left,
	// 		title,
	// 		lipgloss.NewStyle().
	// 			Italic(true).
	// 			PaddingLeft(2).
	// 			Width(sidebar.getIndentedContentWidth()).
	// 			Render("No checks to display..."),
	// 	)
	// }
	//
	// renderedChecks := lipgloss.JoinVertical(lipgloss.Left, checks...)
	// return lipgloss.JoinVertical(
	// 	lipgloss.Left,
	// 	title,
	// 	lipgloss.NewStyle().PaddingLeft(2).Width(sidebar.getIndentedContentWidth()).Render(renderedChecks),
	// )
}

func (m *Model) viewCheckCategory(top, bottom string, isLast bool) string {
	w := m.getIndentedContentWidth()
	part := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, !isLast, false).
		BorderForeground(m.ctx.Theme.FaintBorder).
		Width(w).
		Padding(1)
	title := lipgloss.NewStyle().Bold(true)
	subtitle := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText)

	category := m.ctx.Styles.Common.FailureGlyph + " "
	margin := lipgloss.Width(category)
	category = lipgloss.JoinHorizontal(lipgloss.Top, category, title.Render(top))
	category = lipgloss.JoinVertical(lipgloss.Left, category, subtitle.MarginLeft(margin).Render(bottom))
	return part.Render(category)
}

func (m *Model) viewChecksBar() string {
	w := m.getIndentedContentWidth() - 4
	numSuccess := 8.0
	numFailing := 3.0
	numSkipped := 1.0
	total := numSuccess + numFailing + numSkipped

	succW := int(math.Floor((numSuccess / total) * float64(w)))
	failW := int(math.Floor((numFailing / total) * float64(w)))
	skipW := int(math.Floor((numSkipped / total) * float64(w)))

	succBar := lipgloss.NewStyle().Width(succW).Foreground(m.ctx.Theme.SuccessText).Height(1).Render(strings.Repeat("▃", succW))
	failBar := lipgloss.NewStyle().Width(failW).Foreground(m.ctx.Theme.ErrorText).Height(1).Render(strings.Repeat("▃", failW))
	skipBar := lipgloss.NewStyle().Width(skipW).Foreground(m.ctx.Theme.FaintText).Height(1).Render(strings.Repeat("▃", skipW))

	return lipgloss.JoinHorizontal(lipgloss.Top, skipBar, failBar, succBar)
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
