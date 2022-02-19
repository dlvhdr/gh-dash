package ui

import (
	"github.com/dlvhdr/gh-prs/data"
	"github.com/dlvhdr/gh-prs/ui/components/prssection"
	"github.com/dlvhdr/gh-prs/utils"
)

func (m Model) getCurrSection() *prssection.Model {
	if m.sections == nil || len(m.sections) == 0 {
		return nil
	}
	return m.sections[m.cursor.currSectionId]
}

func (m Model) getCurrPr() *data.PullRequestData {
	section := m.getCurrSection()
	if section == nil ||
		// section.IsLoading ||
		section.NumPrs() == 0 ||
		m.cursor.currPrId > section.NumPrs()-1 {
		return nil
	}

	pr := section.Prs[m.cursor.currPrId]
	return &pr
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

	newPrId := utils.Min(m.cursor.currPrId+1, currSection.NumPrs()-1)
	newPrId = utils.Max(newPrId, 0)
	m.cursor.currPrId = newPrId
}

func (m Model) getSectionAt(id int) *prssection.Model {
	return m.sections[id]
}

func (m Model) getPrevSectionId() int {
	m.cursor.currSectionId = (m.cursor.currSectionId - 1) % len(m.config.PRSections)
	if m.cursor.currSectionId < 0 {
		m.cursor.currSectionId += len(m.config.PRSections)
	}

	return m.cursor.currSectionId
}

func (m Model) getNextSectionId() int {
	return (m.cursor.currSectionId + 1) % len(m.config.PRSections)
}
