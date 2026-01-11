package notificationssection

import (
	"fmt"
	"regexp"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

const SectionType = "notification"

// repoFilterRegex matches "repo:owner/name" patterns in search strings
var repoFilterRegex = regexp.MustCompile(`repo:([^\s]+)`)

// stateFilterRegex matches "is:unread", "is:read", "is:done", "is:all" patterns
var stateFilterRegex = regexp.MustCompile(`is:(unread|read|done|all)`)

// parseRepoFilters extracts repo:owner/name patterns from a search string
func parseRepoFilters(search string) []string {
	matches := repoFilterRegex.FindAllStringSubmatch(search, -1)
	repos := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			repos = append(repos, match[1])
		}
	}
	return repos
}

// NotificationFilters holds parsed notification filters
type NotificationFilters struct {
	RepoFilters       []string
	ReadState         data.NotificationReadState
	IsDone            bool // If true, user asked for is:done which is not retrievable
	ExplicitUnread    bool // If true, user explicitly typed "is:unread" (excludes bookmarked+read)
	IncludeBookmarked bool // If true, include bookmarked items even if read (default view)
}

// parseNotificationFilters extracts all notification filters from search string
func parseNotificationFilters(search string) NotificationFilters {
	filters := NotificationFilters{
		RepoFilters:       parseRepoFilters(search),
		ReadState:         data.NotificationStateUnread, // Default to unread
		IsDone:            false,
		ExplicitUnread:    false,
		IncludeBookmarked: true, // Default view includes bookmarked items
	}

	matches := stateFilterRegex.FindAllStringSubmatch(search, -1)
	hasUnread := false
	hasRead := false
	hasDone := false
	hasAll := false

	for _, match := range matches {
		if len(match) > 1 {
			switch match[1] {
			case "unread":
				hasUnread = true
			case "read":
				hasRead = true
			case "done":
				hasDone = true
			case "all":
				hasAll = true
			}
		}
	}

	if hasDone {
		filters.IsDone = true
	}

	if hasAll || (hasUnread && hasRead) {
		filters.ReadState = data.NotificationStateAll
		filters.IncludeBookmarked = false // Explicit filter, don't auto-include bookmarks
	} else if hasRead {
		filters.ReadState = data.NotificationStateRead
		filters.IncludeBookmarked = false // Explicit filter, don't auto-include bookmarks
	} else if hasUnread {
		// User explicitly typed "is:unread" - don't include bookmarked+read items
		filters.ReadState = data.NotificationStateUnread
		filters.ExplicitUnread = true
		filters.IncludeBookmarked = false
	}
	// Default case: ReadState = Unread, IncludeBookmarked = true

	return filters
}

type SortOrder int

const (
	SortByUpdated SortOrder = iota
	SortByRepo
)

type Model struct {
	section.BaseModel
	Notifications     []notificationrow.Data
	SortOrder         SortOrder
	lastSidebarOpen   bool
	sessionMarkedRead map[string]bool // IDs of notifications marked as read this session (kept visible until manual refresh)
}

func NewModel(
	id int,
	ctx *context.ProgramContext,
	lastUpdated time.Time,
) Model {
	cfg := config.SectionConfig{
		Title:   "  | Notifications",
		Filters: "",
	}

	m := Model{}
	m.BaseModel = section.NewModel(
		ctx,
		section.NewSectionOptions{
			Id:          id,
			Config:      cfg,
			Type:        SectionType,
			Columns:     GetSectionColumns(ctx),
			Singular:    m.GetItemSingularForm(),
			Plural:      m.GetItemPluralForm(),
			LastUpdated: lastUpdated,
			CreatedAt:   lastUpdated,
		},
	)
	// Set 3-line content height for notification rows
	m.Table.SetContentHeight(3)
	// Respect smartFilteringAtLaunch - scope to current repo by default if enabled
	m.IsFilteredByCurrentRemote = ctx.Config.SmartFilteringAtLaunch
	m.SearchValue = m.GetSearchValue()
	m.SearchBar.SetValue(m.SearchValue)
	m.Notifications = []notificationrow.Data{}
	m.sessionMarkedRead = make(map[string]bool)

	return m
}

