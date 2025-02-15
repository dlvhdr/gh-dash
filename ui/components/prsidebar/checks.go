package prsidebar

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/data"
)

func (m *Model) renderChecks() string {
	title := m.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Render(" Checks")
	w := m.getIndentedContentWidth()
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(m.ctx.Theme.ErrorText).Width(w)
	// review := sidebar.viewCheckCategory("Review required", "Code owner review required", false)
	review := m.viewReview()

	stats := m.getChecksStats()

	statStrs := make([]string, 0)
	checksConclusion := m.viewChecksConclusionTitle()
	if stats.failed > 0 {
		statStrs = append(statStrs, fmt.Sprintf("%d failing", stats.failed))
	}
	if stats.inProgress > 0 {
		statStrs = append(statStrs, fmt.Sprintf("%d in progress", stats.inProgress))
	}
	if stats.skipped > 0 {
		statStrs = append(statStrs, fmt.Sprintf("%d skipped", stats.skipped))
	}
	if stats.succeeded > 0 {
		statStrs = append(statStrs, fmt.Sprintf("%d successful", stats.succeeded))
	}

	checks := ""
	if checksConclusion != "" {
		checksBar := m.viewChecksBar()
		checksBottom := lipgloss.JoinVertical(lipgloss.Left, strings.Join(statStrs, ", "), checksBar)
		checks = m.viewCheckCategory(checksConclusion, checksBottom, false)
	}

	mergeTitle := ""
	mergeSub := ""
	numReviewOwners := m.numReviewOwners()
	if m.pr.Data.MergeStateStatus == "CLEAN" {
		mergeTitle = m.ctx.Styles.Common.SuccessGlyph + " No conflicts with base branch"
		mergeSub = "Changes can be cleanly merged"
	} else if m.pr.Data.MergeStateStatus == "BLOCKED" {
		mergeTitle = m.ctx.Styles.Common.FailureGlyph + " Merging is blocked"
		if numReviewOwners > 0 {
			mergeSub = "Waiting on code owner review"
		}
	} else if m.pr.Data.Mergeable == "CONFLICTING" {
		mergeTitle = m.ctx.Styles.Common.FailureGlyph + " This branch has conflicts that must be resolved"
		if m.pr.Data.MergeStateStatus == "CLEAN" {
			mergeSub = "Changes can be cleanly merged"
		}
	}
	merge := m.viewCheckCategory(mergeTitle, mergeSub, true)

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		box.Render(lipgloss.JoinVertical(lipgloss.Left, review, checks, merge)),
	)

}

func (m *Model) viewReview() string {
	pr := m.pr
	if pr.Data == nil {
		return ""
	}

	var title, subtitle string
	reviewCellStyle := lipgloss.NewStyle()
	numReviewOwners := m.numReviewOwners()

	numApproving, numChangesRequested, numPending, numCommented := 0, 0, 0, 0
	for _, node := range pr.Data.LatestReviews.Nodes {
		if node.State == "APPROVED" {
			numApproving++
		} else if node.State == "CHANGES_REQUESTED" {
			numChangesRequested++
		} else if node.State == "PENDING" {
			numPending++
		} else if node.State == "COMMENTED" {
			numCommented++
		}
	}

	if pr.Data.ReviewDecision == "APPROVED" {
		title = lipgloss.JoinHorizontal(lipgloss.Top, m.ctx.Styles.Common.SuccessGlyph, " ", "Changes approved")
		subtitle = fmt.Sprintf("%d approving reviews", numApproving)
	} else if pr.Data.ReviewDecision == "CHANGES_REQUESTED" {
		title = lipgloss.JoinHorizontal(lipgloss.Top, m.ctx.Styles.Common.FailureGlyph, " ", "Changes requested")
		subtitle = fmt.Sprintf("%d requested changes", numChangesRequested)
	} else if pr.Data.ReviewDecision == "REVIEW_REQUIRED" {
		title = reviewCellStyle.Render(pr.Ctx.Styles.Common.WaitingGlyph + " " + "Review Required")
		if numReviewOwners > 0 {
			subtitle = "Code owner review required"
		} else if numCommented > 0 {
			subtitle = fmt.Sprintf("%d reviewers left comments", numCommented)
		}
	}

	return m.viewCheckCategory(title, subtitle, false)
}

