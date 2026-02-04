package tui

import (
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/charmbracelet/log"
	"github.com/cli/go-gh/v2/pkg/browser"
	zone "github.com/lrstanley/bubblezone"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/git"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/branch"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/branchsidebar"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/footer"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/issuessection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/issueview"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationview"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prview"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/reposection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/sidebar"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tabs"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

type Model struct {
	keys             *keys.KeyMap
	sidebar          sidebar.Model
	prView           prview.Model
	issueSidebar     issueview.Model
	branchSidebar    branchsidebar.Model
	notificationView notificationview.Model
	currSectionId    int
	footer           footer.Model
	repo             section.Section
	prs              []section.Section
	issues           []section.Section
	notifications    []section.Section
	tabs             tabs.Model
	ctx              *context.ProgramContext
	taskSpinner      spinner.Model
	tasks            map[string]context.Task
}

func NewModel(location config.Location) Model {
	taskSpinner := spinner.Model{Spinner: spinner.Dot}
	m := Model{
		keys:        keys.Keys,
		sidebar:     sidebar.NewModel(),
		taskSpinner: taskSpinner,
		tasks:       map[string]context.Task{},
	}

	version := "dev"
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		version = info.Main.Version
	}

	m.ctx = &context.ProgramContext{
		RepoPath:   location.RepoPath,
		ConfigFlag: location.ConfigFlag,
		Version:    version,
		StartTask: func(task context.Task) tea.Cmd {
			log.Info("Starting task", "id", task.Id)
			task.StartTime = time.Now()
			m.tasks[task.Id] = task
			rTask := m.renderRunningTask()
			m.footer.SetRightSection(rTask)
			return m.taskSpinner.Tick
		},
	}

	m.taskSpinner.Style = lipgloss.NewStyle().
		Background(m.ctx.Theme.SelectedBackground)

	m.footer = footer.NewModel(m.ctx)
	m.prView = prview.NewModel(m.ctx)
	m.issueSidebar = issueview.NewModel(m.ctx)
	m.branchSidebar = branchsidebar.NewModel(m.ctx)
	m.notificationView = notificationview.NewModel(m.ctx)
	m.notificationView.SetOnConfirmAction(m.executeNotificationAction)
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
				m.ctx.ConfigFlag,
				"err",
				err,
			)
	}

	cfg, err := config.ParseConfig(config.Location{RepoPath: m.ctx.RepoPath, ConfigFlag: m.ctx.ConfigFlag})
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
		prViewCmd       tea.Cmd
		issueSidebarCmd tea.Cmd
		footerCmd       tea.Cmd
		cmds            []tea.Cmd
		currSection     = m.getCurrSection()
		currRowData     = m.getCurrRowData()
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		log.Info("Key pressed", "key", msg.String())
		m.ctx.Error = nil

		if currSection != nil && (currSection.IsSearchFocused() ||
			currSection.IsPromptConfirmationFocused()) {
			cmd = m.updateSection(currSection.GetId(), currSection.GetType(), msg)
			return m, cmd
		}

		if m.prView.IsTextInputBoxFocused() {
			m.prView, cmd = m.prView.Update(msg)
			m.syncSidebar()
			return m, cmd
		}

		if m.issueSidebar.IsTextInputBoxFocused() {
			m.issueSidebar, cmd, _ = m.issueSidebar.Update(msg)
			m.syncSidebar()
			return m, cmd
		}

		if m.footer.ShowConfirmQuit && (msg.String() == "y" || msg.String() == "enter") {
			return m, tea.Quit
		} else if m.footer.ShowConfirmQuit {
			m.footer.SetShowConfirmQuit(false)
			return m, nil
		}

		// Handle notification PR/Issue action confirmation
		if m.notificationView.HasPendingAction() {
			m.notificationView, cmd = m.notificationView.Update(msg)
			m.footer.SetLeftSection("")
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
				cmd = m.onViewedRowChanged()
			}

		case key.Matches(msg, m.keys.NextSection):
			nextSectionId := m.getNextSectionId()
			nextSection := m.getSectionAt(nextSectionId)
			if nextSection != nil {
				m.setCurrSectionId(nextSection.GetId())
				cmd = m.onViewedRowChanged()
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
			cmd = m.onViewedRowChanged()

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
			data.ClearEnrichmentCache()
			currSection.ResetFilters()
			currSection.ResetRows()
			m.syncSidebar()
			currSection.SetIsLoading(true)
			cmds = append(cmds, currSection.FetchNextPageSectionRows()...)

		case key.Matches(msg, m.keys.RefreshAll):
			data.ClearEnrichmentCache()
			newSections, fetchSectionsCmds := m.fetchAllViewSections()
			m.setCurrentViewSections(newSections)
			cmds = append(cmds, fetchSectionsCmds)

		case key.Matches(msg, m.keys.Redraw):
			// can't find a way to just ask to send bubbletea's internal repaintMsg{},
			// so this seems like the lightest-weight alternative
			return m, tea.Batch(tea.ExitAltScreen, tea.EnterAltScreen)

		case key.Matches(msg, m.keys.Search):
			if currSection != nil {
				cmd = currSection.SetIsSearching(true)
				return m, cmd
			}

		case key.Matches(msg, m.keys.Help):
			if !m.footer.ShowAll {
				m.ctx.MainContentHeight = m.ctx.MainContentHeight +
					common.FooterHeight - common.ExpandedHelpHeight
			} else {
				m.ctx.MainContentHeight = m.ctx.MainContentHeight +
					common.ExpandedHelpHeight - common.FooterHeight
			}
			m.footer.ShowAll = !m.footer.ShowAll

		case key.Matches(msg, m.keys.CopyNumber):
			var cmd tea.Cmd
			if currRowData == nil || reflect.ValueOf(currRowData).IsNil() {
				cmd = m.notifyErr("Current selection isn't associated with a PR/Issue")
				return m, cmd
			}
			number := fmt.Sprint(currRowData.GetNumber())
			err := clipboard.WriteAll(number)
			if err != nil {
				cmd = m.notifyErr(fmt.Sprintf("Failed copying to clipboard %v", err))
			} else {
				cmd = m.notify(fmt.Sprintf("Copied %s to clipboard", number))
			}
			return m, cmd

		case key.Matches(msg, m.keys.CopyUrl):
			var cmd tea.Cmd
			if currRowData == nil || reflect.ValueOf(currRowData).IsNil() {
				cmd = m.notifyErr("Current selection isn't associated with a PR/Issue")
				return m, cmd
			}
			url := currRowData.GetUrl()
			err := clipboard.WriteAll(url)
			if err != nil {
				cmd = m.notifyErr(fmt.Sprintf("Failed copying to clipboard %v", err))
			} else {
				cmd = m.notify(fmt.Sprintf("Copied %s to clipboard", url))
			}
			return m, cmd

		case key.Matches(msg, m.keys.Quit):
			if !m.ctx.Config.ConfirmQuit {
				return m, tea.Quit
			}

			m.footer.SetShowConfirmQuit(true)

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
				cmds = append(cmds, m.switchSelectedView())
			}
		case m.ctx.View == config.PRsView:
			switch {
			case key.Matches(msg, keys.PRKeys.PrevSidebarTab),
				key.Matches(msg, keys.PRKeys.NextSidebarTab):
				var scmds []tea.Cmd
				var scmd tea.Cmd
				m.prView, scmd = m.prView.Update(msg)
				scmds = append(scmds, scmd)
				m.syncSidebar()
				return m, tea.Batch(scmds...)

			case key.Matches(msg, m.keys.OpenGithub):
				cmds = append(cmds, m.openBrowser())

			case key.Matches(msg, keys.PRKeys.Approve):
				return m, m.openSidebarForPRInput(m.prView.SetIsApproving)

			case key.Matches(msg, keys.PRKeys.Assign):
				return m, m.openSidebarForPRInput(m.prView.SetIsAssigning)

			case key.Matches(msg, keys.PRKeys.Unassign):
				return m, m.openSidebarForPRInput(m.prView.SetIsUnassigning)

			case key.Matches(msg, keys.PRKeys.Comment):
				return m, m.openSidebarForPRInput(m.prView.SetIsCommenting)

			case key.Matches(msg, keys.PRKeys.Close):
				if currRowData != nil {
					cmd = m.promptConfirmation(currSection, "close")
				}
				return m, cmd

			case key.Matches(msg, keys.PRKeys.Ready):
				if currRowData != nil {
					cmd = m.promptConfirmation(currSection, "ready")
				}
				return m, cmd

			case key.Matches(msg, keys.PRKeys.Reopen):
				if currRowData != nil {
					cmd = m.promptConfirmation(currSection, "reopen")
				}
				return m, cmd

			case key.Matches(msg, keys.PRKeys.Merge):
				if currRowData != nil {
					cmd = m.promptConfirmation(currSection, "merge")
				}
				return m, cmd

			case key.Matches(msg, keys.PRKeys.Update):
				if currRowData != nil {
					cmd = m.promptConfirmation(currSection, "update")
				}
				return m, cmd

			case key.Matches(msg, keys.PRKeys.ViewIssues):
				cmds = append(cmds, m.switchSelectedView())

			case key.Matches(msg, keys.PRKeys.SummaryViewMore):
				m.prView.SetSummaryViewMore()
				m.syncSidebar()
				return m, nil
			}
		case m.ctx.View == config.IssuesView:
			switch {
			case key.Matches(msg, m.keys.OpenGithub):
				cmds = append(cmds, m.openBrowser())

			case key.Matches(msg, keys.IssueKeys.Label):
				return m, m.openSidebarForInput(m.issueSidebar.SetIsLabeling)

			case key.Matches(msg, keys.IssueKeys.Assign):
				return m, m.openSidebarForInput(m.issueSidebar.SetIsAssigning)

			case key.Matches(msg, keys.IssueKeys.Unassign):
				return m, m.openSidebarForInput(m.issueSidebar.SetIsUnassigning)

			case key.Matches(msg, keys.IssueKeys.Comment):
				return m, m.openSidebarForInput(m.issueSidebar.SetIsCommenting)

			case key.Matches(msg, keys.IssueKeys.Close):
				if currRowData != nil {
					cmd = m.promptConfirmation(currSection, "close")
				}
				return m, cmd

			case key.Matches(msg, keys.IssueKeys.Reopen):
				if currRowData != nil {
					cmd = m.promptConfirmation(currSection, "reopen")
				}
				return m, cmd

			case key.Matches(msg, keys.IssueKeys.ViewPRs):
				cmds = append(cmds, m.switchSelectedView())
			}
		case m.ctx.View == config.NotificationsView:
			switch {
			case key.Matches(msg, m.keys.OpenGithub):
				cmds = append(cmds, m.openBrowser())
				return m, tea.Batch(cmds...)

			// PR keybindings when viewing a PR notification
			case m.notificationView.GetSubjectPR() != nil:
				// Check for PR actions first (before updating prView)
				if !m.prView.IsTextInputBoxFocused() {
					action := prview.MsgToAction(msg)
					if action != nil {
						switch action.Type {
						case prview.PRActionApprove:
							return m, m.openSidebarForPRInput(m.prView.SetIsApproving)

						case prview.PRActionAssign:
							return m, m.openSidebarForPRInput(m.prView.SetIsAssigning)

						case prview.PRActionUnassign:
							return m, m.openSidebarForPRInput(m.prView.SetIsUnassigning)

						case prview.PRActionComment:
							return m, m.openSidebarForPRInput(m.prView.SetIsCommenting)

						case prview.PRActionDiff:
							if pr := m.notificationView.GetSubjectPR(); pr != nil {
								cmd = common.DiffPR(pr.GetNumber(), pr.GetRepoNameWithOwner(),
									m.ctx.Config.GetFullScreenDiffPagerEnv())
							}
							return m, cmd

						case prview.PRActionCheckout:
							if pr := m.notificationView.GetSubjectPR(); pr != nil {
								cmd, _ = notificationssection.CheckoutPR(
									m.ctx, pr.GetNumber(), pr.GetRepoNameWithOwner())
							}
							return m, cmd

						case prview.PRActionClose:
							cmd = m.promptConfirmationForNotificationPR("close")
							return m, cmd

						case prview.PRActionReady:
							cmd = m.promptConfirmationForNotificationPR("ready")
							return m, cmd

						case prview.PRActionReopen:
							cmd = m.promptConfirmationForNotificationPR("reopen")
							return m, cmd

						case prview.PRActionMerge:
							cmd = m.promptConfirmationForNotificationPR("merge")
							return m, cmd

						case prview.PRActionUpdate:
							cmd = m.promptConfirmationForNotificationPR("update")
							return m, cmd

						case prview.PRActionSummaryViewMore:
							m.prView.SetSummaryViewMore()
							m.syncSidebar()
							return m, nil
						}
					}
				}

				// Handle 's' key to switch views
				if key.Matches(msg, keys.PRKeys.ViewIssues) {
					cmds = append(cmds, m.switchSelectedView())
				}

				// No action matched - update prView for navigation (tab switching, scrolling)
				var prCmd tea.Cmd
				m.prView, prCmd = m.prView.Update(msg)
				m.syncSidebar()
				cmds = append(cmds, prCmd)

			// Issue keybindings when viewing an Issue notification
			case m.notificationView.GetSubjectIssue() != nil:
				var issueCmd tea.Cmd
				var action *issueview.IssueAction
				m.issueSidebar, issueCmd, action = m.issueSidebar.Update(msg)

				if action != nil {
					switch action.Type {
					case issueview.IssueActionLabel:
						return m, m.openSidebarForInput(m.issueSidebar.SetIsLabeling)

					case issueview.IssueActionAssign:
						return m, m.openSidebarForInput(m.issueSidebar.SetIsAssigning)

					case issueview.IssueActionUnassign:
						return m, m.openSidebarForInput(m.issueSidebar.SetIsUnassigning)

					case issueview.IssueActionComment:
						return m, m.openSidebarForInput(m.issueSidebar.SetIsCommenting)

					case issueview.IssueActionClose:
						cmd = m.promptConfirmationForNotificationIssue("close")
						return m, cmd

					case issueview.IssueActionReopen:
						cmd = m.promptConfirmationForNotificationIssue("reopen")
						return m, cmd
					}
				}

				// Handle 's' key to switch views
				if key.Matches(msg, keys.IssueKeys.ViewPRs) {
					cmds = append(cmds, m.switchSelectedView())
				}

				// Sync sidebar and return issueCmd for navigation
				m.syncSidebar()
				cmds = append(cmds, issueCmd)

			// Notification-specific keybindings (when not viewing PR/Issue content)
			case key.Matches(msg, keys.NotificationKeys.View):
				cmds = append(cmds, m.loadNotificationContent())

			case key.Matches(msg, keys.NotificationKeys.MarkAsDone):
				cmds = append(cmds, m.updateSection(currSection.GetId(), currSection.GetType(), msg))

			case key.Matches(msg, keys.NotificationKeys.MarkAllAsDone):
				cmd = m.promptConfirmation(currSection, "done_all")
				return m, cmd

			case key.Matches(msg, keys.NotificationKeys.Open):
				cmd = m.updateSection(currSection.GetId(), currSection.GetType(), msg)
				return m, cmd

			case key.Matches(msg, keys.NotificationKeys.SortByRepo):
				cmd = m.updateSection(currSection.GetId(), currSection.GetType(), msg)
				return m, cmd

			case key.Matches(msg, keys.PRKeys.ViewIssues):
				cmds = append(cmds, m.switchSelectedView())
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
		m.syncMainContentWidth()

		newSections, fetchSectionsCmds := m.fetchAllViewSections()
		m.setCurrentViewSections(newSections)
		m.tabs.SetCurrSectionId(1)
		cmds = append(cmds, fetchSectionsCmds, m.tabs.Init(), fetchUser,
			m.doRefreshAtInterval(), m.doUpdateFooterAtInterval())

	case intervalRefresh:
		newSections, fetchSectionsCmds := m.fetchAllViewSections()
		m.setCurrentViewSections(newSections)
		cmds = append(cmds, fetchSectionsCmds, m.doRefreshAtInterval())

	case userFetchedMsg:
		m.ctx.User = msg.user

	case constants.TaskFinishedMsg:
		task, ok := m.tasks[msg.TaskId]
		if ok {
			log.Info("Task finished", "id", task.Id)
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

	case prview.EnrichedPrMsg:
		if msg.Err == nil {
			m.prView.SetEnrichedPR(msg.Data)
			m.prs[msg.Id].(*prssection.Model).EnrichPR(msg.Data)
			syncCmd := m.syncSidebar()
			cmds = append(cmds, syncCmd)
		} else {
			log.Error("failed enriching pr", "err", msg.Err)
		}

	case notificationPRFetchedMsg:
		if msg.Err == nil {
			// Convert enriched PR to prrow.Data for display
			prData := msg.PR.ToPullRequestData()
			m.notificationView.SetSubjectPR(&prrow.Data{
				Primary:    &prData,
				Enriched:   msg.PR,
				IsEnriched: true,
			}, msg.NotificationId)
			keys.SetNotificationSubject(keys.NotificationSubjectPR)
			// Update sidebar with PR view
			width := m.sidebar.GetSidebarContentWidth()
			m.prView.SetSectionId(0)
			m.prView.SetRow(m.notificationView.GetSubjectPR())
			m.prView.SetWidth(width)
			m.prView.SetEnrichedPR(msg.PR)
			// Switch to Activity tab and scroll to bottom if there's a latest comment
			// (indicates there's new activity to show)
			if msg.LatestCommentUrl != "" {
				m.prView.GoToActivityTab()
				m.sidebar.SetContent(m.prView.View())
				m.sidebar.ScrollToBottom()
			} else {
				// For notifications without comments (new PRs, state changes, etc.)
				// show the Overview tab without scrolling
				m.prView.GoToFirstTab()
				m.sidebar.SetContent(m.prView.View())
			}
			m.markNotificationAsRead(msg.NotificationId)
		} else {
			log.Error("failed fetching notification PR", "err", msg.Err)
		}

	case notificationIssueFetchedMsg:
		if msg.Err == nil {
			m.notificationView.SetSubjectIssue(&msg.Issue, msg.NotificationId)
			keys.SetNotificationSubject(keys.NotificationSubjectIssue)
			// Update sidebar with Issue view
			width := m.sidebar.GetSidebarContentWidth()
			m.issueSidebar.SetSectionId(0)
			m.issueSidebar.SetRow(m.notificationView.GetSubjectIssue())
			m.issueSidebar.SetWidth(width)
			m.sidebar.SetContent(m.issueSidebar.View())
			// Scroll to bottom if there's a latest comment (indicates new activity)
			if msg.LatestCommentUrl != "" {
				m.sidebar.ScrollToBottom()
			}
			m.markNotificationAsRead(msg.NotificationId)
		} else {
			log.Error("failed fetching notification Issue", "err", msg.Err)
		}

	case notificationssection.UpdateNotificationReadStateMsg:
		m.updateNotificationSections(msg)

	case notificationssection.UpdateNotificationCommentsMsg:
		cmds = append(cmds, m.updateNotificationSections(msg))

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
			cmds = append(cmds, m.onViewedRowChanged())
		}

	case execProcessFinishedMsg, tea.FocusMsg:
		if currSection != nil {
			cmds = append(cmds, currSection.FetchNextPageSectionRows()...)
		}

	case tea.MouseMsg:
		if msg.Action != tea.MouseActionRelease || msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		if zone.Get("donate").InBounds(msg) {
			log.Info("Donate clicked", "msg", msg)
			openCmd := func() tea.Msg {
				b := browser.New("", os.Stdout, os.Stdin)
				err := b.Browse("https://github.com/sponsors/dlvhdr")
				if err != nil {
					return constants.ErrMsg{Err: err}
				}
				return nil
			}
			cmds = append(cmds, openCmd)
		}

	case tea.WindowSizeMsg:
		m.onWindowSizeChanged(msg)

	case updateFooterMsg:
		cmds = append(cmds, cmd, m.doUpdateFooterAtInterval())

	case constants.ErrMsg:
		m.ctx.Error = msg.Err
	}

	m.syncProgramContext()

	var bsCmd tea.Cmd
	m.branchSidebar, bsCmd = m.branchSidebar.Update(msg)
	cmds = append(cmds, bsCmd)

	m.sidebar, sidebarCmd = m.sidebar.Update(msg)

	if m.prView.IsTextInputBoxFocused() {
		m.prView, prViewCmd = m.prView.Update(msg)
		m.syncSidebar()
	}

	if m.issueSidebar.IsTextInputBoxFocused() {
		m.issueSidebar, issueSidebarCmd, _ = m.issueSidebar.Update(msg)
		m.syncSidebar()
	}

	if currSection != nil {
		if currSection.IsPromptConfirmationFocused() {
			m.footer.SetLeftSection(currSection.GetPromptConfirmation())
		}

		if !currSection.IsPromptConfirmationFocused() {
			m.footer.SetLeftSection(currSection.GetPagerContent())
		}
	}

	tm, tabsCmd := m.tabs.Update(msg)
	m.tabs = tm.(tabs.Model)

	sectionCmd := m.updateCurrentSection(msg)
	cmds = append(
		cmds,
		cmd,
		tabsCmd,
		sidebarCmd,
		footerCmd,
		sectionCmd,
		prViewCmd,
		issueSidebarCmd,
	)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.ctx.Config == nil {
		return lipgloss.Place(m.ctx.ScreenWidth, m.ctx.ScreenHeight, lipgloss.Center, lipgloss.Center, "Reading config...")
	}

	s := strings.Builder{}
	if m.ctx.View != config.RepoView {
		s.WriteString(m.tabs.View())
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

	return zone.Scan(s.String())
}

