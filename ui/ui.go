package ui

import (
	"dlvhdr/gh-prs/config"
	"dlvhdr/gh-prs/utils"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	keys     utils.KeyMap
	err      error
	configs  []config.SectionConfig
	data     *[]section
	viewport viewport.Model
	cursor   cursor
	help     help.Model
	ready    bool
	logger   *os.File
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
	return Model{
		keys: utils.Keys,
		help: help.NewModel(),
		cursor: cursor{
			currSectionId: 0,
			currPrId:      0,
		},
		logger: logFile,
	}
}

func initScreen() tea.Msg {
	sections, err := config.ParseSectionsConfig()
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
			m.viewport.SetContent(m.renderPullRequestList())
			return m, nil

		case key.Matches(msg, m.keys.NextSection):
			nextSection := m.getSectionAt(m.getNextSectionId())
			newCursor := cursor{
				currSectionId: nextSection.Id,
				currPrId:      0,
			}
			m.cursor = newCursor
			m.viewport.SetContent(m.renderPullRequestList())
			return m, nil

		case key.Matches(msg, m.keys.Down):
			m.nextPr()
			m.viewport.SetContent(m.renderPullRequestList())
			return m, nil

		case key.Matches(msg, m.keys.Up):
			m.prevPr()
			m.viewport.SetContent(m.renderPullRequestList())
			return m, nil

		case key.Matches(msg, m.keys.Open):
			currPR := func() PullRequest {
				var prs []PullRequest
				for _, section := range *m.data {
					prs = append(prs, section.Prs...)
				}
				return prs[m.cursor.currPrId]
			}()
			utils.OpenBrowser(currPR.Url)
			return m, nil

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
				Id:      i,
				Config:  sectionConfig,
				Spinner: sectionSpinner{Model: s, NumReposFetched: 0},
			})
		}
		m.data = &data
		return m, m.startFetchingSectionsData()

	case repoPullRequestsFetchedMsg:
		section := (*m.data)[msg.SectionId]
		section.Prs = append(section.Prs, msg.Prs...)
		sort.Slice(section.Prs, func(i, j int) bool {
			return section.Prs[i].UpdatedAt.After(section.Prs[j].UpdatedAt)
		})

		section.Spinner.NumReposFetched += 1
		(*m.data)[msg.SectionId] = section
		return m, m.makeRenderPullRequestCmd(msg.SectionId)

	case pullRequestsRenderedMsg:
		section := (*m.data)[msg.sectionId]
		section.Spinner.Model.Finish()
		(*m.data)[msg.sectionId] = section
		m.viewport.SetContent(msg.content)
		return m, nil

	case tickMsg:
		var internalCmd tea.Cmd
		section := (*m.data)[msg.SectionId]
		section.Spinner.Model, internalCmd = section.Spinner.Model.Update(msg.InternalTickMsg)
		(*m.data)[msg.SectionId] = section
		return m, section.Tick(internalCmd)

	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		verticalMargins := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.Model{
				Width:  msg.Width - 2*mainContentPadding,
				Height: msg.Height - verticalMargins,
			}
			m.ready = true

			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width - 2*mainContentPadding
			m.viewport.Height = msg.Height - verticalMargins
		}

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
		PaddingLeft(mainContentPadding).
		PaddingRight(mainContentPadding)

	s := strings.Builder{}
	s.WriteString(m.renderTabs())
	s.WriteString("\n")
	s.WriteString(paddedContentStyle.Render(m.renderTableHeader()))
	s.WriteString("\n")
	s.WriteString(paddedContentStyle.Render(m.viewport.View()))
	s.WriteString("\n")
	s.WriteString(lipgloss.PlaceVertical(2, lipgloss.Bottom, m.help.View(m.keys)))
	return s.String()
}

func (m Model) startFetchingSectionsData() tea.Cmd {
	var cmds []tea.Cmd
	for _, section := range *m.data {
		section := section
		cmds = append(cmds, section.fetchSectionPullRequests()...)
		cmds = append(cmds, section.Tick(spinner.Tick))
	}
	return tea.Batch(cmds...)
}
