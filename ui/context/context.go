package context

import "github.com/dlvhdr/gh-prs/config"

type ProgramContext struct {
	ScreenHeight      int
	ScreenWidth       int
	MainContentWidth  int
	MainContentHeight int
	Config            *config.Config
}
