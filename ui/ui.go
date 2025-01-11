package ui

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/config"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/git"
	"github.com/dlvhdr/gh-dash/v4/ui/common"
	"github.com/dlvhdr/gh-dash/v4/ui/components/branch"
	"github.com/dlvhdr/gh-dash/v4/ui/components/branchsidebar"
	"github.com/dlvhdr/gh-dash/v4/ui/components/footer"
	"github.com/dlvhdr/gh-dash/v4/ui/components/issuesidebar"
	"github.com/dlvhdr/gh-dash/v4/ui/components/issuessection"
	"github.com/dlvhdr/gh-dash/v4/ui/components/prsidebar"
	"github.com/dlvhdr/gh-dash/v4/ui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/ui/components/reposection"
	"github.com/dlvhdr/gh-dash/v4/ui/components/section"
	"github.com/dlvhdr/gh-dash/v4/ui/components/sidebar"
	"github.com/dlvhdr/gh-dash/v4/ui/components/tabs"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/ui/keys"
	"github.com/dlvhdr/gh-dash/v4/ui/theme"
)

type Model struct {
	keys          *keys.KeyMap
	sidebar       sidebar.Model
	prSidebar     prsidebar.Model
	issueSidebar  issuesidebar.Model
	branchSidebar branchsidebar.Model
	currSectionId int
	footer        footer.Model
	repo          section.Section
	prs           []section.Section
	issues        []section.Section
	tabs          tabs.Model
	ctx           *context.ProgramContext
	taskSpinner   spinner.Model
	tasks         map[string]context.Task
}

func NewModel(repoPath string, configPath string) Model {
	taskSpinner := spinner.Model{Spinner: spinner.Dot}
	m := Model{
		keys:        keys.Keys,
		sidebar:     sidebar.NewModel(),
		taskSpinner: taskSpinner,
		tasks:       map[string]context.Task{},
	}

	m.ctx = &context.ProgramContext{
		RepoPath:   repoPath,
		ConfigPath: configPath,
		StartTask: func(task context.Task) tea.Cmd {
			log.Debug("Starting task", "id", task.Id)
			task.StartTime = time.Now()
			m.tasks[task.Id] = task
			rTask := m.renderRunningTask()
			m.footer.SetRightSection(rTask)
			return m.taskSpinner.Tick
		},
	}

	m.taskSpinner.Style = lipgloss.NewStyle().
		Background(m.ctx.Theme.SelectedBackground)

	footer := footer.NewModel(m.ctx)
	m.footer = footer
	m.prSidebar = prsidebar.NewModel(m.ctx)
	m.issueSidebar = issuesidebar.NewModel(m.ctx)
	m.branchSidebar = branchsidebar.NewModel(m.ctx)
	m.tabs = tabs.NewModel(m.ctx)

	return m
}

