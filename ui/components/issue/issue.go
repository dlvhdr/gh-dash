package issue

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/utils"
)

type Issue struct {
	Data  data.IssueData
	Width int
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

func (issue *Issue) renderUpdateAt() string {
	return lipgloss.NewStyle().
		Render(utils.TimeElapsed(issue.Data.UpdatedAt))
}

func (issue *Issue) renderRepoName() string {
	repoName := utils.TruncateString(issue.Data.Repository.Name, 18)
	return lipgloss.NewStyle().
		Render(repoName)
}

func (issue *Issue) renderTitle() string {
	title := issue.Data.Title
	if len(strings.TrimSpace(title)) == 0 {
		title = "-"
	}
	title = fmt.Sprintf("#%d %v",
		issue.Data.Number,
		titleText.Copy().Render(title),
	)

	return title
}

func (issue *Issue) renderOpenedBy() string {
	return lipgloss.NewStyle().Render(issue.Data.Author.Login)
}

func (issue *Issue) renderAssignees() string {
	assignees := make([]string, 0, len(issue.Data.Assignees.Nodes))
	for _, assignee := range issue.Data.Assignees.Nodes {
		assignees = append(assignees, assignee.Login)
	}
	return lipgloss.NewStyle().Render(strings.Join(assignees, ","))
}

func (issue *Issue) renderStatus() string {
	if issue.Data.State == "OPEN" {
		return lipgloss.NewStyle().Foreground(OpenIssue).Render("")
	} else {
		return lipgloss.NewStyle().Foreground(OpenIssue).Render("")
	}
}

func (issue *Issue) renderNumComments() string {
	return lipgloss.NewStyle().Render(fmt.Sprintf("%d", issue.Data.Comments.TotalCount))
}

func (issue *Issue) renderNumReactions() string {
	return lipgloss.NewStyle().Render(fmt.Sprintf("%d", issue.Data.Reactions.TotalCount))
}
