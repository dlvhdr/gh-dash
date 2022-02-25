package listviewport

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/constants"
	"github.com/dlvhdr/gh-prs/utils"
)

type Model struct {
	viewport       viewport.Model
	topBoundId     int
	bottomBoundId  int
	currId         int
	Width          int
	ListItemHeight int
	NumItems       int
	ItemTypeLabel  string
}

func NewModel(dimensions constants.Dimensions, itemTypeLabel string, numItems, listItemHeight int) Model {
	model := Model{
		Width:          dimensions.Width,
		NumItems:       numItems,
		ListItemHeight: listItemHeight,
		currId:         0,
		viewport: viewport.Model{
			Width:  dimensions.Width,
			Height: dimensions.Height,
		},
		topBoundId:    0,
		ItemTypeLabel: itemTypeLabel,
	}
	model.bottomBoundId = utils.Min(model.NumItems-1, model.getNumPrsPerPage()-1)
	return model
}

func (m *Model) SyncViewPort(content string) {
	m.viewport.Width = m.Width
	m.viewport.SetContent(content)
}

func (m *Model) getNumPrsPerPage() int {
	return m.viewport.Height / m.ListItemHeight
}

func (m *Model) GetCurrItem() int {
	return m.currId
}

func (m *Model) NextItem() int {
	atBottomOfViewport := m.currId > m.bottomBoundId
	if atBottomOfViewport {
		m.topBoundId += 1
		m.bottomBoundId += 1
		m.viewport.LineDown(m.ListItemHeight)
	}

	newId := utils.Min(m.currId+1, m.NumItems-1)
	newId = utils.Max(newId, 0)
	m.currId = newId
	return m.currId
}

func (m *Model) PrevItem() int {
	atTopOfViewport := m.currId < m.topBoundId
	if atTopOfViewport {
		m.topBoundId -= 1
		m.bottomBoundId -= 1
		m.viewport.LineUp(m.ListItemHeight)
	}

	m.currId = utils.Max(m.currId-1, 0)
	return m.currId
}

func (m *Model) SetDimensions(dimensions constants.Dimensions) {
	m.viewport.Height = dimensions.Height
	m.viewport.Width = dimensions.Width
}

func (m *Model) View() string {
	pagerContent := ""
	if m.NumItems > 0 {
		pagerContent = fmt.Sprintf(
			"%v %v/%v",
			m.ItemTypeLabel,
			m.currId+1,
			m.NumItems,
		)
	}
	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.viewport.View(),
		pagerStyle.Copy().Render(pagerContent),
	)
}
