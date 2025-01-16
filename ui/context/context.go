package context

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/config"
	"github.com/dlvhdr/gh-dash/v4/ui/theme"
	"github.com/dlvhdr/gh-dash/v4/utils"
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
	RepoPath          string
	RepoUrl           string
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
	switch ctx.View {
	case config.RepoView:
		t := config.RepoView
		configs = append(configs, config.PrsSectionConfig{
			Title:   "Local Branches",
			Filters: "author:@me is:open",
			Limit:   utils.IntPtr(20),
			Type:    &t,
		}.ToSectionConfig())
	case config.PRsView:
		for _, cfg := range ctx.Config.PRSections {
			configs = append(configs, cfg.ToSectionConfig())
		}
	case config.IssuesView:
		for _, cfg := range ctx.Config.IssuesSections {
			configs = append(configs, cfg.ToSectionConfig())
		}
	}

	return append([]config.SectionConfig{{Title: ""}}, configs...)
}
