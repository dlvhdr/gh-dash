# Notifications Feature — Implementation & Design Summary

## Overview

The notifications feature adds a new view to gh-dash that displays GitHub notifications, allowing users to triage their inbox directly from the terminal. The implementation follows the existing patterns established by the PR and Issue views.

## Architecture

### File Structure

```
internal/
├── data/
│   ├── notificationapi.go       # GitHub API interactions for notifications
│   └── bookmarks.go             # Local bookmark storage (singleton)
├── tui/
│   ├── keys/
│   │   └── notificationKeys.go  # Key bindings specific to notifications
│   └── components/
│       ├── notificationrow/
│       │   ├── data.go          # Data model implementing RowData interface
│       │   ├── data_test.go     # Tests for data accessors
│       │   ├── notificationrow.go # Row rendering for the table
│       │   └── notificationrow_test.go # Tests for rendering logic
│       ├── notificationssection/
│       │   ├── notificationssection.go # Main section component
│       │   ├── commands.go      # Tea commands (mark done, mark read, diff, checkout, etc.)
│       │   ├── commands_test.go # Tests for command functions
│       │   └── filters_test.go  # Tests for filter parsing
│       └── notificationview/
│           └── notificationview.go # Detail view in sidebar
```

### Key Design Decisions

#### 1. Explicit View Action for Notifications

Unlike PRs and Issues which auto-fetch content when selected, notifications require an explicit action (Enter key) to view content. This design choice exists because:

- Viewing a notification marks it as read (GitHub API behavior)
- Users should consciously decide when to mark something as read
- Prevents accidental "read" marking when just browsing the list

When a notification is selected but not yet viewed, a prompt is displayed in the Preview pane:

```
      Press [Enter] to view the PR
      (Note: this will mark it as read)

      Other Actions

            [D]  mark as done
            [m]  mark as read
            [u]  unsubscribe
            [b]  toggle bookmark
            [t]  toggle filtering
            [S]  sort by repo
            [o]  open in browser
        [Enter]  view
```

- Keys are displayed with a background highlight
- Actions are displayed in green (success color)
- The note about marking as read appears for all notification types
- For PR/Issue types: "Press Enter to view the PR/Issue"
- For other notification types (Discussion, Release, etc.): "Press Enter to open in browser"

#### 2. Notification Data Flow

```
GitHub REST API
      │
      ▼
notificationapi.go (FetchNotifications)
      │
      ▼
notificationssection.go (stores []notificationrow.Data)
      │
      ├──▶ notificationrow.go (renders table rows)
      │
      └──▶ notificationview.go (renders sidebar detail)
           OR
           renderNotificationPrompt() (shows action prompt)
```

#### 3. Comment Count Tracking

For PR and Issue notifications, the system fetches additional data to show new comment counts:

- Compares `LastReadAt` timestamp with comment timestamps
- Displays count of comments made since user last read the notification
- Scrolls to the latest comment when viewing PR/Issue notifications

#### 4. Actor Display

For PR and Issue notifications, the username of the person who triggered the notification is displayed:

- Fetches the author from `latest_comment_url` (the comment that triggered the notification)
- Falls back to the PR/Issue author for new items without comments
- Helps identify spam without needing to open the notification
- Fetched alongside comment counts (no additional latency)
- Color is configurable via `theme.colors.text.actor` (defaults to secondary text color)

#### 5. Three-Line Row Layout

Each notification row displays three lines of information:

```
repo/name #123 🔖                        +5💬  2d ago
Title of the notification (bold if unread)
@username commented on this pull request
```

(The 🔖 bookmark icon only appears if the notification is bookmarked)

- **Line 1:** Repository name with issue/PR number, bookmark icon if bookmarked (SecondaryText color; bookmark icon in WarningText color)
- **Line 2:** Notification title (PrimaryText, bold for unread notifications)
- **Line 3:** Activity description (FaintText color, generated from reason, type, and actor)

The three lines use distinct colors to create visual hierarchy: line 1 is secondary, line 2 is primary/bold, and line 3 is faint.

**Unread indicator:** A blue dot is displayed below the notification type icon for unread notifications. This is the sole indicator of read/unread status—text is not dimmed for read notifications.

