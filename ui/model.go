package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/config"
	"github.com/dlvhdr/gh-prs/ui/components/tabs"
	"github.com/dlvhdr/gh-prs/utils"
)

func NewModel() Model {
	helpModel := help.NewModel()
	style := lipgloss.NewStyle().Foreground(secondaryText)
	helpModel.Styles = help.Styles{
		ShortDesc:      style.Copy(),
		FullDesc:       style.Copy(),
		ShortSeparator: style.Copy(),
		FullSeparator:  style.Copy(),
		FullKey:        style.Copy(),
		ShortKey:       style.Copy(),
		Ellipsis:       style.Copy(),
	}
	tabsModel := tabs.NewModel()
	return Model{
		keys: utils.Keys,
		help: helpModel,
		cursor: cursor{
			currSectionId: 0,
			currPrId:      0,
		},
		tabs: tabsModel,
	}
}

func (m *Model) updateOnConfigFetched(config config.Config) {
	m.config = &config
	m.context.Config = config
	var data []Section
	for i, sectionConfig := range m.config.PRSections {
		s := spinner.Model{Spinner: spinner.Dot}
		data = append(data, Section{
			Id:        i,
			Config:    sectionConfig,
			Spinner:   s,
			IsLoading: true,
			Limit: func() int {
				if sectionConfig.Limit != nil {
					return *sectionConfig.Limit
				}

				return m.config.Defaults.PrsLimit
			}(),
		})
	}
	m.data = &data
	m.isSidebarOpen = m.config.Defaults.Preview.Open
}

func (m Model) startFetchingSectionsData() tea.Cmd {
	var cmds []tea.Cmd
	for _, section := range *m.data {
		section := section
		cmds = append(cmds, section.fetchSectionPullRequests())
		cmds = append(cmds, section.Tick(spinner.Tick))
	}
	return tea.Batch(cmds...)
}
