package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/utils"
)

type MainViewport struct {
	model         viewport.Model
	topBoundId    int
	bottomBoundId int
}

func (m *Model) syncMainViewPort() {
	m.mainViewport.model.Width = m.calcViewPortWidth()
	// prs := m.section.View()
	// m.mainViewport.model.SetContent(prs)
}

func (m *Model) calcViewPortWidth() int {
	sideBarOffset := 0
	if m.isSidebarOpen {
		sideBarOffset = m.getSidebarWidth()
	}
	return m.width - sideBarOffset
}

func (m *Model) getNumPrsPerPage() int {
	return m.mainViewport.model.Height / prRowHeight
}

func (m *Model) setMainViewPortBounds() {
	currSection := m.getCurrSection()
	if currSection == nil {
		return
	}

	m.mainViewport.topBoundId = 0
	m.mainViewport.bottomBoundId = utils.Min(currSection.NumPrs()-1, m.getNumPrsPerPage()-1)
}

func (m *Model) onLineDown() {
	atBottomOfViewport := m.cursor.currPrId > m.mainViewport.bottomBoundId
	if atBottomOfViewport {
		m.mainViewport.topBoundId += 1
		m.mainViewport.bottomBoundId += 1
		m.mainViewport.model.LineDown(prRowHeight)
	}

	m.sidebarViewport.YOffset = 0
}

func (m *Model) onLineUp() {
	atTopOfViewport := m.cursor.currPrId < m.mainViewport.topBoundId
	if atTopOfViewport {
		m.mainViewport.topBoundId -= 1
		m.mainViewport.bottomBoundId -= 1
		m.mainViewport.model.LineUp(prRowHeight)
	}
	m.sidebarViewport.YOffset = 0
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