Activity descriptions are generated based on the notification reason:
- `comment`: "@username commented on this pull request/issue"
- `review_requested`: "@username requested your review" or "Review requested"
- `mention`: "@username mentioned you"
- `author`: "Activity on your thread"
- `assign`: "You were assigned"
- `state_change`: "Pull request/Issue state changed"
- `ci_activity`: "CI activity"
- `subscribed`: "@username commented on this pull request/issue"

#### 6. Title Sanitization

Notification titles from GitHub's API may contain control characters (e.g., trailing `\r`) that corrupt terminal rendering. The `GetTitle()` method sanitizes titles by:
- Removing carriage return characters (`\r`)
- Replacing newlines (`\n`) with spaces
- Trimming leading/trailing whitespace

#### 7. Multi-Line Row Background

The table component applies cell styling to each line individually in multi-line content. This ensures the background color (for selected rows) extends properly across the entire cell.

To preserve parent background colors, row content uses raw ANSI escape codes for foreground styling without trailing reset sequences. The `utils.GetStylePrefix()` helper extracts ANSI codes from lipgloss styles while stripping the reset, preventing internal resets from breaking the cell's background color.

Title truncation is handled dynamically by the table component based on actual column width, with ellipsis added when content is truncated. This allows titles to adjust when the sidebar is shown/hidden.

#### 8. Bookmark and Done Systems

Both bookmarks and "done" status are tracked locally because GitHub's API doesn't provide these features. They share a common `NotificationIDStore` implementation in `data/bookmarks.go`:

```go
type NotificationIDStore struct {
    ids      map[string]bool
    filePath string
    // ... mutex, name for logging
}
```

**Bookmarks** allow users to keep notifications visible even after marking them as read:

- Stored in `~/.local/state/gh-dash/bookmarks.json`
- Accessed via `data.GetBookmarkStore()` singleton
- Bookmarked notifications appear in the default inbox view even when read
- Bookmarked notifications are styled as read (faint text) but show a bookmark indicator
- When user explicitly searches `is:unread`, bookmarked+read items are excluded

**Done tracking** is necessary because GitHub's "mark as done" API (`DELETE /notifications/threads/{id}`) doesn't actually delete notifications — they still appear in API responses with `all=true`. Without local tracking, done notifications would reappear when filtering to `is:read` or `is:all`:

- Stored in `~/.local/state/gh-dash/done.json`
- Accessed via `data.GetDoneStore()` singleton
- When marking a notification as done, its ID is persisted to the store
- Done notifications are filtered out during fetch, regardless of API response
- Persists across sessions and application restarts

**Pagination with local filtering**: Because done notifications are filtered out locally after fetching from the API, a single page of results may yield very few visible notifications. To handle this, the fetch logic automatically requests additional pages from the API until the requested limit is reached or all pages are exhausted. This ensures users see a full page of results even when many notifications have been marked as done.

#### 9. Unsubscribe

The unsubscribe feature allows users to stop receiving notifications for a thread:

- Uses GitHub's `DELETE /notifications/threads/{id}/subscription` API
- Removes the subscription without marking the notification as done
- Useful for threads that are no longer relevant but shouldn't be deleted

#### 10. State Management

Notification state (read/unread, done) is tracked both:
- Locally in the `notificationrow.Data` struct for immediate UI updates
- Remotely via GitHub API calls for persistence

The `UpdateNotificationMsg` and `UpdateNotificationReadStateMsg` types propagate state changes through the Bubble Tea update cycle.

**Session persistence for read notifications:** When a notification is marked as read (via `m` key or by viewing it), its ID is tracked in `sessionMarkedRead`. These notifications remain visible in the inbox even during automatic refreshes (e.g., when the terminal regains focus). They are only removed when:
- The user performs a manual refresh (Refresh key)
- The user quits and restarts the application
- The notification is explicitly marked as done

#### 12. Command Architecture

Notification commands are organized in `commands.go`, following the pattern established by `prssection` (which has `checkout.go`, `diff.go`). Commands fall into two categories:

**Section methods** — Commands that operate on section state and are invoked via key handling in the section's `Update` method:
- `markAsDone()` — Marks the current notification as done
- `markAllAsDone()` — Marks all visible notifications as done
- `markAsRead()` — Marks the current notification as read
- `markAllAsRead()` — Marks all notifications as read
- `unsubscribe()` — Unsubscribes from the current thread
- `openInBrowser()` — Marks as read and opens in browser

