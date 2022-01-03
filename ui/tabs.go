package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/utils"
)

func (m Model) renderTabs() string {
	var tabs []string
	for i, sectionConfig := range m.configs {
		if m.cursor.currSectionId == i {
			tabs = append(tabs, activeTab.Render(sectionConfig.Title))
		} else {
			tabs = append(tabs, tab.Render(sectionConfig.Title))
		}
	}

	{
		row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
		gap := tabGap.Render(
			strings.Repeat(
				" ",
				utils.Max(0, m.width-lipgloss.Width(row)),
			),
		)

		return tabsRow.Render(lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap))
	}
}
