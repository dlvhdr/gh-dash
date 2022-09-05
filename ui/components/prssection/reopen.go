package prssection

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/utils"
)

func (m *Model) reopen() tea.Cmd {
	pr := m.GetCurrRow()
	prNumber := pr.GetNumber()
	taskId := fmt.Sprintf("reopen_%d", prNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Reopening PR #%d", prNumber),
		FinishedText: fmt.Sprintf("PR #%d has been reopened", prNumber),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		c := exec.Command(
			"gh",
			"pr",
			"reopen",
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
				IsClosed: utils.BoolPtr(false),
			},
		}
	})
}
