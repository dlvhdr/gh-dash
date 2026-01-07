package common

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/provider"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

// DiffPR opens a diff view for a PR/MR using the gh or glab CLI.
// The env parameter should be the result of Config.GetFullScreenDiffPagerEnv().
func DiffPR(prNumber int, repoName string, env []string) tea.Cmd {
	var c *exec.Cmd
	if provider.IsGitLab() {
		c = exec.Command(
			"glab",
			"mr",
			"diff",
			fmt.Sprint(prNumber),
			"--repo",
			repoName,
		)
	} else {
		c = exec.Command(
			"gh",
			"pr",
			"diff",
			fmt.Sprint(prNumber),
			"-R",
			repoName,
		)
	}
	c.Env = env

	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return constants.ErrMsg{Err: err}
		}
		return nil
	})
}
