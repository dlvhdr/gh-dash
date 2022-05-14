package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/help"
	"github.com/dlvhdr/gh-dash/ui/components/issuesidebar"
	"github.com/dlvhdr/gh-dash/ui/components/issuessection"
	"github.com/dlvhdr/gh-dash/ui/components/prsidebar"
	"github.com/dlvhdr/gh-dash/ui/components/prssection"
	"github.com/dlvhdr/gh-dash/ui/components/section"
	"github.com/dlvhdr/gh-dash/ui/components/sidebar"
	"github.com/dlvhdr/gh-dash/ui/components/tabs"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/utils"
)

type Model struct {
	keys          utils.KeyMap
	err           error
	sidebar       sidebar.Model
	currSectionId int
	help          help.Model
	prs           []section.Section
	issues        []section.Section
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
		sidebar:       sidebar.NewModel(),
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
		cmd         tea.Cmd
		sidebarCmd  tea.Cmd
		helpCmd     tea.Cmd
		cmds        []tea.Cmd
		currSection = m.getCurrSection()
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case m.isUserDefinedKeybinding(msg):
			m.executeKeybinding(msg.String())

		case key.Matches(msg, m.keys.PrevSection):
			prevSection := m.getSectionAt(m.getPrevSectionId())
			if prevSection != nil {
				m.setCurrSectionId(prevSection.Id())
				m.onViewedRowChanged()
			}

		case key.Matches(msg, m.keys.NextSection):
			nextSectionId := m.getNextSectionId()
			nextSection := m.getSectionAt(nextSectionId)
			if nextSection != nil {
				m.setCurrSectionId(nextSection.Id())
				m.onViewedRowChanged()
			}

		case key.Matches(msg, m.keys.Down):
			currSection.NextRow()
			m.onViewedRowChanged()

		case key.Matches(msg, m.keys.Up):
			currSection.PrevRow()
			m.onViewedRowChanged()

		case key.Matches(msg, m.keys.TogglePreview):
			m.sidebar.IsOpen = !m.sidebar.IsOpen
			m.syncMainContentWidth()

		case key.Matches(msg, m.keys.OpenGithub):
			var currRow = m.getCurrRowData()
			if currRow != nil {
				utils.OpenBrowser(currRow.GetUrl())
			}

		case key.Matches(msg, m.keys.Refresh):
			cmd = currSection.FetchSectionRows()

		case key.Matches(msg, m.keys.SwitchView):
			m.ctx.View = m.switchSelectedView()
			m.syncMainContentWidth()
			m.setCurrSectionId(0)

			currSections := m.getCurrentViewSections()
			if len(currSections) == 0 {
				newSections, fetchSectionsCmds := m.fetchAllViewSections()
				m.setCurrentViewSections(newSections)
				cmd = fetchSectionsCmds
			}
			m.onViewedRowChanged()

		case key.Matches(msg, m.keys.Quit):
			cmd = tea.Quit

		}

	case initMsg:
		m.ctx.Config = &msg.Config
		m.ctx.View = m.ctx.Config.Defaults.View
		m.sidebar.IsOpen = msg.Config.Defaults.Preview.Open
		m.syncMainContentWidth()
		newSections, fetchSectionsCmds := m.fetchAllViewSections()
		m.setCurrentViewSections(newSections)
		cmd = fetchSectionsCmds

	case section.SectionMsg:
		cmd = m.updateRelevantSection(msg)

		if msg.GetSectionId() == m.currSectionId {
			switch msg.GetSectionType() {
			case prssection.SectionType:
				m.onViewedRowChanged()
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
	currSection := m.getCurrSection()
	mainContent := ""
	if currSection != nil {
		mainContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.getCurrSection().View(),
			m.sidebar.View(),
		)
	} else {
		mainContent = "No sections defined..."
	}
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
	m.tabs.SetCurrSectionId(newSectionId)
}

func (m *Model) onViewedRowChanged() {
	m.syncSidebarPr()
}

func (m *Model) onWindowSizeChanged(msg tea.WindowSizeMsg) {
	m.help.SetWidth(msg.Width)
	m.ctx.ScreenWidth = msg.Width
	m.ctx.ScreenHeight = msg.Height
	m.ctx.MainContentHeight = msg.Height - tabs.TabsHeight - help.FooterHeight
	m.syncMainContentWidth()
}

func (m *Model) syncProgramContext() {
	for _, section := range m.getCurrentViewSections() {
		section.UpdateProgramContext(&m.ctx)
	}
	m.sidebar.UpdateProgramContext(&m.ctx)
}

func (m *Model) updateRelevantSection(msg section.SectionMsg) (cmd tea.Cmd) {
	var updatedSection section.Section

	switch msg.GetSectionType() {
	case prssection.SectionType:
		updatedSection, cmd = m.prs[msg.GetSectionId()].Update(msg)
		m.prs[msg.GetSectionId()] = updatedSection
	case issuessection.SectionType:
		updatedSection, cmd = m.issues[msg.GetSectionId()].Update(msg)
		m.issues[msg.GetSectionId()] = updatedSection
	}

	return cmd
}

func (m *Model) syncMainContentWidth() {
	sideBarOffset := 0
	if m.sidebar.IsOpen {
		sideBarOffset = m.ctx.Config.Defaults.Preview.Width
	}
	m.ctx.MainContentWidth = m.ctx.ScreenWidth - sideBarOffset
}

func (m *Model) syncSidebarPr() {
	currRowData := m.getCurrRowData()
	width := m.sidebar.GetSidebarContentWidth()

	switch data := currRowData.(type) {
	case *data.PullRequestData:
		content := prsidebar.NewModel(data, width).View()
		m.sidebar.SetContent(content)
	case *data.IssueData:
		content := issuesidebar.NewModel(data, width).View()
		m.sidebar.SetContent(content)
	}
}

func (m *Model) fetchAllViewSections() ([]section.Section, tea.Cmd) {
	if m.ctx.View == config.PRsView {
		return prssection.FetchAllSections(m.ctx)
	} else {
		return issuessection.FetchAllSections(m.ctx)
	}
}

func (m *Model) getCurrentViewSections() []section.Section {
	if m.ctx.View == config.PRsView {
		return m.prs
	} else {
		return m.issues
	}
}

func (m *Model) setCurrentViewSections(newSections []section.Section) {
	if m.ctx.View == config.PRsView {
		m.prs = newSections
	} else {
		m.issues = newSections
	}
}

func (m *Model) switchSelectedView() config.ViewType {
	if m.ctx.View == config.PRsView {
		return config.IssuesView
	} else {
		return config.PRsView
	}
}

func (m *Model) isUserDefinedKeybinding(msg tea.KeyMsg) bool {
	if m.ctx.View != config.PRsView {
		return false
	}

	for _, keybinding := range m.ctx.Config.Keybindings.Prs {
		if keybinding.Key == msg.String() {
			return true
		}
	}

	return false
}
