package branch

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/git"
	"github.com/dlvhdr/gh-dash/v4/ui/components"
	"github.com/dlvhdr/gh-dash/v4/ui/components/table"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/utils"
)

type Branch struct {
	Ctx     *context.ProgramContext
	PR      *data.PullRequestData
	Data    git.Branch
	Columns []table.Column
}

func (b *Branch) getTextStyle() lipgloss.Style {
	return components.GetIssueTextStyle(b.Ctx)
}

func (b *Branch) renderReviewStatus() string {
	if b.PR == nil {
		return "-"
	}
	reviewCellStyle := b.getTextStyle()
	if b.PR.ReviewDecision == "APPROVED" {
		reviewCellStyle = reviewCellStyle.Foreground(
			b.Ctx.Theme.SuccessText,
		)
		return reviewCellStyle.Render("󰄬")
	}

	if b.PR.ReviewDecision == "CHANGES_REQUESTED" {
		reviewCellStyle = reviewCellStyle.Foreground(
			b.Ctx.Theme.ErrorText,
		)
		return reviewCellStyle.Render("󰌑")
	}

	return reviewCellStyle.Render(b.Ctx.Styles.Common.WaitingGlyph)
}

func (b *Branch) renderState() string {
	mergeCellStyle := lipgloss.NewStyle()

	if b.PR == nil {
		return mergeCellStyle.Foreground(b.Ctx.Theme.SuccessText).Render("󰜛")
	}

	switch b.PR.State {
	case "OPEN":
		if b.PR.IsDraft {
			return mergeCellStyle.Foreground(b.Ctx.Theme.FaintText).Render(constants.DraftIcon)
		} else {
			return mergeCellStyle.Foreground(b.Ctx.Styles.Colors.OpenPR).Render(constants.OpenIcon)
		}
	case "CLOSED":
		return mergeCellStyle.Foreground(b.Ctx.Styles.Colors.ClosedPR).
			Render(constants.ClosedIcon)
	case "MERGED":
		return mergeCellStyle.Foreground(b.Ctx.Styles.Colors.MergedPR).
			Render(constants.MergedIcon)
	default:
		return mergeCellStyle.Foreground(b.Ctx.Theme.FaintText).Render("-")
	}
}

func (b *Branch) GetStatusChecksRollup() string {
	if b.PR.Mergeable == "CONFLICTING" {
		return "FAILURE"
	}

	accStatus := "SUCCESS"
	commits := b.PR.Commits.Nodes
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

func (b *Branch) renderCiStatus() string {
	if b.PR == nil {
		return "-"
	}

	accStatus := b.GetStatusChecksRollup()
	ciCellStyle := b.getTextStyle()
	if accStatus == "SUCCESS" {
		ciCellStyle = ciCellStyle.Foreground(b.Ctx.Theme.SuccessText)
		return ciCellStyle.Render(constants.SuccessIcon)
	}

	if accStatus == "PENDING" {
		return ciCellStyle.Render(b.Ctx.Styles.Common.WaitingGlyph)
	}

	ciCellStyle = ciCellStyle.Foreground(b.Ctx.Theme.ErrorText)
	return ciCellStyle.Render(constants.FailureIcon)
}

func (b *Branch) renderLines(isSelected bool) string {
	if b.PR == nil {
		return "-"
	}
	deletions := 0
	if b.PR.Deletions > 0 {
		deletions = b.PR.Deletions
	}

	var additionsFg, deletionsFg lipgloss.AdaptiveColor
	additionsFg = b.Ctx.Theme.SuccessText
	deletionsFg = b.Ctx.Theme.ErrorText

	baseStyle := lipgloss.NewStyle()
	if isSelected {
		baseStyle = baseStyle.Background(b.Ctx.Theme.SelectedBackground)
	}

	additionsText := baseStyle.
		Foreground(additionsFg).
		Render(fmt.Sprintf("+%s", components.FormatNumber(b.PR.Additions)))
	deletionsText := baseStyle.
		Foreground(deletionsFg).
		Render(fmt.Sprintf("-%s", components.FormatNumber(deletions)))

	return b.getTextStyle().Render(
		keepSameSpacesOnAddDeletions(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				additionsText,
				baseStyle.Render(" "),
				deletionsText,
			)),
	)
}

func (b *Branch) renderTitle() string {
	return components.RenderIssueTitle(
		b.Ctx,
		b.PR.State,
		b.PR.Title,
		b.PR.Number,
	)
}

func (b *Branch) renderExtendedTitle(isSelected bool) string {
	baseStyle := b.getBaseStyle(isSelected)
	width := b.getMaxWidth()

	title := b.renderLastCommitMsg(isSelected, width)
	branch := b.renderBranch(isSelected, width)
	return baseStyle.Render(lipgloss.JoinVertical(lipgloss.Left, branch, title))
}

