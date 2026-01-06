package markdown

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
)

var markdownStyle *ansi.StyleConfig

func InitializeMarkdownStyle(hasDarkBackground bool) {
	if markdownStyle != nil {
		return
	}
	if hasDarkBackground {
		markdownStyle = &CustomDarkStyleConfig
	} else {
		markdownStyle = &styles.LightStyleConfig
	}
}

func GetMarkdownRenderer(width int) glamour.TermRenderer {
	markdownRenderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(*markdownStyle),
		glamour.WithWordWrap(width),
	)
	if err != nil || markdownRenderer == nil {
		// Return a fallback renderer that just returns input unchanged
		fallback, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())
		if fallback != nil {
			return *fallback
		}
		// If even fallback fails, panic with helpful message
		panic("failed to create markdown renderer: " + err.Error())
	}

	return *markdownRenderer
}
