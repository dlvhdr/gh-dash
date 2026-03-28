package prview

import (
	"fmt"
	"math"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	ghchecks "github.com/dlvhdr/x/gh-checks"
)

type checkSectionStatus int

const (
	statusSuccess checkSectionStatus = iota
	statusFailure
	statusWaiting
	statusNonRequested
)

func (m *Model) renderChecksOverview() string {
	w := m.getIndentedContentWidth()

	if m.pr.Data.Primary.State == "MERGED" {
		return m.viewMergedStatus()
	}

	if m.pr.Data.Primary.State == "CLOSED" {
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

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(w)
	parts := make([]string, 0)
	if review != "" {
		parts = append(parts, review)
	}
	if checks != "" {
		parts = append(parts, checks)
	}
	if merge != "" {
		parts = append(parts, merge)
	}

	return box.Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (m *Model) viewChecksStatus() (string, checkSectionStatus) {
	checks := ""

	if !m.pr.Data.IsEnriched {
		return m.viewCheckCategory(
			m.ctx.Styles.Common.WaitingGlyph,
			"Loading...",
			"",
			false,
		), statusWaiting
	}

	// For GitLab, use pipeline jobs
	if len(m.pr.Data.Enriched.PipelineJobs) > 0 {
		stats := m.getGitLabChecksStats()
		var icon, title string
		var status checkSectionStatus

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

		statStrs := make([]string, 0)
		if stats.failed > 0 {
			statStrs = append(statStrs, fmt.Sprintf("%d failing", stats.failed))
		}
		if stats.inProgress > 0 {
			statStrs = append(statStrs, fmt.Sprintf("%d in progress", stats.inProgress))
		}
		if stats.succeeded > 0 {
			statStrs = append(statStrs, fmt.Sprintf("%d successful", stats.succeeded))
		}
		if stats.skipped > 0 {
			statStrs = append(statStrs, fmt.Sprintf("%d skipped", stats.skipped))
		}

		checks = m.viewCheckCategory(icon, title, strings.Join(statStrs, ", "), true)
		return checks, status
	}

	// For GitHub, use StatusCheckRollup
	stats := m.getChecksStats()
	var icon, title string
	var status checkSectionStatus

	statStrs := make([]string, 0)
	if stats.failed > 0 {
		icon = m.ctx.Styles.Common.FailureGlyph
		title = "Some checks were not successful"
		status = statusFailure
	} else if stats.awaitingApproval > 0 {
		icon = m.ctx.Styles.Common.ActionRequiredGlyph
		title = "Workflows awaiting approval"
		status = statusWaiting
	} else if stats.inProgress > 0 {
		icon = m.ctx.Styles.Common.WaitingGlyph
		title = "Some checks haven’t completed yet"
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
	if stats.awaitingApproval > 0 {
		statStrs = append(statStrs, fmt.Sprintf("%d awaiting approval", stats.awaitingApproval))
	}
	if stats.inProgress > 0 {
		statStrs = append(statStrs, fmt.Sprintf("%d in progress", stats.inProgress))
	}
	if stats.succeeded > 0 {
		statStrs = append(statStrs, fmt.Sprintf("%d successful", stats.succeeded))
	}
	if stats.skipped > 0 {
		statStrs = append(statStrs, fmt.Sprintf("%d skipped", stats.skipped))
	}
	if stats.neutral > 0 {
		statStrs = append(statStrs, fmt.Sprintf("%d neutral", stats.neutral))
	}
	if stats.succeeded > 0 {
		statStrs = append(statStrs, fmt.Sprintf("%d successful", stats.succeeded))
	}
	if title != "" {
		checksBar := m.viewChecksBar()
		checksBottom := lipgloss.JoinVertical(
			lipgloss.Left,
			strings.Join(statStrs, ", "),
			checksBar,
		)
		checks = m.viewCheckCategory(icon, title, checksBottom, false)
	}
	return checks, status
}

func (m *Model) viewMergeStatus() (string, checkSectionStatus) {
	var icon, title, subtitle string
	var status checkSectionStatus
	numReviewOwners := m.numRequestedReviewOwners()
	if m.pr.Data.Primary.MergeStateStatus == "CLEAN" ||
		m.pr.Data.Primary.MergeStateStatus == "UNSTABLE" {
		icon = m.ctx.Styles.Common.SuccessGlyph
		title = "No conflicts with base branch"
		subtitle = "Changes can be cleanly merged"
		status = statusSuccess
	} else if m.pr.Data.Primary.IsDraft {
		icon = m.ctx.Styles.Common.DraftGlyph
		title = "This pull request is still a work in progress"
		subtitle = "Draft pull requests cannot be merged"
		status = statusWaiting
	} else if m.pr.Data.Primary.MergeStateStatus == "BLOCKED" {
		icon = m.ctx.Styles.Common.FailureGlyph
		title = "Merging is blocked"
		if numReviewOwners > 0 {
			subtitle = "Waiting on code owner review"
		}
		status = statusFailure
	} else if m.pr.Data.Primary.Mergeable == "CONFLICTING" {
		icon = m.ctx.Styles.Common.FailureGlyph
		title = "This branch has conflicts that must be resolved"
		status = statusFailure
		if m.pr.Data.Primary.MergeStateStatus == "CLEAN" {
			subtitle = "Changes can be cleanly merged"
		}
	}
	return m.viewCheckCategory(icon, title, subtitle, true), status
}

func (m *Model) viewMergedStatus() string {
	w := m.getIndentedContentWidth()
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.ctx.Styles.Colors.MergedPR).
		Width(w)
	return box.Render(m.viewCheckCategory(
		m.ctx.Styles.Common.MergedGlyph,
		"Pull request successfully merged and closed",
		"The branch has been merged",
		true,
	))
}

func (m *Model) viewClosedStatus() string {
	w := m.getIndentedContentWidth()
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.ctx.Theme.FaintBorder).
		Width(w)
	return box.Render(m.viewCheckCategory(
		"",
		"Closed with unmerged commits",
		"This pull request is closed",
		true,
	))
}

func (m *Model) viewReviewStatus() (string, checkSectionStatus) {
	w := m.getIndentedContentWidth() - 2

	if !m.pr.Data.IsEnriched {
		return m.viewCheckCategory(m.ctx.Styles.Common.WaitingGlyph, "Loading...", "", false), statusWaiting
	}

	reviewRequests := m.pr.Data.Enriched.ReviewRequests.Nodes
	reviews := m.pr.Data.Enriched.Reviews.Nodes

	var icon, title string
	var status checkSectionStatus

	changesRequested := 0
	approved := 0
	pending := len(reviewRequests)

	for _, review := range reviews {
		if review.State == "APPROVED" {
			approved++
		}
		if review.State == "CHANGES_REQUESTED" {
			changesRequested++
		}
	}

	if changesRequested > 0 {
		icon = m.ctx.Styles.Common.FailureGlyph
		title = "Changes requested"
		status = statusFailure
	} else if approved > 0 {
		icon = m.ctx.Styles.Common.SuccessGlyph
		title = fmt.Sprintf("%d approving review", approved)
		if approved > 1 {
			title += "s"
		}
		status = statusSuccess
	} else if pending > 0 {
		icon = m.ctx.Styles.Common.WaitingGlyph
		title = "Review required"
		status = statusWaiting
	} else {
		return "", statusNonRequested
	}

	subs := make([]string, 0)
	if pending > 0 {
		for _, reviewRequest := range reviewRequests {
			name := ""
			if reviewRequest.RequestedReviewer.User.Login != "" {
				name = reviewRequest.RequestedReviewer.User.Login
			}
			if reviewRequest.RequestedReviewer.Team.Name != "" {
				name = reviewRequest.RequestedReviewer.Team.Name
			}
			if reviewRequest.RequestedReviewer.Team.Slug != "" {
				name = reviewRequest.RequestedReviewer.Team.Slug
			}
			subs = append(subs, fmt.Sprintf("%s Review pending from %s", m.ctx.Styles.Common.WaitingGlyph, name))
		}
	}

	for _, review := range reviews {
		if review.State == "APPROVED" {
			subs = append(subs, fmt.Sprintf("%s %s approved", m.ctx.Styles.Common.SuccessGlyph, review.Author.Login))
		}
		if review.State == "CHANGES_REQUESTED" {
			subs = append(subs, fmt.Sprintf("%s %s requested changes", m.ctx.Styles.Common.FailureGlyph, review.Author.Login))
		}
	}

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		" ",
		icon,
		" ",
		lipgloss.NewStyle().Width(w).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				m.ctx.Styles.Common.MainTextStyle.Render(title),
				lipgloss.NewStyle().MarginLeft(2).Width(w).Foreground(m.ctx.Theme.FaintText).Render(
					lipgloss.JoinVertical(lipgloss.Left, subs...),
				),
			),
		),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true).
		BorderForeground(m.ctx.Theme.FaintBorder).
		Width(w).
		Padding(1)

	sTitle := lipgloss.NewStyle().Bold(true)
	sSub := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText)

	category := lipgloss.JoinHorizontal(lipgloss.Top, icon, " ", sTitle.Render(title))

	if subtitle != "" {
		category = lipgloss.JoinVertical(
			lipgloss.Left,
			category,
			sSub.MarginLeft(2).Render(subtitle),
		)
	}
	if category == "" {
		return ""
	}
	return part.Render(category)
}

