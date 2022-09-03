package context

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/config"
)

type State = int

const (
	TaskStart State = iota
	TaskFinished
	TaskError
)

type Task struct {
	Id           string
	StartText    string
	FinishedText string
	State        State
	Error        error
	StartTime    time.Time
}

type ProgramContext struct {
	User              string
	ScreenHeight      int
	ScreenWidth       int
	MainContentWidth  int
	MainContentHeight int
	Config            *config.Config
	View              config.ViewType
	Error             error
	StartTask         func(task Task) tea.Cmd
}

func (ctx *ProgramContext) GetViewSectionsConfig() []config.SectionConfig {
	var configs []config.SectionConfig
	if ctx.View == config.PRsView {
		configs = ctx.Config.PRSections
	} else {
		configs = ctx.Config.IssuesSections
	}

	return append([]config.SectionConfig{{Title: ""}}, configs...)
}
