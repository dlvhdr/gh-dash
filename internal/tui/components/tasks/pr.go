package tasks

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/provider"
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

// prSubCmd returns "mr" for GitLab, "pr" for GitHub.
func prSubCmd() string {
	if provider.IsGitLab() {
		return "mr"
	}
	return "pr"
}

// prLabel returns "MR" for GitLab, "PR" for GitHub.
func prLabel() string {
	if provider.IsGitLab() {
		return "MR"
	}
	return "PR"
}

// repoFlag returns "--repo" for GitLab (glab), "-R" for GitHub (gh).
func repoFlag() string {
	if provider.IsGitLab() {
		return "--repo"
	}
	return "-R"
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

	cliCmd := provider.GetCLICommand()
	startCmd := ctx.StartTask(start)
	return tea.Batch(startCmd, func() tea.Msg {
		log.Info("Running task", "cmd", cliCmd+" "+strings.Join(task.Args, " "))
		c := exec.Command(cliCmd, task.Args...)

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
	label := prLabel()
	return fireTask(ctx, GitHubTask{
		Id: fmt.Sprintf("branch_open_%s", branch),
		Args: []string{
			prSubCmd(),
			"view",
			"--web",
			branch,
			repoFlag(),
			ctx.RepoUrl,
		},
		Section:      section,
		StartText:    fmt.Sprintf("Opening %s for branch %s", label, branch),
		FinishedText: fmt.Sprintf("%s for branch %s has been opened", label, branch),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdatePRMsg{}
		},
	})
}

func ReopenPR(ctx *context.ProgramContext, section SectionIdentifier, pr data.RowData) tea.Cmd {
	prNumber := pr.GetNumber()
	label := prLabel()
	return fireTask(ctx, GitHubTask{
		Id: buildTaskId("pr_reopen", prNumber),
		Args: []string{
			prSubCmd(),
			"reopen",
			fmt.Sprint(prNumber),
			repoFlag(),
			pr.GetRepoNameWithOwner(),
		},
		Section:      section,
		StartText:    fmt.Sprintf("Reopening %s #%d", label, prNumber),
		FinishedText: fmt.Sprintf("%s #%d has been reopened", label, prNumber),
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
	label := prLabel()
	return fireTask(ctx, GitHubTask{
		Id: buildTaskId("pr_close", prNumber),
		Args: []string{
			prSubCmd(),
			"close",
			fmt.Sprint(prNumber),
			repoFlag(),
			pr.GetRepoNameWithOwner(),
		},
		Section:      section,
		StartText:    fmt.Sprintf("Closing %s #%d", label, prNumber),
		FinishedText: fmt.Sprintf("%s #%d has been closed", label, prNumber),
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
	label := prLabel()

	// GitLab uses 'mr update --ready', GitHub uses 'pr ready'
	var args []string
	if provider.IsGitLab() {
		args = []string{
			"mr",
			"update",
			fmt.Sprint(prNumber),
			"--repo",
			pr.GetRepoNameWithOwner(),
			"--ready",
		}
	} else {
		args = []string{
			"pr",
			"ready",
			fmt.Sprint(prNumber),
			"-R",
			pr.GetRepoNameWithOwner(),
		}
	}

	return fireTask(ctx, GitHubTask{
		Id:           buildTaskId("pr_ready", prNumber),
		Args:         args,
		Section:      section,
		StartText:    fmt.Sprintf("Marking %s #%d as ready for review", label, prNumber),
		FinishedText: fmt.Sprintf("%s #%d has been marked as ready for review", label, prNumber),
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
	label := prLabel()
	c := exec.Command(
		provider.GetCLICommand(),
		prSubCmd(),
		"merge",
		fmt.Sprint(prNumber),
		repoFlag(),
		pr.GetRepoNameWithOwner(),
	)

	taskId := fmt.Sprintf("merge_%d", prNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Merging %s #%d", label, prNumber),
		FinishedText: fmt.Sprintf("%s #%d has been merged", label, prNumber),
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
		provider.GetCLICommand(),
		prSubCmd(),
		"create",
		"--title",
		title,
		repoFlag(),
		ctx.RepoUrl,
	)

	label := prLabel()
	taskId := fmt.Sprintf("create_pr_%s", title)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf(`Creating %s "%s"`, label, title),
		FinishedText: fmt.Sprintf(`%s "%s" has been created`, label, title),
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
	label := prLabel()

	// GitLab uses 'mr rebase', GitHub uses 'pr update-branch'
	var args []string
	if provider.IsGitLab() {
		args = []string{
			"mr",
			"rebase",
			fmt.Sprint(prNumber),
			"--repo",
			pr.GetRepoNameWithOwner(),
		}
	} else {
		args = []string{
			"pr",
			"update-branch",
			fmt.Sprint(prNumber),
			"-R",
			pr.GetRepoNameWithOwner(),
		}
	}

	return fireTask(ctx, GitHubTask{
		Id:           buildTaskId("pr_update", prNumber),
		Args:         args,
		Section:      section,
		StartText:    fmt.Sprintf("Updating %s #%d", label, prNumber),
		FinishedText: fmt.Sprintf("%s #%d has been updated", label, prNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdatePRMsg{
				PrNumber: prNumber,
			}
		},
	})
}

func AssignPR(ctx *context.ProgramContext, section SectionIdentifier, pr data.RowData, usernames []string) tea.Cmd {
	prNumber := pr.GetNumber()
	label := prLabel()
	var args []string
	if provider.IsGitLab() {
		args = []string{
			"mr",
			"update",
			fmt.Sprint(prNumber),
			"--repo",
			pr.GetRepoNameWithOwner(),
			"--assignee",
			strings.Join(usernames, ","),
		}
	} else {
		args = []string{
			"pr",
			"edit",
			fmt.Sprint(prNumber),
			"-R",
			pr.GetRepoNameWithOwner(),
		}
		for _, assignee := range usernames {
			args = append(args, "--add-assignee", assignee)
		}
	}
	return fireTask(ctx, GitHubTask{
		Id:           buildTaskId("pr_assign", prNumber),
		Args:         args,
		Section:      section,
		StartText:    fmt.Sprintf("Assigning %s #%d to %s", label, prNumber, usernames),
		FinishedText: fmt.Sprintf("%s #%d has been assigned to %s", label, prNumber, usernames),
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
	label := prLabel()
	var args []string
	if provider.IsGitLab() {
		args = []string{
			"mr",
			"update",
			fmt.Sprint(prNumber),
			"--repo",
			pr.GetRepoNameWithOwner(),
			"--unassign",
		}
	} else {
		args = []string{
			"pr",
			"edit",
			fmt.Sprint(prNumber),
			"-R",
			pr.GetRepoNameWithOwner(),
		}
		for _, assignee := range usernames {
			args = append(args, "--remove-assignee", assignee)
		}
	}
	return fireTask(ctx, GitHubTask{
		Id:           buildTaskId("pr_unassign", prNumber),
		Args:         args,
		Section:      section,
		StartText:    fmt.Sprintf("Unassigning %s from %s #%d", usernames, label, prNumber),
		FinishedText: fmt.Sprintf("%s unassigned from %s #%d", usernames, label, prNumber),
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
	label := prLabel()
	var args []string
	if provider.IsGitLab() {
		args = []string{
			"mr",
			"note",
			fmt.Sprint(prNumber),
			"--repo",
			pr.GetRepoNameWithOwner(),
			"-m",
			body,
		}
	} else {
		args = []string{
			"pr",
			"comment",
			fmt.Sprint(prNumber),
			"-R",
			pr.GetRepoNameWithOwner(),
			"-b",
			body,
		}
	}
	return fireTask(ctx, GitHubTask{
		Id:           buildTaskId("pr_comment", prNumber),
		Args:         args,
		Section:      section,
		StartText:    fmt.Sprintf("Commenting on %s #%d", label, prNumber),
		FinishedText: fmt.Sprintf("Commented on %s #%d", label, prNumber),
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
	label := prLabel()
	var args []string
	if provider.IsGitLab() {
		args = []string{
			"mr",
			"approve",
			fmt.Sprint(prNumber),
			"--repo",
			pr.GetRepoNameWithOwner(),
		}
	} else {
		args = []string{
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
	}
	return fireTask(ctx, GitHubTask{
		Id:           buildTaskId("pr_approve", prNumber),
		Args:         args,
		Section:      section,
		StartText:    fmt.Sprintf("Approving %s #%d", label, prNumber),
		FinishedText: fmt.Sprintf("%s #%d has been approved", label, prNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdatePRMsg{
				PrNumber: prNumber,
			}
		},
	})
}
