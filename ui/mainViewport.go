package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type MainViewport struct {
	model         viewport.Model
	topBoundId    int
	bottomBoundId int
}

func (m *Model) calcMainContentWidth(screenWidth int) int {
	sideBarOffset := 0
	if m.isSidebarOpen {
		sideBarOffset = m.getSidebarWidth()
	}
	return screenWidth - sideBarOffset
}

func (m *Model) RenderMainViewPort() string {
	pagerContent := ""
	numPrs := m.getCurrSection().NumPrs()
	if numPrs > 0 {
		pagerContent = fmt.Sprintf(
			"PR %v/%v",
			m.cursor.currPrId+1,
			m.getCurrSection().NumPrs(),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.mainViewport.model.View(),
		pagerStyle.Copy().Render(pagerContent),
	)
}
