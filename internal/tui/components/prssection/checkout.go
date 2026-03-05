package prssection

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func (m *Model) checkout() (tea.Cmd, error) {
	row := m.GetCurrRow()
	if row == nil {
		return nil, errors.New("no pr selected")
	}

	pr, ok := row.(*prrow.Data)
	if !ok {
		return nil, errors.New("unexpected row data type")
	}

	return common.CheckoutPR(common.CheckoutParams{
		RepoPaths:     m.Ctx.Config.RepoPaths,
		WorktreePaths: m.Ctx.Config.WorktreePaths,
		StartTask: func(ct common.CheckoutTask) tea.Cmd {
			return m.Ctx.StartTask(context.Task{
				Id:           ct.Id,
				StartText:    ct.StartText,
				FinishedText: ct.FinishedText,
				State:        context.TaskStart,
			})
		},
		PRNumber:   pr.GetNumber(),
		RepoName:   pr.GetRepoNameWithOwner(),
		BranchName: pr.Primary.HeadRefName,
	})
}