type initMsg struct {
	Config  config.Config
	RepoUrl string
}

// Message types for notification subject fetching
type notificationPRFetchedMsg struct {
	NotificationId   string
	PR               data.EnrichedPullRequestData
	LatestCommentUrl string
	Err              error
}

type notificationIssueFetchedMsg struct {
	NotificationId   string
	Issue            data.IssueData
	LatestCommentUrl string
	Err              error
}

func (m *Model) setCurrSectionId(newSectionId int) {
	m.currSectionId = newSectionId
	m.tabs.SetCurrSectionId(newSectionId)
}

func (m *Model) updateNotificationSections(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for i := range m.notifications {
		if m.notifications[i] != nil {
			var cmd tea.Cmd
			m.notifications[i], cmd = m.notifications[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (m *Model) markNotificationAsRead(notificationId string) {
	readStateMsg := notificationssection.UpdateNotificationReadStateMsg{
		Id:     notificationId,
		Unread: false,
	}
	m.updateNotificationSections(readStateMsg)
}

func (m *Model) onViewedRowChanged() tea.Cmd {
	m.prView.SetSummaryViewLess()
	m.prView.GoToFirstTab()
	sidebarCmd := m.syncSidebar()
	enrichCmd := m.prView.EnrichCurrRow()
	m.sidebar.ScrollToTop()
	m.notificationView.ResetSubject()
	return tea.Batch(sidebarCmd, enrichCmd)
}

func (m *Model) onWindowSizeChanged(msg tea.WindowSizeMsg) {
	log.Info("window size changed", "width", msg.Width, "height", msg.Height)
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
	m.tabs.UpdateProgramContext(m.ctx)
	m.footer.UpdateProgramContext(m.ctx)
	m.sidebar.UpdateProgramContext(m.ctx)
	m.prView.UpdateProgramContext(m.ctx)
	m.issueSidebar.UpdateProgramContext(m.ctx)
	m.branchSidebar.UpdateProgramContext(m.ctx)
	m.notificationView.UpdateProgramContext(m.ctx)
}

func (m *Model) updateSection(id int, sType string, msg tea.Msg) (cmd tea.Cmd) {
	var updatedSection section.Section
	switch sType {
	case reposection.SectionType:
		m.repo, cmd = m.repo.Update(msg)

	case notificationssection.SectionType:
		if id < len(m.notifications) && m.notifications[id] != nil {
			m.notifications[id], cmd = m.notifications[id].Update(msg)
		}

	case prssection.SectionType:
		updatedSection, cmd = m.prs[id].Update(msg)
		m.prs[id] = updatedSection
	case issuessection.SectionType:
		updatedSection, cmd = m.issues[id].Update(msg)
		m.issues[id] = updatedSection
	}

	currSection := m.getCurrSection()
	if currSection != nil && id == currSection.GetId() {
		if _, ok := msg.(prssection.SectionPullRequestsFetchedMsg); ok {
			cmd = m.onViewedRowChanged()
		}
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
	m.ctx.SidebarOpen = m.sidebar.IsOpen
}

func (m *Model) openSidebarForPRInput(setFunc func(bool) tea.Cmd) tea.Cmd {
	m.prView.GoToFirstTab()
	return m.openSidebarForInput(setFunc)
}

func (m *Model) openSidebarForInput(setFunc func(bool) tea.Cmd) tea.Cmd {
	m.sidebar.IsOpen = true
	cmd := setFunc(true)
	m.syncMainContentWidth()
	m.syncSidebar()
	m.sidebar.ScrollToBottom()
	return cmd
}

func (m *Model) promptConfirmation(currSection section.Section, action string) tea.Cmd {
	if currSection != nil {
		currSection.SetPromptConfirmationAction(action)
		return currSection.SetIsPromptConfirmationShown(true)
	}
	return nil
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
	case *prrow.Data:
		m.prView.SetSectionId(m.currSectionId)
		m.prView.SetRow(row)
		m.prView.SetWidth(width)
		m.sidebar.SetContent(m.prView.View())
	case *data.IssueData:
		m.issueSidebar.SetSectionId(m.currSectionId)
		m.issueSidebar.SetRow(row)
		m.issueSidebar.SetWidth(width)
		m.sidebar.SetContent(m.issueSidebar.View())
	case *notificationrow.Data:
		notifId := row.GetId()

		// Check if we already have cached data for this notification (user already viewed it)
		if m.notificationView.GetSubjectId() == notifId {
			// Use cached data
			if m.notificationView.GetSubjectPR() != nil {
				m.prView.SetSectionId(0)
				m.prView.SetRow(m.notificationView.GetSubjectPR())
				m.prView.SetWidth(width)
				m.sidebar.SetContent(m.prView.View())
			} else if m.notificationView.GetSubjectIssue() != nil {
				m.issueSidebar.SetSectionId(0)
				m.issueSidebar.SetRow(m.notificationView.GetSubjectIssue())
				m.issueSidebar.SetWidth(width)
				m.sidebar.SetContent(m.issueSidebar.View())
			}
			return nil
		}

		// Show prompt to view notification (don't auto-fetch)
		// User must press Enter to view content and mark as read
		m.sidebar.SetContent(m.renderNotificationPrompt(row, width))
	}

	return cmd
}

func (m *Model) renderNotificationPrompt(row *notificationrow.Data, width int) string {
	var content strings.Builder

	subjectType := row.GetSubjectType()
	leftMargin := "      " // Left margin for content

	// Styles
	normalText := lipgloss.NewStyle().Foreground(m.ctx.Theme.PrimaryText)
	faintText := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText)
	// Highlighted key style for main prompt (with background)
	highlightKeyStyle := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.PrimaryText).
		Background(m.ctx.Theme.FaintBorder).
		Padding(0, 1)
	// Simple key style for table (no background)
	keyStyle := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.PrimaryText)
	actionStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.SuccessText)
	headerStyle := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.PrimaryText).
		Bold(true)

	// Determine subject type display name and primary action
	typeName := "PR"
	enterAction := "view"
	if subjectType == "Issue" {
		typeName = "Issue"
	} else if subjectType != "PullRequest" {
		typeName = subjectType
		enterAction = "open in browser"
	}

	// Main prompt: "Press Enter to view the PR" or "Press Enter to open in browser"
	content.WriteString("\n")
	content.WriteString(leftMargin)
	content.WriteString(normalText.Render("Press "))
	content.WriteString(highlightKeyStyle.Render("Enter"))
	if enterAction == "view" {
		content.WriteString(normalText.Render(fmt.Sprintf(" to %s the %s", enterAction, typeName)))
	} else {
		content.WriteString(normalText.Render(fmt.Sprintf(" to %s", enterAction)))
	}
	content.WriteString("\n")

	// Note about marking as read
	content.WriteString(leftMargin)
	content.WriteString(faintText.Render("(Note: this will mark it as read)"))
	content.WriteString("\n")

	content.WriteString("\n")

	// Other Actions header
	content.WriteString(leftMargin)
	content.WriteString(headerStyle.Render("Other Actions"))
	content.WriteString("\n\n")

	// Key-action pairs (simple list without borders)
	actions := []struct {
		key    string
		action string
	}{
		{"D", "mark as done"},
		{"m", "mark as read"},
		{"u", "unsubscribe"},
		{"b", "toggle bookmark"},
		{"t", "toggle filtering"},
		{"S", "sort by repo"},
		{"o", "open in browser"},
	}

	keyWidth := 7 // Width for key column
	for _, a := range actions {
		content.WriteString(leftMargin)
		// Right-align the key in its column
		padding := strings.Repeat(" ", keyWidth-len(a.key))
		content.WriteString(padding)
		content.WriteString(keyStyle.Render(a.key))
		content.WriteString("  ")
		content.WriteString(actionStyle.Render(a.action))
		content.WriteString("\n")
	}

	// Add Enter at the end
	content.WriteString(leftMargin)
	padding := strings.Repeat(" ", keyWidth-len("Enter"))
	content.WriteString(padding)
	content.WriteString(keyStyle.Render("Enter"))
	content.WriteString("  ")
	content.WriteString(actionStyle.Render(enterAction))

	return content.String()
}

