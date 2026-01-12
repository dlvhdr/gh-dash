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
│       │   └── notificationrow.go # Row rendering for the table
│       ├── notificationssection/
│       │   ├── notificationssection.go # Main section component
│       │   ├── commands.go      # Tea commands (mark done, mark read, etc.)
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

            [d]  mark as done
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

To preserve parent background colors, row content uses raw ANSI escape codes for foreground styling without trailing reset sequences. The `getStylePrefix()` helper extracts ANSI codes from lipgloss styles while stripping the reset, preventing internal resets from breaking the cell's background color.

Title truncation is handled dynamically by the table component based on actual column width, with ellipsis added when content is truncated. This allows titles to adjust when the sidebar is shown/hidden.

#### 8. Bookmark System

Bookmarks allow users to keep notifications visible even after marking them as read. Since GitHub's API doesn't support bookmarks, this is implemented locally:

- Bookmarks are stored in `~/.config/gh-dash/bookmarks.json`
- `BookmarkStore` singleton manages bookmark state with thread-safe operations
- Bookmarked notifications appear in the default inbox view even when read
- Bookmarked notifications are styled as read (faint text) but show a bookmark indicator
- When user explicitly searches `is:unread`, bookmarked+read items are excluded

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
| d | Mark as done (removes from inbox) |
| D | Mark all as done |
| m | Mark as read |
| M | Mark all as read |
| u | Unsubscribe from thread |
| b | Toggle bookmark |
| t | Toggle smart filtering (filter to current repo) |
| y | Copy PR/Issue number |
| Y | Copy URL |
| S | Sort by repository |
| o | Open in browser |
| Enter | View notification (fetches content, marks as read) |

## Configuration

Notifications are enabled via `config.yml`:

```yaml
notifications:
  enabled: true
  filters:
    - query: "default"  # Shows all notifications
```

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
- **Bookmark Persistence**: Bookmarks are stored locally and are not synced across machines or with GitHub.
- **Section Configuration**: Unlike PRs and Issues which support multiple user-defined sections with custom filters and titles, notifications currently use a single hardcoded section. User-configurable notification sections may be added in a future release.
