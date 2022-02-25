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
	"github.com/dlvhdr/gh-prs/ui/constants"
	"github.com/dlvhdr/gh-prs/ui/context"
	"github.com/dlvhdr/gh-prs/utils"
)

type Model struct {
	keys            utils.KeyMap
	err             error
	config          *config.Config
	mainViewport    MainViewport
	sidebarViewport viewport.Model
	cursor          cursor
	help            help.Model
	sections        []*prssection.Model
	ready           bool
	isSidebarOpen   bool
	tabs            tabs.Model
	context         context.ProgramContext
}

type cursor struct {
	currSectionId int
	currPrId      int
}

type initMsg struct {
	Config config.Config
}

type errMsg struct {
	error
}

func (e errMsg) Error() string { return e.error.Error() }

type pullRequestsRenderedMsg struct {
	sectionId int
	content   string
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
			newCursor := cursor{
				currSectionId: prevSection.Id,
				currPrId:      0,
			}
			m.cursor = newCursor
			m.syncSidebarViewPort()
			m.tabs.SetCurrSectionId(newCursor.currSectionId)

		case key.Matches(msg, m.keys.NextSection):
			nextSection := m.getSectionAt(m.getNextSectionId())
			newCursor := cursor{
				currSectionId: nextSection.Id,
				currPrId:      0,
			}
			m.cursor = newCursor
			m.syncSidebarViewPort()
			m.tabs.SetCurrSectionId(newCursor.currSectionId)

		case key.Matches(msg, m.keys.Down):
			m.cursor.currPrId = currSection.NextPr()
			m.syncSidebarViewPort()

		case key.Matches(msg, m.keys.Up):
			m.cursor.currPrId = currSection.PrevPr()
			m.syncSidebarViewPort()

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
		m.sidebarViewport.Width = m.getSidebarWidth()
		m.syncSidebarViewPort()

		fetchPRsCmds := make([]tea.Cmd, 0, len(msg.Config.PRSections))
		sections := make([]*prssection.Model, 0, len(msg.Config.PRSections))
		for i, sectionConfig := range msg.Config.PRSections {
			sectionModel := prssection.NewModel(i, &m.context, sectionConfig, nil)
			sections = append(sections, &sectionModel)
			fetchPRsCmds = append(fetchPRsCmds, sectionModel.FetchSectionPullRequests())
		}
		m.sections = sections

		cmd = tea.Batch(fetchPRsCmds...)

	case prssection.SectionMsg:
		for _, section := range m.sections {
			var updatedSection prssection.Model
			if msg.GetSectionId() == section.Id {
				updatedSection, cmd = section.Update(msg)
				m.sections[section.Id] = &updatedSection
			}
		}

	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		verticalMargins := headerHeight + footerHeight + pagerHeight

		if !m.ready {
			m.sidebarViewport = viewport.Model{
				Width:  0,
				Height: msg.Height - verticalMargins + 1,
			}
			m.context.ScreenWidth = msg.Width
			m.context.MainContentWidth = m.calcMainContentWidth(m.context.ScreenWidth)
			m.context.MainContentHeight = msg.Height - verticalMargins - 2
			m.ready = true
		}
		for _, section := range m.sections {
			section.SetDimensions(constants.Dimensions{
				Width:  m.calcMainContentWidth(),
				Height: msg.Height - verticalMargins - 2,
			})
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
	s.WriteString(m.tabs.View(m.context))
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
