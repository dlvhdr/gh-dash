package ui

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh/v2/pkg/browser"

	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

func (m *Model) openBrowser() tea.Cmd {
	taskId := fmt.Sprintf("open_browser_%d", time.Now().Unix())
	task := context.Task{
		Id:           taskId,
		StartText:    "Opening in browser",
		FinishedText: "Opened in browser",
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.ctx.StartTask(task)
	openCmd := func() tea.Msg {
		b := browser.New("", os.Stdout, os.Stdin)
		currRow := m.getCurrRowData()
		if reflect.ValueOf(currRow).IsNil() {
			return constants.TaskFinishedMsg{TaskId: taskId, Err: errors.New("Current selection doesn't have a URL")}
		}
		err := b.Browse(currRow.GetUrl())
		return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
	}
	return tea.Batch(startCmd, openCmd)
}
