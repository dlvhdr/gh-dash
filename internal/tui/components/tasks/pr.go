package tasks

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type SectionIdentifier struct {
	Id   int
	Type string
}

type UpdatePRMsg struct {
	PrNumber         int
	IsClosed         *bool
	NewComment       *data.Comment
	ReadyForReview   *bool
	IsMerged         *bool
	AddedAssignees   *data.Assignees
	RemovedAssignees *data.Assignees
}

type UpdateBranchMsg struct {
	Name      string
	IsCreated *bool
	NewPr     *data.PullRequestData
}

func buildTaskId(prefix string, prNumber int) string {
	return fmt.Sprintf("%s_%d", prefix, prNumber)
}

type GitHubTask struct {
	Id           string
	Args         []string
	Section      SectionIdentifier
	StartText    string
	FinishedText string
	Msg          func(c *exec.Cmd, err error) tea.Msg
}

func fireTask(ctx *context.ProgramContext, task GitHubTask) tea.Cmd {
	start := context.Task{
		Id:           task.Id,
		StartText:    task.StartText,
		FinishedText: task.FinishedText,
		State:        context.TaskStart,
		Error:        nil,
	}

	startCmd := ctx.StartTask(start)
	return tea.Batch(startCmd, func() tea.Msg {
		log.Info("Running task", "cmd", "gh "+strings.Join(task.Args, " "))
		c := exec.Command("gh", task.Args...)

		err := c.Run()
		return constants.TaskFinishedMsg{
			TaskId:      task.Id,
			SectionId:   task.Section.Id,
			SectionType: task.Section.Type,
			Err:         err,
			Msg:         task.Msg(c, err),
		}
	})
}

func OpenBranchPR(ctx *context.ProgramContext, section SectionIdentifier, branch string) tea.Cmd {
	return fireTask(ctx, GitHubTask{
		Id: fmt.Sprintf("branch_open_%s", branch),
		Args: []string{
			"pr",
			"view",
			"--web",
			branch,
			"-R",
			ctx.RepoUrl,
		},
		Section:      section,
		StartText:    fmt.Sprintf("Opening PR for branch %s", branch),
		FinishedText: fmt.Sprintf("PR for branch %s has been opened", branch),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdatePRMsg{}
		},
	})
}

func ReopenPR(ctx *context.ProgramContext, section SectionIdentifier, pr data.RowData) tea.Cmd {
	prNumber := pr.GetNumber()
	return fireTask(ctx, GitHubTask{
		Id: buildTaskId("pr_reopen", prNumber),
		Args: []string{
			"pr",
			"reopen",
			fmt.Sprint(prNumber),
			"-R",
			pr.GetRepoNameWithOwner(),
		},
		Section:      section,
		StartText:    fmt.Sprintf("Reopening PR #%d", prNumber),
		FinishedText: fmt.Sprintf("PR #%d has been reopened", prNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdatePRMsg{
				PrNumber: prNumber,
				IsClosed: utils.BoolPtr(false),
			}
		},
	})
}

func ClosePR(ctx *context.ProgramContext, section SectionIdentifier, pr data.RowData) tea.Cmd {
	prNumber := pr.GetNumber()
	return fireTask(ctx, GitHubTask{
		Id: buildTaskId("pr_close", prNumber),
		Args: []string{
			"pr",
			"close",
			fmt.Sprint(prNumber),
			"-R",
			pr.GetRepoNameWithOwner(),
		},
		Section:      section,
		StartText:    fmt.Sprintf("Closing PR #%d", prNumber),
		FinishedText: fmt.Sprintf("PR #%d has been closed", prNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdatePRMsg{
				PrNumber: prNumber,
				IsClosed: utils.BoolPtr(true),
			}
		},
	})
}

func PRReady(ctx *context.ProgramContext, section SectionIdentifier, pr data.RowData) tea.Cmd {
	prNumber := pr.GetNumber()
	return fireTask(ctx, GitHubTask{
		Id: buildTaskId("pr_ready", prNumber),
		Args: []string{
			"pr",
			"ready",
			fmt.Sprint(prNumber),
			"-R",
			pr.GetRepoNameWithOwner(),
		},
		Section:      section,
		StartText:    fmt.Sprintf("Marking PR #%d as ready for review", prNumber),
		FinishedText: fmt.Sprintf("PR #%d has been marked as ready for review", prNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdatePRMsg{
				PrNumber:       prNumber,
				ReadyForReview: utils.BoolPtr(true),
			}
		},
	})
}