func (m *Model) initScreen() tea.Msg {
	showError := func(err error) {
		styles := log.DefaultStyles()
		styles.Key = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)
		styles.Separator = lipgloss.NewStyle()

		logger := log.New(os.Stderr)
		logger.SetStyles(styles)
		logger.SetTimeFormat(time.RFC3339)
		logger.SetReportTimestamp(true)
		logger.SetPrefix("Reading config file")
		logger.SetReportCaller(true)

		logger.
			Fatal(
				"failed parsing config file",
				"location",
				m.ctx.ConfigPath,
				"err",
				err,
			)
	}

	cfg, err := config.ParseConfig(m.ctx.ConfigPath, m.ctx.RepoPath)
	if err != nil {
		showError(err)
		return initMsg{Config: cfg}
	}

	var url string
	if config.IsFeatureEnabled(config.FF_REPO_VIEW) && m.ctx.RepoPath != "" {
		res, err := git.GetOriginUrl(m.ctx.RepoPath)
		if err != nil {
			showError(err)
			return initMsg{Config: cfg}
		}
		url = res
	}

	err = keys.Rebind(
		cfg.Keybindings.Universal,
		cfg.Keybindings.Issues,
		cfg.Keybindings.Prs,
		cfg.Keybindings.Branches,
	)
	if err != nil {
		showError(err)
	}

	return initMsg{Config: cfg, RepoUrl: url}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.initScreen, tea.EnterAltScreen)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd             tea.Cmd
		tabsCmd         tea.Cmd
		sidebarCmd      tea.Cmd
		prSidebarCmd    tea.Cmd
		issueSidebarCmd tea.Cmd
		footerCmd       tea.Cmd
		cmds            []tea.Cmd
		currSection     = m.getCurrSection()
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		log.Debug("Key pressed", "key", msg.String())
		m.ctx.Error = nil

		if currSection != nil && (currSection.IsSearchFocused() || currSection.IsPromptConfirmationFocused()) {
			cmd = m.updateSection(currSection.GetId(), currSection.GetType(), msg)
			return m, cmd
		}

		if m.prSidebar.IsTextInputBoxFocused() {
			m.prSidebar, cmd = m.prSidebar.Update(msg)
			m.syncSidebar()
			return m, cmd
		}

		if m.issueSidebar.IsTextInputBoxFocused() {
			m.issueSidebar, cmd = m.issueSidebar.Update(msg)
			m.syncSidebar()
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
			prevRow := currSection.CurrRow()
			nextRow := currSection.NextRow()
			if prevRow != nextRow && nextRow == currSection.NumRows()-1 && m.ctx.View != config.RepoView {
				cmds = append(cmds, currSection.FetchNextPageSectionRows()...)
			}
			cmd = m.onViewedRowChanged()

		case key.Matches(msg, m.keys.Up):
			currSection.PrevRow()
			m.onViewedRowChanged()

		case key.Matches(msg, m.keys.FirstLine):
			currSection.FirstItem()
			cmd = m.onViewedRowChanged()

		case key.Matches(msg, m.keys.LastLine):
			if currSection.CurrRow()+1 < currSection.NumRows() {
				cmds = append(cmds, currSection.FetchNextPageSectionRows()...)
			}
			currSection.LastItem()
			cmd = m.onViewedRowChanged()

		case key.Matches(msg, m.keys.TogglePreview):
			m.sidebar.IsOpen = !m.sidebar.IsOpen
			m.syncMainContentWidth()

		case key.Matches(msg, m.keys.Refresh):
			currSection.ResetFilters()
			currSection.ResetRows()
			currSection.SetIsLoading(true)
			cmds = append(cmds, currSection.FetchNextPageSectionRows()...)

		case key.Matches(msg, m.keys.RefreshAll):
			newSections, fetchSectionsCmds := m.fetchAllViewSections()
			m.setCurrentViewSections(newSections)
			cmds = append(cmds, fetchSectionsCmds)

		case key.Matches(msg, m.keys.Search):
			if currSection != nil {
				cmd = currSection.SetIsSearching(true)
				return m, cmd
			}

		case key.Matches(msg, m.keys.Help):
			if !m.footer.ShowAll {
				m.ctx.MainContentHeight = m.ctx.MainContentHeight + common.FooterHeight - common.ExpandedHelpHeight
			} else {
				m.ctx.MainContentHeight = m.ctx.MainContentHeight + common.ExpandedHelpHeight - common.FooterHeight
			}

		case key.Matches(msg, m.keys.CopyNumber):
			row := m.getCurrRowData()
			var cmd tea.Cmd
			if reflect.ValueOf(row).IsNil() {
				cmd = m.notifyErr("Current selection isn't associated with a PR/Issue")
				return m, cmd
			}
			number := fmt.Sprint(row.GetNumber())
			err := clipboard.WriteAll(number)
			if err != nil {
				cmd = m.notifyErr(fmt.Sprintf("Failed copying to clipboard %v", err))
			} else {
				cmd = m.notify(fmt.Sprintf("Copied %s to clipboard", number))
			}
			return m, cmd

		case key.Matches(msg, m.keys.CopyUrl):
			var cmd tea.Cmd
			row := m.getCurrRowData()
			if reflect.ValueOf(row).IsNil() {
				cmd = m.notifyErr("Current selection isn't associated with a PR/Issue")
				return m, cmd
			}
			url := row.GetUrl()
			err := clipboard.WriteAll(url)
			if err != nil {
				cmd = m.notifyErr(fmt.Sprintf("Failed copying to clipboard %v", err))
			} else {
				cmd = m.notify(fmt.Sprintf("Copied %s to clipboard", url))
			}
			return m, cmd

		case key.Matches(msg, m.keys.Quit):
			if m.ctx.Config.ConfirmQuit {
				m.footer, cmd = m.footer.Update(msg)
				return m, cmd
			}
			cmd = tea.Quit

		case m.ctx.View == config.RepoView:
			switch {

			case key.Matches(msg, m.keys.OpenGithub):
				cmds = append(cmds, m.repo.(*reposection.Model).OpenGithub())

			case key.Matches(msg, keys.BranchKeys.Delete):
				if currSection != nil {
					currSection.SetPromptConfirmationAction("delete")
					cmd = currSection.SetIsPromptConfirmationShown(true)
				}
				return m, cmd

			case key.Matches(msg, keys.BranchKeys.New):
				if currSection != nil {
					currSection.SetPromptConfirmationAction("new")
					cmd = currSection.SetIsPromptConfirmationShown(true)
				}
				return m, cmd

			case key.Matches(msg, keys.BranchKeys.CreatePr):
				if currSection != nil {
					currSection.SetPromptConfirmationAction("create_pr")
					cmd = currSection.SetIsPromptConfirmationShown(true)
				}
				return m, cmd

			case key.Matches(msg, keys.BranchKeys.ViewPRs):
				m.ctx.View = m.switchSelectedView()
				m.syncMainContentWidth()
				m.setCurrSectionId(m.getCurrentViewDefaultSection())
				m.tabs.UpdateSectionsConfigs(m.ctx)

				currSections := m.getCurrentViewSections()
				if len(currSections) == 0 {
					newSections, fetchSectionsCmds := m.fetchAllViewSections()
					m.setCurrentViewSections(newSections)
					cmd = fetchSectionsCmds
				}
				m.onViewedRowChanged()

			}
		case m.ctx.View == config.PRsView:
			switch {
			case key.Matches(msg, m.keys.OpenGithub):
				cmds = append(cmds, m.openBrowser())

			case key.Matches(msg, keys.PRKeys.Approve):
				m.sidebar.IsOpen = true
				cmd = m.prSidebar.SetIsApproving(true)
				m.syncMainContentWidth()
				m.syncSidebar()
				m.sidebar.ScrollToBottom()
				return m, cmd

			case key.Matches(msg, keys.PRKeys.Assign):
				m.sidebar.IsOpen = true
				cmd = m.prSidebar.SetIsAssigning(true)
				m.syncMainContentWidth()
				m.syncSidebar()
				m.sidebar.ScrollToBottom()
				return m, cmd

			case key.Matches(msg, keys.PRKeys.Unassign):
				m.sidebar.IsOpen = true
				cmd = m.prSidebar.SetIsUnassigning(true)
				m.syncMainContentWidth()
				m.syncSidebar()
				m.sidebar.ScrollToBottom()
				return m, cmd

			case key.Matches(msg, keys.PRKeys.Comment):
				m.sidebar.IsOpen = true
				cmd = m.prSidebar.SetIsCommenting(true)
				m.syncMainContentWidth()
				m.syncSidebar()
				m.sidebar.ScrollToBottom()
				return m, cmd

			case key.Matches(msg, keys.PRKeys.Close):
				if currSection != nil {
					currSection.SetPromptConfirmationAction("close")
					cmd = currSection.SetIsPromptConfirmationShown(true)
				}
				return m, cmd

			case key.Matches(msg, keys.PRKeys.Ready):
				if currSection != nil {
					currSection.SetPromptConfirmationAction("ready")
					cmd = currSection.SetIsPromptConfirmationShown(true)
				}
				return m, cmd

			case key.Matches(msg, keys.PRKeys.Reopen):
				if currSection != nil {
					currSection.SetPromptConfirmationAction("reopen")
					cmd = currSection.SetIsPromptConfirmationShown(true)
				}
				return m, cmd

			case key.Matches(msg, keys.PRKeys.Merge):
				if currSection != nil {
					currSection.SetPromptConfirmationAction("merge")
					cmd = currSection.SetIsPromptConfirmationShown(true)
				}
				return m, cmd

			case key.Matches(msg, keys.PRKeys.Update):
				if currSection != nil {
					currSection.SetPromptConfirmationAction("update")
					cmd = currSection.SetIsPromptConfirmationShown(true)
				}
				return m, cmd

			case key.Matches(msg, keys.PRKeys.ViewIssues):
				m.ctx.View = m.switchSelectedView()
				m.syncMainContentWidth()
				m.setCurrSectionId(m.getCurrentViewDefaultSection())
				m.tabs.UpdateSectionsConfigs(m.ctx)

				currSections := m.getCurrentViewSections()
				if len(currSections) == 0 {
					newSections, fetchSectionsCmds := m.fetchAllViewSections()
					m.setCurrentViewSections(newSections)
					cmd = fetchSectionsCmds
				}
				m.onViewedRowChanged()
			}
		case m.ctx.View == config.IssuesView:
			switch {
			case key.Matches(msg, m.keys.OpenGithub):
				cmds = append(cmds, m.openBrowser())

			case key.Matches(msg, keys.IssueKeys.Assign):
				m.sidebar.IsOpen = true
				cmd = m.issueSidebar.SetIsAssigning(true)
				m.syncMainContentWidth()
				m.syncSidebar()
				m.sidebar.ScrollToBottom()
				return m, cmd

			case key.Matches(msg, keys.IssueKeys.Unassign):
				m.sidebar.IsOpen = true
				cmd = m.issueSidebar.SetIsUnassigning(true)
				m.syncMainContentWidth()
				m.syncSidebar()
				m.sidebar.ScrollToBottom()
				return m, cmd

			case key.Matches(msg, keys.IssueKeys.Comment):
				m.sidebar.IsOpen = true
				cmd = m.issueSidebar.SetIsCommenting(true)
				m.syncMainContentWidth()
				m.syncSidebar()
				m.sidebar.ScrollToBottom()
				return m, cmd

			case key.Matches(msg, keys.IssueKeys.Close):
				if currSection != nil {
					currSection.SetPromptConfirmationAction("close")
					cmd = currSection.SetIsPromptConfirmationShown(true)
				}
				return m, cmd

			case key.Matches(msg, keys.IssueKeys.Reopen):
				if currSection != nil {
					currSection.SetPromptConfirmationAction("reopen")
					cmd = currSection.SetIsPromptConfirmationShown(true)
				}
				return m, cmd

			case key.Matches(msg, keys.IssueKeys.ViewPRs):
				m.ctx.View = m.switchSelectedView()
				m.syncMainContentWidth()
				m.setCurrSectionId(m.getCurrentViewDefaultSection())
				m.tabs.UpdateSectionsConfigs(m.ctx)

				currSections := m.getCurrentViewSections()
				if len(currSections) == 0 {
					newSections, fetchSectionsCmds := m.fetchAllViewSections()
					m.setCurrentViewSections(newSections)
					cmd = fetchSectionsCmds
				}
				m.onViewedRowChanged()
			}

		}

	case initMsg:
		m.ctx.Config = &msg.Config
		m.ctx.RepoUrl = msg.RepoUrl
		m.ctx.Theme = theme.ParseTheme(m.ctx.Config)
		m.ctx.Styles = context.InitStyles(m.ctx.Theme)
		m.ctx.View = m.ctx.Config.Defaults.View
		m.currSectionId = m.getCurrentViewDefaultSection()
		m.sidebar.IsOpen = msg.Config.Defaults.Preview.Open
		m.tabs.UpdateSectionsConfigs(m.ctx)
		m.syncMainContentWidth()
		newSections, fetchSectionsCmds := m.fetchAllViewSections()
		m.setCurrentViewSections(newSections)
		cmds = append(cmds, fetchSectionsCmds, fetchUser, m.doRefreshAtInterval(), m.doUpdateFooterAtInterval())

	case intervalRefresh:
		newSections, fetchSectionsCmds := m.fetchAllViewSections()
		m.setCurrentViewSections(newSections)
		cmds = append(cmds, fetchSectionsCmds, m.doRefreshAtInterval())

	case userFetchedMsg:
		m.ctx.User = msg.user

	case constants.TaskFinishedMsg:
		task, ok := m.tasks[msg.TaskId]
		if ok {
			log.Debug("Task finished", "id", task.Id)
			if msg.Err != nil {
				log.Error("Task finished with error", "id", task.Id, "err", msg.Err)
				task.State = context.TaskError
				task.Error = msg.Err
			} else {
				task.State = context.TaskFinished
			}
			now := time.Now()
			task.FinishedTime = &now
			m.tasks[msg.TaskId] = task
			clear := tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return constants.ClearTaskMsg{TaskId: msg.TaskId}
			})
			cmds = append(cmds, clear)

			scmd := m.updateSection(msg.SectionId, msg.SectionType, msg.Msg)
			cmds = append(cmds, scmd)

			syncCmd := m.syncSidebar()
			cmds = append(cmds, syncCmd)
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

	case section.SectionMsg:
		cmd = m.updateRelevantSection(msg)

		if msg.Id == m.currSectionId {
			m.onViewedRowChanged()
		}

	case execProcessFinishedMsg, tea.FocusMsg:
		if currSection != nil {
			cmds = append(cmds, currSection.FetchNextPageSectionRows()...)
		}

	case tea.WindowSizeMsg:
		m.onWindowSizeChanged(msg)

	case updateFooterMsg:
		m.footer, cmd = m.footer.Update(msg)
		cmds = append(cmds, cmd, m.doUpdateFooterAtInterval())

	case constants.ErrMsg:
		m.ctx.Error = msg.Err
	}

	m.syncProgramContext()

	var bsCmd tea.Cmd
	m.branchSidebar, bsCmd = m.branchSidebar.Update(msg)
	cmds = append(cmds, bsCmd)

	m.sidebar, sidebarCmd = m.sidebar.Update(msg)

	if m.prSidebar.IsTextInputBoxFocused() {
		m.prSidebar, prSidebarCmd = m.prSidebar.Update(msg)
		m.syncSidebar()
	}

	if m.issueSidebar.IsTextInputBoxFocused() {
		m.issueSidebar, issueSidebarCmd = m.issueSidebar.Update(msg)
		m.syncSidebar()
	}

	m.footer, footerCmd = m.footer.Update(msg)
	if currSection != nil {
		if currSection.IsPromptConfirmationFocused() {
			m.footer.SetLeftSection(currSection.GetPromptConfirmation())
		}

		if !currSection.IsPromptConfirmationFocused() {
			m.footer.SetLeftSection(currSection.GetPagerContent())
		}
	}

	m.tabs.UpdateSectionCounts(m.getCurrentViewSections())

	sectionCmd := m.updateCurrentSection(msg)
	cmds = append(
		cmds,
		cmd,
		tabsCmd,
		sidebarCmd,
		footerCmd,
		sectionCmd,
		prSidebarCmd,
		issueSidebarCmd,
	)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.ctx.Config == nil {
		return "Reading config...\n"
	}

	s := strings.Builder{}
	if m.ctx.View != config.RepoView {
		s.WriteString(m.tabs.View(m.ctx))
	}
	s.WriteString("\n")
	content := "No sections defined"
	currSection := m.getCurrSection()
	if currSection != nil {
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.getCurrSection().View(),
			m.sidebar.View(),
		)
	}
	s.WriteString(content)
	s.WriteString("\n")
	if m.ctx.Error != nil {
		s.WriteString(
			m.ctx.Styles.Common.ErrorStyle.
				Width(m.ctx.ScreenWidth).
				Render(fmt.Sprintf("%s %s",
					m.ctx.Styles.Common.FailureGlyph,
					lipgloss.NewStyle().
						Foreground(m.ctx.Theme.ErrorText).
						Render(m.ctx.Error.Error()),
				)),
		)
	} else {
		s.WriteString(m.footer.View())
	}

	return s.String()
}

