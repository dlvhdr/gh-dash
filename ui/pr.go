package ui

import (
	"dlvhdr/gh-prs/data"
	"dlvhdr/gh-prs/utils"
	"fmt"

	"github.com/charmbracelet/lipgloss"
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
		return reviewCellStyle.Foreground(lipgloss.Color("42")).Render("")
	}

	if pr.Data.ReviewDecision == "CHANGES_REQUESTED" {
		return reviewCellStyle.Foreground(lipgloss.Color("196")).Render("")
	}

	return reviewCellStyle.Faint(true).Render("")
}

func (pr PullRequest) renderMergeableStatus(isSelected bool) string {
	mergeCellStyle := makeRuneCellStyle(isSelected)
	switch pr.Data.Mergeable {
	case "MERGEABLE":
		return mergeCellStyle.Foreground(lipgloss.Color("42")).Render("")
	case "CONFLICTING":
		return mergeCellStyle.Foreground(lipgloss.Color("196")).Render("")
	case "UNKNOWN":
		fallthrough
	default:
		return mergeCellStyle.Faint(true).Render("")
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
		if status == "PENDING" ||
			status == "QUEUED" ||
			status == "IN_PROGRESS" ||
			status == "WAITING" {
			accStatus = "PENDING"
		}
	}

	ciCellStyle := makeRuneCellStyle(isSelected).Width(ciCellWidth)
	if accStatus == "SUCCESS" {
		return ciCellStyle.Foreground(lipgloss.Color("42")).Render("")
	}

	if accStatus == "PENDING" {
		return ciCellStyle.Foreground(lipgloss.Color("214")).Render("")
	}

	return ciCellStyle.Foreground(lipgloss.Color("196")).Render("")
}

func (pr PullRequest) renderLines(isSelected bool) string {
	separator := makeCellStyle(isSelected).Faint(true).PaddingLeft(1).PaddingRight(1).Render("/")
	added := makeCellStyle(isSelected).Render(fmt.Sprintf("%d", pr.Data.Additions))
	deletions := 0
	if pr.Data.Deletions > 0 {
		deletions = pr.Data.Deletions
	}
	removed := makeCellStyle(isSelected).Render(
		fmt.Sprintf("-%d", deletions),
	)

	return makeCellStyle(isSelected).
		Width(linesCellWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, added, separator, removed))
}

func (pr PullRequest) renderTitle(viewportWidth int, isSelected bool) string {
	number := lipgloss.NewStyle().Width(6).Faint(true).Render(
		fmt.Sprintf("#%s", fmt.Sprintf("%d", pr.Data.Number)),
	)

	totalWidth := getTitleWidth(viewportWidth)
	title := lipgloss.NewStyle().Render(pr.Data.Title)

	return makeCellStyle(isSelected).
		Width(totalWidth - 1).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, title, number))
}

func (pr PullRequest) renderAuthor(isSelected bool) string {
	return makeCellStyle(isSelected).Width(prAuthorCellWidth).Render(
		utils.TruncateString(pr.Data.Author.Login, prAuthorCellWidth-cellPadding),
	)
}

func (pr PullRequest) renderRepoName(isSelected bool) string {
	return makeCellStyle(isSelected).
		Width(prRepoCellWidth).
		Render(fmt.Sprintf("%-20s", utils.TruncateString(pr.Data.HeadRepository.Name, 20)))
}

func (pr PullRequest) renderUpdateAt(isSelected bool) string {
	return makeCellStyle(isSelected).
		Width(updatedAtCellWidth).
		Render(utils.TimeElapsed(pr.Data.UpdatedAt))
}

func renderSelectionPointer(isSelected bool) string {
	return makeRuneCellStyle(isSelected).
		Width(emptyCellWidth).
		Render(func() string {
			if isSelected {
				return selectionPointerStyle.Render("")
			} else {
				return " "
			}
		}())
}

func (pr PullRequest) render(isSelected bool, viewPortWidth int) string {
	selectionPointerCell := renderSelectionPointer(isSelected)
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
			selectionPointerCell,
			reviewCell,
			prTitleCell,
			mergeableCell,
			ciCell,
			linesCell,
			prAuthorCell,
			prRepoCell,
			updatedAtCell,
		))
}
