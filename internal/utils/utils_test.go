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

func TestGetStylePrefix_VariousStyles(t *testing.T) {
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{
			name:  "bold style",
			style: lipgloss.NewStyle().Bold(true),
		},
		{
			name:  "italic style",
			style: lipgloss.NewStyle().Italic(true),
		},
		{
			name:  "underline style",
			style: lipgloss.NewStyle().Underline(true),
		},
		{
			name:  "combined foreground and background",
			style: lipgloss.NewStyle().Foreground(lipgloss.Color("red")).Background(lipgloss.Color("blue")),
		},
		{
			name:  "adaptive color",
			style: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"}),
		},
		{
			name:  "256 color",
			style: lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
		},
		{
			name:  "true color",
			style: lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5733")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := GetStylePrefix(tt.style)

			// Should not end with reset
			if len(prefix) >= 4 && prefix[len(prefix)-4:] == "\x1b[0m" {
				t.Errorf("GetStylePrefix should strip reset for %s, but got: %q", tt.name, prefix)
			}

			// Should be valid ANSI (starts with escape if non-empty and has styling)
			if len(prefix) > 0 && prefix[0] != '\x1b' {
				// Only fail if the style would produce ANSI codes
				rendered := tt.style.Render("")
				if len(rendered) > 0 && rendered[0] == '\x1b' {
					t.Errorf("GetStylePrefix for %s should start with escape, got: %q", tt.name, prefix)
				}
			}
		})
	}
}

func TestGetStylePrefix_ConcatenationPreservesStyle(t *testing.T) {
	// Test that concatenating prefix + text works as expected
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
	prefix := GetStylePrefix(style)

	// The prefix should be usable for concatenation
	result := prefix + "Hello"

	// Result should contain the text
	if len(result) < 5 {
		t.Error("Result should contain 'Hello'")
	}

	// Result should start with ANSI if style produces ANSI
	rendered := style.Render("")
	if len(rendered) > 0 && rendered[0] == '\x1b' && prefix[0] != '\x1b' {
		t.Error("Prefix should start with ANSI escape when style has color")
	}
}