**Standalone functions** — Commands that require data from outside the section (e.g., the PR shown in the sidebar). These are called from `ui.go` with the necessary parameters:
- `DiffPR(ctx, prNumber, repoName)` — Opens a diff view for a PR
- `CheckoutPR(ctx, prNumber, repoName)` — Checks out a PR branch locally

This split exists because diff and checkout operate on the PR/Issue content shown in the `notificationView` sidebar, not the notification row itself. When viewing a PR notification, the sidebar displays the full PR details, and diff/checkout actions use that data. The section doesn't have access to this enriched data, so `ui.go` extracts the PR details from `notificationView` and passes them to the standalone functions.

Key events are routed through `ui.go`, which either:
1. Passes them to the section via `updateSection()` for section methods
2. Handles them directly and calls standalone functions with the required data

### Interface Compliance

`notificationrow.Data` implements the `data.RowData` interface:

```go
type RowData interface {
    GetRepoNameWithOwner() string
    GetNumber() int
    GetTitle() string
    GetUrl() string
}
```

- `GetNumber()` extracts the PR/Issue number from the subject URL
- `GetUrl()` constructs the GitHub web URL from the API URL

### Styling Consistency

Common styling functions were extracted to `common/styles.go`:

- `RenderPreviewHeader()` — renders the repo/type header line with background
- `RenderPreviewTitle()` — renders the title block with background highlight

These are used by PR view, Issue view, notification view, and notification prompt for consistent appearance.

### Table Column Alignment

The table component was extended to support per-column alignment via an `Align` property on the `Column` struct. This allows the comment count column to be right-aligned.

## Key Bindings

| Key | Action |
|-----|--------|
| D | Mark as done (removes from inbox) |
| Alt+d | Mark all as done |
| m | Mark as read |
| M | Mark all as read |
| u | Unsubscribe from thread |
| b | Toggle bookmark |
| t | Toggle smart filtering (filter to current repo) |
| y | Copy PR/Issue number |
| Y | Copy URL |
| S | Sort by repository |
| s | Switch to PRs view |
| o | Open in browser |
| Enter | View notification (fetches content, marks as read) |

### PR/Issue Keybindings in Notifications View

When viewing a PR notification in the preview pane, all PR-specific keybindings become available:

| Key | Action |
|-----|--------|
| v | Approve PR |
| a | Assign |
| A | Unassign |
| c | Comment |
| d | View diff |
| C/Space | Checkout branch |
| x | Close PR |
| X | Reopen PR |
| W | Mark ready for review |
| m | Merge PR |
| u | Update from base branch |
| w | Watch checks |
| [ | Previous sidebar tab |
| ] | Next sidebar tab |
| e | Expand description |

Similarly, when viewing an Issue notification, Issue-specific keybindings are available:

| Key | Action |
|-----|--------|
| L | Add/remove labels |
| a | Assign |
| A | Unassign |
| c | Comment |
| x | Close issue |
| X | Reopen issue |

The `?` help display dynamically updates to show the applicable keybindings based on what type of notification content is being viewed.

#### Confirmation Prompts for Destructive Actions

When viewing a PR or Issue notification, destructive actions (close, reopen, merge, etc.) require confirmation before execution. This uses a footer-based confirmation mechanism separate from the section-level confirmation used in PR/Issue views:

1. User presses action key (e.g., `x` for close)
2. Footer displays: "Are you sure you want to close PR #123? (y/N)"
3. User presses `y`, `Y`, or `Enter` to confirm, any other key cancels
4. Action executes via the `tasks` package (same as PR/Issue views)

This design is necessary because:
- The notification section doesn't understand PR/Issue-specific actions
- PR/Issue data is stored in `notificationView`, not in the section
- Actions operate on the notification's subject PR/Issue, not the notification itself

The confirmation state is managed by `notificationView.Model`:
- `pendingAction` field tracks the pending action (e.g., "pr_close", "issue_reopen")
- `SetPendingPRAction()` / `SetPendingIssueAction()` set the pending action and return the confirmation prompt text
- `Update()` method handles confirmation key presses (y/Y/Enter to confirm, any other key cancels)
- `onConfirmAction` callback is invoked when confirmed, which `ui.go` sets to `executeNotificationAction()`

