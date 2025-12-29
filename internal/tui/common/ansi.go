package common

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func ApplyBackgroundAfterReset(value string, bgColor lipgloss.AdaptiveColor) string {
	bg := bgColor.Light
	if bg == "" {
		bg = bgColor.Dark
	}
	if bg == "" {
		return value
	}

	color := lipgloss.ColorProfile().Color(bg)
	if color == nil {
		return value
	}
	seq := color.Sequence(true)
	if seq == "" {
		return value
	}

	reset := termenv.CSI + termenv.ResetSeq + "m"
	if !strings.Contains(value, reset) {
		return value
	}
	bgSeq := termenv.CSI + seq + "m"
	return strings.ReplaceAll(value, reset, reset+bgSeq)
}
