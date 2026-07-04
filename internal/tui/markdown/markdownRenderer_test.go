package markdown

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetMarkdownRendererNilContext(t *testing.T) {
	markdownStyle = nil
	markdownStyleSource = ""

	require.NotPanics(t, func() {
		renderer := GetMarkdownRenderer(80, nil)
		rendered, err := renderer.Render("# hello")
		require.NoError(t, err)
		require.NotEmpty(t, rendered)
	})
}

func TestInitializeMarkdownStyleNilContext(t *testing.T) {
	markdownStyle = nil
	markdownStyleSource = ""

	require.NotPanics(t, func() {
		InitializeMarkdownStyle(nil)
	})

	require.NotNil(t, markdownStyle)
}
