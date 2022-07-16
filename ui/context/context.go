package context

import (
	"github.com/dlvhdr/gh-dash/config"
)

type ProgramContext struct {
	ScreenHeight      int
	ScreenWidth       int
	MainContentWidth  int
	MainContentHeight int
	Config            *config.Config
	View              config.ViewType
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
