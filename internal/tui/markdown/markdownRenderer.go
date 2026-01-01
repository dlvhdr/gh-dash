package markdown

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
)

var markdownStyle *ansi.StyleConfig
var currentBgColor string

func InitializeMarkdownStyle(hasDarkBackground bool) {
	if hasDarkBackground {
		markdownStyle = &CustomDarkStyleConfig
	} else {
		markdownStyle = &styles.LightStyleConfig
	}
}

func SetBackgroundColor(bgColor string) {
	currentBgColor = bgColor
}

func GetMarkdownRenderer(width int) glamour.TermRenderer {
	style := *markdownStyle
	// Apply the current background color if set
	if currentBgColor != "" {
		bg := stringPtr(currentBgColor)
		// Document and block elements
		style.Document.StylePrimitive.BackgroundColor = bg
		style.Paragraph.StylePrimitive.BackgroundColor = bg
		style.BlockQuote.StylePrimitive.BackgroundColor = bg
		style.List.StylePrimitive.BackgroundColor = bg
		style.Item.BackgroundColor = bg
		style.Enumeration.BackgroundColor = bg

		// Headings
		style.Heading.StylePrimitive.BackgroundColor = bg
		style.H1.StylePrimitive.BackgroundColor = bg
		style.H2.StylePrimitive.BackgroundColor = bg
		style.H3.StylePrimitive.BackgroundColor = bg
		style.H4.StylePrimitive.BackgroundColor = bg
		style.H5.StylePrimitive.BackgroundColor = bg
		style.H6.StylePrimitive.BackgroundColor = bg

		// Text formatting
		style.Text.BackgroundColor = bg
		style.Emph.BackgroundColor = bg
		style.Strong.BackgroundColor = bg
		style.Strikethrough.BackgroundColor = bg

		// Code
		style.Code.StylePrimitive.BackgroundColor = bg

		// Links and images
		style.Link.BackgroundColor = bg
		style.LinkText.BackgroundColor = bg
		style.Image.BackgroundColor = bg
		style.ImageText.BackgroundColor = bg

		// Other elements
		style.HorizontalRule.BackgroundColor = bg
		style.Task.StylePrimitive.BackgroundColor = bg
		style.DefinitionDescription.BackgroundColor = bg
		style.Table.StyleBlock.StylePrimitive.BackgroundColor = bg
	}

	markdownRenderer, _ := glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(width),
	)

	return *markdownRenderer
}