// loadNotificationContent fetches and displays notification content, marking it as read
func (m *Model) loadNotificationContent() tea.Cmd {
	currRowData := m.getCurrRowData()
	row, ok := currRowData.(*notificationrow.Data)
	if !ok || row == nil {
		return nil
	}

	notifId := row.GetId()
	subjectType := row.GetSubjectType()
	subjectUrl := row.GetUrl()
	latestCommentUrl := row.GetLatestCommentUrl()

	// Show loading indicator
	width := m.sidebar.GetSidebarContentWidth()
	m.notificationView.SetRow(row)
	m.notificationView.SetWidth(width)
	m.sidebar.SetContent(m.notificationView.View())

	switch subjectType {
	case "PullRequest":
		return tea.Batch(
			func() tea.Msg {
				_ = data.MarkNotificationRead(notifId)
				return notificationssection.UpdateNotificationReadStateMsg{
					Id:     notifId,
					Unread: false,
				}
			},
			func() tea.Msg {
				pr, err := data.FetchPullRequest(subjectUrl)
				return notificationPRFetchedMsg{
					NotificationId:   notifId,
					PR:               pr,
					LatestCommentUrl: latestCommentUrl,
					Err:              err,
				}
			},
		)
	case "Issue":
		return tea.Batch(
			func() tea.Msg {
				_ = data.MarkNotificationRead(notifId)
				return notificationssection.UpdateNotificationReadStateMsg{
					Id:     notifId,
					Unread: false,
				}
			},
			func() tea.Msg {
				issue, err := data.FetchIssue(subjectUrl)
				return notificationIssueFetchedMsg{
					NotificationId:   notifId,
					Issue:            issue,
					LatestCommentUrl: latestCommentUrl,
					Err:              err,
				}
			},
		)
	default:
		// For discussions, releases, etc. - mark as read and open in browser
		// since we can't show rich content for these types
		return tea.Batch(
			func() tea.Msg {
				_ = data.MarkNotificationRead(notifId)
				return notificationssection.UpdateNotificationReadStateMsg{
					Id:     notifId,
					Unread: false,
				}
			},
			m.openBrowser(),
		)
	}
}

