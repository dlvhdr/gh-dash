package ui

import "github.com/charmbracelet/lipgloss"

func (m Model) renderSidebar() string {
	return sideBarStyle.Copy().
		Height(m.viewport.Height + lipgloss.Height(titleCellStyle.Render("Title")) + 1).
		Render("I ARE SIDEBAR")
}
