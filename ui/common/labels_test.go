package common_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/common"
)

var (
	defaultStyle = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
)

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
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := common.RenderLabels(tc.width, tc.labels, tc.baseStyle)
			require.Equal(t, tc.want, got)
		})
	}
}
