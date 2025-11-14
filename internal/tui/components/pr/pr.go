package pr

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	checks "github.com/dlvhdr/x/gh-checks"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/git"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type PullRequest struct {
	Ctx            *context.ProgramContext
	Data           *data.PullRequestData
	Branch         git.Branch
	Columns        []table.Column
	ShowAuthorIcon bool
}

func (pr *PullRequest) getTextStyle() lipgloss.Style {
	return components.GetIssueTextStyle(pr.Ctx)
}

func (pr *PullRequest) renderReviewStatus() string {
	if pr.Data == nil {
		return "-"
	}
	reviewCellStyle := pr.getTextStyle()
	if pr.Data.ReviewDecision == "APPROVED" {
		reviewCellStyle = reviewCellStyle.Foreground(
			pr.Ctx.Theme.SuccessText,
		)
		return reviewCellStyle.Render("󰄬")
	}

	if pr.Data.ReviewDecision == "CHANGES_REQUESTED" {
		reviewCellStyle = reviewCellStyle.Foreground(
			pr.Ctx.Theme.ErrorText,
		)
		return reviewCellStyle.Render("")
	}

	if pr.Data.Reviews.TotalCount > 0 {
		return reviewCellStyle.Render(pr.Ctx.Styles.Common.CommentGlyph)
	}

	return reviewCellStyle.Render(pr.Ctx.Styles.Common.WaitingGlyph)
}

