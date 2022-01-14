package ui

func (m *Model) syncSidebarViewPort() {
	m.sidebarViewport.Width = m.getSidebarWidth()
	m.setSidebarViewportContent()
}
