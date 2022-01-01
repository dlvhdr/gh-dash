package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/data"
	"github.com/dlvhdr/gh-prs/utils"
)

type PullRequest struct {
	Data data.PullRequestData
}

type repoPullRequestsFetchedMsg struct {
	SectionId int
	Prs       []PullRequest
}

func (pr PullRequest) renderReviewStatus(isSelected bool) string {
	reviewCellStyle := makeRuneCellStyle(isSelected)
	if pr.Data.ReviewDecision == "APPROVED" {
		return reviewCellStyle.Copy().Foreground(successText).Render("")
	}

	if pr.Data.ReviewDecision == "CHANGES_REQUESTED" {
		return reviewCellStyle.Copy().Foreground(warningText).Render("")
	}

	return reviewCellStyle.Copy().Foreground(faintText).Render("")
}

func (pr PullRequest) renderMergeableStatus(isSelected bool) string {
	mergeCellStyle := makeRuneCellStyle(isSelected)
	switch pr.Data.Mergeable {
	case "MERGEABLE":
		return mergeCellStyle.Foreground(successText).Render("")
	case "CONFLICTING":
		return mergeCellStyle.Foreground(warningText).Render("")
	case "UNKNOWN":
		fallthrough
	default:
		return mergeCellStyle.Foreground(faintText).Render("")
	}
}

func (pr PullRequest) renderCiStatus(isSelected bool) string {
	accStatus := "SUCCESS"
	mostRecentCommit := pr.Data.Commits.Nodes[0].Commit
	for _, statusCheck := range mostRecentCommit.StatusCheckRollup.Contexts.Nodes {
		conclusion := statusCheck.CheckRun.Conclusion
		if conclusion == "FAILURE" || conclusion == "TIMED_OUT" || conclusion == "STARTUP_FAILURE" {
			accStatus = "FAILURE"
			break
		}

		status := statusCheck.CheckRun.Status
		if data.IsStatusWaiting(string(status)) {
			accStatus = "PENDING"
		}
	}

	ciCellStyle := makeCellStyle(isSelected).Width(ciCellWidth).MaxWidth(ciCellWidth)
	if accStatus == "SUCCESS" {
		return ciCellStyle.Foreground(successText).Render("")
	}

	if accStatus == "PENDING" {
		return ciCellStyle.Foreground(faintText).Render("")
	}

	return ciCellStyle.Foreground(warningText).Render("")
}

func (pr PullRequest) renderLines(isSelected bool) string {
	separator := makeCellStyle(isSelected).Faint(true).PaddingLeft(1).PaddingRight(1).Render("/")
	added := makeCellStyle(isSelected).PaddingLeft(0).PaddingRight(0).Render(fmt.Sprintf("%d", pr.Data.Additions))
	deletions := 0
	if pr.Data.Deletions > 0 {
		deletions = pr.Data.Deletions
	}
	removed := makeCellStyle(isSelected).PaddingLeft(0).PaddingRight(0).Render(
		fmt.Sprintf("-%d", deletions),
	)

	return makeCellStyle(isSelected).
		Width(linesCellWidth).
		MaxWidth(linesCellWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, added, separator, removed))
}

func (pr PullRequest) renderTitle(viewportWidth int, isSelected bool) string {
	number := makeCellStyle(isSelected).
		MaxWidth(6).
		Foreground(secondaryText).
		PaddingLeft(0).
		PaddingRight(0).
		Align(lipgloss.Left).
		Render(
			fmt.Sprintf("#%s", fmt.Sprintf("%d", pr.Data.Number)),
		)

	totalWidth := getTitleWidth(viewportWidth)
	title := makeCellStyle(isSelected).Render(utils.TruncateString(pr.Data.Title, totalWidth-6))

	return makeCellStyle(isSelected).
		Width(totalWidth).
		MaxWidth(totalWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, number, title))
}

func (pr PullRequest) renderAuthor(isSelected bool) string {
	return makeCellStyle(isSelected).Width(prAuthorCellWidth).Render(
		utils.TruncateString(pr.Data.Author.Login, prAuthorCellWidth-cellPadding),
	)
}

func (pr PullRequest) renderRepoName(isSelected bool) string {
	repoName := utils.TruncateString(pr.Data.HeadRepository.Name, 18)
	return makeCellStyle(isSelected).
		Width(prRepoCellWidth).
		Render(repoName)
}

func (pr PullRequest) renderUpdateAt(isSelected bool) string {
	return makeCellStyle(isSelected).
		Width(updatedAtCellWidth).
		MaxWidth(updatedAtCellWidth).
		Render(utils.TimeElapsed(pr.Data.UpdatedAt))
}

func (pr PullRequest) renderState() string {
	switch pr.Data.State {
	case "OPEN":
		return "Open"
	case "CLOSED":
		return "Closed"
	case "MERGED":
		return "Merged"
	default:
		return ""
	}
}

func (pr PullRequest) render(isSelected bool, viewPortWidth int) string {
	reviewCell := pr.renderReviewStatus(isSelected)
	mergeableCell := pr.renderMergeableStatus(isSelected)
	ciCell := pr.renderCiStatus(isSelected)
	linesCell := pr.renderLines(isSelected)
	prTitleCell := pr.renderTitle(viewPortWidth, isSelected)
	prAuthorCell := pr.renderAuthor(isSelected)
	prRepoCell := pr.renderRepoName(isSelected)
	updatedAtCell := pr.renderUpdateAt(isSelected)

	rowStyle := pullRequestStyle.Copy()
	return rowStyle.
		Width(viewPortWidth).
		MaxWidth(viewPortWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Left,
			updatedAtCell,
			reviewCell,
			prRepoCell,
			prTitleCell,
			prAuthorCell,
			mergeableCell,
			ciCell,
			linesCell,
		))
}
