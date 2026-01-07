package issueview

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/provider"
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

	labelsMap := make(map[string]bool)
	for _, label := range labels {
		labelsMap[label] = true
	}

	existingLabelsColorMap := make(map[string]string)
	for _, label := range m.issue.Data.Labels.Nodes {
		existingLabelsColorMap[label.Name] = label.Color
	}

	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		var c *exec.Cmd
		if provider.IsGitLab() {
			// GitLab: use glab issue update --label
			commandArgs := []string{
				"issue",
				"update",
				fmt.Sprint(issueNumber),
				"--repo",
				issue.GetRepoNameWithOwner(),
				"--label",
				strings.Join(labels, ","),
			}
			// Remove labels not in the new list
			for _, label := range m.issue.Data.Labels.Nodes {
				if _, ok := labelsMap[label.Name]; !ok {
					commandArgs = append(commandArgs, "--unlabel", label.Name)
				}
			}
			c = exec.Command("glab", commandArgs...)
		} else {
			commandArgs := []string{
				"issue",
				"edit",
				fmt.Sprint(issueNumber),
				"-R",
				issue.GetRepoNameWithOwner(),
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
			c = exec.Command("gh", commandArgs...)
		}

		err := c.Run()

		returnedLabels := data.IssueLabels{Nodes: []data.Label{}}
		for _, label := range labels {
			returnedLabels.Nodes = append(returnedLabels.Nodes, data.Label{
				Name:  label,
				Color: existingLabelsColorMap[label],
			})
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
