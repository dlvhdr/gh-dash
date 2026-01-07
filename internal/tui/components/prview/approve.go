package prview

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/provider"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func (m *Model) approve(comment string) tea.Cmd {
	pr := m.pr.Data.Primary
	prNumber := pr.GetNumber()
	taskId := fmt.Sprintf("pr_approve_%d", prNumber)

	label := "PR"
	if provider.IsGitLab() {
		label = "MR"
	}

	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Approving %s #%d", label, prNumber),
		FinishedText: fmt.Sprintf("%s #%d has been approved", label, prNumber),
		State:        context.TaskStart,
		Error:        nil,
	}

	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		var c *exec.Cmd
		if provider.IsGitLab() {
			// GitLab uses 'glab mr approve'
			commandArgs := []string{
				"mr",
				"approve",
				fmt.Sprint(prNumber),
				"--repo",
				pr.GetRepoNameWithOwner(),
			}
			c = exec.Command("glab", commandArgs...)
		} else {
			commandArgs := []string{
				"pr",
				"review",
				"-R",
				pr.GetRepoNameWithOwner(),
				fmt.Sprint(prNumber),
				"--approve",
			}
			if comment != "" {
				commandArgs = append(commandArgs, "--body", comment)
			}
			c = exec.Command("gh", commandArgs...)
		}

		err := c.Run()
		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: prssection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: tasks.UpdatePRMsg{
				PrNumber: prNumber,
			},
		}
	})
}
