package pr

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/data"
	"github.com/dlvhdr/gh-prs/ui/components/table"
	"github.com/dlvhdr/gh-prs/ui/constants"
	"github.com/dlvhdr/gh-prs/ui/styles"
	"github.com/dlvhdr/gh-prs/utils"
)

type PullRequest struct {
	Data data.PullRequestData
}

type sectionPullRequestsFetchedMsg struct {
	SectionId int
	Prs       []PullRequest
}

func (pr PullRequest) renderReviewStatus(isSelected bool) string {
	reviewCellStyle := lipgloss.NewStyle()
	if pr.Data.ReviewDecision == "APPROVED" {
		return reviewCellStyle.Foreground(successText).Render("")
	}

	if pr.Data.ReviewDecision == "CHANGES_REQUESTED" {
		return reviewCellStyle.Foreground(styles.DefaultTheme.WarningText).Render("")
	}

	return reviewCellStyle.Render(constants.WaitingGlyph)
}

func (pr PullRequest) renderMergeableStatus(isSelected bool) string {
	mergeCellStyle := lipgloss.NewStyle()
	switch pr.Data.Mergeable {
	case "MERGEABLE":
		return mergeCellStyle.Foreground(successText).Render("")
	case "CONFLICTING":
		return mergeCellStyle.Render(constants.FailureGlyph)
	case "UNKNOWN":
		fallthrough
	default:
		return mergeCellStyle.Foreground(styles.DefaultTheme.FaintText).Render("-")
	}
}

func (pr PullRequest) GetStatusChecksRollup() string {
	accStatus := "SUCCESS"
	mostRecentCommit := pr.Data.Commits.Nodes[0].Commit
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
		}

		if isConclusionAFailure(conclusion) {
			accStatus = "FAILURE"
			break
		}
	}

	return accStatus
}

func (pr PullRequest) renderCiStatus(isSelected bool) string {
	accStatus := pr.GetStatusChecksRollup()
	ciCellStyle := lipgloss.NewStyle()
	if accStatus == "SUCCESS" {
		return ciCellStyle.Render(constants.SuccessGlyph)
	}

	if accStatus == "PENDING" {
		return ciCellStyle.Render(constants.WaitingGlyph)
	}

	return ciCellStyle.Render(constants.FailureGlyph)
}

func (pr PullRequest) renderLines(isSelected bool) string {
	deletions := 0
	if pr.Data.Deletions > 0 {
		deletions = pr.Data.Deletions
	}

	return lipgloss.NewStyle().Render(
		fmt.Sprintf("%d / -%d", pr.Data.Additions, deletions),
	)
}

func (pr PullRequest) renderTitle(viewportWidth int, isSelected bool) string {
	return lipgloss.NewStyle().
		Inline(true).
		Foreground(styles.DefaultTheme.SecondaryText).
		Render(
			fmt.Sprintf("#%d %s",
				pr.Data.Number,
				titleText.Copy().Render(pr.Data.Title),
			),
		)
}

func (pr PullRequest) renderAuthor(isSelected bool) string {
	return lipgloss.NewStyle().Render(pr.Data.Author.Login)
}

func (pr PullRequest) renderRepoName(isSelected bool) string {
	repoName := utils.TruncateString(pr.Data.HeadRepository.Name, 18)
	return lipgloss.NewStyle().
		Render(repoName)
}

func (pr PullRequest) renderUpdateAt(isSelected bool) string {
	return lipgloss.NewStyle().
		Render(utils.TimeElapsed(pr.Data.UpdatedAt))
}

func (pr PullRequest) RenderState() string {
	switch pr.Data.State {
	case "OPEN":
		return " Open"
	case "CLOSED":
		return "﫧Closed"
	case "MERGED":
		return " Merged"
	default:
		return ""
	}
}

func (pr PullRequest) Render(isSelected bool, viewPortWidth int) table.Row {
	return table.Row{
		pr.renderUpdateAt(isSelected),
		pr.renderReviewStatus(isSelected),
		pr.renderRepoName(isSelected),
		pr.renderTitle(viewPortWidth, isSelected),
		pr.renderAuthor(isSelected),
		pr.renderMergeableStatus(isSelected),
		pr.renderCiStatus(isSelected),
		pr.renderLines(isSelected),
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
