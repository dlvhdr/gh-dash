package prview

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/provider"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func (m *Model) unassign(usernames []string) tea.Cmd {
	pr := m.pr.Data.Primary
	prNumber := pr.GetNumber()
	taskId := fmt.Sprintf("pr_unassign_%d", prNumber)

	label := "PR"
	if provider.IsGitLab() {
		label = "MR"
	}

	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Unassigning %s from %s #%d", usernames, label, prNumber),
		FinishedText: fmt.Sprintf("%s unassigned from %s #%d", usernames, label, prNumber),
		State:        context.TaskStart,
		Error:        nil,
	}

	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		var c *exec.Cmd
		if provider.IsGitLab() {
			// GitLab: unassign by setting empty assignee (glab mr update --unassign)
			commandArgs := []string{
				"mr",
				"update",
				fmt.Sprint(prNumber),
				"--repo",
				pr.GetRepoNameWithOwner(),
				"--unassign",
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
				commandArgs = append(commandArgs, "--remove-assignee")
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
				PrNumber:         prNumber,
				RemovedAssignees: &returnedAssignees,
			},
		}
	})
}