func (m *Model) fetchAllViewSections() ([]section.Section, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	cmds = append(cmds, m.tabs.SetAllLoading()...)

	switch m.ctx.View {
	case config.RepoView:
		var cmd tea.Cmd
		s, cmd := reposection.FetchAllBranches(m.ctx)
		cmds = append(cmds, cmd)
		m.repo = &s
		return nil, tea.Batch(cmds...)
	case config.NotificationsView:
		s, notifCmd := notificationssection.FetchAllSections(m.ctx, m.notifications)
		cmds = append(cmds, notifCmd)
		m.notifications = s
		return s, tea.Batch(cmds...)
	case config.PRsView:
		s, prcmds := prssection.FetchAllSections(m.ctx, m.prs)
		cmds = append(cmds, prcmds)
		return s, tea.Batch(cmds...)
	default:
		s, issuecmds := issuessection.FetchAllSections(m.ctx)
		cmds = append(cmds, issuecmds)
		return s, tea.Batch(cmds...)
	}
}

func (m *Model) getCurrentViewSections() []section.Section {
	switch m.ctx.View {
	case config.RepoView:
		if m.repo == nil {
			return []section.Section{}
		}
		return []section.Section{m.repo}
	case config.NotificationsView:
		if len(m.notifications) == 0 {
			return []section.Section{}
		}
		return m.notifications
	case config.PRsView:
		return m.prs
	default:
		return m.issues
	}
}

