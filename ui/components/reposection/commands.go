package reposection

import (
	"fmt"
	"sync"
	"time"

	gitm "github.com/aymanbagabas/git-module"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/git"
	"github.com/dlvhdr/gh-dash/v4/ui/components/tasks"
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
		repo, err := git.GetRepo(m.Ctx.RepoPath)
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
		repo, err = git.GetRepo(m.Ctx.RepoPath)
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

type pushOptions struct {
	force bool
}

func (m *Model) push(opts pushOptions) (tea.Cmd, error) {
	b := m.getCurrBranch()

	taskId := fmt.Sprintf("push_%s_%d", b.Data.Name, time.Now().Unix())
	withForceText := func() string {
		if opts.force {
			return " with force"
		}
		return ""
	}
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Pushing branch %s%s", b.Data.Name, withForceText()),
		FinishedText: fmt.Sprintf("Branch %s has been pushed%s", b.Data.Name, withForceText()),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		var err error
		args := []string{}
		if opts.force {
			args = append(args, "--force")
		}
		if len(b.Data.Remotes) == 0 {
			args = append(args, "--set-upstream")
			err = gitm.Push(
				m.Ctx.RepoPath,
				"origin",
				b.Data.Name,
				gitm.PushOptions{CommandOptions: gitm.CommandOptions{Args: args}},
			)
		} else {
			err = gitm.Push(
				m.Ctx.RepoPath,
				b.Data.Remotes[0],
				b.Data.Name,
				gitm.PushOptions{CommandOptions: gitm.CommandOptions{Args: args}},
			)
		}
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
		}
		repo, err := git.GetRepo(m.Ctx.RepoPath)
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
		err := gitm.Checkout(m.Ctx.RepoPath, b.Data.Name)
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
		}
		repo, err := git.GetRepo(m.Ctx.RepoPath)
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
		}

		return constants.TaskFinishedMsg{
			SectionId:   0,
			SectionType: SectionType,
			TaskId:      taskId,
			Msg:         repoMsg{repo: repo, resetSelection: true},
			Err:         err,
		}
	}), nil
}

type repoMsg struct {
	repo           *git.Repo
	resetSelection bool
}

func (m *Model) readRepoCmd() []tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	branchesTaskId := fmt.Sprintf("fetching_branches_%d", time.Now().Unix())
	if m.Ctx.RepoPath != "" {
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
		repo, err := git.GetRepo(m.Ctx.RepoPath)
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
	if m.Ctx.RepoPath == "" {
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
		repo, err := git.FetchRepo(m.Ctx.RepoPath)
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
		StartText:    "Fetching PRs",
		FinishedText: "PRs fetched",
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.PrsLimit
		}
		res, err := data.FetchPullRequests(fmt.Sprintf("author:@me repo:%s", git.GetRepoShortName(m.Ctx.RepoUrl)), *limit, nil)
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

func (m *Model) fetchPRCmd(branch string) []tea.Cmd {
	prsTaskId := fmt.Sprintf("fetching_pr_for_branch_%s_%d", branch, time.Now().Unix())
	task := context.Task{
		Id:           prsTaskId,
		StartText:    fmt.Sprintf("Fetching PR for branch %s", branch),
		FinishedText: "PR fetched",
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return []tea.Cmd{startCmd, func() tea.Msg {
		res, err := data.FetchPullRequests(fmt.Sprintf("author:@me repo:%s head:%s", git.GetRepoShortName(m.Ctx.RepoUrl), branch), 1, nil)
		log.Debug("Fetching PRs", "res", res)
		if err != nil {
			return constants.TaskFinishedMsg{
				SectionId:   0,
				SectionType: SectionType,
				TaskId:      prsTaskId,
				Err:         err,
			}
		}

		if len(res.Prs) != 1 {
			return constants.TaskFinishedMsg{
				SectionId:   0,
				SectionType: SectionType,
				TaskId:      prsTaskId,
				Err:         fmt.Errorf("expected 1 PR, got %d", len(res.Prs)),
			}
		}

		return constants.TaskFinishedMsg{
			SectionId:   0,
			SectionType: SectionType,
			TaskId:      prsTaskId,
			Msg: tasks.UpdateBranchMsg{
				Name:  branch,
				NewPr: &res.Prs[0],
			},
		}
	}}
}

type RefreshBranchesMsg struct {
	id   int
	time time.Time
}

type RefreshPrsMsg struct {
	id   int
	time time.Time
}

var (
	lastID int
	idMtx  sync.Mutex
)

// Return the next ID we should use on the Model.
func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}

func (m *Model) tickRefreshBranchesCmd() tea.Cmd {
	return tea.Tick(time.Second*time.Duration(m.Ctx.Config.Repo.BranchesRefetchIntervalSeconds), func(t time.Time) tea.Msg {
		return RefreshBranchesMsg{id: m.refreshId, time: t}
	})
}

func (m *Model) tickFetchPrsCmd() tea.Cmd {
	return tea.Tick(time.Second*time.Duration(m.Ctx.Config.Repo.PrsRefetchIntervalSeconds), func(t time.Time) tea.Msg {
		return RefreshPrsMsg{id: m.refreshId, time: t}
	})
}

func (m *Model) onRefreshBranchesMsg() []tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	cmds = append(cmds, m.readRepoCmd()...)
	cmds = append(cmds, m.tickRefreshBranchesCmd())
	return cmds
}

func (m *Model) onRefreshPrsMsg() []tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	cmds = append(cmds, m.fetchRepoCmd()...)
	cmds = append(cmds, m.tickFetchPrsCmd())
	return cmds
}

func (m *Model) OpenGithub() tea.Cmd {
	row := m.CurrRow()
	b := m.getFilteredBranches()[row]
	return tasks.OpenBranchPR(m.Ctx, tasks.SectionIdentifier{Id: 0, Type: SectionType}, b.Data.Name)
}

func (m *Model) deleteBranch() tea.Cmd {
	b := m.getCurrBranch()

	taskId := fmt.Sprintf("delete_%s_%d", b.Data.Name, time.Now().Unix())
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Deleting branch %s", b.Data.Name),
		FinishedText: fmt.Sprintf("Branch %s has been deleted", b.Data.Name),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		err := gitm.DeleteBranch(m.Ctx.RepoPath, b.Data.Name, gitm.DeleteBranchOptions{Force: true})
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
		}
		repo, err := git.GetRepo(m.Ctx.RepoPath)
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
	})
}

func (m *Model) newBranch(name string) tea.Cmd {
	taskId := fmt.Sprintf("create_branch_%s_%d", name, time.Now().Unix())
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Creating branch %s", name),
		FinishedText: fmt.Sprintf("Branch %s has been created", name),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		err := gitm.Checkout(m.Ctx.RepoPath, name, gitm.CheckoutOptions{BaseBranch: m.repo.HeadBranchName})
		if err != nil {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
		}
		repo, err := git.GetRepo(m.Ctx.RepoPath)
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
	})
}