func MergePR(ctx *context.ProgramContext, section SectionIdentifier, pr data.RowData) tea.Cmd {
	prNumber := pr.GetNumber()
	c := exec.Command(
		"gh",
		"pr",
		"merge",
		fmt.Sprint(prNumber),
		"-R",
		pr.GetRepoNameWithOwner(),
	)

	taskId := fmt.Sprintf("merge_%d", prNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Merging PR #%d", prNumber),
		FinishedText: fmt.Sprintf("PR #%d has been merged", prNumber),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := ctx.StartTask(task)

	return tea.Batch(startCmd, tea.ExecProcess(c, func(err error) tea.Msg {
		isMerged := err == nil && c.ProcessState.ExitCode() == 0

		return constants.TaskFinishedMsg{
			SectionId:   section.Id,
			SectionType: section.Type,
			TaskId:      taskId,
			Err:         err,
			Msg: UpdatePRMsg{
				PrNumber: prNumber,
				IsMerged: &isMerged,
			},
		}
	}))
}

func CreatePR(ctx *context.ProgramContext, section SectionIdentifier, branchName string, title string) tea.Cmd {
	c := exec.Command(
		"gh",
		"pr",
		"create",
		"--title",
		title,
		"-R",
		ctx.RepoUrl,
	)

	taskId := fmt.Sprintf("create_pr_%s", title)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf(`Creating PR "%s"`, title),
		FinishedText: fmt.Sprintf(`PR "%s" has been created`, title),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := ctx.StartTask(task)

	return tea.Batch(startCmd, tea.ExecProcess(c, func(err error) tea.Msg {
		isCreated := err == nil && c.ProcessState.ExitCode() == 0

		return constants.TaskFinishedMsg{
			SectionId:   section.Id,
			SectionType: section.Type,
			TaskId:      taskId,
			Err:         nil,
			Msg:         UpdateBranchMsg{Name: branchName, IsCreated: &isCreated},
		}
	}))
}

func UpdatePR(ctx *context.ProgramContext, section SectionIdentifier, pr data.RowData) tea.Cmd {
	prNumber := pr.GetNumber()
	return fireTask(ctx, GitHubTask{
		Id: buildTaskId("pr_update", prNumber),
		Args: []string{
			"pr",
			"update-branch",
			fmt.Sprint(prNumber),
			"-R",
			pr.GetRepoNameWithOwner(),
		},
		Section:      section,
		StartText:    fmt.Sprintf("Updating PR #%d", prNumber),
		FinishedText: fmt.Sprintf("PR #%d has been updated", prNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdatePRMsg{
				PrNumber: prNumber,
				IsClosed: utils.BoolPtr(true),
			}
		},
	})
}

func AssignPR(ctx *context.ProgramContext, section SectionIdentifier, pr data.RowData, usernames []string) tea.Cmd {
	prNumber := pr.GetNumber()
	args := []string{
		"pr",
		"edit",
		fmt.Sprint(prNumber),
		"-R",
		pr.GetRepoNameWithOwner(),
	}
	for _, assignee := range usernames {
		args = append(args, "--add-assignee", assignee)
	}
	return fireTask(ctx, GitHubTask{
		Id:           buildTaskId("pr_assign", prNumber),
		Args:         args,
		Section:      section,
		StartText:    fmt.Sprintf("Assigning pr #%d to %s", prNumber, usernames),
		FinishedText: fmt.Sprintf("pr #%d has been assigned to %s", prNumber, usernames),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			returnedAssignees := data.Assignees{Nodes: []data.Assignee{}}
			for _, assignee := range usernames {
				returnedAssignees.Nodes = append(returnedAssignees.Nodes, data.Assignee{Login: assignee})
			}
			return UpdatePRMsg{
				PrNumber:       prNumber,
				AddedAssignees: &returnedAssignees,
			}
		},
	})
}

func UnassignPR(ctx *context.ProgramContext, section SectionIdentifier, pr data.RowData, usernames []string) tea.Cmd {
	prNumber := pr.GetNumber()
	args := []string{
		"pr",
		"edit",
		fmt.Sprint(prNumber),
		"-R",
		pr.GetRepoNameWithOwner(),
	}
	for _, assignee := range usernames {
		args = append(args, "--remove-assignee", assignee)
	}
	return fireTask(ctx, GitHubTask{
		Id:           buildTaskId("pr_unassign", prNumber),
		Args:         args,
		Section:      section,
		StartText:    fmt.Sprintf("Unassigning %s from pr #%d", usernames, prNumber),
		FinishedText: fmt.Sprintf("%s unassigned from pr #%d", usernames, prNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			returnedAssignees := data.Assignees{Nodes: []data.Assignee{}}
			for _, assignee := range usernames {
				returnedAssignees.Nodes = append(returnedAssignees.Nodes, data.Assignee{Login: assignee})
			}
			return UpdatePRMsg{
				PrNumber:         prNumber,
				RemovedAssignees: &returnedAssignees,
			}
		},
	})
}

func CommentOnPR(ctx *context.ProgramContext, section SectionIdentifier, pr data.RowData, body string) tea.Cmd {
	prNumber := pr.GetNumber()
	return fireTask(ctx, GitHubTask{
		Id: buildTaskId("pr_comment", prNumber),
		Args: []string{
			"pr",
			"comment",
			fmt.Sprint(prNumber),
			"-R",
			pr.GetRepoNameWithOwner(),
			"-b",
			body,
		},
		Section:      section,
		StartText:    fmt.Sprintf("Commenting on PR #%d", prNumber),
		FinishedText: fmt.Sprintf("Commented on PR #%d", prNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdatePRMsg{
				PrNumber: prNumber,
				NewComment: &data.Comment{
					Author:    struct{ Login string }{Login: ctx.User},
					Body:      body,
					UpdatedAt: time.Now(),
				},
			}
		},
	})
}

func ApprovePR(ctx *context.ProgramContext, section SectionIdentifier, pr data.RowData, comment string) tea.Cmd {
	prNumber := pr.GetNumber()
	args := []string{
		"pr",
		"review",
		"-R",
		pr.GetRepoNameWithOwner(),
		fmt.Sprint(prNumber),
		"--approve",
	}
	if comment != "" {
		args = append(args, "--body", comment)
	}
	return fireTask(ctx, GitHubTask{
		Id:           buildTaskId("pr_approve", prNumber),
		Args:         args,
		Section:      section,
		StartText:    fmt.Sprintf("Approving pr #%d", prNumber),
		FinishedText: fmt.Sprintf("pr #%d has been approved", prNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdatePRMsg{
				PrNumber: prNumber,
			}
		},
	})
}