func (m *Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.IsSearchFocused() {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.SearchBar.SetValue(m.SearchValue)
				blinkCmd := m.SetIsSearching(false)
				return m, blinkCmd

			case tea.KeyEnter:
				m.SearchValue = m.SearchBar.Value()
				m.SetIsSearching(false)
				m.ResetRows()
				return m, tea.Batch(m.FetchNextPageSectionRows()...)
			}

			break
		}

		if m.IsPromptConfirmationFocused() {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.PromptConfirmationBox.Reset()
				cmd = m.SetIsPromptConfirmationShown(false)
				return m, cmd

			case tea.KeyEnter:
				input := m.PromptConfirmationBox.Value()
				action := m.GetPromptConfirmationAction()
				if input == "Y" || input == "y" {
					switch action {
					case "done":
						cmd = m.markAsDone()
					case "done_all":
						cmd = m.markAllAsDone()
					}
				}

				m.PromptConfirmationBox.Reset()
				blinkCmd := m.SetIsPromptConfirmationShown(false)

				return m, tea.Batch(cmd, blinkCmd)
			}
			break
		}

		switch {
		case key.Matches(msg, keys.NotificationKeys.MarkAsDone):
			if m.GetCurrRow() != nil {
				cmd = m.markAsDone()
			}
			return m, cmd

		case key.Matches(msg, keys.NotificationKeys.MarkAsRead):
			if m.GetCurrRow() != nil {
				cmd = m.markAsRead()
			}
			return m, cmd

		case key.Matches(msg, keys.NotificationKeys.MarkAllAsRead):
			cmd = m.markAllAsRead()
			return m, cmd

		case key.Matches(msg, keys.NotificationKeys.Unsubscribe):
			if m.GetCurrRow() != nil {
				cmd = m.unsubscribe()
			}
			return m, cmd

		case key.Matches(msg, keys.NotificationKeys.ToggleBookmark):
			if notification := m.GetCurrNotification(); notification != nil {
				data.GetBookmarkStore().ToggleBookmark(notification.GetId())
				// Rebuild rows to update bookmark indicator
				m.Table.SetRows(m.BuildRows())
			}
			return m, nil

		case key.Matches(msg, keys.NotificationKeys.SortByRepo):
			m.toggleSortOrder()
			m.Table.SetRows(m.BuildRows())
			return m, nil

		case key.Matches(msg, keys.NotificationKeys.ToggleSmartFiltering):
			if !m.HasRepoNameInConfiguredFilter() {
				m.IsFilteredByCurrentRemote = !m.IsFilteredByCurrentRemote
			}
			searchValue := m.GetSearchValue()
			if m.SearchValue != searchValue {
				m.SearchValue = searchValue
				m.SearchBar.SetValue(searchValue)
				m.SetIsSearching(false)
				m.ResetRows()
				return m, tea.Batch(m.FetchNextPageSectionRows()...)
			}
		}

	case UpdateNotificationMsg:
		if msg.IsRemoved {
			for i, n := range m.Notifications {
				if n.GetId() == msg.Id {
					m.Notifications = append(m.Notifications[:i], m.Notifications[i+1:]...)
					break
				}
			}
			m.TotalCount = len(m.Notifications)
			m.SetIsLoading(false)
			m.Table.SetRows(m.BuildRows())
			m.UpdateTotalItemsCount(m.TotalCount)
		}

	case UpdateNotificationReadStateMsg:
		// Update the notification's read state
		for i := range m.Notifications {
			if m.Notifications[i].GetId() == msg.Id {
				m.Notifications[i].Notification.Unread = msg.Unread
				// Track notifications marked as read this session so they remain visible
				if !msg.Unread {
					m.sessionMarkedRead[msg.Id] = true
				}
				m.Table.SetRows(m.BuildRows())
				break
			}
		}

	case UpdateNotificationCommentsMsg:
		// Update the notification with fetched data
		log.Debug("UpdateNotificationCommentsMsg received", "id", msg.Id, "count", msg.NewCommentsCount, "state", msg.SubjectState, "actor", msg.Actor)
		for i := range m.Notifications {
			if m.Notifications[i].GetId() == msg.Id {
				m.Notifications[i].NewCommentsCount = msg.NewCommentsCount
				m.Notifications[i].SubjectState = msg.SubjectState
				m.Notifications[i].IsDraft = msg.IsDraft
				m.Notifications[i].Actor = msg.Actor
				// Generate activity description based on reason, type, and actor
				m.Notifications[i].ActivityDescription = notificationrow.GenerateActivityDescription(
					m.Notifications[i].GetReason(),
					m.Notifications[i].GetSubjectType(),
					msg.Actor,
				)
				m.Table.SetRows(m.BuildRows())
				log.Debug("Updated notification", "id", msg.Id, "count", msg.NewCommentsCount, "state", msg.SubjectState, "actor", msg.Actor)
				break
			}
		}

	case SectionNotificationsFetchedMsg:
		if m.LastFetchTaskId == msg.TaskId {
			m.Notifications = msg.Notifications
			m.TotalCount = msg.TotalCount
			m.SetIsLoading(false)
			m.Table.SetRows(m.BuildRows())
			m.UpdateLastUpdated(time.Now())
			m.UpdateTotalItemsCount(m.TotalCount)

			// Start background fetches for comment counts
			fetchCmds := m.fetchAllCommentCounts()
			cmd = tea.Batch(fetchCmds...)
		}

	case ClearAllNotificationsMsg:
		// Clear all notifications after marking all as done
		m.Notifications = []notificationrow.Data{}
		m.TotalCount = 0
		m.SetIsLoading(false)
		m.Table.SetRows(m.BuildRows())
		m.UpdateTotalItemsCount(0)

	case MarkAllAsReadMsg:
		// Mark all notifications as read (update their state)
		for i := range m.Notifications {
			m.Notifications[i].Notification.Unread = false
			// Track all as marked read this session so they remain visible
			m.sessionMarkedRead[m.Notifications[i].GetId()] = true
		}
		m.Table.SetRows(m.BuildRows())
	}

	search, searchCmd := m.SearchBar.Update(msg)
	m.SearchBar = search

	prompt, promptCmd := m.PromptConfirmationBox.Update(msg)
	m.PromptConfirmationBox = prompt

	tbl, tableCmd := m.Table.Update(msg)
	m.Table = tbl

	return m, tea.Batch(cmd, searchCmd, promptCmd, tableCmd)
}

