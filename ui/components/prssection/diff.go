package prssection

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
)

func (m Model) diff() tea.Cmd {
	c := exec.Command(
		"gh",
		"pr",
		"diff",
		fmt.Sprint(m.GetCurrRow().GetNumber()),
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