func (m *Model) viewCheckCategory(top, bottom string, isLast bool) string {
	w := m.getIndentedContentWidth()
	part := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, !isLast, false).
		BorderForeground(m.ctx.Theme.FaintBorder).
		Width(w).
		Padding(1)
	title := lipgloss.NewStyle().Bold(true)

	category := title.Render(top)
	if bottom != "" {
		subtitle := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText)
		category = lipgloss.JoinVertical(lipgloss.Left, category, subtitle.MarginLeft(2).Render(bottom))
	}
	return part.Render(category)
}

func (m *Model) viewChecksBar() string {
	w := m.getIndentedContentWidth() - 4
	stats := m.getChecksStats()
	total := float64(stats.failed + stats.skipped + stats.succeeded + stats.inProgress)
	numSections := 0
	if stats.failed > 0 {
		numSections++
	}
	if stats.inProgress > 0 {
		numSections++
	}
	if stats.skipped > 0 {
		numSections++
	}
	if stats.succeeded > 0 {
		numSections++
	}
	// subtract num of spacers
	w -= numSections - 1

	sections := make([]string, 0)
	if stats.failed > 0 {
		failWidth := int(math.Floor((float64(stats.failed) / total) * float64(w)))
		sections = append(sections, lipgloss.NewStyle().Width(failWidth).Foreground(m.ctx.Theme.ErrorText).Height(1).Render(strings.Repeat("▃", failWidth)))
	}
	if stats.inProgress > 0 {
		ipWidth := int(math.Floor((float64(stats.inProgress) / total) * float64(w)))
		sections = append(sections, lipgloss.NewStyle().Width(ipWidth).Foreground(m.ctx.Theme.WarningText).Height(1).Render(strings.Repeat("▃", ipWidth)))
	}
	if stats.skipped > 0 {
		skipWidth := int(math.Floor((float64(stats.skipped) / total) * float64(w)))
		sections = append(sections, lipgloss.NewStyle().Width(skipWidth).Foreground(m.ctx.Theme.FaintText).Height(1).Render(strings.Repeat("▃", skipWidth)))
	}
	if stats.succeeded > 0 {
		succWidth := int(math.Floor((float64(stats.succeeded) / total) * float64(w)))
		sections = append(sections, lipgloss.NewStyle().Width(succWidth).Foreground(m.ctx.Theme.SuccessText).Height(1).Render(strings.Repeat("▃", succWidth)))
	}

	return strings.Join(sections, " ")
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

type checksStats struct {
	succeeded  int
	failed     int
	skipped    int
	inProgress int
}

func (m *Model) getChecksStats() checksStats {
	var res checksStats
	commits := m.pr.Data.Commits.Nodes
	if len(commits) == 0 {
		return res
	}

	lastCommit := commits[0]
	for _, node := range lastCommit.Commit.StatusCheckRollup.Contexts.Nodes {
		if node.Typename == "CheckRun" {
			checkRun := node.CheckRun
			conclusion := string(checkRun.Conclusion)
			if data.IsStatusWaiting(string(checkRun.Status)) {
				res.inProgress++
			} else if data.IsConclusionAFailure(conclusion) {
				res.failed++
			} else if data.IsConclusionASkip(conclusion) {
				res.skipped++
			} else if data.IsConclusionASuccess(conclusion) {
				res.succeeded++
			}
		}
	}

	return res
}

func (m *Model) viewChecksConclusionTitle() string {
	stats := m.getChecksStats()
	if stats.failed > 0 {
		return m.ctx.Styles.Common.FailureGlyph + " Some checks were not successful"
	}
	if stats.inProgress > 0 {
		return m.ctx.Styles.Common.WaitingGlyph + " Some checks haven't completed yet"
	}
	if stats.succeeded > 0 {
		return m.ctx.Styles.Common.SuccessGlyph + " All checks have passed"
	}

	return ""
}

func (m *Model) numReviewOwners() int {
	numOwners := 0
	for _, node := range m.pr.Data.ReviewRequests.Nodes {
		if node.AsCodeOwner {
			numOwners++
		}
	}
	return numOwners
}
