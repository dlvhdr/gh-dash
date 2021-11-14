package msgs

import (
	"dlvhdr/gh-prs/config"

	tea "github.com/charmbracelet/bubbletea"
)

type InitMsg struct {
	Config []config.Section
}

type TickMsg struct {
	SectionId       int
	InternalTickMsg tea.Msg
}

type ErrMsg struct {
	error
}

func (e ErrMsg) Error() string { return e.error.Error() }
