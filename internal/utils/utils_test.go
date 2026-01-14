package utils

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestGetStylePrefix(t *testing.T) {
	tests := []struct {
		name     string
		style    lipgloss.Style
		wantLen  int
		wantSame bool // whether input == output (no reset to strip)
	}{
		{
			name:    "empty style returns empty or minimal ANSI",
			style:   lipgloss.NewStyle(),
			wantLen: 0,
		},
		{
			name:    "style with foreground color",
			style:   lipgloss.NewStyle().Foreground(lipgloss.Color("red")),
			wantLen: 5, // At least some ANSI codes
		},
		{
			name:    "style with background color",
			style:   lipgloss.NewStyle().Background(lipgloss.Color("blue")),
			wantLen: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStylePrefix(tt.style)
			// The result should not end with reset sequence
			if len(result) >= 4 && result[len(result)-4:] == "\x1b[0m" {
				t.Error("GetStylePrefix should strip trailing reset sequence")
			}
		})
	}
}

func TestGetStylePrefix_StripsReset(t *testing.T) {
	// Create a style that will produce a reset sequence
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	rendered := style.Render("")

	// Verify the raw render has a reset
	hasReset := len(rendered) >= 4 && rendered[len(rendered)-4:] == "\x1b[0m"

	prefix := GetStylePrefix(style)

	// Prefix should not have reset even if rendered did
	prefixHasReset := len(prefix) >= 4 && prefix[len(prefix)-4:] == "\x1b[0m"
	if prefixHasReset {
		t.Errorf("GetStylePrefix should strip reset, but got: %q", prefix)
	}

	// If original had reset, prefix should be shorter
	if hasReset && len(prefix) >= len(rendered) {
		t.Error("GetStylePrefix should return shorter string when stripping reset")
	}
}
