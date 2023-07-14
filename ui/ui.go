package ui

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/common"
	"github.com/dlvhdr/gh-dash/ui/components/dashboard"
	"github.com/dlvhdr/gh-dash/ui/components/footer"
	"github.com/dlvhdr/gh-dash/ui/components/issuesidebar"
	"github.com/dlvhdr/gh-dash/ui/components/prsidebar"
	"github.com/dlvhdr/gh-dash/ui/components/section"
	"github.com/dlvhdr/gh-dash/ui/components/sectionsview"
	"github.com/dlvhdr/gh-dash/ui/components/sidebar"
	"github.com/dlvhdr/gh-dash/ui/components/tabs"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/keys"
	"github.com/dlvhdr/gh-dash/ui/theme"
)

type Model struct {
	keys          keys.KeyMap
	sidebar       sidebar.Model
	prSidebar     prsidebar.Model
	issueSidebar  issuesidebar.Model
	currSectionId int
	footer        footer.Model
	sectionsView  sectionsview.Model
	ready         bool
	ctx           context.ProgramContext
	taskSpinner   spinner.Model
	tasks         map[string]context.Task
	tabs          tabs.Model
	dashboard     dashboard.Model
}

func NewModel(configPath string) Model {
	tabsModel := tabs.NewModel()
	m := Model{
		keys:          keys.Keys,
		currSectionId: 1,
		tabs:          tabsModel,
		taskSpinner:   spinner.Model{Spinner: spinner.Dot},
		tasks:         map[string]context.Task{},
	}

	m.ctx = context.ProgramContext{
		ConfigPath: configPath,
		StartTask: func(task context.Task) tea.Cmd {
			log.Debug("Starting task", "id", task.Id)
			task.StartTime = time.Now()
			m.tasks[task.Id] = task
			rTask := m.renderRunningTask()
			m.footer.SetRightSection(rTask)
			return m.taskSpinner.Tick
		},
		Notify: func(message string) tea.Cmd {
			log.Debug("Notifying", message)
			return m.notify(message)
		},
	}
	m.taskSpinner.Style = lipgloss.NewStyle().
		Background(m.ctx.Theme.SelectedBackground)

	footer := footer.NewModel(m.ctx)
	m.footer = footer

	m.sidebar = sidebar.NewModel(&m.ctx)
	m.sectionsView = sectionsview.NewModel(&m.ctx)
	m.prSidebar = prsidebar.NewModel(m.ctx)
	m.issueSidebar = issuesidebar.NewModel(m.ctx)
	m.dashboard = dashboard.NewModel(m.ctx)

	return m
}

