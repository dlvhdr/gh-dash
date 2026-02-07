package tasks

import (
	"fmt"
	"os/exec"
	"time"

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

func AssignIssue(ctx *context.ProgramContext, section SectionIdentifier, issue data.RowData, usernames []string) tea.Cmd {
	issueNumber := issue.GetNumber()
	args := []string{
		"issue",
		"edit",
		fmt.Sprint(issueNumber),
		"-R",
		issue.GetRepoNameWithOwner(),
	}
	for _, assignee := range usernames {
		args = append(args, "--add-assignee", assignee)
	}
	return fireTask(ctx, GitHubTask{
		Id:           fmt.Sprintf("issue_assign_%d", issueNumber),
		Args:         args,
		Section:      section,
		StartText:    fmt.Sprintf("Assigning issue #%d to %s", issueNumber, usernames),
		FinishedText: fmt.Sprintf("Issue #%d has been assigned to %s", issueNumber, usernames),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			returnedAssignees := data.Assignees{Nodes: []data.Assignee{}}
			for _, assignee := range usernames {
				returnedAssignees.Nodes = append(returnedAssignees.Nodes, data.Assignee{Login: assignee})
			}
			return UpdateIssueMsg{
				IssueNumber:    issueNumber,
				AddedAssignees: &returnedAssignees,
			}
		},
	})
}

func UnassignIssue(ctx *context.ProgramContext, section SectionIdentifier, issue data.RowData, usernames []string) tea.Cmd {
	issueNumber := issue.GetNumber()
	args := []string{
		"issue",
		"edit",
		fmt.Sprint(issueNumber),
		"-R",
		issue.GetRepoNameWithOwner(),
	}
	for _, assignee := range usernames {
		args = append(args, "--remove-assignee", assignee)
	}
	return fireTask(ctx, GitHubTask{
		Id:           fmt.Sprintf("issue_unassign_%d", issueNumber),
		Args:         args,
		Section:      section,
		StartText:    fmt.Sprintf("Unassigning %s from issue #%d", usernames, issueNumber),
		FinishedText: fmt.Sprintf("%s unassigned from issue #%d", usernames, issueNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			returnedAssignees := data.Assignees{Nodes: []data.Assignee{}}
			for _, assignee := range usernames {
				returnedAssignees.Nodes = append(returnedAssignees.Nodes, data.Assignee{Login: assignee})
			}
			return UpdateIssueMsg{
				IssueNumber:      issueNumber,
				RemovedAssignees: &returnedAssignees,
			}
		},
	})
}

func CommentOnIssue(ctx *context.ProgramContext, section SectionIdentifier, issue data.RowData, body string) tea.Cmd {
	issueNumber := issue.GetNumber()
	return fireTask(ctx, GitHubTask{
		Id: fmt.Sprintf("issue_comment_%d", issueNumber),
		Args: []string{
			"issue",
			"comment",
			fmt.Sprint(issueNumber),
			"-R",
			issue.GetRepoNameWithOwner(),
			"-b",
			body,
		},
		Section:      section,
		StartText:    fmt.Sprintf("Commenting on issue #%d", issueNumber),
		FinishedText: fmt.Sprintf("Commented on issue #%d", issueNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdateIssueMsg{
				IssueNumber: issueNumber,
				NewComment: &data.IssueComment{
					Author:    struct{ Login string }{Login: ctx.User},
					Body:      body,
					UpdatedAt: time.Now(),
				},
			}
		},
	})
}

func LabelIssue(ctx *context.ProgramContext, section SectionIdentifier, issue data.RowData, labels []string, existingLabels []data.Label) tea.Cmd {
	issueNumber := issue.GetNumber()
	args := []string{
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

	existingLabelsColorMap := make(map[string]string)
	for _, label := range existingLabels {
		existingLabelsColorMap[label.Name] = label.Color
	}

	for _, label := range existingLabels {
		if _, ok := labelsMap[label.Name]; !ok {
			args = append(args, "--remove-label", label.Name)
		}
	}

	for _, label := range labels {
		args = append(args, "--add-label", label)
	}

	return fireTask(ctx, GitHubTask{
		Id:           fmt.Sprintf("issue_label_%d", issueNumber),
		Args:         args,
		Section:      section,
		StartText:    fmt.Sprintf("Labeling issue #%d to %s", issueNumber, labels),
		FinishedText: fmt.Sprintf("Issue #%d has been labeled with %s", issueNumber, labels),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			returnedLabels := data.IssueLabels{Nodes: []data.Label{}}
			for _, label := range labels {
				returnedLabels.Nodes = append(returnedLabels.Nodes, data.Label{
					Name:  label,
					Color: existingLabelsColorMap[label],
				})
			}
			return UpdateIssueMsg{
				IssueNumber: issueNumber,
				Labels:      &returnedLabels,
			}
		},
	})
}
