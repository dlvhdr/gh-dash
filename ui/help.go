package ui

import "github.com/charmbracelet/lipgloss"

func (m Model) renderHelp() string {
	return helpStyle.Copy().
		Width(m.ctx.ScreenWidth).
		Render(lipgloss.PlaceVertical(footerHeight, lipgloss.Top, m.help.View(m.keys)))
}
