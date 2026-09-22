package common

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

// DiffPR opens a diff view for a PR using the configured pager.
// The env parameter should be the result of Config.GetFullScreenDiffPagerEnv().
func DiffPR(prNumber int, repoName string, env []string) tea.Cmd {
	return diffPR(prNumber, repoName, env, data.FetchPullRequestDiff)
}

type diffFetcher func(prNumber int, repoName string) (string, error)

func diffPR(prNumber int, repoName string, env []string, fetchDiff diffFetcher) tea.Cmd {
	return func() tea.Msg {
		diff, err := fetchDiff(prNumber, repoName)
		if err != nil {
			return constants.ErrMsg{Err: err}
		}

		c := pagerCommand(diffPager(env), diff, env)
		return tea.ExecProcess(c, func(err error) tea.Msg {
			if err != nil {
				return constants.ErrMsg{Err: err}
			}
			return nil
		})()
	}
}

func diffPager(env []string) string {
	for i := len(env) - 1; i >= 0; i-- {
		pager, ok := strings.CutPrefix(env[i], "GH_PAGER=")
		if ok && strings.TrimSpace(pager) != "" {
			return pager
		}
	}

	return "less"
}

func pagerCommand(pager string, diff string, env []string) *exec.Cmd {
	return pagerCommandForGOOS(runtime.GOOS, pager, diff, env)
}

func pagerCommandForGOOS(goos string, pager string, diff string, env []string) *exec.Cmd {
	var c *exec.Cmd
	if goos == "windows" {
		shell := os.Getenv("COMSPEC")
		if shell == "" {
			shell = "cmd"
		}
		c = exec.Command(shell, "/C", pager)
	} else {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "sh"
		}
		c = exec.Command(shell, "-c", pager)
	}
	c.Env = env
	c.Stdin = strings.NewReader(diff)

	return c
}
