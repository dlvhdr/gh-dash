package debug

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/dlvhdr/gh-dash/ui/constants"
)

func LogMsg(msg tea.Msg) {
	switch msg := msg.(type) {
	case constants.TaskFinishedMsg, spinner.TickMsg:
		return
	default:
		log.Debug("Msg", msg)
	}
}
