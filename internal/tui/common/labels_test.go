package common_test

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
)

var defaultStyle = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)

func TestRenderLabels(t *testing.T) {
	testCases := map[string]struct {
		width     int
		labels    []data.Label
		baseStyle lipgloss.Style
		want      string
	}{
		"one label, one row": {
			width:     20,
			baseStyle: defaultStyle,
			labels: []data.Label{
				{
					Name: "label-1",
				},
			},
			want: " label-1  ",
		},
		"two labels, one row": {
			width:     20,
			baseStyle: defaultStyle,
			labels: []data.Label{
				{
					Name: "l-1",
				},
				{
					Name: "l-2",
				},
			},
			want: " l-1   l-2  ",
		},
		"two labels, two rows": {
			width:     10,
			baseStyle: defaultStyle,
			labels: []data.Label{
				{
					Name: "label-1",
				},
				{
					Name: "label-2",
				},
			},
			want: " label-1  \n          \n label-2  ",
		},
		"three labels, two rows": {
			width:     20,
			baseStyle: defaultStyle,
			labels: []data.Label{
				{
					Name: "l-1",
				},
				{
					Name: "l-2",
				},
				{
					Name: "label-3",
				},
			},
			want: " l-1   l-2  \n            \n label-3    ",
		},
		"two labels, two rows, labels equal width": {
			width:     5,
			baseStyle: defaultStyle,
			labels: []data.Label{
				{
					Name: "l-1",
				},
				{
					Name: "l-2",
				},
			},
			want: " l-1  \n      \n l-2  ",
		},
		"two labels, one row, labels equal width": {
			width:     11,
			baseStyle: defaultStyle,
			labels: []data.Label{
				{
					Name: "l-1",
				},
				{
					Name: "l-2",
				},
			},
			want: " l-1   l-2  ",
		},
		"emoji shortcode": {
			width:     20,
			baseStyle: defaultStyle,
			labels: []data.Label{
				{
					Name: ":chicken:",
				},
			},
			want: " 🐔  ",
		},
		"mixed emoji shortcode text": {
			width:     20,
			baseStyle: defaultStyle,
			labels: []data.Label{
				{
					Name: "team :rocket:",
				},
			},
			want: " team 🚀  ",
		},
		"unknown shortcode stays raw": {
			width:     24,
			baseStyle: defaultStyle,
			labels: []data.Label{
				{
					Name: ":not-real:",
				},
			},
			want: " :not-real:  ",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := common.RenderLabels(tc.labels, common.LabelOpts{
				Width:     tc.width,
				PillStyle: tc.baseStyle,
			})
			require.Equal(t, tc.want, ansi.Strip(got))
		})
	}
}

func TestRenderLabelsRespectsMaxRows(t *testing.T) {
	got := common.RenderLabels(
		[]data.Label{
			{Name: "bug"},
			{Name: "fix"},
			{Name: "chore"},
		},
		common.LabelOpts{
			Width:     12,
			MaxRows:   1,
			PillStyle: defaultStyle,
		},
	)

	require.Equal(t, " bug   +2 ", ansi.Strip(got))
}

func TestRenderLabelsUsesUpstreamRendering(t *testing.T) {
	got := common.RenderLabels(
		[]data.Label{
			{Name: "bug", Color: "ff0000"},
			{Name: "fix", Color: "00ff00"},
		},
		common.LabelOpts{
			Width:     12,
			MaxRows:   1,
			PillStyle: lipgloss.NewStyle(),
		},
	)

	require.NotEmpty(t, got)
	require.Contains(t, ansi.Strip(got), "bug")
	require.NotContains(t, ansi.Strip(got), "[")
}

func TestRenderLabelsOversizedFirstLabelRespectsMaxRows(t *testing.T) {
	got := common.RenderLabels(
		[]data.Label{
			{Name: "verylonglabel"},
			{Name: "bug"},
		},
		common.LabelOpts{
			Width:     8,
			MaxRows:   1,
			PillStyle: defaultStyle,
		},
	)

	stripped := ansi.Strip(got)
	require.Equal(t, 0, strings.Count(stripped, "\n"))
	require.NotContains(t, stripped, "+")
}

func TestRenderLabelsRendersEmojiShortcodes(t *testing.T) {
	got := common.RenderLabels(
		[]data.Label{
			{Name: ":chicken:", Color: "ff0000"},
			{Name: ":rocket:", Color: "00ff00"},
		},
		common.LabelOpts{
			Width:     16,
			MaxRows:   1,
			PillStyle: defaultStyle,
		},
	)

	require.Contains(t, ansi.Strip(got), "🐔")
	require.Contains(t, ansi.Strip(got), "🚀")
	require.NotContains(t, ansi.Strip(got), ":chicken:")
}
