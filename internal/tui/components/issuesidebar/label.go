package issuesidebar

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/issuessection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func (m *Model) label(labels []string) tea.Cmd {
	issue := m.issue.Data
	issueNumber := issue.GetNumber()
	taskId := fmt.Sprintf("issue_label_%d", issueNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Labeling issue #%d to %s", issueNumber, labels),
		FinishedText: fmt.Sprintf("Issue #%d has been labeled with %s", issueNumber, labels),
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
	labelsMap := make(map[string]bool)
	for _, label := range labels {
		labelsMap[label] = true
	}

	for _, label := range m.issue.Data.Labels.Nodes {
		if _, ok := labelsMap[label.Name]; !ok {
			commandArgs = append(commandArgs, "--remove-label")
			commandArgs = append(commandArgs, label.Name)
		}
	}

	for _, label := range labels {
		commandArgs = append(commandArgs, "--add-label")
		commandArgs = append(commandArgs, label)
	}

	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		c := exec.Command("gh", commandArgs...)

		err := c.Run()

		returnedLabels := data.IssueLabels{Nodes: []data.Label{}}
		for _, label := range labels {
			returnedLabels.Nodes = append(returnedLabels.Nodes, data.Label{Name: label})
		}
		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: issuessection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: issuessection.UpdateIssueMsg{
				IssueNumber: issueNumber,
				Labels:      &returnedLabels,
			},
		}
	})
}
