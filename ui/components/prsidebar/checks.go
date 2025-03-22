package prsidebar

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/data"
)

type checkSectionStatus int

const (
	statusSuccess checkSectionStatus = iota
	statusFailure
	statusWaiting
)

func (m *Model) renderChecksOverview() string {
	w := m.getIndentedContentWidth()

	if m.pr.Data.State == "MERGED" {
		return m.viewMergedStatus()
	}

	if m.pr.Data.State == "CLOSED" {
		return m.viewClosedStatus()
	}

	review, rStatus := m.viewReviewStatus()
	checks, cStatus := m.viewChecksStatus()
	merge, mStatus := m.viewMergeStatus()

	borderColor := m.ctx.Theme.FaintBorder
	if rStatus == statusFailure || cStatus == statusFailure || mStatus == statusFailure {
		borderColor = m.ctx.Theme.ErrorText
	} else if rStatus == statusSuccess && cStatus == statusSuccess && mStatus == statusSuccess {
		borderColor = m.ctx.Theme.SuccessText
	}

	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Width(w)

	return box.Render(lipgloss.JoinVertical(lipgloss.Left, review, checks, merge))
}

func (m *Model) viewChecksStatus() (string, checkSectionStatus) {
	checks := ""
	stats := m.getChecksStats()
	var icon, title string
	var status checkSectionStatus

	statStrs := make([]string, 0)
	if stats.failed > 0 {
		icon = m.ctx.Styles.Common.FailureGlyph
		title = "Some checks were not successful"
		status = statusFailure
	} else if stats.inProgress > 0 {
		icon = m.ctx.Styles.Common.WaitingGlyph
		title = "Some checks haven't completed yet"
		status = statusWaiting
	} else if stats.succeeded > 0 {
		icon = m.ctx.Styles.Common.SuccessGlyph
		title = "All checks have passed"
		status = statusSuccess
	} else {
		return "", statusWaiting
	}

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
	if title != "" {
		checksBar := m.viewChecksBar()
		checksBottom := lipgloss.JoinVertical(lipgloss.Left, strings.Join(statStrs, ", "), checksBar)
		checks = m.viewCheckCategory(icon, title, checksBottom, false)
	}
	return checks, status
}

func (m *Model) viewMergeStatus() (string, checkSectionStatus) {
	var icon, title, subtitle string
	var status checkSectionStatus
	numReviewOwners := m.numRequestedReviewOwners()
	if m.pr.Data.MergeStateStatus == "CLEAN" || m.pr.Data.MergeStateStatus == "UNSTABLE" {
		icon = m.ctx.Styles.Common.SuccessGlyph
		title = "No conflicts with base branch"
		subtitle = "Changes can be cleanly merged"
		status = statusSuccess
	} else if m.pr.Data.IsDraft {
		icon = m.ctx.Styles.Common.DraftGlyph
		title = "This pull request is still a work in progress"
		subtitle = "Draft pull requests cannot be merged"
		status = statusWaiting
	} else if m.pr.Data.MergeStateStatus == "BLOCKED" {
		icon = m.ctx.Styles.Common.FailureGlyph
		title = "Merging is blocked"
		if numReviewOwners > 0 {
			subtitle = "Waiting on code owner review"
		}
		status = statusFailure
	} else if m.pr.Data.Mergeable == "CONFLICTING" {
		icon = m.ctx.Styles.Common.FailureGlyph
		title = "This branch has conflicts that must be resolved"
		status = statusFailure
		if m.pr.Data.MergeStateStatus == "CLEAN" {
			subtitle = "Changes can be cleanly merged"
		}
	}
	return m.viewCheckCategory(icon, title, subtitle, true), status
}

func (m *Model) viewMergedStatus() string {
	w := m.getIndentedContentWidth()
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(m.ctx.Styles.Colors.MergedPR).Width(w)
	return box.Render(m.viewCheckCategory(
		m.ctx.Styles.Common.MergedGlyph,
		"Pull request successfully merged and closed",
		"The branch has been merged",
		true,
	))
}

func (m *Model) viewClosedStatus() string {
	w := m.getIndentedContentWidth()
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(m.ctx.Theme.FaintBorder).Width(w)
	return box.Render(m.viewCheckCategory(
		"",
		"Closed with unmerged commits",
		"This pull request is closed",
		true,
	))
}

func (m *Model) viewReviewStatus() (string, checkSectionStatus) {
	pr := m.pr
	if pr.Data == nil {
		return "", statusWaiting
	}

	var icon, title, subtitle string
	var status checkSectionStatus
	numReviewOwners := m.numRequestedReviewOwners()

	numApproving, numChangesRequested, numPending, numCommented := 0, 0, 0, 0
	for _, node := range pr.Data.Reviews.Nodes {
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
		icon = m.ctx.Styles.Common.SuccessGlyph
		title = "Changes approved"
		subtitle = fmt.Sprintf("%d approving reviews", numApproving)
		status = statusSuccess
	} else if pr.Data.ReviewDecision == "CHANGES_REQUESTED" {
		icon = m.ctx.Styles.Common.FailureGlyph
		title = "Changes requested"
		subtitle = fmt.Sprintf("%d requested changes", numChangesRequested)
		status = statusFailure
	} else if pr.Data.ReviewDecision == "REVIEW_REQUIRED" {
		icon = pr.Ctx.Styles.Common.WaitingGlyph
		title = "Review Required"

		branchRules := m.pr.Data.Repository.BranchProtectionRules.Nodes
		if len(branchRules) > 0 && branchRules[0].RequiresCodeOwnerReviews && numApproving < 1 {
			subtitle = "Code owner review required"
			status = statusFailure
		} else if numApproving < numReviewOwners {
			subtitle = "Code owner review required"
			status = statusFailure
		} else if len(branchRules) > 0 && numApproving < branchRules[0].RequiredApprovingReviewCount {
			subtitle = fmt.Sprintf("Need %d more approval", branchRules[0].RequiredApprovingReviewCount-numApproving)
			status = statusWaiting
		} else if numCommented > 0 {
			subtitle = fmt.Sprintf("%d reviewers left comments", numCommented)
			status = statusWaiting
		}
	}

	return m.viewCheckCategory(icon, title, subtitle, false), status
}

func (m *Model) viewCheckCategory(icon, title, subtitle string, isLast bool) string {
	w := m.getIndentedContentWidth()
	part := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, !isLast, false).
		BorderForeground(m.ctx.Theme.FaintBorder).
		Width(w).
		Padding(1)

	sTitle := lipgloss.NewStyle().Bold(true)
	sSub := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText)

	category := lipgloss.JoinHorizontal(lipgloss.Top, icon, " ", sTitle.Render(title))

	if subtitle != "" {
		category = lipgloss.JoinVertical(lipgloss.Left, category, sSub.MarginLeft(2).Render(subtitle))
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

func (sidebar *Model) renderChecks() string {
	title := sidebar.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Render(" All Checks")

	commits := sidebar.pr.Data.Commits.Nodes
	if len(commits) == 0 {
		return ""
	}

	var checks []string

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

func (m *Model) numRequestedReviewOwners() int {
	numOwners := 0

	for _, node := range m.pr.Data.ReviewRequests.Nodes {
		if node.AsCodeOwner {
			numOwners++
		}
	}

	return numOwners
}