func GetSectionColumns(ctx *context.ProgramContext) []table.Column {
	return []table.Column{
		{
			Title: "Type",
			Width: utils.IntPtr(6), // Type icon
			Align: func() *lipgloss.Position { p := lipgloss.Center; return &p }(),
		},
		{
			Title: "Title",
			Grow:  utils.BoolPtr(true), // 3-line title block (includes bookmark icon)
		},
		{
			Title: "Activity",
			Width: utils.IntPtr(10), // Comments count with icon
			Align: func() *lipgloss.Position { p := lipgloss.Right; return &p }(),
		},
		{
			Title: "󱦻    ",          // Trailing padding to center when right-aligned
			Width: utils.IntPtr(12), // Updated at (e.g., "12mo ago")
			Align: func() *lipgloss.Position { p := lipgloss.Right; return &p }(),
		},
	}
}

func (m *Model) toggleSortOrder() {
	if m.SortOrder == SortByUpdated {
		m.SortOrder = SortByRepo
	} else {
		m.SortOrder = SortByUpdated
	}
	m.sortNotifications()
}

func (m *Model) sortNotifications() {
	switch m.SortOrder {
	case SortByRepo:
		slices.SortFunc(m.Notifications, func(a, b notificationrow.Data) int {
			repoA := a.Notification.Repository.FullName
			repoB := b.Notification.Repository.FullName
			if repoA < repoB {
				return -1
			}
			if repoA > repoB {
				return 1
			}
			// Secondary sort by updated time (most recent first)
			if a.Notification.UpdatedAt.After(b.Notification.UpdatedAt) {
				return -1
			}
			if a.Notification.UpdatedAt.Before(b.Notification.UpdatedAt) {
				return 1
			}
			return 0
		})
	case SortByUpdated:
		slices.SortFunc(m.Notifications, func(a, b notificationrow.Data) int {
			if a.Notification.UpdatedAt.After(b.Notification.UpdatedAt) {
				return -1
			}
			if a.Notification.UpdatedAt.Before(b.Notification.UpdatedAt) {
				return 1
			}
			return 0
		})
	}
}

