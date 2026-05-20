package markdown

import "testing"

func TestGetMarkdownRendererWithoutInitializedStyle(t *testing.T) {
	markdownStyle = nil
	t.Cleanup(func() { markdownStyle = nil })

	renderer := GetMarkdownRenderer(80)

	if _, err := renderer.Render("# title"); err != nil {
		t.Fatalf("render markdown: %v", err)
	}
}
