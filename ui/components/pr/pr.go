package pr

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/styles"
	"github.com/dlvhdr/gh-dash/utils"
)

type PullRequest struct {
	Data data.PullRequestData
}

type sectionPullRequestsFetchedMsg struct {
	SectionId int
	Prs       []PullRequest
}

func (pr *PullRequest) getTextStyle() lipgloss.Style {
	return components.GetIssueTextStyle(pr.Data.State)
}

func (pr *PullRequest) renderReviewStatus() string {
	reviewCellStyle := pr.getTextStyle()
	if pr.Data.ReviewDecision == "APPROVED" {
		if pr.Data.State == "OPEN" {
			reviewCellStyle = reviewCellStyle.Foreground(successText)
		}
		return reviewCellStyle.Render("")
	}

	if pr.Data.ReviewDecision == "CHANGES_REQUESTED" {
		if pr.Data.State == "OPEN" {
			reviewCellStyle = reviewCellStyle.Foreground(styles.DefaultTheme.WarningText)
		}
		return reviewCellStyle.Render("")
	}

	return reviewCellStyle.Render(constants.WaitingGlyph)
}

func (pr *PullRequest) renderState() string {
	mergeCellStyle := lipgloss.NewStyle()
	switch pr.Data.State {
	case "OPEN":
		if pr.Data.IsDraft {
			return mergeCellStyle.Foreground(styles.DefaultTheme.FaintText).Render("")
		} else {
			return mergeCellStyle.Foreground(openPR).Render("")
		}
	case "CLOSED":
		return mergeCellStyle.Foreground(closedPR).Render("")
	case "MERGED":
		return mergeCellStyle.Foreground(mergedPR).Render("")
	default:
		return mergeCellStyle.Foreground(styles.DefaultTheme.FaintText).Render("-")
	}
}

func (pr *PullRequest) GetStatusChecksRollup() string {
	if pr.Data.Mergeable == "CONFLICTING" {
		return "FAILURE"
	}

	accStatus := "SUCCESS"
	commits := pr.Data.Commits.Nodes
	if len(commits) == 0 {
		return "PENDING"
	}

	mostRecentCommit := commits[0].Commit
	for _, statusCheck := range mostRecentCommit.StatusCheckRollup.Contexts.Nodes {
		var conclusion string
		if statusCheck.Typename == "CheckRun" {
			conclusion = string(statusCheck.CheckRun.Conclusion)
			status := string(statusCheck.CheckRun.Status)
			if isStatusWaiting(status) {
				accStatus = "PENDING"
			}
		} else if statusCheck.Typename == "StatusContext" {
			conclusion = string(statusCheck.StatusContext.State)
			if isStatusWaiting(conclusion) {
				accStatus = "PENDING"
			}
		}

		if isConclusionAFailure(conclusion) {
			accStatus = "FAILURE"
			break
		}
	}

	return accStatus
}

func (pr *PullRequest) renderCiStatus() string {

	accStatus := pr.GetStatusChecksRollup()
	ciCellStyle := pr.getTextStyle()
	if accStatus == "SUCCESS" {
		if pr.Data.State == "OPEN" {
			ciCellStyle = ciCellStyle.Foreground(styles.DefaultTheme.SuccessText)
		}
		return ciCellStyle.Render(constants.SuccessIcon)
	}

	if accStatus == "PENDING" {
		return ciCellStyle.Render(constants.WaitingGlyph)
	}

	if pr.Data.State == "OPEN" {
		ciCellStyle = ciCellStyle.Foreground(styles.DefaultTheme.WarningText)
	}
	return ciCellStyle.Render(constants.FailureIcon)
}

func (pr *PullRequest) renderLines() string {
	deletions := 0
	if pr.Data.Deletions > 0 {
		deletions = pr.Data.Deletions
	}

	return pr.getTextStyle().Render(
		fmt.Sprintf("%d / -%d", pr.Data.Additions, deletions),
	)
}

func (pr *PullRequest) renderTitle() string {
	return components.RenderIssueTitle(pr.Data.State, pr.Data.Title, pr.Data.Number)
}

func (pr *PullRequest) renderAuthor() string {
	return pr.getTextStyle().Render(pr.Data.Author.Login)
}

func (pr *PullRequest) renderRepoName() string {
	repoName := utils.TruncateString(pr.Data.HeadRepository.Name, 18)
	return pr.getTextStyle().Render(repoName)
}

func (pr *PullRequest) renderUpdateAt() string {
	return pr.getTextStyle().
		Render(utils.TimeElapsed(pr.Data.UpdatedAt))
}

func (pr *PullRequest) RenderState() string {
	switch pr.Data.State {
	case "OPEN":
		if pr.Data.IsDraft {
			return " Draft"
		} else {
			return " Open"
		}
	case "CLOSED":
		return "﫧Closed"
	case "MERGED":
		return " Merged"
	default:
		return ""
	}
}

func (pr *PullRequest) ToTableRow() table.Row {
	return table.Row{
		pr.renderUpdateAt(),
		pr.renderState(),
		pr.renderRepoName(),
		pr.renderTitle(),
		pr.renderAuthor(),
		pr.renderReviewStatus(),
		pr.renderCiStatus(),
		pr.renderLines(),
	}
}

func isConclusionAFailure(conclusion string) bool {
	return conclusion == "FAILURE" || conclusion == "TIMED_OUT" || conclusion == "STARTUP_FAILURE"
}

func isStatusWaiting(status string) bool {
	return status == "PENDING" ||
		status == "QUEUED" ||
		status == "IN_PROGRESS" ||
		status == "WAITING"
}
