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

When a notification is selected but not yet viewed, a prompt is displayed showing available actions:

```
   Key    Action
┌───────┬─────────────────┐
│   d   │ mark done       │
│   m   │ mark read       │
│   u   │ unsubscribe     │
│   b   │ toggle bookmark │
│   o   │ open in browser │
│ Enter │ view            │
└───────┴─────────────────┘
```

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
- Appends `@username` to the notification title
- Helps identify spam without needing to open the notification
- Fetched alongside comment counts (no additional latency)
- Color is configurable via `theme.colors.text.actor` (defaults to secondary text color)

#### 5. Bookmark System

Bookmarks allow users to keep notifications visible even after marking them as read. Since GitHub's API doesn't support bookmarks, this is implemented with local storage:

- Bookmarks are stored in `~/.config/gh-dash/bookmarks.json`
- `BookmarkStore` singleton manages bookmark state with thread-safe operations
- Bookmarked notifications appear in the default inbox view even when read
- Bookmarked notifications are styled as read (faint text) but show a bookmark indicator
- When user explicitly searches `is:unread`, bookmarked+read items are excluded

#### 6. Unsubscribe

The unsubscribe feature allows users to stop receiving notifications for a thread:

- Uses GitHub's `DELETE /notifications/threads/{id}/subscription` API
- Removes the subscription without marking the notification as done
- Useful for threads that are no longer relevant but shouldn't be deleted

#### 7. State Management

Notification state (read/unread, done) is tracked both:
- Locally in the `notificationrow.Data` struct for immediate UI updates
- Remotely via GitHub API calls for persistence

The `UpdateNotificationMsg` and `UpdateNotificationReadStateMsg` types propagate state changes through the Bubble Tea update cycle.

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

## Limitations

- **Mark as Unread**: GitHub's REST API does not support marking notifications as unread, so this feature is not available. Bookmarks provide a workaround by keeping items visible in the inbox.
- **Discussion/Release Content**: Only PR and Issue notifications can display detailed content in the sidebar; other types open directly in the browser.
- **CheckSuite/CI Notifications**: GitHub's API returns `subject.url=null` for CheckSuite notifications, so we cannot link to the specific commit checks page. These notifications open the repository's /actions page instead.
- **Bookmark Persistence**: Bookmarks are stored locally and are not synced across machines or with GitHub.
