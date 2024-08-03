package pr

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/components"
	"github.com/dlvhdr/gh-dash/v4/ui/components/table"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/utils"
)

type PullRequest struct {
	Ctx  *context.ProgramContext
	Data data.PullRequestData
}

func (pr *PullRequest) getTextStyle() lipgloss.Style {
	return components.GetIssueTextStyle(pr.Ctx, pr.Data.State)
}

func (pr *PullRequest) renderReviewStatus() string {
	reviewCellStyle := pr.getTextStyle()
	if pr.Data.ReviewDecision == "APPROVED" {
		reviewCellStyle = reviewCellStyle.Foreground(
			pr.Ctx.Theme.SuccessText,
		)
		return reviewCellStyle.Render("󰄬")
	}

	if pr.Data.ReviewDecision == "CHANGES_REQUESTED" {
		reviewCellStyle = reviewCellStyle.Foreground(
			pr.Ctx.Theme.WarningText,
		)
		return reviewCellStyle.Render("󰌑")
	}

	return reviewCellStyle.Render(pr.Ctx.Styles.Common.WaitingGlyph)
}

func (pr *PullRequest) renderState() string {
	mergeCellStyle := lipgloss.NewStyle()
	switch pr.Data.State {
	case "OPEN":
		if pr.Data.IsDraft {
			return mergeCellStyle.Foreground(pr.Ctx.Theme.FaintText).Render(constants.DraftIcon)
		} else {
			return mergeCellStyle.Foreground(pr.Ctx.Styles.Colors.OpenPR).Render(constants.OpenIcon)
		}
	case "CLOSED":
		return mergeCellStyle.Foreground(pr.Ctx.Styles.Colors.ClosedPR).
			Render(constants.ClosedIcon)
	case "MERGED":
		return mergeCellStyle.Foreground(pr.Ctx.Styles.Colors.MergedPR).
			Render(constants.MergedIcon)
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
		ciCellStyle = ciCellStyle.Foreground(pr.Ctx.Theme.SuccessText)
		return ciCellStyle.Render(constants.SuccessIcon)
	}

	if accStatus == "PENDING" {
		return ciCellStyle.Render(pr.Ctx.Styles.Common.WaitingGlyph)
	}

	ciCellStyle = ciCellStyle.Foreground(pr.Ctx.Theme.WarningText)
	return ciCellStyle.Render(constants.FailureIcon)
}

func (pr *PullRequest) renderLines(isSelected bool) string {
	deletions := 0
	if pr.Data.Deletions > 0 {
		deletions = pr.Data.Deletions
	}

	var additionsFg, deletionsFg lipgloss.AdaptiveColor
	additionsFg = pr.Ctx.Theme.SuccessText
	deletionsFg = pr.Ctx.Theme.WarningText

	baseStyle := lipgloss.NewStyle()
	if isSelected {
		baseStyle = baseStyle.Background(pr.Ctx.Theme.SelectedBackground)
	}

	additionsText := baseStyle.Copy().
		Foreground(additionsFg).
		Render(fmt.Sprintf("+%s", components.FormatNumber(pr.Data.Additions)))
	deletionsText := baseStyle.Copy().
		Foreground(deletionsFg).
		Render(fmt.Sprintf("-%s", components.FormatNumber(deletions)))

	return pr.getTextStyle().Render(
		keepSameSpacesOnAddDeletions(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				additionsText,
				baseStyle.Render(" "),
				deletionsText,
			)),
	)
}

func keepSameSpacesOnAddDeletions(str string) string {
	strAsList := strings.Split(str, " ")
	return fmt.Sprintf(
		"%7s",
		strAsList[0],
	) + " " + fmt.Sprintf(
		"%7s",
		strAsList[1],
	)
}

func (pr *PullRequest) renderTitle() string {
	return components.RenderIssueTitle(
		pr.Ctx,
		pr.Data.State,
		pr.Data.Title,
		pr.Data.Number,
	)
}

func (pr *PullRequest) renderExtendedTitle(isSelected bool) string {
	baseStyle := lipgloss.NewStyle()
	if isSelected {
		baseStyle = baseStyle.Foreground(pr.Ctx.Theme.SecondaryText).Background(pr.Ctx.Theme.SelectedBackground)
	}
	repoName := baseStyle.Render(pr.Data.Repository.NameWithOwner)
	prNumber := baseStyle.Render(fmt.Sprintf("#%d ", pr.Data.Number))
	top := lipgloss.JoinHorizontal(lipgloss.Top, repoName, baseStyle.Render(" "), prNumber)
	title := pr.Data.Title
	width := max(lipgloss.Width(top), lipgloss.Width(title))
	top = baseStyle.Foreground(pr.Ctx.Theme.SecondaryText).Width(width).Render(top)
	title = baseStyle.Foreground(pr.Ctx.Theme.PrimaryText).Width(width).Render(title)

	return baseStyle.Render(lipgloss.JoinVertical(lipgloss.Left, top, title))
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
	repoName := ""
	if pr.Ctx.Config.Theme.Ui.Table.Multiline {
		repoName = pr.Data.Repository.NameWithOwner
	} else {
		repoName = pr.Data.HeadRepository.Name
	}
	return pr.getTextStyle().Copy().Foreground(pr.Ctx.Theme.FaintText).Render(repoName)
}

func (pr *PullRequest) renderUpdateAt() string {
	timeFormat := pr.Ctx.Config.Defaults.DateFormat

	updatedAtOutput := ""
	if timeFormat == "" || timeFormat == "relative" {
		updatedAtOutput = utils.TimeElapsed(pr.Data.UpdatedAt)
	} else {
		updatedAtOutput = pr.Data.UpdatedAt.Format(timeFormat)
	}

	return pr.getTextStyle().Copy().Foreground(pr.Ctx.Theme.FaintText).Render(updatedAtOutput)
}

func (pr *PullRequest) renderBaseName() string {
	return pr.getTextStyle().Render(pr.Data.BaseRefName)
}

func (pr *PullRequest) RenderState() string {
	switch pr.Data.State {
	case "OPEN":
		if pr.Data.IsDraft {
			return constants.DraftIcon + " Draft"
		} else {
			return constants.OpenIcon + " Open"
		}
	case "CLOSED":
		return constants.ClosedIcon + " Closed"
	case "MERGED":
		return constants.MergedIcon + " Merged"
	default:
		return ""
	}
}

func (pr *PullRequest) ToTableRow(isSelected bool) table.Row {
	if pr.Ctx.Config.Theme.Ui.Table.Multiline {
		return table.Row{
			pr.renderState(),
			pr.renderExtendedTitle(isSelected),
			pr.renderBaseName(),
			pr.renderAssignees(),
			pr.renderReviewStatus(),
			pr.renderCiStatus(),
			pr.renderLines(isSelected),
			pr.renderUpdateAt(),
		}
	}

	return table.Row{
		pr.renderState(),
		pr.renderRepoName(),
		pr.renderTitle(),
		pr.renderAuthor(),
		pr.renderBaseName(),
		pr.renderAssignees(),
		pr.renderReviewStatus(),
		pr.renderCiStatus(),
		pr.renderLines(isSelected),
		pr.renderUpdateAt(),
	}
}

func isConclusionAFailure(conclusion string) bool {
	return conclusion == "FAILURE" || conclusion == "TIMED_OUT" ||
		conclusion == "STARTUP_FAILURE"
}

func isStatusWaiting(status string) bool {
	return status == "PENDING" ||
		status == "QUEUED" ||
		status == "IN_PROGRESS" ||
		status == "WAITING"
}
