package prssection

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

func (m Model) diff() tea.Cmd {
	currRowData := m.GetCurrRow()
	if currRowData == nil {
		return nil
	}

	c := exec.Command(
		"gh",
		"pr",
		"diff",
		fmt.Sprint(currRowData.GetNumber()),
		"-R",
		m.GetCurrRow().GetRepoNameWithOwner(),
	)
	c.Env = m.Ctx.Config.GetFullScreenDiffPagerEnv()

	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return constants.ErrMsg{Err: err}
		}
		return nil
	})
}
