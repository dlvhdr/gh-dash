package ui

func (m Model) calcViewPortWidth() int {
	sideBarOffset := 0
	if m.isSidebarOpen {
		sideBarOffset = m.getSidebarWidth()
	}
	return m.width - sideBarOffset
}

func (m *Model) syncMainViewPort() {
	m.mainViewport.Width = m.calcViewPortWidth()
	m.mainViewport.SetContent(m.renderPullRequestList())
}

func (m *Model) syncSidebarViewPort() {
	m.sidebarViewport.Width = m.getSidebarWidth()
	m.setSidebarViewportContent()
}
