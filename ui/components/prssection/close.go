package prssection

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/utils"
)

func (m *Model) close() tea.Cmd {
	pr := m.GetCurrRow()
	prNumber := pr.GetNumber()
	taskId := fmt.Sprintf("pr_close_%d", prNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Closing PR #%d", prNumber),
		FinishedText: fmt.Sprintf("PR #%d has been closed", prNumber),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		c := exec.Command(
			"gh",
			"pr",
			"close",
			fmt.Sprint(m.GetCurrRow().GetNumber()),
			"-R",
			m.GetCurrRow().GetRepoNameWithOwner(),
		)

		err := c.Run()
		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: UpdatePRMsg{
				PrNumber: prNumber,
				IsClosed: utils.BoolPtr(true),
			},
		}
	})
}