func (m *Model) viewChecksBar() string {
	w := m.getIndentedContentWidth() - 4
	stats := m.getChecksStats()
	total := float64(
		stats.failed + stats.skipped + stats.neutral + stats.succeeded + stats.inProgress + stats.awaitingApproval,
	)
	numSections := 0
	if stats.failed > 0 {
		numSections++
	}
	if stats.awaitingApproval > 0 {
		numSections++
	}
	if stats.inProgress > 0 {
		numSections++
	}
	if stats.skipped > 0 || stats.neutral > 0 {
		numSections++
	}
	if stats.succeeded > 0 {
		numSections++
	}
	// subtract number of spacers
	w -= numSections - 1
	if w < 0 {
		w = 0
	}

	sections := make([]string, 0)
	if stats.failed > 0 {
		failWidth := int(math.Floor((float64(stats.failed) / total) * float64(w)))
		sections = append(sections, lipgloss.NewStyle().Width(failWidth).Foreground(
			m.ctx.Theme.ErrorText).Height(1).Render(strings.Repeat("▃", failWidth)))
	}
	if stats.awaitingApproval > 0 {
		awWidth := int(math.Floor((float64(stats.awaitingApproval) / total) * float64(w)))
		sections = append(sections, lipgloss.NewStyle().Width(awWidth).Foreground(
			m.ctx.Theme.WarningText).Height(1).Render(strings.Repeat("▃", awWidth)))
	}
	if stats.inProgress > 0 {
		ipWidth := int(math.Floor((float64(stats.inProgress) / total) * float64(w)))
		sections = append(sections, lipgloss.NewStyle().Width(ipWidth).Foreground(
			m.ctx.Theme.WarningText).Height(1).Render(strings.Repeat("▃", ipWidth)))
	}
	if stats.skipped > 0 || stats.neutral > 0 {
		skipWidth := int(math.Floor((float64(stats.skipped+stats.neutral) / total) * float64(w)))
		sections = append(sections, lipgloss.NewStyle().Width(skipWidth).Foreground(
			m.ctx.Theme.FaintText).Height(1).Render(strings.Repeat("▃", skipWidth)))
	}
	if stats.succeeded > 0 {
		succWidth := int(math.Floor((float64(stats.succeeded) / total) * float64(w)))
		sections = append(sections, lipgloss.NewStyle().Width(succWidth).Foreground(
			m.ctx.Theme.SuccessText).Height(1).Render(strings.Repeat("▃", succWidth)))
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

	if m.pr.Data.Primary.Mergeable == "MERGEABLE" {
		if m.pr.Data.Primary.MergeStateStatus == "BLOCKED" {
			return m.viewCheckCategory(m.ctx.Styles.Common.WaitingGlyph, "Merging is blocked", "", false), statusWaiting
		}
		return m.viewCheckCategory(m.ctx.Styles.Common.SuccessGlyph, "This branch has no conflicts with the base branch", "Merging can be performed automatically", false), statusSuccess
	}

	if m.pr.Data.Primary.Mergeable == "CONFLICTING" {
		return m.viewCheckCategory(m.ctx.Styles.Common.FailureGlyph, "This branch has conflicts that must be resolved", "", false), statusFailure
	}

	return m.viewCheckCategory(m.ctx.Styles.Common.WaitingGlyph, "Checking for ability to merge automatically", "", false), statusWaiting
}

func (m *Model) renderStatusContextConclusion(
	statusContext data.StatusContext,
) (CheckCategory, string) {
	conclusionStr := string(statusContext.State)
	if ghchecks.IsStatusWaiting(conclusionStr) {
		return CheckWaiting, m.ctx.Styles.Common.WaitingGlyph
	}

	if ghchecks.IsConclusionAFailure(conclusionStr) {
		return CheckFailure, m.ctx.Styles.Common.FailureGlyph
	}

	return CheckSuccess, m.ctx.Styles.Common.SuccessGlyph
}

func (m *Model) viewClosedStatus() string {
	w := m.getIndentedContentWidth()
	closed := lipgloss.NewStyle().Foreground(m.ctx.Theme.ErrorText).Render(" Closed")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.ctx.Theme.ErrorText).
		Width(w).
		Padding(1).
		Render(closed)
}

