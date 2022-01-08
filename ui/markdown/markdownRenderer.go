package markdown

import "github.com/charmbracelet/glamour"

func GetMarkdownRenderer(width int) glamour.TermRenderer {
	markdownStyle := CustomDarkStyleConfig
	markdownRenderer, _ := glamour.NewTermRenderer(
		glamour.WithStyles(markdownStyle),
		glamour.WithWordWrap(width),
	)

	return *markdownRenderer
}
