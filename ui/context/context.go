package context

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/config"
	"github.com/dlvhdr/gh-dash/v4/ui/theme"
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
	FinishedTime *time.Time
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
	log.Debug("View", "view", ctx.View)
	switch ctx.View {
	case config.PRsView:
		log.Debug("sections", "prs", ctx.Config.PRSections)
		for _, cfg := range ctx.Config.PRSections {
			configs = append(configs, cfg.ToSectionConfig())
		}
	case config.IssuesView:
		log.Debug("HelPPPPPPPPPPPPPPP", "config", ctx.Config, "view", ctx.View)
		log.Debug("sections", "issues", ctx.Config.IssuesSections)
		for _, cfg := range ctx.Config.IssuesSections {
			configs = append(configs, cfg.ToSectionConfig())
		}
	}

	return append([]config.SectionConfig{{Title: ""}}, configs...)
}