func (m Model) BuildRows() []table.Row {
	var rows []table.Row
	for i := range m.Notifications {
		notification := &m.Notifications[i]
		notificationModel := notificationrow.Notification{Ctx: m.Ctx, Data: notification}
		rows = append(rows, notificationModel.ToTableRow())
	}

	if rows == nil {
		rows = []table.Row{}
	}

	return rows
}

func (m *Model) NumRows() int {
	return len(m.Notifications)
}

func (m *Model) GetCurrRow() data.RowData {
	idx := m.Table.GetCurrItem()
	if idx < 0 || idx >= len(m.Notifications) {
		return nil
	}
	return &m.Notifications[idx]
}

func (m *Model) GetCurrNotification() *notificationrow.Data {
	idx := m.Table.GetCurrItem()
	if idx < 0 || idx >= len(m.Notifications) {
		return nil
	}
	return &m.Notifications[idx]
}

func (m *Model) FetchNextPageSectionRows() []tea.Cmd {
	if m == nil {
		return nil
	}

	var cmds []tea.Cmd

	// Parse filters from search value (includes repo filter if smartFilteringAtLaunch is enabled)
	filters := parseNotificationFilters(m.GetSearchValue())

	// Handle is:done filter - these notifications cannot be retrieved
	if filters.IsDone {
		m.Notifications = []notificationrow.Data{}
		m.TotalCount = 0
		m.SetIsLoading(false)
		m.Table.SetRows(m.BuildRows())
		m.UpdateTotalItemsCount(0)
		// Return a message that will be shown to the user
		return []tea.Cmd{func() tea.Msg {
			return constants.TaskFinishedMsg{
				SectionId:   m.Id,
				SectionType: m.Type,
				TaskId:      "done_filter",
				Err:         fmt.Errorf("done notifications cannot be retrieved"),
			}
		}}
	}

	taskId := fmt.Sprintf("fetching_notifications_%d_%s", m.Id, time.Now().String())
	m.LastFetchTaskId = taskId
	task := context.Task{
		Id:           taskId,
		StartText:    "Fetching notifications",
		FinishedText: "Notifications have been fetched",
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	cmds = append(cmds, startCmd)

	// Capture session-marked-read IDs for the closure
	sessionMarkedRead := m.sessionMarkedRead
	hasSessionMarkedRead := len(sessionMarkedRead) > 0

	fetchCmd := func() tea.Msg {
		limit := 50 // Default limit for notifications

		// Check if we need to include bookmarked items
		bookmarkStore := data.GetBookmarkStore()
		bookmarkedIds := bookmarkStore.GetBookmarkedIds()
		hasBookmarks := len(bookmarkedIds) > 0

		// If we want to include bookmarks/session-marked-read and have some, fetch all notifications
		// so we can include read+bookmarked or read+session-marked items
		readState := filters.ReadState
		if (filters.IncludeBookmarked && hasBookmarks) || hasSessionMarkedRead {
			readState = data.NotificationStateAll
		}

		res, err := data.FetchNotifications(limit, filters.RepoFilters, readState)
		if err != nil {
			return constants.TaskFinishedMsg{
				SectionId:   m.Id,
				SectionType: m.Type,
				TaskId:      taskId,
				Err:         err,
			}
		}

		// Filter notifications based on bookmark settings and session state
		notifications := make([]notificationrow.Data, 0, len(res.Notifications))
		for _, n := range res.Notifications {
			include := false

			// Always include notifications marked as read this session (until manual refresh)
			if sessionMarkedRead[n.Id] {
				include = true
			} else if filters.IncludeBookmarked && hasBookmarks {
				// Default view: include if unread OR bookmarked
				isBookmarked := bookmarkStore.IsBookmarked(n.Id)
				include = n.Unread || isBookmarked
			} else {
				// Explicit filter: follow the ReadState filter
				switch filters.ReadState {
				case data.NotificationStateUnread:
					include = n.Unread
				case data.NotificationStateRead:
					include = !n.Unread
				case data.NotificationStateAll:
					include = true
				}
			}

			if include {
				notifications = append(notifications, notificationrow.Data{
					Notification: n,
					// Generate initial activity description (will be updated with actor later)
					ActivityDescription: notificationrow.GenerateActivityDescription(n.Reason, n.Subject.Type, ""),
				})
			}
		}

		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: m.Type,
			TaskId:      taskId,
			Msg: SectionNotificationsFetchedMsg{
				Notifications: notifications,
				TotalCount:    len(notifications),
				TaskId:        taskId,
			},
		}
	}
	cmds = append(cmds, fetchCmd)

	return cmds
}

