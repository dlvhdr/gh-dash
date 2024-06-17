package issuesidebar

import (
	"fmt"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/components/issuessection"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

func (m *Model) comment(body string) tea.Cmd {
	issue := m.issue.Data
	issueNumber := issue.GetNumber()
	taskId := fmt.Sprintf("issue_comment_%d", issueNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Commenting on issue #%d", issueNumber),
		FinishedText: fmt.Sprintf("Commented on issue #%d", issueNumber),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		c := exec.Command(
			"gh",
			"issue",
			"comment",
			fmt.Sprint(issueNumber),
			"-R",
			issue.GetRepoNameWithOwner(),
			"-b",
			body,
		)

		err := c.Run()
		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: issuessection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: issuessection.UpdateIssueMsg{
				IssueNumber: issueNumber,
				NewComment: &data.IssueComment{
					Author:    struct{ Login string }{Login: m.ctx.User},
					Body:      body,
					UpdatedAt: time.Now(),
				},
			},
		}
	})
}
