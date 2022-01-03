package ui

import "github.com/charmbracelet/lipgloss"

var (
	waitingGlyph = lipgloss.NewStyle().Foreground(faintBorder).Render("")
	failureGlyph = lipgloss.NewStyle().Foreground(warningText).Render("")
	successGlyph = lipgloss.NewStyle().Foreground(successText).Render("")
)
