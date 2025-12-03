package prview

import (
	"fmt"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func (m *Model) comment(body string) tea.Cmd {
	pr := m.pr.Data.Primary
	prNumber := pr.GetNumber()
	taskId := fmt.Sprintf("pr_comment_%d", prNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Commenting on PR #%d", prNumber),
		FinishedText: fmt.Sprintf("Commented on PR #%d", prNumber),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		c := exec.Command(
			"gh",
			"pr",
			"comment",
			fmt.Sprint(prNumber),
			"-R",
			pr.GetRepoNameWithOwner(),
			"-b",
			body,
		)

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