func (m *Model) initScreen() tea.Msg {
	var err error

	config, err := config.ParseConfig(m.ctx.ConfigPath)
	if err != nil {
		log.KeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)
		log.SeparatorStyle = lipgloss.NewStyle()
		log.NewWithOptions(
			os.Stderr,
			log.Options{
				TimeFormat:      time.RFC3339,
				ReportTimestamp: true,
				Prefix:          "Reading config file",
				ReportCaller:    true,
			},
		).
			Fatal(
				"failed parsing config file",
				"location",
				m.ctx.ConfigPath,
				"err",
				err,
			)
	}

	return common.ConfigReadMsg{Config: config}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.initScreen, tea.EnterAltScreen)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd             tea.Cmd
		sidebarCmd      tea.Cmd
		prSidebarCmd    tea.Cmd
		prsViewCmd      tea.Cmd
		issueSidebarCmd tea.Cmd
		footerCmd       tea.Cmd
		dashboardCmd    tea.Cmd
		tabsCmd         tea.Cmd
		cmds            []tea.Cmd
		taskInternalMsg tea.Msg
		taskCmd         tea.Cmd
		currSection     = m.sectionsView.GetCurrSection()
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		log.Debug("Key pressed", "key", msg.String())
		m.ctx.Error = nil

		if currSection != nil && currSection.IsSearchFocused() {
			cmd = m.updateSection(msg)
			return m, cmd
		}

		if m.prSidebar.IsTextInputBoxFocused() {
			m.prSidebar, cmd = m.prSidebar.Update(msg)
			return m, cmd
		}

		if m.issueSidebar.IsTextInputBoxFocused() {
			m.issueSidebar, cmd = m.issueSidebar.Update(msg)
			return m, cmd
		}

		switch {

		case key.Matches(msg, m.keys.TogglePreview):
			m.sidebar.IsOpen = !m.sidebar.IsOpen

		case key.Matches(msg, m.keys.SwitchView):
			m.ctx.View = m.getNextView()
			m.tabs.GoToFirstSection()
			if m.ctx.View != config.DashboardsView {
				m.sectionsView.UpdateProgramContext(&m.ctx)
				cmd = m.sectionsView.FetchAllViewSections(false)
				m.sectionsView.GoToFirstSection()
			}

		case key.Matches(msg, keys.PRKeys.Comment), key.Matches(msg, keys.IssueKeys.Comment):
			m.sidebar.IsOpen = true
			if m.ctx.View == config.PRsView {
				cmd = m.prSidebar.SetIsCommenting(true)
			} else {
				cmd = m.issueSidebar.SetIsCommenting(true)
			}
			m.sidebar.ScrollToBottom()
			return m, cmd

		case key.Matches(msg, keys.IssueKeys.Assign), key.Matches(msg, keys.PRKeys.Assign):
			m.sidebar.IsOpen = true
			if m.ctx.View == config.PRsView {
				cmd = m.prSidebar.SetIsAssigning(true)
			} else {
				cmd = m.issueSidebar.SetIsAssigning(true)
			}
			m.sidebar.ScrollToBottom()
			return m, cmd

		case key.Matches(msg, keys.IssueKeys.Unassign), key.Matches(msg, keys.PRKeys.Unassign):
			m.sidebar.IsOpen = true
			if m.ctx.View == config.PRsView {
				cmd = m.prSidebar.SetIsUnassigning(true)
			} else {
				cmd = m.issueSidebar.SetIsUnassigning(true)
			}
			m.sidebar.ScrollToBottom()
			return m, cmd

		case key.Matches(msg, m.keys.Help):
			if !m.footer.ShowAll {
				m.ctx.MainContentHeight = m.ctx.MainContentHeight + common.FooterHeight - common.ExpandedHelpHeight
			} else {
				m.ctx.MainContentHeight = m.ctx.MainContentHeight + common.ExpandedHelpHeight - common.FooterHeight
			}

		case key.Matches(msg, m.keys.Quit):
			cmd = tea.Quit

		}

	case common.ConfigReadMsg:
		m.ctx.Config = &msg.Config
		m.ctx.Theme = theme.ParseTheme(m.ctx.Config)
		m.ctx.Styles = context.InitStyles(m.ctx.Theme)
		m.ctx.View = m.ctx.Config.Defaults.View
		m.sidebar.IsOpen = msg.Config.Defaults.Preview.Open
		m.sidebar.UpdateProgramContext(&m.ctx)
		m.dashboard.UpdateProgramContext(&m.ctx)
		m.sectionsView.UpdateProgramContext(&m.ctx)
		cmds = append(cmds, fetchUser, m.doRefreshAtInterval())

	case userFetchedMsg:
		m.ctx.User = msg.user

	case constants.TaskFinishedMsg:
		task, ok := m.tasks[msg.TaskId]
		if ok {
			log.Debug("Task finished", "id", task.Id)
			if msg.Err != nil {
				task.State = context.TaskError
				task.Error = msg.Err
			} else {
				task.State = context.TaskFinished
			}
			now := time.Now()
			task.FinishedTime = &now
			m.tasks[msg.TaskId] = task
			tickCmd := tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return constants.ClearTaskMsg{TaskId: msg.TaskId}
			})

			var prsViewTaskCmd tea.Cmd
			m.sectionsView, prsViewTaskCmd = m.sectionsView.Update(section.SectionMsg{Id: msg.SectionId, Type: msg.SectionType, InternalMsg: msg.Msg})
			taskInternalMsg = msg.Msg
			cmds = append(cmds, prsViewTaskCmd, tickCmd)
		}

	case spinner.TickMsg:
		if len(m.tasks) > 0 {
			taskSpinner, internalTickCmd := m.taskSpinner.Update(msg)
			m.taskSpinner = taskSpinner
			rTask := m.renderRunningTask()
			m.footer.SetRightSection(rTask)
			cmd = internalTickCmd
		}

	case constants.ClearTaskMsg:
		m.footer.SetRightSection("")
		delete(m.tasks, msg.TaskId)

	case tea.WindowSizeMsg:
		m.onWindowSizeChanged(msg)

	case constants.ErrMsg:
		m.ctx.Error = msg.Err
	}

	var newRow data.RowData
	oldRow := m.sectionsView.GetCurrRow()
	if m.ctx.View != config.DashboardsView {
		m.sectionsView, prsViewCmd = m.sectionsView.Update(msg)
		newRow = m.sectionsView.GetCurrRow()
	}
	if isSameRow(oldRow, newRow) {
		m.onViewedRowChanged()
	}

	m.syncSidebar(newRow)
	m.sidebar, sidebarCmd = m.sidebar.Update(msg)

	if m.prSidebar.IsTextInputBoxFocused() {
		m.prSidebar, prSidebarCmd = m.prSidebar.Update(msg)
	}

	if m.issueSidebar.IsTextInputBoxFocused() {
		m.issueSidebar, issueSidebarCmd = m.issueSidebar.Update(msg)
	}

	m.footer, footerCmd = m.footer.Update(msg)
	if currSection != nil {
		m.footer.SetLeftSection(currSection.GetPagerContent())
	}
	m.dashboard, dashboardCmd = m.dashboard.Update(msg)
	m.dashboard, taskCmd = m.dashboard.Update(taskInternalMsg)

	m.tabs, tabsCmd = m.tabs.Update(msg)

	m.syncProgramContext()

	cmds = append(
		cmds,
		cmd,
		sidebarCmd,
		footerCmd,
		prsViewCmd,
		prSidebarCmd,
		issueSidebarCmd,
		dashboardCmd,
		taskCmd,
		tabsCmd,
	)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.ctx.Config == nil || m.ctx.MainContentHeight == 0 {
		return "Reading config...\n"
	}

	s := strings.Builder{}
	tabs := m.tabs.View(m.ctx)
	s.WriteString(tabs)
	s.WriteString("\n")
	mainContent := ""
	if m.ctx.View == config.DashboardsView {
		mainContent = m.dashboard.View()
	} else {
		mainContent = m.sectionsView.View()
	}
	rest := lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainContent,
		m.sidebar.View(),
	)
	s.WriteString(rest)
	s.WriteString("\n")
	if m.ctx.Error != nil {
		s.WriteString(
			m.ctx.Styles.Common.ErrorStyle.
				Width(m.ctx.ScreenWidth).
				Render(fmt.Sprintf("%s %s",
					m.ctx.Styles.Common.FailureGlyph,
					lipgloss.NewStyle().
						Foreground(m.ctx.Theme.WarningText).
						Render(m.ctx.Error.Error()),
				)),
		)
	} else {
		s.WriteString(m.footer.View())
	}

	return s.String()
}

