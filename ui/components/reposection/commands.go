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
		if len(b.Data.Remotes) == 0 {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: fmt.Errorf("No remotes found for branch %s", b.Data.Name)}
		}
		err := gitm.Push(*m.Ctx.RepoPath, b.Data.Remotes[0], b.Data.Name)
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
	err  error
}

func (m *Model) readRepoCmd() []tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	branchesTaskId := fmt.Sprintf("fetching_branches_%d", time.Now().Unix())
	if m.Ctx.RepoPath != nil {
		branchesTask := context.Task{
			Id:        branchesTaskId,
			StartText: "Reading local branches",
			FinishedText: fmt.Sprintf(
				`Read branches successfully for "%s"`,
				*m.Ctx.RepoPath,
			),
			State: context.TaskStart,
			Error: nil,
		}
		bCmd := m.Ctx.StartTask(branchesTask)
		cmds = append(cmds, bCmd)
	}
	cmds = append(cmds, func() tea.Msg {
		repo, err := git.GetRepo(*m.Ctx.RepoPath)
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

func (m *Model) fetchPRsCmd() []tea.Cmd {
	prsTaskId := fmt.Sprintf("fetching_pr_branches_%d", time.Now().Unix())
	cmds := make([]tea.Cmd, 0)
	task := context.Task{
		Id:           prsTaskId,
		StartText:    "Fetching PRs for your branches",
		FinishedText: "PRs for your branches have been fetched",
		State:        context.TaskStart,
		Error:        nil,
	}
	cmds = append(cmds, m.Ctx.StartTask(task))
	cmds = append(cmds, func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.PrsLimit
		}
		res, err := data.FetchPullRequests("author:@me", *limit, nil)
		// TODO: enrich with branches only for section with branches
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
	return cmds
}
