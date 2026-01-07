package prssection

import (
	"bytes"
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/gen2brain/beeep"

	"github.com/dlvhdr/gh-dash/v4/internal/provider"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func (m *Model) watchChecks() tea.Cmd {
	pr := m.GetCurrRow()
	if pr == nil {
		return nil
	}

	prNumber := pr.GetNumber()
	title := pr.GetTitle()
	repoNameWithOwner := pr.GetRepoNameWithOwner()
	prData := pr.(*prrow.Data)
	taskId := fmt.Sprintf("pr_reopen_%d", prNumber)

	label := "PR"
	if provider.IsGitLab() {
		label = "MR"
	}

	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Watching checks for %s #%d", label, prNumber),
		FinishedText: fmt.Sprintf("Watching checks for %s #%d", label, prNumber),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		var c *exec.Cmd
		if provider.IsGitLab() {
			// GitLab: use glab ci status to watch pipeline
			c = exec.Command(
				"glab",
				"ci",
				"status",
				"--live",
				"--repo",
				repoNameWithOwner,
				"--branch",
				prData.Primary.HeadRefName,
			)
		} else {
			c = exec.Command(
				"gh",
				"pr",
				"checks",
				"--watch",
				"--fail-fast",
				fmt.Sprint(prNumber),
				"-R",
				repoNameWithOwner,
			)
		}

		var outb, errb bytes.Buffer
		c.Stdout = &outb
		c.Stderr = &errb

		err := c.Start()
		go func() {
			err := c.Wait()
			if err != nil {
				log.Error("Error waiting for watch command to finish", "err", err,
					"stderr", errb.String(), "stdout", outb.String())
			}

			renderedPr := prrow.PullRequest{Ctx: m.Ctx, Data: prData}
			checksRollup := " Checks are pending"
			switch renderedPr.GetStatusChecksRollup() {
			case "SUCCESS":
				checksRollup = "✅ Checks have passed"
			case "FAILURE":
				checksRollup = "❌ Checks have failed"
			}

			err = beeep.Notify(
				fmt.Sprintf("gh-dash: %s", title),
				fmt.Sprintf("PR #%d in %s\n%s", prNumber, repoNameWithOwner, checksRollup),
				"",
			)
			if err != nil {
				log.Error("Error showing system notification", "err", err)
			}
		}()

		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: tasks.UpdatePRMsg{
				PrNumber: prNumber,
			},
		}
	})
}