func (pr *Branch) renderAuthor() string {
	return pr.getTextStyle().Render(pr.PR.Author.Login)
}

func (b *Branch) renderAssignees() string {
	if b.PR == nil {
		return ""
	}
	assignees := make([]string, 0, len(b.PR.Assignees.Nodes))
	for _, assignee := range b.PR.Assignees.Nodes {
		assignees = append(assignees, assignee.Login)
	}
	return b.getTextStyle().Render(strings.Join(assignees, ","))
}

func (b *Branch) renderRepoName() string {
	repoName := ""
	if !b.Ctx.Config.Theme.Ui.Table.Compact {
		repoName = b.PR.Repository.NameWithOwner
	} else {
		repoName = b.PR.HeadRepository.Name
	}
	return b.getTextStyle().Foreground(b.Ctx.Theme.FaintText).Render(repoName)
}

func (b *Branch) renderUpdateAt() string {
	timeFormat := b.Ctx.Config.Defaults.DateFormat

	updatedAtOutput := ""
	t := b.Data.LastUpdatedAt
	if b.PR != nil {
		t = &b.PR.UpdatedAt
	}

	if t == nil {
		return ""
	}

	if timeFormat == "" || timeFormat == "relative" {
		updatedAtOutput = utils.TimeElapsed(*t)
	} else {
		updatedAtOutput = t.Format(timeFormat)
	}

	return b.getTextStyle().Foreground(b.Ctx.Theme.FaintText).Render(updatedAtOutput)
}

func (b *Branch) renderBaseName() string {
	if b.PR == nil {
		return ""
	}
	return b.getTextStyle().Render(b.PR.BaseRefName)
}

func (b *Branch) RenderState() string {
	switch b.PR.State {
	case "OPEN":
		if b.PR.IsDraft {
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

func (b *Branch) ToTableRow(isSelected bool) table.Row {
	if !b.Ctx.Config.Theme.Ui.Table.Compact {
		return table.Row{
			b.renderState(),
			b.renderExtendedTitle(isSelected),
			b.renderBaseName(),
			b.renderAssignees(),
			b.renderReviewStatus(),
			b.renderCiStatus(),
			b.renderLines(isSelected),
			b.renderUpdateAt(),
		}
	}

	return table.Row{
		b.renderState(),
		b.renderRepoName(),
		b.renderTitle(),
		b.renderAuthor(),
		b.renderBaseName(),
		b.renderAssignees(),
		b.renderReviewStatus(),
		b.renderCiStatus(),
		b.renderLines(isSelected),
		b.renderUpdateAt(),
	}
}

func (b *Branch) renderBranch(isSelected bool, width int) string {
	baseStyle := b.getBaseStyle(isSelected)
	name := b.Data.Name
	if b.Data.IsCheckedOut {
		name = baseStyle.Foreground(b.Ctx.Theme.SuccessText).Render(name)
	} else {
		name = baseStyle.Foreground(b.Ctx.Theme.PrimaryText).Render(name)
	}
	return baseStyle.MaxHeight(1).Width(width).MaxWidth(width).Render(lipgloss.JoinHorizontal(
		lipgloss.Top,
		name,
		b.renderCommitsAheadBehind(isSelected),
	))

}

func (b *Branch) getBaseStyle(isSelected bool) lipgloss.Style {
	baseStyle := lipgloss.NewStyle()
	if isSelected {
		baseStyle = baseStyle.Background(b.Ctx.Theme.SelectedBackground)
	}

	return baseStyle
}

func (b *Branch) getMaxWidth() int {
	var titleColumn table.Column
	for _, column := range b.Columns {
		if column.Grow != nil && *column.Grow {
			titleColumn = column
		}
	}
	return titleColumn.ComputedWidth - 2
}

func (b *Branch) renderCommitsAheadBehind(isSelected bool) string {
	baseStyle := b.getBaseStyle(isSelected)

	commitsAhead := ""
	commitsBehind := ""
	if b.Data.CommitsAhead > 0 {
		commitsAhead = baseStyle.Foreground(b.Ctx.Theme.WarningText).Render(fmt.Sprintf(" ↑%d", b.Data.CommitsAhead))
	}
	if b.Data.CommitsBehind > 0 {
		commitsBehind = baseStyle.Foreground(b.Ctx.Theme.WarningText).Render(fmt.Sprintf(" ↓%d", b.Data.CommitsBehind))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, commitsAhead, commitsBehind)
}

func (b *Branch) renderLastCommitMsg(isSelected bool, width int) string {
	baseStyle := b.getBaseStyle(isSelected)
	title := "-"
	if b.Data.LastCommitMsg != nil {
		title = *b.Data.LastCommitMsg
	}
	return baseStyle.Foreground(b.Ctx.Theme.SecondaryText).Width(width).MaxWidth(width).Render(title)
}