func (sidebar *Model) renderChecks() string {
	title := sidebar.ctx.Styles.Common.MainTextStyle.MarginBottom(1).
		Underline(true).
		Render(" All Checks")

	commits := sidebar.pr.Data.Enriched.Commits.Nodes
	if len(commits) == 0 {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"Loading...",
		)
	}

	failures := make([]string, 0)
	waiting := make([]string, 0)
	rest := make([]string, 0)
	awaitingApproval := make([]string, 0)
	pending := make([]string, 0)

	lastCommit := commits[0]

	// Collect check suites that don't appear in statusCheckRollup
	for _, suite := range lastCommit.Commit.CheckSuites.Nodes {
		workflowName := strings.TrimSpace(string(suite.WorkflowRun.Workflow.Name))
		if workflowName == "" {
			workflowName = strings.TrimSpace(string(suite.App.Name))
		}
		if workflowName == "" {
			workflowName = "Workflow"
		}

		if suite.Conclusion == "ACTION_REQUIRED" {
			// Workflow requires approval before it can run
			check := lipgloss.JoinHorizontal(
				lipgloss.Top,
				sidebar.ctx.Styles.Common.ActionRequiredGlyph,
				" ",
				workflowName,
			)
			awaitingApproval = append(awaitingApproval, check)
		} else if suite.Status == "QUEUED" || suite.Status == "PENDING" || suite.Status == "WAITING" {
			// Workflow is queued/pending (will run automatically)
			check := lipgloss.JoinHorizontal(
				lipgloss.Top,
				sidebar.ctx.Styles.Common.WaitingGlyph,
				" ",
				workflowName,
			)
			pending = append(pending, check)
		}
	}

	// Build a set of reported check names to compare against required checks
	reportedChecks := make(map[string]bool)

	for _, node := range lastCommit.Commit.StatusCheckRollup.Contexts.Nodes {
		var category CheckCategory
		var check string
		var checkName string
		switch node.Typename {
		case "CheckRun":
			checkRun := node.CheckRun
			var renderedStatus string
			category, renderedStatus = sidebar.renderCheckRunConclusion(checkRun)
			checkName = string(checkRun.Name)
			name := renderCheckRunName(checkRun)
			check = lipgloss.JoinHorizontal(lipgloss.Top, renderedStatus, " ", name)
		case "StatusContext":
			statusContext := node.StatusContext
			var status string
			category, status = sidebar.renderStatusContextConclusion(statusContext)
			checkName = string(statusContext.Context)
			check = lipgloss.JoinHorizontal(
				lipgloss.Top,
				status,
				" ",
				renderStatusContextName(statusContext),
			)
		}

		reportedChecks[checkName] = true

		switch category {
		case CheckWaiting:
			waiting = append(waiting, check)
		case CheckFailure:
			failures = append(failures, check)
		default:
			rest = append(rest, check)
		}
	}

	// Check for required status checks that haven't been reported yet
	branchRules := sidebar.pr.Data.Primary.Repository.BranchProtectionRules.Nodes
	if len(branchRules) > 0 {
		for _, requiredContext := range branchRules[0].RequiredStatusCheckContexts {
			contextName := string(requiredContext)
			if !reportedChecks[contextName] {
				// Required check hasn't been reported yet
				check := lipgloss.JoinHorizontal(
					lipgloss.Top,
					sidebar.ctx.Styles.Common.WaitingGlyph,
					" ",
					contextName,
				)
				pending = append(pending, check)
			}
		}
	}

	if len(awaitingApproval)+len(pending)+len(waiting)+len(failures)+len(rest) == 0 {
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

	parts := make([]string, 0)

	// Show awaiting approval workflows first
	if len(awaitingApproval) > 0 {
		sectionHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(sidebar.ctx.Theme.WarningText).
			Render(fmt.Sprintf("Awaiting Approval (%d)", len(awaitingApproval)))
		parts = append(parts, sectionHeader)
		parts = append(parts, awaitingApproval...)
		parts = append(parts, "") // spacing
	}

	// Show pending workflows
	if len(pending) > 0 {
		sectionHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(sidebar.ctx.Theme.WarningText).
			Render(fmt.Sprintf("Pending (%d)", len(pending)))
		parts = append(parts, sectionHeader)
		parts = append(parts, pending...)
		parts = append(parts, "") // spacing
	}

	parts = append(parts, failures...)
	parts = append(parts, waiting...)
	parts = append(parts, rest...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().PaddingLeft(2).Width(sidebar.getIndentedContentWidth()).Render(
			lipgloss.JoinVertical(lipgloss.Left, parts...)),
	)
}

