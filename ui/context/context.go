package context

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/ui/theme"
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
	ConfigPath        string
	View              config.ViewType
	Error             error
	StartTask         func(task Task) tea.Cmd
	Theme             theme.Theme
	Styles            Styles
}

func (ctx *ProgramContext) GetViewSectionsConfig() []config.SectionConfig {
	var configs []config.SectionConfig
	if ctx.View == config.PRsView {
		for _, cfg := range ctx.Config.PRSections {
			configs = append(configs, cfg.ToSectionConfig())
		}
	} else {
		for _, cfg := range ctx.Config.IssuesSections {
			configs = append(configs, cfg.ToSectionConfig())
		}
	}

	return append([]config.SectionConfig{{Title: ""}}, configs...)
}
