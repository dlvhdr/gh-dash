package ui

import "github.com/dlvhdr/gh-prs/utils"

func (m Model) getCurrSection() *section {
	if m.data == nil || len(*m.data) == 0 {
		return nil
	}
	return &(*m.data)[m.cursor.currSectionId]
}

func (m *Model) prevPr() {
	currSection := m.getCurrSection()
	if currSection == nil {
		return
	}

	newPrId := utils.Max(m.cursor.currPrId-1, 0)
	m.cursor.currPrId = newPrId
}

func (m *Model) nextPr() {
	currSection := m.getCurrSection()
	if currSection == nil {
		return
	}

	newPrId := utils.Min(m.cursor.currPrId+1, currSection.numPrs()-1)
	m.cursor.currPrId = newPrId
}

func (m Model) getSectionAt(id int) *section {
	return &(*m.data)[id]
}

func (m Model) getPrevSectionId() int {
	m.cursor.currSectionId = (m.cursor.currSectionId - 1) % len(m.configs)
	if m.cursor.currSectionId < 0 {
		m.cursor.currSectionId += len(m.configs)
	}

	return m.cursor.currSectionId
}

func (m Model) getNextSectionId() int {
	return (m.cursor.currSectionId + 1) % len(m.configs)
}