type checksStats struct {
	succeeded        int
	neutral          int
	failed           int
	skipped          int
	inProgress       int
	awaitingApproval int
}

func (m *Model) getStatusCheckRollupStats(rollup data.StatusCheckRollupStats) checksStats {
	allChecks := make([]data.ContextCountByState, 0)
	allChecks = append(allChecks, rollup.Contexts.CheckRunCountsByState...)
	allChecks = append(allChecks, rollup.Contexts.StatusContextCountsByState...)

	return m.getStatsFromChecks(allChecks)
}

func (m *Model) getChecksStats() checksStats {
	commits := m.pr.Data.Enriched.Commits.Nodes
	if len(commits) == 0 {
		return checksStats{}
	}

	lastCommit := commits[0]
	allChecks := make([]data.ContextCountByState, 0)
	allChecks = append(
		allChecks,
		lastCommit.Commit.StatusCheckRollup.Contexts.CheckRunCountsByState...)
	allChecks = append(
		allChecks,
		lastCommit.Commit.StatusCheckRollup.Contexts.StatusContextCountsByState...)

	return m.getStatsFromChecks(allChecks)
}

// getGitLabChecksStats calculates stats from GitLab pipeline jobs
func (m *Model) getGitLabChecksStats() checksStats {
	stats := checksStats{}
	for _, job := range m.pr.Data.Enriched.PipelineJobs {
		switch job.Status {
		case "success":
			stats.succeeded++
		case "failed":
			stats.failed++
		case "running", "pending", "created":
			stats.inProgress++
		case "skipped":
			stats.skipped++
		case "canceled":
			stats.failed++
		}
	}
	return stats
}

	// Count check suites that don't appear in statusCheckRollup
	for _, suite := range lastCommit.Commit.CheckSuites.Nodes {
		if suite.Conclusion == "ACTION_REQUIRED" {
			res.awaitingApproval++
		} else if suite.Status == "QUEUED" || suite.Status == "PENDING" || suite.Status == "WAITING" {
			res.inProgress++
		}
	}

	return res
}

