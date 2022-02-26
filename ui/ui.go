package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/config"
	"github.com/dlvhdr/gh-prs/ui/components/prssection"
	"github.com/dlvhdr/gh-prs/ui/components/tabs"
	"github.com/dlvhdr/gh-prs/ui/context"
	"github.com/dlvhdr/gh-prs/utils"
)

type Model struct {
	keys            utils.KeyMap
	err             error
	config          *config.Config
	sidebarViewport viewport.Model
	currSectionId   int
	help            help.Model
	sections        []*prssection.Model
	ready           bool
	isSidebarOpen   bool
	tabs            tabs.Model
	ctx             context.ProgramContext
}

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
		keys:          utils.Keys,
		help:          helpModel,
		currSectionId: 0,
		tabs:          tabsModel,
	}
}

func initScreen() tea.Msg {
	config, err := config.ParseConfig()
	if err != nil {
		return errMsg{err}
	}

	return initMsg{Config: config}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(initScreen, tea.EnterAltScreen)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	currSection := m.getCurrSection()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.PrevSection):
			prevSection := m.getSectionAt(m.getPrevSectionId())
			m.setCurrSectionId(prevSection.Id)
			m.onViewedPrChanged()

		case key.Matches(msg, m.keys.NextSection):
			nextSection := m.getSectionAt(m.getNextSectionId())
			m.setCurrSectionId(nextSection.Id)
			m.onViewedPrChanged()

		case key.Matches(msg, m.keys.Down):
			currSection.NextPr()
			m.onViewedPrChanged()
			return m, nil

		case key.Matches(msg, m.keys.Up):
			currSection.PrevPr()
			m.onViewedPrChanged()
			return m, nil

		case key.Matches(msg, m.keys.TogglePreview):
			m.isSidebarOpen = !m.isSidebarOpen
			m.syncSidebarViewPort()

		case key.Matches(msg, m.keys.OpenGithub):
			currPR := m.getCurrPr()
			if currPR != nil {
				utils.OpenBrowser(currPR.Url)
			}

		case key.Matches(msg, m.keys.Refresh):
			cmd = currSection.FetchSectionPullRequests()

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, m.keys.Quit):
			cmd = tea.Quit
		}

	case initMsg:
		m.updateOnConfigFetched(msg.Config)
		m.syncSidebarViewPort()

		newSections, fetchPRsCmds := m.fetchAllSections()
		m.sections = newSections
		cmd = fetchPRsCmds

	case prssection.SectionMsg:
		for _, section := range m.sections {
			var updatedSection prssection.Model
			if msg.GetSectionId() == section.Id {
				updatedSection, cmd = section.Update(msg)
				m.sections[section.Id] = &updatedSection
			}

			if msg.GetSectionId() == m.currSectionId {
				switch msg.(type) {
				case prssection.SectionPullRequestsFetchedMsg:
					m.syncSidebarViewPort()
				}
			}
		}

	case tea.WindowSizeMsg:
		verticalMargins := headerHeight + footerHeight

		contentHeight := msg.Height - verticalMargins
		if !m.ready {
			m.sidebarViewport = viewport.Model{
				Width:  0,
				Height: contentHeight,
			}
			m.ready = true
		}

		m.help.Width = msg.Width
		m.ctx.ScreenWidth = msg.Width
		m.ctx.MainContentWidth = m.calcMainContentWidth()
		m.ctx.MainContentHeight = contentHeight - 2
		for _, section := range m.sections {
			section.UpdateProgramContext(&m.ctx)
		}

	case errMsg:
		m.err = msg
	}

	newSidebarViewport, sidebarCmd := m.sidebarViewport.Update(msg)
	m.sidebarViewport = newSidebarViewport

	cmds = append(cmds, cmd, sidebarCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	if m.config == nil {
		return "Reading config...\n"
	}

	s := strings.Builder{}
	s.WriteString(m.tabs.View(m.ctx))
	s.WriteString("\n")
	s.WriteString(lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.getCurrSection().View(),
		m.renderSidebar(),
	))
	s.WriteString("\n")
	s.WriteString(m.renderHelp())
	return s.String()
}

type initMsg struct {
	Config config.Config
}

type errMsg struct {
	error
}

func (e errMsg) Error() string { return e.error.Error() }

func (m *Model) setCurrSectionId(newSectionId int) {
	m.currSectionId = newSectionId
	m.tabs.SetCurrSectionId(m.currSectionId)
}

func (m *Model) onViewedPrChanged() {
	m.syncSidebarViewPort()
}

func (m *Model) fetchAllSections() (newSections []*prssection.Model, fetchCmds tea.Cmd) {
	fetchPRsCmds := make([]tea.Cmd, 0, len(m.config.PRSections))
	sections := make([]*prssection.Model, 0, len(m.config.PRSections))
	for i, sectionConfig := range m.config.PRSections {
		sectionModel := prssection.NewModel(i, &m.ctx, sectionConfig, nil)
		sections = append(sections, &sectionModel)
		fetchPRsCmds = append(fetchPRsCmds, sectionModel.FetchSectionPullRequests())
	}
	return sections, tea.Batch(fetchPRsCmds...)
}

func (m *Model) updateOnConfigFetched(config config.Config) {
	m.config = &config
	m.ctx.Config = config
	m.isSidebarOpen = m.config.Defaults.Preview.Open
}

func (m *Model) calcMainContentWidth() int {
	sideBarOffset := 0
	if m.isSidebarOpen {
		sideBarOffset = m.getSidebarWidth()
	}
	return m.ctx.ScreenWidth - sideBarOffset
}

func (m *Model) syncSidebarViewPort() {
	m.ctx.MainContentWidth = m.calcMainContentWidth()
	section := m.getCurrSection()
	if section != nil {
		section.UpdateProgramContext(&m.ctx)
	}
	m.sidebarViewport.Width = m.getSidebarWidth()
	m.setSidebarViewportContent()
}
