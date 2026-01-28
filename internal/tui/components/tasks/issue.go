package tasks

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type UpdateIssueMsg struct {
	IssueNumber      int
	Labels           *data.IssueLabels
	NewComment       *data.IssueComment
	IsClosed         *bool
	AddedAssignees   *data.Assignees
	RemovedAssignees *data.Assignees
}

func CloseIssue(ctx *context.ProgramContext, section SectionIdentifier, issue data.RowData) tea.Cmd {
	issueNumber := issue.GetNumber()
	return fireTask(ctx, GitHubTask{
		Id: fmt.Sprintf("issue_close_%d", issueNumber),
		Args: []string{
			"issue",
			"close",
			fmt.Sprint(issueNumber),
			"-R",
			issue.GetRepoNameWithOwner(),
		},
		Section:      section,
		StartText:    fmt.Sprintf("Closing issue #%d", issueNumber),
		FinishedText: fmt.Sprintf("Issue #%d has been closed", issueNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdateIssueMsg{
				IssueNumber: issueNumber,
				IsClosed:    utils.BoolPtr(true),
			}
		},
	})
}

func ReopenIssue(ctx *context.ProgramContext, section SectionIdentifier, issue data.RowData) tea.Cmd {
	issueNumber := issue.GetNumber()
	return fireTask(ctx, GitHubTask{
		Id: fmt.Sprintf("issue_reopen_%d", issueNumber),
		Args: []string{
			"issue",
			"reopen",
			fmt.Sprint(issueNumber),
			"-R",
			issue.GetRepoNameWithOwner(),
		},
		Section:      section,
		StartText:    fmt.Sprintf("Reopening issue #%d", issueNumber),
		FinishedText: fmt.Sprintf("Issue #%d has been reopened", issueNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdateIssueMsg{
				IssueNumber: issueNumber,
				IsClosed:    utils.BoolPtr(false),
			}
		},
	})
}
