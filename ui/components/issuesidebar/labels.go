package issuesidebar

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/ui/components/issuessection"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

func (m *Model) addLabels(labels []string) tea.Cmd {
	issue := m.issue.Data
	issueNumber := issue.GetNumber()
	taskId := fmt.Sprintf("issue_add_labels_%d", issueNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Adding labels to issue #%d: %s", issueNumber, labels),
		FinishedText: fmt.Sprintf("Issue #%d labels added: %s", issueNumber, labels),
		State:        context.TaskStart,
		Error:        nil,
	}

	commandArgs := []string{
		"issue",
		"edit",
		fmt.Sprint(issueNumber),
		"-R",
		issue.GetRepoNameWithOwner(),
	}
	for _, label := range labels {
		commandArgs = append(commandArgs, "--add-label")
		commandArgs = append(commandArgs, label)
	}

	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		c := exec.Command("gh", commandArgs...)

		err := c.Run()
		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: issuessection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: issuessection.UpdateIssueMsg{
				IssueNumber: issueNumber,
			},
		}
	})
}
