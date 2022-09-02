package ui

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
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
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/keys"
	"github.com/dlvhdr/gh-dash/ui/styles"
	"github.com/dlvhdr/gh-dash/utils"
)

type Model struct {
	keys          keys.KeyMap
	sidebar       sidebar.Model
	currSectionId int
	help          help.Model
	prs           []section.Section
	issues        []section.Section
	ready         bool
	isSidebarOpen bool
	tabs          tabs.Model
	ctx           context.ProgramContext
	taskSpinner   spinner.Model
	tasks         map[string]context.Task
}

func NewModel() Model {
	tabsModel := tabs.NewModel()
	m := Model{
		keys:          keys.Keys,
		help:          help.NewModel(),
		currSectionId: 1,
		tabs:          tabsModel,
		sidebar:       sidebar.NewModel(),
		taskSpinner:   spinner.Model{Spinner: spinner.Dot},
		tasks:         map[string]context.Task{},
	}

	m.ctx = context.ProgramContext{StartTask: func(task context.Task) tea.Cmd {
		task.StartTime = time.Now()
		m.tasks[task.Id] = task
		return m.taskSpinner.Tick
	}}
	return m
}

func initScreen() tea.Msg {
	config, err := config.ParseConfig()
	if err != nil {
		log.Fatal(err)
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
		m.ctx.Error = nil

		if currSection != nil && currSection.IsSearchFocused() {
			cmd = m.updateSection(currSection.GetId(), currSection.GetType(), msg)
			return m, cmd
		}

		switch {
		case m.isUserDefinedKeybinding(msg):
			cmd = m.executeKeybinding(msg.String())
			return m, cmd

		case key.Matches(msg, m.keys.PrevSection):
			prevSection := m.getSectionAt(m.getPrevSectionId())
			if prevSection != nil {
				m.setCurrSectionId(prevSection.GetId())
				m.onViewedRowChanged()
			}

		case key.Matches(msg, m.keys.NextSection):
			nextSectionId := m.getNextSectionId()
			nextSection := m.getSectionAt(nextSectionId)
			if nextSection != nil {
				m.setCurrSectionId(nextSection.GetId())
				m.onViewedRowChanged()
			}

		case key.Matches(msg, m.keys.Down):
			currSection.NextRow()
			m.onViewedRowChanged()

		case key.Matches(msg, m.keys.Up):
			currSection.PrevRow()
			m.onViewedRowChanged()

		case key.Matches(msg, m.keys.FirstLine):
			currSection.FirstItem()
			m.onViewedRowChanged()

		case key.Matches(msg, m.keys.LastLine):
			currSection.LastItem()
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
			currSection.ResetFilters()
			cmd = currSection.FetchSectionRows()

		case key.Matches(msg, m.keys.SwitchView):
			m.ctx.View = m.switchSelectedView()
			m.syncMainContentWidth()
			m.setCurrSectionId(1)

			currSections := m.getCurrentViewSections()
			if len(currSections) == 0 {
				newSections, fetchSectionsCmds := m.fetchAllViewSections()
				m.setCurrentViewSections(newSections)
				cmd = fetchSectionsCmds
			}
			m.onViewedRowChanged()

		case key.Matches(msg, m.keys.Search):
			if currSection != nil {
				cmd = currSection.SetIsSearching(true)
				return m, cmd
			}

		case key.Matches(msg, m.keys.Help):
			if !m.help.ShowAll {
				m.ctx.MainContentHeight = m.ctx.MainContentHeight + styles.FooterHeight - styles.ExpandedHelpHeight
			} else {
				m.ctx.MainContentHeight = m.ctx.MainContentHeight + styles.ExpandedHelpHeight - styles.FooterHeight
			}

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

	case constants.TaskFinishedMsg:
		task, ok := m.tasks[msg.TaskId]
		if ok {
			task.State = context.TaskFinished
			m.tasks[msg.TaskId] = task
			cmd = tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return constants.ClearTaskMsg{TaskId: msg.TaskId}
			})
		}

	case spinner.TickMsg:
		if len(m.tasks) > 0 {
			taskSpinner, internalTickCmd := m.taskSpinner.Update(msg)
			m.taskSpinner = taskSpinner
			cmd = internalTickCmd
		}

	case constants.ClearTaskMsg:
		delete(m.tasks, msg.TaskId)

	case section.SectionMsg:
		cmd = m.updateRelevantSection(msg)

		if msg.Id == m.currSectionId {
			switch msg.Type {
			case prssection.SectionType:
				m.onViewedRowChanged()
			}
		}

	case tea.WindowSizeMsg:
		m.onWindowSizeChanged(msg)

	case constants.ErrMsg:
		m.ctx.Error = msg.Err
	}

	m.syncProgramContext()

	m.sidebar, sidebarCmd = m.sidebar.Update(msg)
	m.help, helpCmd = m.help.Update(msg)
	sectionCmd := m.updateCurrentSection(msg)
	cmds = append(cmds, cmd, sidebarCmd, helpCmd, sectionCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
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
	if m.ctx.Error != nil {
		s.WriteString(
			styles.ErrorStyle.
				Width(m.ctx.ScreenWidth).
				Render(fmt.Sprintf("%s %s",
					constants.FailureGlyph,
					lipgloss.NewStyle().
						Foreground(styles.DefaultTheme.WarningText).
						Render(m.ctx.Error.Error()),
				)),
		)
	} else if len(m.tasks) > 0 {
		s.WriteString(m.renderRunningTask())
	} else {
		s.WriteString(m.help.View(m.ctx))
	}

	return s.String()
}

