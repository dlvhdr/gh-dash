package issuerow

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type Issue struct {
	Ctx            *context.ProgramContext
	Data           data.IssueData
	ShowAuthorIcon bool
}

func (issue *Issue) ToTableRow() table.Row {
	return table.Row{
		issue.renderStatus(),
		issue.renderRepoName(),
		issue.renderType(),
		issue.renderTitle(),
		issue.renderOpenedBy(),
		issue.renderAssignees(),
		issue.renderLinkedPrs(),
		issue.renderNumComments(),
		issue.renderNumReactions(),
		issue.renderUpdateAt(),
		issue.renderCreatedAt(),
	}
}

func (issue *Issue) getTextStyle() lipgloss.Style {
	return components.GetIssueTextStyle(issue.Ctx)
}

func (issue *Issue) renderUpdateAt() string {
	timeFormat := issue.Ctx.Config.Defaults.DateFormat

	updatedAtOutput := ""
	if timeFormat == "" || timeFormat == "relative" {
		updatedAtOutput = utils.TimeElapsed(issue.Data.UpdatedAt)
	} else {
		updatedAtOutput = issue.Data.UpdatedAt.Format(timeFormat)
	}

	return issue.getTextStyle().Render(updatedAtOutput)
}

func (issue *Issue) renderCreatedAt() string {
	timeFormat := issue.Ctx.Config.Defaults.DateFormat

	createdAtOutput := ""
	if timeFormat == "" || timeFormat == "relative" {
		createdAtOutput = utils.TimeElapsed(issue.Data.CreatedAt)
	} else {
		createdAtOutput = issue.Data.CreatedAt.Format(timeFormat)
	}

	return issue.getTextStyle().Render(createdAtOutput)
}

func (issue *Issue) renderRepoName() string {
	repoName := issue.Data.Repository.Name
	return issue.getTextStyle().Render(repoName)
}

func (issue *Issue) renderType() string {
	issueType := issue.Data.IssueType
	if issueType == nil {
		return ""
	}
	return lipgloss.NewStyle().
		Foreground(common.IssueTypeColor(issueType.Color)).
		Render(issueType.Name)
}

func (issue *Issue) renderTitle() string {
	return components.RenderIssueTitle(
		issue.Ctx,
		issue.Data.State,
		issue.Data.Title,
		issue.Data.Number,
	)
}

func (issue *Issue) renderOpenedBy() string {
	return issue.getTextStyle().Render(issue.Data.GetAuthor(issue.Ctx.Theme, issue.ShowAuthorIcon))
}

func (issue *Issue) renderAssignees() string {
	assignees := make([]string, 0, len(issue.Data.Assignees.Nodes))
	for _, assignee := range issue.Data.Assignees.Nodes {
		assignees = append(assignees, assignee.Login)
	}
	return issue.getTextStyle().Render(strings.Join(assignees, ","))
}

func (issue *Issue) renderStatus() string {
	if issue.Data.State == "OPEN" {
		return lipgloss.NewStyle().Foreground(issue.Ctx.Styles.Colors.OpenIssue).Render("")
	} else {
		return issue.getTextStyle().Render("")
	}
}

// LinkedPrsColor is the color for the issue's linked PRs: the color of the state their State() sums up to.
func (issue *Issue) LinkedPrsColor() compat.AdaptiveColor {
	switch issue.Data.LinkedPrs.State() {
	case "OPEN":
		return issue.Ctx.Styles.Colors.OpenPR
	case "DRAFT":
		return issue.Ctx.Theme.FaintText
	default:
		return issue.Ctx.Styles.Colors.MergedPR
	}
}

// renderLinkedPrs renders the number of PRs linked to the issue, after a PR icon in LinkedPrsColor() -- or nothing, if
// it has none.
func (issue *Issue) renderLinkedPrs() string {
	count := issue.Data.LinkedPrs.TotalCount
	if count == 0 {
		return ""
	}

	// The table paints a selected row's background around each cell's content, so a reset between the icon and the
	// count would clear that background behind the count.
	icon := lipgloss.NewStyle().Foreground(issue.LinkedPrsColor()).Render(constants.OpenIcon)
	icon = strings.ReplaceAll(icon, "\x1b[0m", "")
	icon = strings.ReplaceAll(icon, "\x1b[m", "")

	return icon + issue.Ctx.Styles.Common.FaintTextStyle.Render(fmt.Sprintf(" %d", count))
}

func (issue *Issue) renderNumComments() string {
	return issue.getTextStyle().Render(fmt.Sprintf("%d", issue.Data.Comments.TotalCount))
}

func (issue *Issue) renderNumReactions() string {
	return issue.getTextStyle().Render(fmt.Sprintf("%d", issue.Data.Reactions.TotalCount))
}
