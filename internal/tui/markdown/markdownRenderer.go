package markdown

import (
	"charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
	log "charm.land/log/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

var (
	markdownStyle       *ansi.StyleConfig
	markdownStyleSource string
)

func InitializeMarkdownStyle(ctx *context.ProgramContext) {
	if markdownStyle != nil && markdownStyleSource == "bubbletea" {
		log.Debugf("InitializeMarkdownStyle: keeping existing bubbletea style")
		return
	}

	hasDarkBackground := true
	backgroundSource := "default"
	if ctx != nil {
		hasDarkBackground = ctx.HasDarkBackground
		backgroundSource = ctx.BackgroundSource
	}

	if hasDarkBackground {
		markdownStyle = &CustomDarkStyleConfig
	} else {
		markdownStyle = &styles.LightStyleConfig
	}
	markdownStyleSource = backgroundSource

	log.Debugf(
		"InitializeMarkdownStyle: assigned ctx.hasDarkBackground=%t, markdownStyleSource=%q",
		hasDarkBackground,
		markdownStyleSource,
	)
}

func GetMarkdownRenderer(width int, ctx *context.ProgramContext) glamour.TermRenderer {
	if markdownStyle == nil {
		InitializeMarkdownStyle(ctx)
	}

	markdownRenderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(*markdownStyle),
		glamour.WithWordWrap(width),
	)
	if err != nil || markdownRenderer == nil {
		// Return a fallback renderer that just returns input unchanged
		fallback, _ := glamour.NewTermRenderer()
		if fallback != nil {
			return *fallback
		}
		// If even fallback fails, panic with helpful message
		msg := "unknown error"
		if err != nil {
			msg = err.Error()
		}
		panic("failed to create markdown renderer: " + msg)
	}

	return *markdownRenderer
}