func (m *Model) onViewedRowChanged() {
	m.sidebar.ScrollToTop()
}

func (m *Model) onWindowSizeChanged(msg tea.WindowSizeMsg) {
	m.ctx.ScreenWidth = msg.Width
	m.ctx.ScreenHeight = msg.Height
	m.ctx.MainContentHeight = msg.Height - common.TabsHeight - common.FooterHeight
	m.syncMainContentWidth()
}

func (m *Model) syncProgramContext() {
	m.syncMainContentWidth()
	m.footer.UpdateProgramContext(&m.ctx)
	m.sidebar.UpdateProgramContext(&m.ctx)
	m.prSidebar.UpdateProgramContext(&m.ctx)
	m.issueSidebar.UpdateProgramContext(&m.ctx)
	m.dashboard.UpdateProgramContext(&m.ctx)
	m.sectionsView.UpdateProgramContext(&m.ctx)
}

func (m *Model) updateSection(msg tea.Msg) (cmd tea.Cmd) {
	m.sectionsView, cmd = m.sectionsView.Update(msg)

	return cmd
}

func (m *Model) syncMainContentWidth() {
	sideBarOffset := 0
	if m.sidebar.IsOpen {
		sideBarOffset = m.ctx.Config.Defaults.Preview.Width
	}

	m.ctx.MainContentWidth = m.ctx.ScreenWidth - sideBarOffset
}

