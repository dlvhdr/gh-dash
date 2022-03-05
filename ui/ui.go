package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/config"
	"github.com/dlvhdr/gh-prs/ui/components/help"
	"github.com/dlvhdr/gh-prs/ui/components/prsidebar"
	"github.com/dlvhdr/gh-prs/ui/components/prssection"
	"github.com/dlvhdr/gh-prs/ui/components/tabs"
	"github.com/dlvhdr/gh-prs/ui/context"
	"github.com/dlvhdr/gh-prs/utils"
)

type Model struct {
	keys          utils.KeyMap
	err           error
	sidebar       prsidebar.Model
	currSectionId int
	help          help.Model
	sections      []*prssection.Model
	ready         bool
	isSidebarOpen bool
	tabs          tabs.Model
	ctx           context.ProgramContext
}

func NewModel() Model {
	tabsModel := tabs.NewModel()
	return Model{
		keys:          utils.Keys,
		help:          help.NewModel(),
		currSectionId: 0,
		tabs:          tabsModel,
		sidebar:       prsidebar.NewModel(),
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
		cmd        tea.Cmd
		sidebarCmd tea.Cmd
		helpCmd    tea.Cmd
		cmds       []tea.Cmd
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

		case key.Matches(msg, m.keys.Up):
			currSection.PrevPr()
			m.onViewedPrChanged()

		case key.Matches(msg, m.keys.TogglePreview):
			m.sidebar.IsOpen = !m.sidebar.IsOpen
			m.syncMainContentWidth()

		case key.Matches(msg, m.keys.OpenGithub):
			currPR := m.getCurrPr()
			if currPR != nil {
				utils.OpenBrowser(currPR.Url)
			}

		case key.Matches(msg, m.keys.Refresh):
			cmd = currSection.FetchSectionPullRequests()

		case key.Matches(msg, m.keys.Quit):
			cmd = tea.Quit

		}

	case initMsg:
		m.ctx.Config = &msg.Config
		m.sidebar.IsOpen = msg.Config.Defaults.Preview.Open
		m.syncMainContentWidth()

		newSections, fetchPRsCmds := m.fetchAllSections()
		m.sections = newSections
		cmd = fetchPRsCmds

	case prssection.SectionMsg:
		updatedSection, sectionCmd := m.updateRelevantSection(msg)
		m.sections[updatedSection.Id] = &updatedSection
		cmd = sectionCmd

		if msg.GetSectionId() == m.currSectionId {
			switch msg.(type) {
			case prssection.SectionPullRequestsFetchedMsg:
				m.syncSidebarPr()
			}
		}

	case tea.WindowSizeMsg:
		m.onWindowSizeChanged(msg)

	case errMsg:
		m.err = msg
	}

	m.syncProgramContext()
	m.sidebar, sidebarCmd = m.sidebar.Update(msg)
	m.help, helpCmd = m.help.Update(msg)
	cmds = append(cmds, cmd, sidebarCmd, helpCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	if m.ctx.Config == nil {
		return "Reading config...\n"
	}

	s := strings.Builder{}
	s.WriteString(m.tabs.View(m.ctx))
	s.WriteString("\n")
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.getCurrSection().View(),
		m.sidebar.View(),
	)
	s.WriteString(mainContent)
	s.WriteString("\n")
	s.WriteString(m.help.View(m.ctx))
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
	m.syncSidebarPr()
}

func (m *Model) fetchAllSections() (newSections []*prssection.Model, fetchCmds tea.Cmd) {
	fetchPRsCmds := make([]tea.Cmd, 0, len(m.ctx.Config.PRSections))
	sections := make([]*prssection.Model, 0, len(m.ctx.Config.PRSections))
	for i, sectionConfig := range m.ctx.Config.PRSections {
		sectionModel := prssection.NewModel(i, &m.ctx, sectionConfig, nil)
		sections = append(sections, &sectionModel)
		fetchPRsCmds = append(fetchPRsCmds, sectionModel.FetchSectionPullRequests())
	}
	return sections, tea.Batch(fetchPRsCmds...)
}

func (m *Model) onWindowSizeChanged(msg tea.WindowSizeMsg) {
	m.help.SetWidth(msg.Width)
	m.ctx.ScreenWidth = msg.Width
	m.ctx.ScreenHeight = msg.Height
	m.ctx.MainContentHeight = msg.Height - tabs.TabsHeight - help.FooterHeight
	m.syncMainContentWidth()
}

func (m *Model) syncProgramContext() {
	for _, section := range m.sections {
		section.UpdateProgramContext(&m.ctx)
	}
	m.sidebar.UpdateProgramContext(&m.ctx)
}

func (m *Model) updateRelevantSection(msg prssection.SectionMsg) (updatedSection prssection.Model, cmd tea.Cmd) {
	for _, section := range m.sections {
		if msg.GetSectionId() == section.Id {
			updatedSection, cmd = section.Update(msg)
		}
	}

	return updatedSection, cmd
}

func (m *Model) syncMainContentWidth() {
	sideBarOffset := 0
	if m.sidebar.IsOpen {
		sideBarOffset = m.ctx.Config.Defaults.Preview.Width
	}
	m.ctx.MainContentWidth = m.ctx.ScreenWidth - sideBarOffset
}

func (m *Model) syncSidebarPr() {
	currPr := m.getCurrPr()
	m.sidebar.SetPrData(currPr)
}
