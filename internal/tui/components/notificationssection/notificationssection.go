package notificationssection

import (
	"fmt"
	"regexp"
	"slices"
	"sync"
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

// reasonFilterRegex matches "reason:value" patterns in search strings
var reasonFilterRegex = regexp.MustCompile(`reason:([^\s]+)`)

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
	ReasonFilters     []string // Notification reasons to filter by (e.g., "author", "mention")
	ReadState         data.NotificationReadState
	IsDone            bool // If true, user asked for is:done which is not retrievable
	ExplicitUnread    bool // If true, user explicitly typed "is:unread" (excludes bookmarked+read)
	IncludeBookmarked bool // If true, include bookmarked items even if read (default view)
}

// parseReasonFilters extracts reason:value patterns from a search string
// Handles "reason:participating" as a meta-filter and normalizes hyphenated names
func parseReasonFilters(search string) []string {
	matches := reasonFilterRegex.FindAllStringSubmatch(search, -1)
	reasons := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			reason := match[1]
			// Expand "participating" meta-filter to multiple reasons
			if reason == "participating" {
				reasons = append(reasons,
					data.ReasonAuthor,
					data.ReasonComment,
					data.ReasonMention,
					data.ReasonReviewRequested,
					data.ReasonAssign,
					data.ReasonStateChange,
				)
			} else {
				// Normalize hyphenated names to match GitHub API values
				switch reason {
				case "review-requested":
					reasons = append(reasons, data.ReasonReviewRequested)
				case "team-mention":
					reasons = append(reasons, data.ReasonTeamMention)
				case "ci-activity":
					reasons = append(reasons, data.ReasonCIActivity)
				case "security-alert":
					reasons = append(reasons, data.ReasonSecurityAlert)
				case "state-change":
					reasons = append(reasons, data.ReasonStateChange)
				default:
					reasons = append(reasons, reason)
				}
			}
		}
	}
	return reasons
}

