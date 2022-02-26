package context

import "github.com/dlvhdr/gh-prs/config"

type ProgramContext struct {
	ScreenWidth       int
	MainContentWidth  int
	MainContentHeight int
	Config            config.Config
}
