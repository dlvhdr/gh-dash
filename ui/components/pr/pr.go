package pr

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/utils"
)

type PullRequest struct {
	Ctx  *context.ProgramContext
	Data data.PullRequestData
}

type sectionPullRequestsFetchedMsg struct {
	SectionId int
	Prs       []PullRequest
}

func (pr *PullRequest) getTextStyle() lipgloss.Style {
	return components.GetIssueTextStyle(pr.Ctx, pr.Data.State)
}

func (pr *PullRequest) renderReviewStatus() string {
	reviewCellStyle := pr.getTextStyle()
	if pr.Data.ReviewDecision == "APPROVED" {
		if pr.Data.State == "OPEN" {
			reviewCellStyle = reviewCellStyle.Foreground(pr.Ctx.Theme.SuccessText)
		}
		return reviewCellStyle.Render("󰄬")
	}

	if pr.Data.ReviewDecision == "CHANGES_REQUESTED" {
		if pr.Data.State == "OPEN" {
			reviewCellStyle = reviewCellStyle.Foreground(pr.Ctx.Theme.WarningText)
		}
		return reviewCellStyle.Render("󰌑")
	}

	return reviewCellStyle.Render(pr.Ctx.Styles.Common.WaitingGlyph)
}

func (pr *PullRequest) renderState() string {
	mergeCellStyle := lipgloss.NewStyle()
	switch pr.Data.State {
	case "OPEN":
		if pr.Data.IsDraft {
			return mergeCellStyle.Foreground(pr.Ctx.Theme.FaintText).Render("")
		} else {
			return mergeCellStyle.Foreground(pr.Ctx.Styles.Colors.OpenPR).Render("")
		}
	case "CLOSED":
		return mergeCellStyle.Foreground(pr.Ctx.Styles.Colors.ClosedPR).Render("")
	case "MERGED":
		return mergeCellStyle.Foreground(pr.Ctx.Styles.Colors.MergedPR).Render("")
	default:
		return mergeCellStyle.Foreground(pr.Ctx.Theme.FaintText).Render("-")
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
			ciCellStyle = ciCellStyle.Foreground(pr.Ctx.Theme.SuccessText)
		}
		return ciCellStyle.Render(constants.SuccessIcon)
	}

	if accStatus == "PENDING" {
		return ciCellStyle.Render(pr.Ctx.Styles.Common.WaitingGlyph)
	}

	if pr.Data.State == "OPEN" {
		ciCellStyle = ciCellStyle.Foreground(pr.Ctx.Theme.WarningText)
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
	return components.RenderIssueTitle(pr.Ctx, pr.Data.State, pr.Data.Title, pr.Data.Number)
}

func (pr *PullRequest) renderAuthor() string {
	return pr.getTextStyle().Render(pr.Data.Author.Login)
}

func (pr *PullRequest) renderAssignees() string {
	assignees := make([]string, 0, len(pr.Data.Assignees.Nodes))
	for _, assignee := range pr.Data.Assignees.Nodes {
		assignees = append(assignees, assignee.Login)
	}
	return pr.getTextStyle().Render(strings.Join(assignees, ","))
}

func (pr *PullRequest) renderRepoName() string {
	repoName := pr.Data.HeadRepository.Name
	return pr.getTextStyle().Render(repoName)
}

func (pr *PullRequest) renderUpdateAt() string {
	return pr.getTextStyle().
		Render(utils.TimeElapsed(pr.Data.UpdatedAt))
}

func (pr *PullRequest) renderBaseName() string {
	return pr.getTextStyle().Render(pr.Data.BaseRefName)
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
		return "󰗨Closed"
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
		pr.renderAssignees(),
		pr.renderBaseName(),
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
