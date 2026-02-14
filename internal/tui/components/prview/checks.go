package prview

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

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

	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Width(w)
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
		return m.viewCheckCategory(m.ctx.Styles.Common.WaitingGlyph, "Loading...", "", false), statusWaiting
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
	if stats.succeeded > 0 {
		statStrs = append(statStrs, fmt.Sprintf("%d successful", stats.succeeded))
	}
	if stats.skipped > 0 {
		statStrs = append(statStrs, fmt.Sprintf("%d skipped", stats.skipped))
	}
	if stats.neutral > 0 {
		statStrs = append(statStrs, fmt.Sprintf("%d neutral", stats.neutral))
	}

	checks = m.viewCheckCategory(icon, title, strings.Join(statStrs, ", "), true)
	return checks, status
}

func (m *Model) viewCheckCategory(icon string, title string, meta string, border bool) string {
	w := m.getIndentedContentWidth() - 2
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		" ",
		icon,
		" ",
		lipgloss.NewStyle().Width(w).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				m.ctx.Styles.Common.MainTextStyle.Render(title),
				lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Render(meta),
			),
		),
	)
	if border {
		return lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true).
			BorderForeground(m.ctx.Theme.FaintBorder).
			Width(w).
			Padding(1, 0).
			Render(content)
	}
	return lipgloss.NewStyle().Width(w).Padding(1, 0).Render(content)
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
		Padding(1, 0).
		Render(content), status
}

func (m *Model) viewMergeStatus() (string, checkSectionStatus) {
	if !m.pr.Data.IsEnriched {
		return m.viewCheckCategory(m.ctx.Styles.Common.WaitingGlyph, "Loading...", "", false), statusWaiting
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

func (m *Model) viewMergedStatus() string {
	w := m.getIndentedContentWidth()
	merged := lipgloss.NewStyle().Foreground(m.ctx.Theme.SuccessText).Render(" Merged")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.ctx.Theme.SuccessText).
		Width(w).
		Padding(1).
		Render(merged)
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
	title := sidebar.ctx.Styles.Common.MainTextStyle.MarginBottom(1).Underline(true).Render(" All Checks")

	failures := make([]string, 0)
	waiting := make([]string, 0)
	rest := make([]string, 0)

	// Check for GitLab pipeline jobs first
	if len(sidebar.pr.Data.Enriched.PipelineJobs) > 0 {
		for _, job := range sidebar.pr.Data.Enriched.PipelineJobs {
			var category CheckCategory
			var icon string

			switch job.Status {
			case "success":
				category = CheckSuccess
				icon = sidebar.ctx.Styles.Common.SuccessGlyph
			case "failed":
				category = CheckFailure
				icon = sidebar.ctx.Styles.Common.FailureGlyph
			case "running", "pending", "created":
				category = CheckWaiting
				icon = sidebar.ctx.Styles.Common.WaitingGlyph
			case "skipped":
				category = CheckSuccess
				icon = lipgloss.NewStyle().Foreground(sidebar.ctx.Theme.FaintText).Render("⊘")
			case "canceled":
				category = CheckFailure
				icon = lipgloss.NewStyle().Foreground(sidebar.ctx.Theme.WarningText).Render("⊘")
			default:
				category = CheckWaiting
				icon = sidebar.ctx.Styles.Common.WaitingGlyph
			}

			jobName := lipgloss.NewStyle().Foreground(sidebar.ctx.Theme.FaintText).Render(fmt.Sprintf("[%s]", job.Stage))
			check := lipgloss.JoinHorizontal(lipgloss.Top, icon, " ", job.Name, " ", jobName)

			switch category {
			case CheckWaiting:
				waiting = append(waiting, check)
			case CheckFailure:
				failures = append(failures, check)
			default:
				rest = append(rest, check)
			}
		}
	} else {
		// GitHub checks via GraphQL
		commits := sidebar.pr.Data.Enriched.Commits.Nodes
		if len(commits) == 0 {
			return lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"Loading...",
			)
		}

		lastCommit := commits[0]
		for _, node := range lastCommit.Commit.StatusCheckRollup.Contexts.Nodes {
			var category CheckCategory
			var check string
			switch node.Typename {
			case "CheckRun":
				checkRun := node.CheckRun
				var renderedStatus string
				category, renderedStatus = sidebar.renderCheckRunConclusion(checkRun)
				name := renderCheckRunName(checkRun)
				check = lipgloss.JoinHorizontal(lipgloss.Top, renderedStatus, " ", name)
			case "StatusContext":
				statusContext := node.StatusContext
				var status string
				category, status = sidebar.renderStatusContextConclusion(statusContext)
				check = lipgloss.JoinHorizontal(lipgloss.Top, status, " ", renderStatusContextName(statusContext))
			}

			switch category {
			case CheckWaiting:
				waiting = append(waiting, check)
			case CheckFailure:
				failures = append(failures, check)
			default:
				rest = append(rest, check)
			}
		}
	}

	if len(waiting)+len(failures)+len(rest) == 0 {
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
	succeeded  int
	neutral    int
	failed     int
	skipped    int
	inProgress int
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
	allChecks = append(allChecks, lastCommit.Commit.StatusCheckRollup.Contexts.CheckRunCountsByState...)
	allChecks = append(allChecks, lastCommit.Commit.StatusCheckRollup.Contexts.StatusContextCountsByState...)

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

func (m *Model) getStatsFromChecks(allChecks []data.ContextCountByState) checksStats {
	stats := checksStats{}
	for _, check := range allChecks {
		switch check.State {
		case ghchecks.CheckRunStateSuccess:
			stats.succeeded += int(check.Count)
		case ghchecks.CheckRunStateFailure:
			stats.failed += int(check.Count)
		case ghchecks.CheckRunStateNeutral:
			stats.neutral += int(check.Count)
		case ghchecks.CheckRunStateSkipped:
			stats.skipped += int(check.Count)
		case ghchecks.CheckRunStateActionRequired,
			ghchecks.CheckRunStateCancelled,
			ghchecks.CheckRunStateStartupFailure,
			ghchecks.CheckRunStateStale,
			ghchecks.CheckRunStateTimedOut:
			stats.failed += int(check.Count)
		case ghchecks.CheckRunStateCompleted:
			stats.succeeded += int(check.Count)
		case ghchecks.CheckRunStateInProgress:
			stats.inProgress += int(check.Count)
		case ghchecks.CheckRunStatePending:
			stats.inProgress += int(check.Count)
		case ghchecks.CheckRunStateQueued:
			stats.inProgress += int(check.Count)
		case ghchecks.CheckRunStateWaiting:
			stats.inProgress += int(check.Count)
		}
	}
	return stats
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
