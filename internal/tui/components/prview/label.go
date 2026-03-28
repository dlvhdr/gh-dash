package prview

import (
	"fmt"
	"os/exec"

	tea "charm.land/bubbletea/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func (m *Model) label(labels []string) tea.Cmd {
	pr := m.pr.Data.Primary
	prNumber := pr.GetNumber()
	taskId := fmt.Sprintf("pr_label_%d", prNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Labeling pr #%d to %s", prNumber, labels),
		FinishedText: fmt.Sprintf("pr #%d has been labeled with %s", prNumber, labels),
		State:        context.TaskStart,
		Error:        nil,
	}

	commandArgs := []string{
		"pr",
		"edit",
		fmt.Sprint(prNumber),
		"-R",
		pr.GetRepoNameWithOwner(),
	}
	labelsMap := make(map[string]bool)
	for _, label := range labels {
		labelsMap[label] = true
	}

	existingLabelsColorMap := make(map[string]string)
	for _, label := range m.pr.Data.Primary.Labels.Nodes {
		existingLabelsColorMap[label.Name] = label.Color
	}

	for _, label := range m.pr.Data.Primary.Labels.Nodes {
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

		returnedLabels := data.PRLabels{Nodes: []data.Label{}}
		for _, label := range labels {
			returnedLabels.Nodes = append(returnedLabels.Nodes, data.Label{
				Name:  label,
				Color: existingLabelsColorMap[label],
			})
		}
		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: prssection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: tasks.UpdatePRMsg{
				PrNumber: prNumber,
				Labels:   &returnedLabels,
			},
		}
	})
}
