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

	if ctx.HasDarkBackground {
		markdownStyle = &CustomDarkStyleConfig
	} else {
		markdownStyle = &styles.LightStyleConfig
	}
	markdownStyleSource = ctx.BackgroundSource

	log.Debugf(
		"InitializeMarkdownStyle: assigned ctx.hasDarkBackground=%t, markdownStyleSource=%q",
		ctx.HasDarkBackground,
		markdownStyleSource,
	)
}

func GetMarkdownRenderer(width int, ctx *context.ProgramContext) glamour.TermRenderer {
	if ctx == nil {
		// If context is nil, use a default renderer without custom styling
		markdownRenderer, err := glamour.NewTermRenderer(
			glamour.WithWordWrap(width),
		)
		if err != nil || markdownRenderer == nil {
			fallback, _ := glamour.NewTermRenderer()
			if fallback != nil {
				return *fallback
			}
			panic("failed to create markdown renderer: " + err.Error())
		}
		return *markdownRenderer
	}
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
		panic("failed to create markdown renderer: " + err.Error())
	}

	return *markdownRenderer
}
