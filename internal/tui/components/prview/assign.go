package prview

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/provider"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func (m *Model) assign(usernames []string) tea.Cmd {
	pr := m.pr.Data.Primary
	prNumber := pr.GetNumber()
	taskId := fmt.Sprintf("pr_assign_%d", prNumber)

	label := "PR"
	if provider.IsGitLab() {
		label = "MR"
	}

	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Assigning %s #%d to %s", label, prNumber, usernames),
		FinishedText: fmt.Sprintf("%s #%d has been assigned to %s", label, prNumber, usernames),
		State:        context.TaskStart,
		Error:        nil,
	}

	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		var c *exec.Cmd
		if provider.IsGitLab() {
			// GitLab uses 'glab mr update --assignee'
			commandArgs := []string{
				"mr",
				"update",
				fmt.Sprint(prNumber),
				"--repo",
				pr.GetRepoNameWithOwner(),
				"--assignee",
				strings.Join(usernames, ","),
			}
			c = exec.Command("glab", commandArgs...)
		} else {
			commandArgs := []string{
				"pr",
				"edit",
				fmt.Sprint(prNumber),
				"-R",
				pr.GetRepoNameWithOwner(),
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
			SectionType: prssection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: tasks.UpdatePRMsg{
				PrNumber:       prNumber,
				AddedAssignees: &returnedAssignees,
			},
		}
	})
}
