package prssection

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
)

func (m Model) diff() tea.Cmd {
	currRowData := m.GetCurrRow()
	if currRowData == nil {
		return nil
	}

	return common.DiffPR(currRowData.GetNumber(), currRowData.GetRepoNameWithOwner(), m.Ctx.Config.GetFullScreenDiffPagerEnv())
}