func (m *Model) UpdateLastUpdated(t time.Time) {
	m.Table.UpdateLastUpdated(t)
}

func (m *Model) ResetRows() {
	m.Notifications = nil
	// Clear session-marked-read on manual refresh - user explicitly wants fresh data
	m.sessionMarkedRead = make(map[string]bool)
	m.BaseModel.ResetRows()
}

func FetchNotifications(
	ctx *context.ProgramContext,
) (section.Section, tea.Cmd) {
	sectionModel := NewModel(1, ctx, time.Now())
	fetchCmd := tea.Batch(sectionModel.FetchNextPageSectionRows()...)
	return &sectionModel, fetchCmd
}

// SectionNotificationsFetchedMsg contains the result of fetching notifications from the GitHub API.
// This message is sent when the initial fetch or a refresh completes.
type SectionNotificationsFetchedMsg struct {
	Notifications []notificationrow.Data
	TotalCount    int
	TaskId        string
}

// UpdateNotificationMsg signals that a notification's state has changed.
// If IsRemoved is true, the notification should be removed from the list (marked as done).
type UpdateNotificationMsg struct {
	Id        string
	IsRemoved bool
}

// UpdateNotificationCommentsMsg carries additional notification metadata fetched asynchronously.
// This includes comment counts, PR/Issue state, draft status, and the actor who triggered the notification.
type UpdateNotificationCommentsMsg struct {
	Id               string
	NewCommentsCount int
	SubjectState     string // OPEN, CLOSED, MERGED
	IsDraft          bool
	Actor            string // Username who triggered the notification
}

func (m Model) GetItemSingularForm() string {
	return "Notification"
}

func (m Model) GetItemPluralForm() string {
	return "Notifications"
}

func (m Model) GetTotalCount() int {
	return m.TotalCount
}

func (m *Model) GetIsLoading() bool {
	return m.IsLoading
}

func (m *Model) SetIsLoading(val bool) {
	m.IsLoading = val
	m.Table.SetIsLoading(val)
}

func (m Model) GetPagerContent() string {
	pagerContent := ""
	if m.TotalCount > 0 {
		pagerContent = fmt.Sprintf(
			"%v %v • %v %v/%v",
			constants.WaitingIcon,
			m.LastUpdated().Format("01/02 15:04:05"),
			m.SingularForm,
			m.Table.GetCurrItem()+1,
			m.TotalCount,
		)
	}
	pager := m.Ctx.Styles.ListViewPort.PagerStyle.Render(pagerContent)
	return pager
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}

	// Rebuild columns if sidebar state changed
	if ctx.SidebarOpen != m.lastSidebarOpen {
		m.lastSidebarOpen = ctx.SidebarOpen
		m.Table.Columns = GetSectionColumns(ctx)
	}

	m.BaseModel.UpdateProgramContext(ctx)
}