type CheckCategory int

const (
	CheckSuccess CheckCategory = iota
	CheckFailure
	CheckWaiting
)

func (sidebar *Model) renderCheckRunConclusion(checkRun data.CheckRun) (CheckCategory, string) {
	switch checkRun.Conclusion {
	case ghchecks.CheckRunStateSuccess:
		return CheckSuccess, sidebar.ctx.Styles.Common.SuccessGlyph
	case ghchecks.CheckRunStateFailure:
		return CheckFailure, sidebar.ctx.Styles.Common.FailureGlyph
	case ghchecks.CheckRunStateStartupFailure:
		return CheckFailure, sidebar.ctx.Styles.Common.FailureGlyph
	case ghchecks.CheckRunStateSkipped:
		skipped := lipgloss.NewStyle().Foreground(sidebar.ctx.Theme.FaintText).Render("⊘")
		return CheckSuccess, skipped
	case ghchecks.CheckRunStateStale:
		return CheckWaiting, sidebar.ctx.Styles.Common.WaitingGlyph
	case ghchecks.CheckRunStateNeutral:
		neutral := lipgloss.NewStyle().Foreground(sidebar.ctx.Theme.FaintText).Render("◦")
		return CheckSuccess, neutral
	case ghchecks.CheckRunStateCancelled:
		cancelled := lipgloss.NewStyle().Foreground(sidebar.ctx.Theme.FaintText).Render("⊘")
		return CheckFailure, cancelled
	case ghchecks.CheckRunStateActionRequired:
		return CheckWaiting, sidebar.ctx.Styles.Common.WaitingGlyph
	case ghchecks.CheckRunStateTimedOut:
		return CheckFailure, sidebar.ctx.Styles.Common.FailureGlyph
	}

	return CheckWaiting, sidebar.ctx.Styles.Common.WaitingGlyph
}