// parseNotificationFilters extracts all notification filters from search string.
// When includeRead is true (the default config), the default read state is "all"
// instead of "unread", matching GitHub's default behavior.
func parseNotificationFilters(search string, includeRead bool) NotificationFilters {
	defaultReadState := data.NotificationStateUnread
	if includeRead {
		defaultReadState = data.NotificationStateAll
	}
	filters := NotificationFilters{
		RepoFilters:       parseRepoFilters(search),
		ReasonFilters:     parseReasonFilters(search),
		ReadState:         defaultReadState,
		IsDone:            false,
		ExplicitUnread:    false,
		IncludeBookmarked: !includeRead, // Only auto-include bookmarks when filtering to unread
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
	sessionMarkedDone map[string]bool // IDs of notifications marked as done this session (excluded until manual refresh)
}

func NewModel(
	id int,
	ctx *context.ProgramContext,
	cfg config.NotificationsSectionConfig,
	lastUpdated time.Time,
) Model {
	sectionCfg := cfg.ToSectionConfig()

	m := Model{}
	m.BaseModel = section.NewModel(
		ctx,
		section.NewSectionOptions{
			Id:          id,
			Config:      sectionCfg,
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
	m.sessionMarkedDone = make(map[string]bool)

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
				m.SyncSmartFilterWithSearchValue()
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
				if input == "" || input == "Y" || input == "y" {
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
				m.NextRow()
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

		case key.Matches(msg, keys.NotificationKeys.Open):
			if m.GetCurrRow() != nil {
				cmd = m.openInBrowser()
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
			if m.HasCurrentRepoNameInConfiguredFilter() || !m.HasRepoNameInConfiguredFilter() {
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
			// Track as done so it doesn't reappear on refresh (GitHub API still returns it with all=true)
			m.sessionMarkedDone[msg.Id] = true
			// Also remove from sessionMarkedRead
			delete(m.sessionMarkedRead, msg.Id)
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
		log.Debug("UpdateNotificationCommentsMsg received", "id", msg.Id, "count",
			msg.NewCommentsCount, "state", msg.SubjectState, "actor", msg.Actor)
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
				log.Debug("Updated notification", "id", msg.Id, "count",
					msg.NewCommentsCount, "state", msg.SubjectState, "actor", msg.Actor)
				break
			}
		}

	case UpdateNotificationUrlMsg:
		// Update the notification with async-resolved URL (e.g., for CheckSuite)
		log.Debug("UpdateNotificationUrlMsg received", "id", msg.Id, "url", msg.ResolvedUrl)
		for i := range m.Notifications {
			if m.Notifications[i].GetId() == msg.Id {
				m.Notifications[i].ResolvedUrl = msg.ResolvedUrl
				m.Table.SetRows(m.BuildRows())
				log.Debug("Updated notification URL", "id", msg.Id, "url", msg.ResolvedUrl)
				break
			}
		}

	case SectionNotificationsFetchedMsg:
		if m.LastFetchTaskId == msg.TaskId {
			if m.PageInfo != nil {
				// Append to existing notifications (pagination)
				m.Notifications = append(m.Notifications, msg.Notifications...)
			} else {
				// First page, replace
				m.Notifications = msg.Notifications
			}
			m.TotalCount = len(m.Notifications)
			m.PageInfo = &msg.PageInfo
			m.SetIsLoading(false)
			m.Table.SetRows(m.BuildRows())
			m.UpdateLastUpdated(time.Now())
			m.UpdateTotalItemsCount(m.TotalCount)

			// Start background fetches for comment counts (only for new notifications)
			fetchCmds := m.fetchCommentCountsForNotifications(msg.Notifications)
			cmd = tea.Batch(fetchCmds...)
		}

	case ClearAllNotificationsMsg:
		// Clear all notifications after marking all as done, then refetch
		m.Notifications = []notificationrow.Data{}
		m.TotalCount = 0
		m.PageInfo = nil
		m.sessionMarkedDone = make(map[string]bool)
		m.SetIsLoading(true)
		m.Table.SetRows(m.BuildRows())
		m.UpdateTotalItemsCount(0)
		cmd = tea.Batch(m.FetchNextPageSectionRows()...)

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

	// Check if there's a next page (skip if we already know there isn't)
	if m.PageInfo != nil && !m.PageInfo.HasNextPage {
		return nil
	}

	var cmds []tea.Cmd

	// Parse filters from search value (includes repo filter if smartFilteringAtLaunch is enabled)
	filters := parseNotificationFilters(m.GetSearchValue(), m.Ctx.Config.IncludeReadNotifications)

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

	// Capture session state for the closure
	sessionMarkedRead := m.sessionMarkedRead
	hasSessionMarkedRead := len(sessionMarkedRead) > 0
	sessionMarkedDone := m.sessionMarkedDone

	// Capture current page info for pagination
	pageInfo := m.PageInfo

	// Capture config limit for the closure
	limit := m.Ctx.Config.Defaults.NotificationsLimit

	// Build reason filter map for O(1) lookup
	reasonFilterMap := make(map[string]bool, len(filters.ReasonFilters))
	for _, reason := range filters.ReasonFilters {
		reasonFilterMap[reason] = true
	}

	fetchCmd := func() tea.Msg {
		// Check if we need to include bookmarked items
		// Build a map for O(1) lookups in the filter loop
		bookmarkStore := data.GetBookmarkStore()
		bookmarkedIds := bookmarkStore.GetBookmarkedIds()
		hasBookmarks := len(bookmarkedIds) > 0
		bookmarkedIdMap := make(map[string]bool, len(bookmarkedIds))
		for _, id := range bookmarkedIds {
			bookmarkedIdMap[id] = true
		}

		// Use the filter's read state directly - don't switch to "all" just for bookmarks/session items
		// Bookmarked and session-marked-read items will be fetched separately by thread ID
		readState := filters.ReadState

		// Initialize done store for filtering
		doneStore := data.GetDoneStore()

		// Track accumulated notifications across multiple pages.
		// We may need to fetch additional pages if many notifications are filtered out
		// (e.g., marked as done locally). The loop continues until we have enough
		// notifications to display or run out of pages from the API.
		notifications := make([]notificationrow.Data, 0, limit)
		currentPageInfo := pageInfo
		var lastPageInfo data.PageInfo
		isFirstPage := pageInfo == nil
		for {
			res, err := data.FetchNotifications(limit, filters.RepoFilters, readState, currentPageInfo)
			if err != nil {
				return constants.TaskFinishedMsg{
					SectionId:   m.Id,
					SectionType: m.Type,
					TaskId:      taskId,
					Err:         err,
				}
			}
			lastPageInfo = res.PageInfo

			// Build a set of IDs we fetched from the API
			fetchedIds := make(map[string]bool, len(res.Notifications))
			for _, n := range res.Notifications {
				fetchedIds[n.Id] = true
			}

			// On first page, fetch any bookmarked/session-marked-read notifications that are missing
			// (they may have aged out of the default notifications list or been marked as read)
			if isFirstPage {
				isFirstPage = false

				// Collect all missing IDs that need to be fetched
				missingIds := make([]string, 0)
				if filters.IncludeBookmarked && hasBookmarks {
					for _, bookmarkId := range bookmarkedIds {
						if !fetchedIds[bookmarkId] {
							missingIds = append(missingIds, bookmarkId)
							fetchedIds[bookmarkId] = true // Mark as fetched to avoid duplicates
						}
					}
				}
				if hasSessionMarkedRead {
					for id := range sessionMarkedRead {
						if !fetchedIds[id] {
							missingIds = append(missingIds, id)
							fetchedIds[id] = true
						}
					}
				}

				// Fetch all missing notifications in parallel
				if len(missingIds) > 0 {
					// Build repo filter map for O(1) lookup
					repoFilterMap := make(map[string]bool, len(filters.RepoFilters))
					for _, repo := range filters.RepoFilters {
						repoFilterMap[repo] = true
					}

					type fetchResult struct {
						notification *data.NotificationData
						err          error
					}
					results := make(chan fetchResult, len(missingIds))

					var wg sync.WaitGroup
					for _, id := range missingIds {
						wg.Add(1)
						go func(threadId string) {
							defer wg.Done()
							notification, err := data.FetchNotificationByThreadId(threadId)
							results <- fetchResult{notification: notification, err: err}
						}(id)
					}

					// Close results channel when all goroutines complete
					go func() {
						wg.Wait()
						close(results)
					}()

					// Collect results
					for result := range results {
						if result.err != nil {
							log.Debug("Failed to fetch missing notification", "err", result.err)
							continue
						}
						if result.notification == nil {
							continue
						}
						// Apply repo filter if set
						if len(repoFilterMap) > 0 && !repoFilterMap[result.notification.Repository.FullName] {
							continue
						}
						res.Notifications = append(res.Notifications, *result.notification)
					}
				}
			}

			// Filter notifications based on bookmark settings and session state
			for _, n := range res.Notifications {
				// Skip notifications marked as done (GitHub API still returns them with all=true)
				// Check both persistent store and session state
				if doneStore.IsDone(n.Id) || sessionMarkedDone[n.Id] {
					continue
				}

				include := false

				// Always include notifications marked as read this session (until manual refresh)
				if sessionMarkedRead[n.Id] {
					include = true
				} else if filters.IncludeBookmarked && hasBookmarks {
					// Default view: include if unread OR bookmarked (O(1) map lookup)
					include = n.Unread || bookmarkedIdMap[n.Id]
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

				// Apply reason filter if specified (O(1) map lookup)
				if include && len(reasonFilterMap) > 0 {
					include = reasonFilterMap[n.Reason]
				}

				if include {
					notifications = append(notifications, notificationrow.Data{
						Notification: n,
						// Generate initial activity description (will be updated with actor later)
						ActivityDescription: notificationrow.GenerateActivityDescription(n.Reason, n.Subject.Type, ""),
					})
				}
			}

			// Check if we have enough notifications or if we've run out of pages
			if len(notifications) >= limit || !lastPageInfo.HasNextPage {
				break
			}

			// Need more notifications - fetch the next page
			currentPageInfo = &lastPageInfo
			log.Debug("Fetching additional page due to done filtering",
				"currentCount", len(notifications),
				"targetLimit", limit,
				"nextPage", lastPageInfo.EndCursor)
		}

		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: m.Type,
			TaskId:      taskId,
			Msg: SectionNotificationsFetchedMsg{
				Notifications: notifications,
				TotalCount:    len(notifications),
				TaskId:        taskId,
				PageInfo:      lastPageInfo,
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
	// Clear session state on manual refresh - user explicitly wants fresh data
	m.sessionMarkedRead = make(map[string]bool)
	m.sessionMarkedDone = make(map[string]bool)
	m.BaseModel.ResetRows()
}

// FetchAllSections creates and fetches all notification sections based on config.
// Returns sections and a batch command to fetch all data.
func FetchAllSections(
	ctx *context.ProgramContext,
	existing []section.Section,
) (sections []section.Section, fetchAllCmd tea.Cmd) {
	sectionConfigs := ctx.Config.NotificationsSections
	fetchCmds := make([]tea.Cmd, 0, len(sectionConfigs))
	sections = make([]section.Section, 0, len(sectionConfigs))

	for i, sectionConfig := range sectionConfigs {
		sectionModel := NewModel(
			i+1, // ID 0 is reserved for search section
			ctx,
			sectionConfig,
			time.Now(),
		)

		// Preserve existing data and filter state if refreshing
		if len(existing) > i+1 && existing[i+1] != nil {
			if oldSection, ok := existing[i+1].(*Model); ok {
				sectionModel.Notifications = oldSection.Notifications
				sectionModel.LastFetchTaskId = oldSection.LastFetchTaskId
				sectionModel.sessionMarkedRead = oldSection.sessionMarkedRead
				sectionModel.sessionMarkedDone = oldSection.sessionMarkedDone
				// Preserve user's filter state - don't reset on refresh
				sectionModel.IsFilteredByCurrentRemote = oldSection.IsFilteredByCurrentRemote
				sectionModel.SearchValue = oldSection.SearchValue
				sectionModel.SearchBar.SetValue(oldSection.SearchValue)
			}
		}

		sections = append(sections, &sectionModel)
		fetchCmds = append(fetchCmds, sectionModel.FetchNextPageSectionRows()...)
	}

	return sections, tea.Batch(fetchCmds...)
}

// SectionNotificationsFetchedMsg contains the result of fetching notifications from the GitHub API.
// This message is sent when the initial fetch or a refresh completes.
type SectionNotificationsFetchedMsg struct {
	Notifications []notificationrow.Data
	TotalCount    int
	TaskId        string
	PageInfo      data.PageInfo
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

// UpdateNotificationUrlMsg carries a resolved URL for notifications where the URL
// cannot be determined synchronously (e.g., CheckSuite notifications).
type UpdateNotificationUrlMsg struct {
	Id          string
	ResolvedUrl string
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

// fetchCommentCountsForNotifications returns commands to fetch comment counts for the given notifications
func (m *Model) fetchCommentCountsForNotifications(notifications []notificationrow.Data) []tea.Cmd {
	var cmds []tea.Cmd

	log.Debug("fetchCommentCountsForNotifications called", "numNotifications", len(notifications))

	for _, notif := range notifications {
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
		case "CheckSuite":
			// CheckSuite notifications have subject.url=null in GitHub's API.
			// We fetch recent workflow runs and find the best match by timestamp.
			id := notifId
			repo := notif.Notification.Repository.FullName
			updatedAt := notif.Notification.UpdatedAt
			title := notif.Notification.Subject.Title
			cmds = append(cmds, func() tea.Msg {
				log.Debug("Fetching workflow run for CheckSuite", "id", id, "repo", repo)
				url, err := data.FetchRecentWorkflowRun(repo, updatedAt, title)
				if err != nil {
					log.Error("Failed to fetch workflow run", "id", id, "err", err)
					return nil
				}
				if url == "" {
					log.Debug("No matching workflow run found", "id", id)
					return nil
				}
				log.Debug("Found workflow run URL", "id", id, "url", url)
				return UpdateNotificationUrlMsg{
					Id:          id,
					ResolvedUrl: url,
				}
			})
		}
	}

	log.Debug("fetchCommentCountsForNotifications returning", "numCmds", len(cmds))
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