This encapsulation keeps confirmation logic close to the view that displays it, while `ui.go` coordinates between the footer prompt and action execution.

## Configuration

### Search Section

Like PRs and Issues, the Notifications view includes a search section (indicated by a magnifying glass icon 🔍) as the first tab. This serves as a scratchpad for one-off searches without modifying your configured sections.

- Default filter: `archived:false`
- Respects `smartFilteringAtLaunch`: when enabled and running from a git repository, the search automatically scopes to that repo
- Use the `/` key to focus the search bar and enter custom queries
- Supports all notification filters: `is:unread`, `is:read`, `repo:owner/name`, `reason:*`

### Notification Sections

Notifications support multiple configurable sections, similar to PRs and Issues. Each section appears as a tab and filters notifications by reason:

```yaml
notificationsSections:
  - title: All
    filters: ""
  - title: Created
    filters: "reason:author"
  - title: Participating
    filters: "reason:participating"
  - title: Mentioned
    filters: "reason:mention"
  - title: Review Requested
    filters: "reason:review-requested"
  - title: Assigned
    filters: "reason:assign"
  - title: Subscribed
    filters: "reason:subscribed"
  - title: Team Mentioned
    filters: "reason:team-mention"
```

These are the default sections. Users can customize by defining their own `notificationsSections` in `config.yml`.

#### Reason Filters

The `reason:` filter matches GitHub's notification reason field:

| Filter | Description |
|--------|-------------|
| `reason:author` | Activity on threads you created |
| `reason:comment` | Someone commented on a thread you're subscribed to |
| `reason:mention` | You were @mentioned |
| `reason:review-requested` | Your review was requested on a PR |
| `reason:assign` | You were assigned |
| `reason:subscribed` | Activity on threads you're watching |
| `reason:team-mention` | Your team was @mentioned |
| `reason:state-change` | Thread state changed (merged, closed, etc.) |
| `reason:ci-activity` | CI workflow activity |
| `reason:participating` | Meta-filter: expands to author, comment, mention, review-requested, assign, state-change |

Reason filters are applied client-side after fetching from GitHub's API.

### Fetch Limit

The initial fetch limit is controlled by `defaults.notificationsLimit` (default: 20, matching PRs and Issues). Additional notifications are fetched automatically as the user scrolls through the list.

```yaml
defaults:
  notificationsLimit: 20
```

### Smart Filtering

Notifications respect the global `smartFilteringAtLaunch` setting (enabled by default). When enabled and running from within a git repository, notifications are automatically scoped to that repository. The search bar displays `repo:owner/name` to indicate this filtering.

Users can:
- Press `t` to toggle filtering on/off for the current session
- Set `smartFilteringAtLaunch: false` in config to disable this behavior globally

#### 11. CheckSuite URL Resolution

GitHub's API returns `subject.url=null` for CheckSuite notifications, making it impossible to directly link to the specific workflow run. To work around this:

1. Initially, CheckSuite notifications link to the repository's `/actions` page as a fallback
2. Asynchronously, we fetch recent workflow runs from `/repos/{owner}/{repo}/actions/runs`
3. We find the workflow run closest in time to the notification's `updated_at` timestamp
4. Once resolved, the notification's URL is updated to point to the specific workflow run

This async resolution uses the existing `UpdateNotificationUrlMsg` message type, following the same pattern as async comment count fetching for PRs and Issues.

## Limitations

- **Mark as Unread**: GitHub's REST API does not support marking notifications as unread, so this feature is not available. Bookmarks provide a workaround by keeping items visible in the inbox.
- **Discussion/Release Content**: Only PR and Issue notifications can display detailed content in the sidebar; other types open directly in the browser.
- **Local State Persistence**: Bookmarks and done status are stored locally (`~/.local/state/gh-dash/`) and are not synced across machines or with GitHub.
- **Done Notifications in API**: GitHub's "mark as done" doesn't delete notifications — they still appear in API responses with `all=true`. We track done IDs locally to filter them out.
- **Server-Side Reason Filtering**: GitHub's notification API does not support filtering by reason on the server side. Reason filters are applied client-side after fetching notifications, which means all notifications are fetched before filtering.
