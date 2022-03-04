package ui

import (
	"github.com/dlvhdr/gh-prs/data"
	"github.com/dlvhdr/gh-prs/ui/components/prssection"
)

func (m Model) getCurrSection() *prssection.Model {
	if m.sections == nil || len(m.sections) == 0 {
		return nil
	}
	return m.sections[m.currSectionId]
}

func (m Model) getCurrPr() *data.PullRequestData {
	section := m.getCurrSection()
	if section == nil {
		return nil
	}
	return section.GetCurrPr()
}

func (m Model) getSectionAt(id int) *prssection.Model {
	return m.sections[id]
}

func (m Model) getPrevSectionId() int {
	m.currSectionId = (m.currSectionId - 1) % len(m.ctx.Config.PRSections)
	if m.currSectionId < 0 {
		m.currSectionId += len(m.ctx.Config.PRSections)
	}

	return m.currSectionId
}

func (m Model) getNextSectionId() int {
	return (m.currSectionId + 1) % len(m.ctx.Config.PRSections)
}
