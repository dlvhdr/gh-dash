package prssection

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/internal/provider"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

func (m Model) diff() tea.Cmd {
	currRowData := m.GetCurrRow()
	if currRowData == nil {
		return nil
	}

	var c *exec.Cmd
	if provider.IsGitLab() {
		// Use glab mr diff for GitLab
		c = exec.Command(
			"glab",
			"mr",
			"diff",
			fmt.Sprint(currRowData.GetNumber()),
			"--repo",
			m.GetCurrRow().GetRepoNameWithOwner(),
		)
	} else {
		// Use gh pr diff for GitHub
		c = exec.Command(
			"gh",
			"pr",
			"diff",
			fmt.Sprint(currRowData.GetNumber()),
			"-R",
			m.GetCurrRow().GetRepoNameWithOwner(),
		)
	}
	c.Env = m.Ctx.Config.GetFullScreenDiffPagerEnv()

	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return constants.ErrMsg{Err: err}
		}
		return nil
	})
}