func (pr *PullRequest) renderState() string {
	mergeCellStyle := lipgloss.NewStyle()

	if pr.Data == nil {
		return mergeCellStyle.Foreground(pr.Ctx.Theme.SuccessText).Render("󰜛")
	}

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

func (pr *PullRequest) GetStatusChecksRollup() checks.CommitState {
	commits := pr.Data.Commits.Nodes
	if len(commits) == 0 {
		return checks.CommitStateUnknown
	}

	return commits[0].Commit.StatusCheckRollup.State
}

func (pr *PullRequest) renderCiStatus() string {
	if pr.Data == nil {
		return "-"
	}

	accStatus := pr.GetStatusChecksRollup()
	ciCellStyle := pr.getTextStyle()

	switch accStatus {
	case checks.CommitStateSuccess:
		ciCellStyle = ciCellStyle.Foreground(pr.Ctx.Theme.SuccessText)
		return ciCellStyle.Render(constants.SuccessIcon)
	case checks.CommitStateExpected, checks.CommitStatePending:
		return ciCellStyle.Render(pr.Ctx.Styles.Common.WaitingGlyph)
	case checks.CommitStateError, checks.CommitStateFailure:
		ciCellStyle = ciCellStyle.Foreground(pr.Ctx.Theme.ErrorText)
		return ciCellStyle.Render(constants.FailureIcon)
	default:
		ciCellStyle = ciCellStyle.Foreground(pr.Ctx.Theme.FaintText)
		return ciCellStyle.Render(constants.EmptyIcon)
	}
}

func (pr *PullRequest) RenderLines(isSelected bool) string {
	if pr.Data == nil {
		return "-"
	}
	deletions := max(pr.Data.Deletions, 0)

	var additionsFg, deletionsFg lipgloss.AdaptiveColor
	additionsFg = pr.Ctx.Theme.SuccessText
	deletionsFg = pr.Ctx.Theme.ErrorText

	baseStyle := lipgloss.NewStyle()
	if isSelected {
		baseStyle = baseStyle.Background(pr.Ctx.Theme.SelectedBackground)
	}

	additionsText := baseStyle.
		Foreground(additionsFg).
		Render(fmt.Sprintf("+%s", components.FormatNumber(pr.Data.Additions)))
	deletionsText := baseStyle.
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

	author := baseStyle.Render(fmt.Sprintf("@%s",
		pr.Data.GetAuthor(pr.Ctx.Theme, pr.ShowAuthorIcon)))
	top := lipgloss.JoinHorizontal(lipgloss.Top, pr.Data.Repository.NameWithOwner,
		fmt.Sprintf(" #%d by %s", pr.Data.Number, author))
	branchHidden := pr.Ctx.Config.Defaults.Layout.Prs.Base.Hidden
	if branchHidden == nil || !*branchHidden {
		branch := baseStyle.Render(pr.Data.HeadRefName)
		top = lipgloss.JoinHorizontal(lipgloss.Top, top, baseStyle.Render(" · "), branch)
	}
	title := pr.Data.Title
	var titleColumn table.Column
	for _, column := range pr.Columns {
		if column.Grow != nil && *column.Grow {
			titleColumn = column
		}
	}
	width := titleColumn.ComputedWidth - 2
	top = baseStyle.Foreground(pr.Ctx.Theme.SecondaryText).Width(width).MaxWidth(width).Height(1).MaxHeight(1).Render(top)
	title = baseStyle.Foreground(pr.Ctx.Theme.PrimaryText).Width(width).MaxWidth(width).Height(1).MaxHeight(1).Render(title)

	return baseStyle.Render(lipgloss.JoinVertical(lipgloss.Left, top, title))
}

func (pr *PullRequest) renderAuthor() string {
	return pr.getTextStyle().Render(pr.Data.GetAuthor(pr.Ctx.Theme, pr.ShowAuthorIcon))
}

func (pr *PullRequest) renderAssignees() string {
	if pr.Data == nil {
		return ""
	}
	assignees := make([]string, 0, len(pr.Data.Assignees.Nodes))
	for _, assignee := range pr.Data.Assignees.Nodes {
		assignees = append(assignees, assignee.Login)
	}
	return pr.getTextStyle().Render(strings.Join(assignees, ","))
}

func (pr *PullRequest) renderRepoName() string {
	repoName := ""
	if !pr.Ctx.Config.Theme.Ui.Table.Compact {
		repoName = pr.Data.Repository.NameWithOwner
	} else {
		repoName = pr.Data.HeadRepository.Name
	}
	return pr.getTextStyle().Foreground(pr.Ctx.Theme.FaintText).Render(repoName)
}

func (pr *PullRequest) renderUpdateAt() string {
	timeFormat := pr.Ctx.Config.Defaults.DateFormat

	updatedAtOutput := ""
	t := pr.Branch.LastUpdatedAt
	if pr.Data != nil {
		t = &pr.Data.UpdatedAt
	}

	if t == nil {
		return ""
	}

	if timeFormat == "" || timeFormat == "relative" {
		updatedAtOutput = utils.TimeElapsed(*t)
	} else {
		updatedAtOutput = t.Format(timeFormat)
	}

	return pr.getTextStyle().Foreground(pr.Ctx.Theme.FaintText).Render(updatedAtOutput)
}

func (pr *PullRequest) renderCreatedAt() string {
	timeFormat := pr.Ctx.Config.Defaults.DateFormat

	createdAtOutput := ""
	t := pr.Branch.CreatedAt
	if pr.Data != nil {
		t = &pr.Data.CreatedAt
	}

	if t == nil {
		return ""
	}

	if timeFormat == "" || timeFormat == "relative" {
		createdAtOutput = utils.TimeElapsed(*t)
	} else {
		createdAtOutput = t.Format(timeFormat)
	}

	return pr.getTextStyle().Foreground(pr.Ctx.Theme.FaintText).Render(createdAtOutput)
}

func (pr *PullRequest) renderBaseName() string {
	if pr.Data == nil {
		return ""
	}
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

func (pr *PullRequest) RenderMergeStateStatus() string {
	switch pr.Data.MergeStateStatus {
	case "CLEAN":
		return constants.SuccessIcon + " Up-to-date"
	case "BLOCKED":
		return constants.BlockedIcon + " Blocked"
	case "BEHIND":
		return constants.BehindIcon + " Behind"
	default:
		return ""
	}
}

func (pr *PullRequest) ToTableRow(isSelected bool) table.Row {
	if !pr.Ctx.Config.Theme.Ui.Table.Compact {
		return table.Row{
			pr.renderState(),
			pr.renderExtendedTitle(isSelected),
			pr.renderAssignees(),
			pr.renderBaseName(),
			pr.renderReviewStatus(),
			pr.renderCiStatus(),
			pr.RenderLines(isSelected),
			pr.renderUpdateAt(),
			pr.renderCreatedAt(),
		}
	}

	return table.Row{
		pr.renderState(),
		pr.renderRepoName(),
		pr.renderTitle(),
		pr.renderAuthor(),
		pr.renderAssignees(),
		pr.renderBaseName(),
		pr.renderReviewStatus(),
		pr.renderCiStatus(),
		pr.RenderLines(isSelected),
		pr.renderUpdateAt(),
		pr.renderCreatedAt(),
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
