package issue

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/utils"
)

type Issue struct {
	Ctx  *context.ProgramContext
	Data data.IssueData
}

func (issue *Issue) ToTableRow() table.Row {
	return table.Row{
		issue.renderUpdateAt(),
		issue.renderStatus(),
		issue.renderRepoName(),
		issue.renderTitle(),
		issue.renderOpenedBy(),
		issue.renderAssignees(),
		issue.renderNumComments(),
		issue.renderNumReactions(),
	}
}

func (issue *Issue) getTextStyle() lipgloss.Style {
	return components.GetIssueTextStyle(issue.Ctx, issue.Data.State)
}

func (issue *Issue) renderUpdateAt() string {
	return issue.getTextStyle().Render(utils.TimeElapsed(issue.Data.UpdatedAt))
}

func (issue *Issue) renderRepoName() string {
	repoName := issue.Data.Repository.Name
	return issue.getTextStyle().Render(repoName)
}

func (issue *Issue) renderTitle() string {
	return components.RenderIssueTitle(issue.Ctx, issue.Data.State, issue.Data.Title, issue.Data.Number)
}

func (issue *Issue) renderOpenedBy() string {
	return issue.getTextStyle().Render(issue.Data.Author.Login)
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

func (issue *Issue) renderNumComments() string {
	return issue.getTextStyle().Render(fmt.Sprintf("%d", issue.Data.Comments.TotalCount))
}

func (issue *Issue) renderNumReactions() string {
	return issue.getTextStyle().Render(fmt.Sprintf("%d", issue.Data.Reactions.TotalCount))
}
