package reposection

import (
	"fmt"
	"time"

	gitm "github.com/aymanbagabas/git-module"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/git"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

type UpdatePRMsg struct {
	PrNumber         int
	IsClosed         *bool
	NewComment       *data.Comment
	ReadyForReview   *bool
	IsMerged         *bool
	AddedAssignees   *data.Assignees
	RemovedAssignees *data.Assignees
}

func (m *Model) fastForward() (tea.Cmd, error) {
	b := m.getCurrBranch()

	taskId := fmt.Sprintf("fast-forward_%s_%d", b.Data.Name, time.Now().Unix())
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Fast-forwarding branch %s", b.Data.Name),
		FinishedText: fmt.Sprintf("Branch %s has been fast-forwarded", b.Data.Name),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		var err error
		repo, err := git.GetRepo(*m.Ctx.RepoPath)
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
		}

		if b.Data.IsCheckedOut {
			err = repo.Pull(gitm.PullOptions{
				All:            false,
				Remote:         "origin",
				Branch:         b.Data.Name,
				CommandOptions: gitm.CommandOptions{Args: []string{"--ff-only", "--no-edit"}},
			})
		} else {
			err = repo.Fetch(gitm.FetchOptions{CommandOptions: gitm.CommandOptions{Args: []string{
				"--no-write-fetch-head",
				"origin",
				b.Data.Name + ":" + b.Data.Name,
			}}})
		}
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
		}
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
		}

		return constants.TaskFinishedMsg{
			SectionId:   0,
			SectionType: SectionType,
			TaskId:      taskId,
			Msg:         repoMsg{repo: repo},
			Err:         err,
		}
	}), nil
}

func (m *Model) push() (tea.Cmd, error) {
	b := m.getCurrBranch()

	taskId := fmt.Sprintf("push_%s_%d", b.Data.Name, time.Now().Unix())
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Pushing branch %s", b.Data.Name),
		FinishedText: fmt.Sprintf("Branch %s has been pushed", b.Data.Name),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		var err error
		if len(b.Data.Remotes) == 0 {
			err = gitm.Push(*m.Ctx.RepoPath, "origin", b.Data.Name, gitm.PushOptions{CommandOptions: gitm.CommandOptions{Args: []string{"--set-upstream"}}})
		} else {
			err = gitm.Push(*m.Ctx.RepoPath, b.Data.Remotes[0], b.Data.Name)
		}
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
		}
		repo, err := git.GetRepo(*m.Ctx.RepoPath)
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
		}

		return constants.TaskFinishedMsg{
			SectionId:   0,
			SectionType: SectionType,
			TaskId:      taskId,
			Msg:         repoMsg{repo: repo},
			Err:         err,
		}
	}), nil
}

func (m *Model) checkout() (tea.Cmd, error) {
	b := m.getCurrBranch()

	taskId := fmt.Sprintf("checkout_%s_%d", b.Data.Name, time.Now().Unix())
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Checking out branch %s", b.Data.Name),
		FinishedText: fmt.Sprintf("Branch %s has been checked out", b.Data.Name),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		err := gitm.Checkout(*m.Ctx.RepoPath, b.Data.Name)
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
		}
		repo, err := git.GetRepo(*m.Ctx.RepoPath)
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
		}

		return constants.TaskFinishedMsg{
			SectionId:   0,
			SectionType: SectionType,
			TaskId:      taskId,
			Msg:         repoMsg{repo: repo},
			Err:         err,
		}
	}), nil
}

type repoMsg struct {
	repo *git.Repo
}

func (m *Model) readRepoCmd() []tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	branchesTaskId := fmt.Sprintf("fetching_branches_%d", time.Now().Unix())
	if m.Ctx.RepoPath != nil {
		branchesTask := context.Task{
			Id:           branchesTaskId,
			StartText:    "Reading local branches",
			FinishedText: "Branches read",
			State:        context.TaskStart,
			Error:        nil,
		}
		bCmd := m.Ctx.StartTask(branchesTask)
		cmds = append(cmds, bCmd)
	}
	cmds = append(cmds, func() tea.Msg {
		repo, err := git.GetRepo(*m.Ctx.RepoPath)
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: branchesTaskId, Err: err}
		}
		return constants.TaskFinishedMsg{
			SectionId:   0,
			SectionType: SectionType,
			TaskId:      branchesTaskId,
			Msg:         repoMsg{repo: repo},
			Err:         err,
		}
	})
	return cmds
}

func (m *Model) fetchRepoCmd() []tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	fetchTaskId := fmt.Sprintf("git_fetch_repo_%d", time.Now().Unix())
	if m.Ctx.RepoPath == nil {
		return []tea.Cmd{}
	}
	fetchTask := context.Task{
		Id:           fetchTaskId,
		StartText:    "Fetching branches from origin",
		FinishedText: "Fetched origin branches",
		State:        context.TaskStart,
		Error:        nil,
	}
	cmds = append(cmds, m.Ctx.StartTask(fetchTask))
	cmds = append(cmds, func() tea.Msg {
		repo, err := git.FetchRepo(*m.Ctx.RepoPath)
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: fetchTaskId, Err: err}
		}
		return constants.TaskFinishedMsg{
			SectionId:   0,
			SectionType: SectionType,
			TaskId:      fetchTaskId,
			Msg:         repoMsg{repo: repo},
			Err:         err,
		}
	})
	return cmds
}

func (m *Model) fetchPRsCmd() tea.Cmd {
	prsTaskId := fmt.Sprintf("fetching_pr_branches_%d", time.Now().Unix())
	task := context.Task{
		Id:           prsTaskId,
		StartText:    "Fetching PRs for your branches",
		FinishedText: "PRs for your branches have been fetched",
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.PrsLimit
		}
		res, err := data.FetchPullRequests(fmt.Sprintf("author:@me repo:%s", git.GetRepoShortName(*m.Ctx.RepoUrl)), *limit, nil)
		if err != nil {
			return constants.TaskFinishedMsg{
				SectionId:   0,
				SectionType: SectionType,
				TaskId:      prsTaskId,
				Err:         err,
			}
		}
		return constants.TaskFinishedMsg{
			SectionId:   0,
			SectionType: SectionType,
			TaskId:      prsTaskId,
			Msg: SectionPullRequestsFetchedMsg{
				Prs:        res.Prs,
				TotalCount: res.TotalCount,
				PageInfo:   res.PageInfo,
				TaskId:     prsTaskId,
			},
		}
	})
}

type RefreshBranchesMsg time.Time

type FetchMsg time.Time

const refreshIntervalSec = 30

const fetchIntervalSec = 60

func (m *Model) tickRefreshBranchesCmd() tea.Cmd {
	return tea.Tick(time.Second*refreshIntervalSec, func(t time.Time) tea.Msg {
		return RefreshBranchesMsg(t)
	})
}

func (m *Model) tickFetchCmd() tea.Cmd {
	return tea.Tick(time.Second*fetchIntervalSec, func(t time.Time) tea.Msg {
		return FetchMsg(t)
	})
}

func (m *Model) onRefreshBranchesMsg() []tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	cmds = append(cmds, m.readRepoCmd()...)
	cmds = append(cmds, m.tickRefreshBranchesCmd())
	return cmds
}

func (m *Model) onFetchMsg() []tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	cmds = append(cmds, m.fetchRepoCmd()...)
	cmds = append(cmds, m.tickRefreshBranchesCmd())
	return cmds
}