type initMsg struct {
	Config  config.Config
	RepoUrl string
}

func (m *Model) setCurrSectionId(newSectionId int) {
	m.currSectionId = newSectionId
	m.tabs.SetCurrSectionId(newSectionId)
}

func (m *Model) onViewedRowChanged() tea.Cmd {
	cmd := m.syncSidebar()
	m.sidebar.ScrollToTop()
	return cmd
}

func (m *Model) onWindowSizeChanged(msg tea.WindowSizeMsg) {
	m.footer.SetWidth(msg.Width)
	m.ctx.ScreenWidth = msg.Width
	m.ctx.ScreenHeight = msg.Height
	if m.footer.ShowAll {
		m.ctx.MainContentHeight = msg.Height - common.TabsHeight - common.ExpandedHelpHeight
	} else {
		m.ctx.MainContentHeight = msg.Height - common.TabsHeight - common.FooterHeight
	}
	m.syncMainContentWidth()
}

func (m *Model) syncProgramContext() {
	for _, section := range m.getCurrentViewSections() {
		section.UpdateProgramContext(m.ctx)
	}
	m.footer.UpdateProgramContext(m.ctx)
	m.sidebar.UpdateProgramContext(m.ctx)
	m.prSidebar.UpdateProgramContext(m.ctx)
	m.issueSidebar.UpdateProgramContext(m.ctx)
	m.branchSidebar.UpdateProgramContext(m.ctx)
}

