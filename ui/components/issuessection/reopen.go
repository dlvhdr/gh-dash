package issuessection

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/utils"
)

func (m *Model) reopen() tea.Cmd {
	issue := m.GetCurrRow()
	issueNumber := issue.GetNumber()
	taskId := fmt.Sprintf("issue_reopen_%d", issueNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Reopening issue #%d", issueNumber),
		FinishedText: fmt.Sprintf("Issue #%d has been reopened", issueNumber),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		c := exec.Command(
			"gh",
			"issue",
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
			Msg: UpdateIssueMsg{
				IssueNumber: issueNumber,
				IsClosed:    utils.BoolPtr(false),
			},
		}
	})
}
