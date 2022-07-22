package markdown

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
)

var markdownStyle *ansi.StyleConfig

func InitializeMarkdownStyle(hasDarkBackground bool) {
	if markdownStyle != nil {
		return
	}
	if hasDarkBackground {
		markdownStyle = &CustomDarkStyleConfig
	} else {
		markdownStyle = &glamour.LightStyleConfig
	}
}

func GetMarkdownRenderer(width int) glamour.TermRenderer {
	markdownRenderer, _ := glamour.NewTermRenderer(
		glamour.WithStyles(*markdownStyle),
		glamour.WithWordWrap(width),
	)

	return *markdownRenderer
}
