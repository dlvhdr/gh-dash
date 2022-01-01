package ui

func (m Model) calcViewPortWidth() int {
	sideBarOffset := 0
	if m.isSidebarOpen {
		sideBarOffset = sideBarWidth
	}
	return m.width - sideBarOffset
}

func (m *Model) syncViewPort() {
	m.viewport.Width = m.calcViewPortWidth()
	m.viewport.SetContent(m.renderPullRequestList())
}