func (sidebar *Model) renderStatusContextConclusion(statusContext data.StatusContext) (CheckCategory, string) {
	state := string(statusContext.State)
	switch strings.ToUpper(state) {
	case "SUCCESS":
		return CheckSuccess, sidebar.ctx.Styles.Common.SuccessGlyph
	case "FAILURE", "ERROR":
		return CheckFailure, sidebar.ctx.Styles.Common.FailureGlyph
	}

	return CheckWaiting, sidebar.ctx.Styles.Common.WaitingGlyph
}

func renderCheckRunName(checkRun data.CheckRun) string {
	if checkRun.CheckSuite.WorkflowRun.Workflow.Name != "" {
		return fmt.Sprintf("%s / %s", checkRun.CheckSuite.WorkflowRun.Workflow.Name, checkRun.Name)
	}

	if checkRun.CheckSuite.Creator.Login != "" {
		return fmt.Sprintf("%s (%s)", checkRun.Name, checkRun.CheckSuite.Creator.Login)
	}

	return string(checkRun.Name)
}

func renderStatusContextName(statusContext data.StatusContext) string {
	if statusContext.Creator.Login != "" {
		return fmt.Sprintf("%s (%s)", statusContext.Context, statusContext.Creator.Login)
	}
	return string(statusContext.Context)
}

func sizePerSection(total, size int) int {
	return int(math.Max(1, math.Floor(float64(total)/float64(size))))
}
