package prview

import (
	"fmt"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/provider"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func (m *Model) comment(body string) tea.Cmd {
	pr := m.pr.Data.Primary
	prNumber := pr.GetNumber()
	taskId := fmt.Sprintf("pr_comment_%d", prNumber)

	label := "PR"
	if provider.IsGitLab() {
		label = "MR"
	}

	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Commenting on %s #%d", label, prNumber),
		FinishedText: fmt.Sprintf("Commented on %s #%d", label, prNumber),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		var c *exec.Cmd
		if provider.IsGitLab() {
			c = exec.Command(
				"glab",
				"mr",
				"note",
				fmt.Sprint(prNumber),
				"--repo",
				pr.GetRepoNameWithOwner(),
				"-m",
				body,
			)
		} else {
			c = exec.Command(
				"gh",
				"pr",
				"comment",
				fmt.Sprint(prNumber),
				"-R",
				pr.GetRepoNameWithOwner(),
				"-b",
				body,
			)
		}

		err := c.Run()
		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: prssection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: tasks.UpdatePRMsg{
				PrNumber: prNumber,
				NewComment: &data.Comment{
					Author:    struct{ Login string }{Login: m.ctx.User},
					Body:      body,
					UpdatedAt: time.Now(),
				},
			},
		}
	})
}
