package ui

import (
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/config"
	"github.com/dlvhdr/gh-prs/utils"
)

type Model struct {
	keys          utils.KeyMap
	err           error
	configs       []config.SectionConfig
	data          *[]section
	viewport      viewport.Model
	cursor        cursor
	help          help.Model
	ready         bool
	isSidebarOpen bool
	width         int
	logger        *os.File
}

type cursor struct {
	currSectionId int
	currPrId      int
}

type initMsg struct {
	Config []config.SectionConfig
}

type errMsg struct {
	error
}

func (e errMsg) Error() string { return e.error.Error() }

type pullRequestsRenderedMsg struct {
	sectionId int
	content   string
}

func NewModel(logFile *os.File) Model {
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
	return Model{
		keys: utils.Keys,
		help: helpModel,
		cursor: cursor{
			currSectionId: 0,
			currPrId:      0,
		},
		logger: logFile,
	}
}

func initScreen() tea.Msg {
	sections, err := config.ParseConfig()
	if err != nil {
		return errMsg{err}
	}

	return initMsg{Config: sections}
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
			m.syncViewPort()
			return m, nil

		case key.Matches(msg, m.keys.NextSection):
			nextSection := m.getSectionAt(m.getNextSectionId())
			newCursor := cursor{
				currSectionId: nextSection.Id,
				currPrId:      0,
			}
			m.cursor = newCursor
			m.syncViewPort()
			return m, nil

		case key.Matches(msg, m.keys.Down):
			m.nextPr()
			m.syncViewPort()
			return m, nil

		case key.Matches(msg, m.keys.Up):
			m.prevPr()
			m.syncViewPort()
			return m, nil

		case key.Matches(msg, m.keys.TogglePreview):
			m.isSidebarOpen = !m.isSidebarOpen
			m.syncViewPort()
			return m, nil
		case key.Matches(msg, m.keys.OpenGithub):
			currPR := m.getCurrPr()
			if currPR == nil {
				return m, nil
			}
			utils.OpenBrowser(currPR.Data.Url)
			return m, nil
		case key.Matches(msg, m.keys.Refresh):
			var newData []section
			for _, section := range *m.data {
				if section.IsLoading {
					return m, nil
				}

				section.Spinner = spinner.Model{Spinner: spinner.Dot}
				section.IsLoading = true
				section.Prs = []PullRequest{}
				newData = append(newData, section)
			}
			m.data = &newData
			return m, m.startFetchingSectionsData()

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil

		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}

	case initMsg:
		m.configs = msg.Config
		var data []section
		for i, sectionConfig := range m.configs {
			s := spinner.Model{Spinner: spinner.Dot}
			data = append(data, section{
				Id:        i,
				Config:    sectionConfig,
				Spinner:   s,
				IsLoading: true,
			})
		}
		m.data = &data
		return m, m.startFetchingSectionsData()

	case tickMsg:
		var internalCmd tea.Cmd
		section := (*m.data)[msg.SectionId]
		if !section.IsLoading {
			return m, nil
		}
		section.Spinner, internalCmd = section.Spinner.Update(msg.InternalTickMsg)
		if internalCmd == nil {
			return m, nil
		}

		(*m.data)[msg.SectionId] = section
		return m, section.Tick(internalCmd)

	case repoPullRequestsFetchedMsg:
		section := (*m.data)[msg.SectionId]
		section.Prs = msg.Prs
		sort.Slice(section.Prs, func(i, j int) bool {
			return section.Prs[i].Data.UpdatedAt.After(section.Prs[j].Data.UpdatedAt)
		})
		(*m.data)[msg.SectionId] = section
		return m, m.makeRenderPullRequestCmd(msg.SectionId)

	case pullRequestsRenderedMsg:
		section := (*m.data)[msg.sectionId]
		section.Spinner.Finish()
		section.IsLoading = false
		(*m.data)[msg.sectionId] = section
		m.viewport.SetContent(msg.content)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.help.Width = msg.Width
		verticalMargins := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.Model{
				Width:  m.calcViewPortWidth(),
				Height: msg.Height - verticalMargins,
			}
			m.ready = true

			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Height = msg.Height - verticalMargins
			m.syncViewPort()
		}
		return m, nil
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	if m.configs == nil {
		return "Reading config...\n"
	}

	paddedContentStyle := lipgloss.NewStyle().
		PaddingTop(0).
		PaddingLeft(mainContentPadding).
		PaddingRight(mainContentPadding)

	s := strings.Builder{}
	s.WriteString(m.renderTabs())
	s.WriteString("\n")
	table := lipgloss.JoinVertical(lipgloss.Top, paddedContentStyle.Render(m.renderTableHeader()), m.renderCurrentSection())
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, table, m.renderSidebar()))
	s.WriteString("\n")
	s.WriteString(m.renderHelp())
	return s.String()
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