type initMsg struct {
	Config config.Config
}

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
	m.ctx.MainContentHeight = msg.Height - tabs.TabsHeight - styles.FooterHeight
	m.syncMainContentWidth()
}

func (m *Model) syncProgramContext() {
	for _, section := range m.getCurrentViewSections() {
		section.UpdateProgramContext(&m.ctx)
	}
	m.sidebar.UpdateProgramContext(&m.ctx)
}

func (m *Model) updateSection(id int, sType string, msg tea.Msg) (cmd tea.Cmd) {
	var updatedSection section.Section
	switch sType {
	case prssection.SectionType:
		updatedSection, cmd = m.prs[id].Update(msg)
		m.prs[id] = updatedSection
	case issuessection.SectionType:
		updatedSection, cmd = m.issues[id].Update(msg)
		m.issues[id] = updatedSection
	}

	return cmd
}

func (m *Model) updateRelevantSection(msg section.SectionMsg) (cmd tea.Cmd) {
	return m.updateSection(msg.Id, msg.Type, msg)
}

func (m *Model) updateCurrentSection(msg tea.Msg) (cmd tea.Cmd) {
	section := m.getCurrSection()
	if section == nil {
		return nil
	}
	return m.updateSection(section.GetId(), section.GetType(), msg)
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

	if currRowData == nil {
		m.sidebar.SetContent("")
		return
	}

	switch row := currRowData.(type) {
	case *data.PullRequestData:
		content := prsidebar.NewModel(row, width).View()
		m.sidebar.SetContent(content)
	case *data.IssueData:
		content := issuesidebar.NewModel(row, width).View()
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
		search := prssection.NewModel(
			0,
			&m.ctx,
			config.SectionConfig{Title: "", Filters: "archived:false"},
		)
		m.prs = append([]section.Section{&search}, newSections...)
	} else {
		search := issuessection.NewModel(
			0,
			&m.ctx,
			config.SectionConfig{Title: "", Filters: ""},
		)
		m.issues = append([]section.Section{&search}, newSections...)
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

func (m *Model) renderRunningTask() string {
	tasks := make([]context.Task, 0, len(m.tasks))
	for _, value := range m.tasks {
		tasks = append(tasks, value)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].StartTime.After(tasks[j].StartTime)
	})
	task := tasks[0]

	var status string
	switch task.State {
	case context.TaskStart:
		status = fmt.Sprintf("%s %s", m.taskSpinner.View(), task.StartText)
	case context.TaskError:
		status = task.Error.Error()
	case context.TaskFinished:
		status = task.FinishedText
	}

	return styles.FooterStyle.Width(m.ctx.ScreenWidth).Copy().Render(status)
}
