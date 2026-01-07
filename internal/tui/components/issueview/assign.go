package issueview

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/provider"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/issuessection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func (m *Model) assign(usernames []string) tea.Cmd {
	issue := m.issue.Data
	issueNumber := issue.GetNumber()
	taskId := fmt.Sprintf("issue_assign_%d", issueNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Assigning issue #%d to %s", issueNumber, usernames),
		FinishedText: fmt.Sprintf("Issue #%d has been assigned to %s", issueNumber, usernames),
		State:        context.TaskStart,
		Error:        nil,
	}

	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		var c *exec.Cmd
		if provider.IsGitLab() {
			commandArgs := []string{
				"issue",
				"update",
				fmt.Sprint(issueNumber),
				"--repo",
				issue.GetRepoNameWithOwner(),
				"--assignee",
				strings.Join(usernames, ","),
			}
			c = exec.Command("glab", commandArgs...)
		} else {
			commandArgs := []string{
				"issue",
				"edit",
				fmt.Sprint(issueNumber),
				"-R",
				issue.GetRepoNameWithOwner(),
			}
			for _, assignee := range usernames {
				commandArgs = append(commandArgs, "--add-assignee")
				commandArgs = append(commandArgs, assignee)
			}
			c = exec.Command("gh", commandArgs...)
		}

		err := c.Run()
		returnedAssignees := data.Assignees{Nodes: []data.Assignee{}}
		for _, assignee := range usernames {
			returnedAssignees.Nodes = append(returnedAssignees.Nodes, data.Assignee{Login: assignee})
		}
		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: issuessection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: issuessection.UpdateIssueMsg{
				IssueNumber:    issueNumber,
				AddedAssignees: &returnedAssignees,
			},
		}
	})
}
