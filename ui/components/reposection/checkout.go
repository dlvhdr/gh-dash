package reposection

import (
	"fmt"

	gitm "github.com/aymanbagabas/git-module"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/git"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

func (m *Model) checkout() (tea.Cmd, error) {
	b := m.GetCurrBranch()

	taskId := fmt.Sprintf("checkout_%s", b.Data.Name)
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
			SectionId:   m.Id,
			SectionType: m.Type,
			TaskId:      taskId,
			Msg:         repoMsg{repo: repo},
			Err:         err,
		}
	}), nil
}
