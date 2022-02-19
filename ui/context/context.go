package context

import "github.com/dlvhdr/gh-prs/config"

type ProgramContext struct {
	ScreenWidth       int
	MainViewportWidth int
	Config            config.Config
}