func (m *Model) updateSection(id int, sType string, msg tea.Msg) (cmd tea.Cmd) {
	var updatedSection section.Section
	switch sType {
	case reposection.SectionType:
		m.repo, cmd = m.repo.Update(msg)

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

func (m *Model) syncSidebar() tea.Cmd {
	currRowData := m.getCurrRowData()
	width := m.sidebar.GetSidebarContentWidth()
	var cmd tea.Cmd

	if currRowData == nil {
		m.sidebar.SetContent("")
		return nil
	}

	switch row := currRowData.(type) {
	case branch.BranchData:
		cmd = m.branchSidebar.SetRow(&row)
		m.sidebar.SetContent(m.branchSidebar.View())
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

	return cmd
}

func (m *Model) fetchAllViewSections() ([]section.Section, tea.Cmd) {
	if m.ctx.View == config.RepoView {
		var cmd tea.Cmd
		s, cmd := reposection.FetchAllBranches(m.ctx)
		m.repo = &s
		return nil, cmd
	} else if m.ctx.View == config.PRsView {
		return prssection.FetchAllSections(m.ctx, m.prs)
	} else {
		return issuessection.FetchAllSections(m.ctx)
	}
}

func (m *Model) getCurrentViewSections() []section.Section {
	if m.ctx.View == config.RepoView {
		return []section.Section{m.repo}
	} else if m.ctx.View == config.PRsView {
		return m.prs
	} else {
		return m.issues
	}
}

func (m *Model) getCurrentViewDefaultSection() int {
	if m.ctx.View == config.RepoView {
		return 0
	} else if m.ctx.View == config.PRsView {
		return 1
	} else {
		return 1
	}
}

func (m *Model) setCurrentViewSections(newSections []section.Section) {
	if newSections == nil {
		return
	}
	if m.ctx.View == config.PRsView {
		search := prssection.NewModel(
			0,
			m.ctx,
			config.PrsSectionConfig{
				Title:   "",
				Filters: "archived:false",
			},
			time.Now(),
		)
		m.prs = append([]section.Section{&search}, newSections...)
	} else {
		search := issuessection.NewModel(
			0,
			m.ctx,
			config.IssuesSectionConfig{
				Title:   "",
				Filters: "",
			},
			time.Now(),
		)
		m.issues = append([]section.Section{&search}, newSections...)
	}
}

func (m *Model) switchSelectedView() config.ViewType {
	repoFF := config.IsFeatureEnabled(config.FF_REPO_VIEW)

	if repoFF {
		switch true {
		case m.ctx.View == config.RepoView:
			return config.PRsView
		case m.ctx.View == config.PRsView:
			return config.IssuesView
		case m.ctx.View == config.IssuesView:
			return config.RepoView
		}
	}

	switch true {
	case m.ctx.View == config.PRsView:
		return config.IssuesView
	default:
		return config.PRsView
	}
}

func (m *Model) isUserDefinedKeybinding(msg tea.KeyMsg) bool {
	for _, keybinding := range m.ctx.Config.Keybindings.Universal {
		if keybinding.Builtin == "" && keybinding.Key == msg.String() {
			return true
		}
	}

	if m.ctx.View == config.IssuesView {
		for _, keybinding := range m.ctx.Config.Keybindings.Issues {
			if keybinding.Builtin == "" && keybinding.Key == msg.String() {
				return true
			}
		}
	}

	if m.ctx.View == config.PRsView {
		for _, keybinding := range m.ctx.Config.Keybindings.Prs {
			if keybinding.Builtin == "" && keybinding.Key == msg.String() {
				return true
			}
		}
	}

	if m.ctx.View == config.RepoView {
		for _, keybinding := range m.ctx.Config.Keybindings.Branches {
			if keybinding.Builtin == "" && keybinding.Key == msg.String() {
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
		currTaskStatus = lipgloss.NewStyle().
			Background(m.ctx.Theme.SelectedBackground).
			Render(
				fmt.Sprintf(
					"%s%s",
					m.taskSpinner.View(),
					task.StartText,
				))
	case context.TaskError:
		currTaskStatus = lipgloss.NewStyle().
			Foreground(m.ctx.Theme.ErrorText).
			Background(m.ctx.Theme.SelectedBackground).
			Render(fmt.Sprintf("%s %s", constants.FailureIcon, task.Error.Error()))
	case context.TaskFinished:
		currTaskStatus = lipgloss.NewStyle().
			Foreground(m.ctx.Theme.SuccessText).
			Background(m.ctx.Theme.SelectedBackground).
			Render(fmt.Sprintf("%s %s", constants.SuccessIcon, task.FinishedText))
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
		MaxHeight(1).
		Background(m.ctx.Theme.SelectedBackground).
		Render(strings.TrimSpace(lipgloss.JoinHorizontal(lipgloss.Top, stats, currTaskStatus)))
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
	if m.ctx.Config.Defaults.RefetchIntervalMinutes == 0 {
		return nil
	}

	return tea.Tick(
		time.Minute*time.Duration(m.ctx.Config.Defaults.RefetchIntervalMinutes),
		func(t time.Time) tea.Msg {
			return intervalRefresh(t)
		},
	)
}

type updateFooterMsg struct{}

func (m *Model) doUpdateFooterAtInterval() tea.Cmd {
	return tea.Tick(
		time.Second*10,
		func(t time.Time) tea.Msg {
			return updateFooterMsg{}
		},
	)
}