func (m *Model) getCurrentViewDefaultSection() int {
	switch m.ctx.View {
	case config.RepoView:
		return 0
	case config.NotificationsView:
		return 1 // First notification section after search section
	case config.PRsView:
		return 1
	default:
		return 1
	}
}

func (m *Model) setCurrentViewSections(newSections []section.Section) {
	if newSections == nil {
		return
	}

	// Handle notifications view with search section like PRs/Issues
	if m.ctx.View == config.NotificationsView {
		missingSearchSection := len(newSections) == 0 ||
			(len(newSections) > 0 && newSections[0].GetId() != 0)
		s := make([]section.Section, 0)
		if missingSearchSection {
			// Check if we have an existing search section to preserve
			if len(m.notifications) > 0 && m.notifications[0] != nil && m.notifications[0].GetId() == 0 {
				// Preserve existing search section with its filter state
				s = append(s, m.notifications[0])
			} else {
				// Create new search section only if none exists
				search := notificationssection.NewModel(
					0,
					m.ctx,
					config.NotificationsSectionConfig{
						Title:   "",
						Filters: "archived:false",
					},
					time.Now(),
				)
				s = append(s, &search)
			}
		}
		m.notifications = append(s, newSections...)
		m.tabs.SetSections(m.notifications)
		return
	}

	missingSearchSection := len(newSections) == 0 || (len(newSections) > 0 && newSections[0].GetId() != 0)
	s := make([]section.Section, 0)
	if m.ctx.View == config.PRsView {
		if missingSearchSection {
			search := prssection.NewModel(
				0,
				m.ctx,
				config.PrsSectionConfig{
					Title:   "",
					Filters: "archived:false",
				},
				time.Now(),
				time.Now(),
			)
			s = append(s, &search)
		}
		m.prs = append(s, newSections...)
		newSections = m.prs
	} else {
		if missingSearchSection {
			search := issuessection.NewModel(
				0,
				m.ctx,
				config.IssuesSectionConfig{
					Title:   "",
					Filters: "",
				},
				time.Now(),
				time.Now(),
			)
			s = append(s, &search)
		}
		m.issues = append(s, newSections...)
		newSections = m.issues
	}

	m.tabs.SetSections(newSections)
}