// fetchAllCommentCounts returns commands to fetch comment counts for all notifications
func (m *Model) fetchAllCommentCounts() []tea.Cmd {
	var cmds []tea.Cmd

	log.Debug("fetchAllCommentCounts called", "numNotifications", len(m.Notifications))

	for _, notif := range m.Notifications {
		// Copy values for closure capture
		notifId := notif.GetId()
		subjectType := notif.GetSubjectType()
		subjectUrl := notif.GetUrl()
		lastReadAt := notif.Notification.LastReadAt
		apiUrl := notif.Notification.Subject.Url

		log.Debug("Processing notification", "id", notifId, "type", subjectType, "webUrl", subjectUrl, "apiUrl", apiUrl)

		latestCommentUrl := notif.Notification.Subject.LatestCommentUrl

		// Only fetch for PR and Issue types
		switch subjectType {
		case "PullRequest":
			// Capture variables for closure
			id, url, readAt, commentUrl := notifId, subjectUrl, lastReadAt, latestCommentUrl
			cmds = append(cmds, func() tea.Msg {
				log.Debug("Fetching PR for comment count", "url", url)
				pr, err := data.FetchPullRequest(url)
				if err != nil {
					log.Error("Failed to fetch PR for comment count", "url", url, "err", err)
					return nil
				}
				count := countNewPRComments(pr, readAt)
				actor, _ := data.FetchCommentAuthor(commentUrl)
				if actor == "" {
					actor = pr.Author.Login
				}
				log.Debug("Got PR comment count", "id", id, "count", count, "state", pr.State, "actor", actor)
				return UpdateNotificationCommentsMsg{
					Id:               id,
					NewCommentsCount: count,
					SubjectState:     pr.State,
					IsDraft:          pr.IsDraft,
					Actor:            actor,
				}
			})
		case "Issue":
			// Capture variables for closure
			id, url, readAt, commentUrl := notifId, subjectUrl, lastReadAt, latestCommentUrl
			cmds = append(cmds, func() tea.Msg {
				log.Debug("Fetching Issue for comment count", "url", url)
				issue, err := data.FetchIssue(url)
				if err != nil {
					log.Error("Failed to fetch Issue for comment count", "url", url, "err", err)
					return nil
				}
				count := countNewIssueComments(issue, readAt)
				actor, _ := data.FetchCommentAuthor(commentUrl)
				if actor == "" {
					actor = issue.Author.Login
				}
				log.Debug("Got Issue comment count", "id", id, "count", count, "state", issue.State, "actor", actor)
				return UpdateNotificationCommentsMsg{
					Id:               id,
					NewCommentsCount: count,
					SubjectState:     issue.State,
					Actor:            actor,
				}
			})
			// Note: CheckSuite notifications have subject.url=null in GitHub's API,
			// so we can't fetch commit SHA for a more specific link. Falls back to /actions.
		}
	}

	log.Debug("fetchAllCommentCounts returning", "numCmds", len(cmds))
	return cmds
}

// countNewPRComments counts comments in a PR that are newer than lastReadAt
// If lastReadAt is nil (never read), counts all comments
func countNewPRComments(pr data.EnrichedPullRequestData, lastReadAt *time.Time) int {
	count := 0

	for _, comment := range pr.Comments.Nodes {
		if lastReadAt == nil || comment.UpdatedAt.After(*lastReadAt) {
			count++
		}
	}

	for _, thread := range pr.ReviewThreads.Nodes {
		for _, comment := range thread.Comments.Nodes {
			if lastReadAt == nil || comment.UpdatedAt.After(*lastReadAt) {
				count++
			}
		}
	}

	for _, review := range pr.Reviews.Nodes {
		if lastReadAt == nil || review.UpdatedAt.After(*lastReadAt) {
			count++
		}
	}

	return count
}

// countNewIssueComments counts comments in an Issue that are newer than lastReadAt
// If lastReadAt is nil (never read), counts all comments
func countNewIssueComments(issue data.IssueData, lastReadAt *time.Time) int {
	count := 0
	for _, comment := range issue.Comments.Nodes {
		if lastReadAt == nil || comment.UpdatedAt.After(*lastReadAt) {
			count++
		}
	}

	return count
}
