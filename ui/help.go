package ui

func (m Model) renderHelp() string {
	return helpStyle.Copy().
		Width(m.ctx.ScreenWidth).
		Render(m.help.View(m.keys))
}
