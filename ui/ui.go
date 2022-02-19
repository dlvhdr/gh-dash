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
	mainViewport    MainViewport
	sidebarViewport viewport.Model
	cursor          cursor
	help            help.Model
	sections        []*prssection.Model
	ready           bool
	isSidebarOpen   bool
	width           int
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
			m.syncMainViewPort()
			m.syncSidebarViewPort()
			m.setMainViewPortBounds()
			m.tabs.SetCurrSectionId(newCursor.currSectionId)

			return m, nil

		case key.Matches(msg, m.keys.NextSection):
			nextSection := m.getSectionAt(m.getNextSectionId())
			newCursor := cursor{
				currSectionId: nextSection.Id,
				currPrId:      0,
			}
			m.cursor = newCursor
			m.syncSidebarViewPort()
			m.syncMainViewPort()
			m.setMainViewPortBounds()
			m.tabs.SetCurrSectionId(newCursor.currSectionId)

			return m, cmd

		case key.Matches(msg, m.keys.Down):
			m.nextPr()
			m.syncMainViewPort()
			m.onLineDown()
			m.syncSidebarViewPort()
			return m, nil

		case key.Matches(msg, m.keys.Up):
			m.prevPr()
			m.syncMainViewPort()
			m.syncSidebarViewPort()
			m.onLineUp()
			return m, nil

		case key.Matches(msg, m.keys.TogglePreview):
			m.isSidebarOpen = !m.isSidebarOpen
			m.syncMainViewPort()
			m.syncSidebarViewPort()
			return m, nil
		case key.Matches(msg, m.keys.OpenGithub):
			currPR := m.getCurrPr()
			if currPR == nil {
				return m, nil
			}
			utils.OpenBrowser(currPR.Url)
			return m, nil
		// case key.Matches(msg, m.keys.Refresh):
		// 	var newData []Section
		// 	for _, section := range *m.data {
		// 		if section.IsLoading {
		// 			return m, nil
		// 		}
		//
		// 		// section.IsLoading = true
		// 		section.Prs = []PullRequest{}
		// 		newData = append(newData, section)
		// 	}
		// 	m.data = &newData
		// 	m.syncSidebarViewPort()
		// 	return m, m.startFetchingSectionsData()
		//
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil

		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}

	case initMsg:
		m.updateOnConfigFetched(msg.Config)
		m.mainViewport.model.Width = m.calcViewPortWidth()
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

		return m, tea.Batch(fetchPRsCmds...)

	// case tickMsg:
	// 	var internalCmd tea.Cmd
	// 	section := (*m.data)[msg.SectionId]
	// 	if !section.IsLoading {
	// 		return m, nil
	// 	}
	// 	section.Spinner, internalCmd = section.Spinner.Update(msg.InternalTickMsg)
	// 	if internalCmd == nil {
	// 		return m, nil
	// 	}
	//
	// 	(*m.data)[msg.SectionId] = section
	// 	return m, section.Tick(internalCmd)

	// case sectionPullRequestsFetchedMsg:
	// 	section := (*m.data)[msg.SectionId]
	// 	section.Prs = msg.Prs
	// 	sort.Slice(section.Prs, func(i, j int) bool {
	// 		return section.Prs[i].UpdatedAt.After(section.Prs[j].UpdatedAt)
	// 	})
	// 	(*m.data)[msg.SectionId] = section
	// 	m.setMainViewPortBounds()
	// 	return m, m.makeRenderPullRequestCmd(msg.SectionId)
	//
	// case pullRequestsRenderedMsg:
	// 	section := (*m.data)[msg.sectionId]
	// 	section.Spinner.Finish()
	// 	section.IsLoading = false
	// 	(*m.data)[msg.sectionId] = section
	// 	m.mainViewport.model.SetContent(msg.content)
	// 	m.syncSidebarViewPort()
	// 	return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.help.Width = msg.Width
		verticalMargins := headerHeight + footerHeight + pagerHeight

		if !m.ready {
			m.mainViewport = MainViewport{
				model: viewport.Model{
					Width:  m.calcViewPortWidth(),
					Height: msg.Height - verticalMargins - 1,
				},
			}
			m.sidebarViewport = viewport.Model{
				Width:  0,
				Height: msg.Height - verticalMargins + 1,
			}
			m.context.ScreenWidth = m.width
			m.context.MainViewportWidth = m.mainViewport.model.Width
			m.ready = true
		} else {
			m.mainViewport.model.Height = msg.Height - verticalMargins - 1
			m.sidebarViewport.Height = msg.Height - verticalMargins + 1
			m.syncMainViewPort()
			m.syncSidebarViewPort()
		}
		return m, nil
	case errMsg:
		m.err = msg
		return m, nil
	}

	updatedSections := make([]*prssection.Model, 0, len(m.sections))
	for _, section := range m.sections {
		updatedSection, cmd := section.Update(msg)
		updatedSections = append(updatedSections, &updatedSection)
		cmds = append(cmds, cmd)
	}
	m.sections = updatedSections

	m.sidebarViewport, cmd = m.sidebarViewport.Update(msg)
	cmds = append(cmds, cmd)
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
		lipgloss.Top,
		m.getCurrSection().View(),
		m.renderSidebar()),
	)
	s.WriteString("\n")
	s.WriteString(m.renderHelp())
	return s.String()
}