func (m *Model) syncSidebar(rowData data.RowData) {
	width := m.sidebar.GetSidebarContentWidth()

	if rowData == nil {
		m.sidebar.SetContent("")
		return
	}

	switch row := rowData.(type) {
	case *data.PullRequestData:
		m.prSidebar.SetSectionId(m.currSectionId)
		m.prSidebar.SetRow(row)
		m.prSidebar.SetWidth(width)
		m.sidebar.SetContent(m.prSidebar.View())
	case *data.IssueData:
		m.issueSidebar.SetSectionId(m.currSectionId)
		m.issueSidebar.SetRow(row)
		m.issueSidebar.SetWidth(width)
		m.sidebar.SetContent(m.issueSidebar.View())
	}
}

func (m *Model) getNextView() config.ViewType {
	if m.ctx.View == config.DashboardsView {
		return config.PRsView
	} else if m.ctx.View == config.PRsView {
		return config.IssuesView
	} else {
		return config.DashboardsView
	}
}

func (m *Model) isUserDefinedKeybinding(msg tea.KeyMsg) bool {
	if m.ctx.View == config.IssuesView {
		for _, keybinding := range m.ctx.Config.Keybindings.Issues {
			if keybinding.Key == msg.String() {
				return true
			}
		}
	}

	if m.ctx.View == config.PRsView {
		for _, keybinding := range m.ctx.Config.Keybindings.Prs {
			if keybinding.Key == msg.String() {
				return true
			}
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
		if tasks[i].FinishedTime != nil && tasks[j].FinishedTime == nil {
			return false
		}
		if tasks[j].FinishedTime != nil && tasks[i].FinishedTime == nil {
			return true
		}
		if tasks[j].FinishedTime != nil && tasks[i].FinishedTime != nil {
			return tasks[i].FinishedTime.After(*tasks[j].FinishedTime)
		}

		return tasks[i].StartTime.After(tasks[j].StartTime)
	})
	task := tasks[0]

	var currTaskStatus string
	switch task.State {
	case context.TaskStart:
		currTaskStatus =

			lipgloss.NewStyle().
				Background(m.ctx.Theme.SelectedBackground).
				Render(
					fmt.Sprintf(
						"%s%s",
						m.taskSpinner.View(),
						task.StartText,
					))
	case context.TaskError:
		currTaskStatus = lipgloss.NewStyle().
			Foreground(m.ctx.Theme.WarningText).
			Background(m.ctx.Theme.SelectedBackground).
			Render(fmt.Sprintf(" %s", task.Error.Error()))
	case context.TaskFinished:
		currTaskStatus = lipgloss.NewStyle().
			Foreground(m.ctx.Theme.SuccessText).
			Background(m.ctx.Theme.SelectedBackground).
			Render(fmt.Sprintf(" %s", task.FinishedText))
	}

	var numProcessing int
	for _, task := range m.tasks {
		if task.State == context.TaskStart {
			numProcessing += 1
		}
	}

	stats := ""
	if numProcessing > 1 {
		stats = lipgloss.NewStyle().
			Foreground(m.ctx.Theme.FaintText).
			Background(m.ctx.Theme.SelectedBackground).
			Render(fmt.Sprintf("[ %d] ", numProcessing))
	}

	return lipgloss.NewStyle().
		Padding(0, 1).
		Background(m.ctx.Theme.SelectedBackground).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, stats, currTaskStatus))
}

type userFetchedMsg struct {
	user string
}

func fetchUser() tea.Msg {
	user, err := data.CurrentLoginName()
	if err != nil {
		return constants.ErrMsg{
			Err: err,
		}
	}

	return userFetchedMsg{
		user: user,
	}
}

type intervalRefresh time.Time

func (m *Model) doRefreshAtInterval() tea.Cmd {
	return tea.Tick(
		time.Minute*time.Duration(m.ctx.Config.Defaults.RefetchIntervalMinutes),
		func(t time.Time) tea.Msg {
			return intervalRefresh(t)
		},
	)
}
