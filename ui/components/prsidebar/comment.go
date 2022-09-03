package prsidebar

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/prssection"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/styles"
)

func (m *Model) renderCommentBox() string {
	return lipgloss.NewStyle().
		Width(m.width).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.DefaultTheme.SecondaryBorder).
		MarginTop(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				fmt.Sprint("Leave a comment...\n"),
				m.textArea.View(),
				lipgloss.NewStyle().
					MarginTop(1).
					Render(m.commentHelp.ShortHelpView(commentKeys)),
			),
		)

}

var commentKeys = []key.Binding{
	key.NewBinding(key.WithKeys(tea.KeyCtrlD.String()), key.WithHelp("Ctrl+d", "submit")),
	key.NewBinding(key.WithKeys(tea.KeyCtrlC.String(), tea.KeyEsc.String()), key.WithHelp("Ctrl+c/esc", "cancel")),
}

func (m *Model) comment(body string) tea.Cmd {
	pr := m.pr.Data
	prNumber := pr.GetNumber()
	taskId := fmt.Sprintf("comment_%d", prNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Commenting on PR #%d", prNumber),
		FinishedText: fmt.Sprintf("Commented on PR #%d", prNumber),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		c := exec.Command(
			"gh",
			"pr",
			"comment",
			fmt.Sprint(prNumber),
			"-R",
			pr.GetRepoNameWithOwner(),
			"-b",
			body,
		)

		err := c.Run()
		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: prssection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: prssection.UpdatePRMsg{
				PrNumber: prNumber,
				NewComment: &data.Comment{
					Author:    struct{ Login string }{Login: m.ctx.User},
					Body:      body,
					UpdatedAt: time.Now(),
				},
			},
		}
	})
}