func (m *Model) switchSelectedView() tea.Cmd {
	repoFF := config.IsFeatureEnabled(config.FF_REPO_VIEW)

	// Reset notification subject when leaving notifications view
	if m.ctx.View == config.NotificationsView {
		keys.SetNotificationSubject(keys.NotificationSubjectNone)
		m.notificationView.ClearSubject()
	}

	// View cycle: Notifications → PRs → Issues (→ Repo if enabled) → Notifications
	if repoFF {
		switch m.ctx.View {
		case config.NotificationsView:
			m.ctx.View = config.PRsView
		case config.PRsView:
			m.ctx.View = config.IssuesView
		case config.IssuesView:
			m.ctx.View = config.RepoView
		case config.RepoView:
			m.ctx.View = config.NotificationsView
		}
	} else {
		switch m.ctx.View {
		case config.NotificationsView:
			m.ctx.View = config.PRsView
		case config.PRsView:
			m.ctx.View = config.IssuesView
		default:
			m.ctx.View = config.NotificationsView
		}
	}

	m.syncMainContentWidth()
	m.setCurrSectionId(m.getCurrentViewDefaultSection())

	var cmds []tea.Cmd
	currSections := m.getCurrentViewSections()
	if len(currSections) == 0 {
		newSections, fetchSectionsCmds := m.fetchAllViewSections()
		currSections = newSections
		cmds = append(cmds, m.tabs.SetAllLoading()...)
		cmds = append(cmds, fetchSectionsCmds)
	}
	m.setCurrentViewSections(currSections)
	cmds = append(cmds, m.onViewedRowChanged())

	return tea.Batch(cmds...)
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
		Height(1).
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

// promptConfirmationForNotificationPR shows a confirmation prompt for PR actions
// when viewing a PR from a notification. This is separate from section-based
// confirmation because the notification section doesn't know about PR actions.
func (m *Model) promptConfirmationForNotificationPR(action string) tea.Cmd {
	prompt := m.notificationView.SetPendingPRAction(action)
	if prompt == "" {
		return nil
	}
	m.footer.SetLeftSection(m.ctx.Styles.ListViewPort.PagerStyle.Render(prompt))
	return nil
}

// promptConfirmationForNotificationIssue shows a confirmation prompt for Issue actions
// when viewing an Issue from a notification.
func (m *Model) promptConfirmationForNotificationIssue(action string) tea.Cmd {
	prompt := m.notificationView.SetPendingIssueAction(action)
	if prompt == "" {
		return nil
	}
	m.footer.SetLeftSection(m.ctx.Styles.ListViewPort.PagerStyle.Render(prompt))
	return nil
}

// executeNotificationAction executes a PR/Issue action after user confirmation
func (m *Model) executeNotificationAction(action string) tea.Cmd {
	if action == "" {
		return nil
	}

	sid := tasks.SectionIdentifier{Id: m.currSectionId, Type: notificationssection.SectionType}
	pr := m.notificationView.GetSubjectPR()
	issue := m.notificationView.GetSubjectIssue()

	switch action {
	case "pr_close":
		if pr != nil {
			return tasks.ClosePR(m.ctx, sid, pr)
		}
	case "pr_reopen":
		if pr != nil {
			return tasks.ReopenPR(m.ctx, sid, pr)
		}
	case "pr_ready":
		if pr != nil {
			return tasks.PRReady(m.ctx, sid, pr)
		}
	case "pr_merge":
		if pr != nil {
			return tasks.MergePR(m.ctx, sid, pr)
		}
	case "pr_update":
		if pr != nil {
			return tasks.UpdatePR(m.ctx, sid, pr)
		}
	case "issue_close":
		if issue != nil {
			return tasks.CloseIssue(m.ctx, sid, issue)
		}
	case "issue_reopen":
		if issue != nil {
			return tasks.ReopenIssue(m.ctx, sid, issue)
		}
	}

	return nil
}
