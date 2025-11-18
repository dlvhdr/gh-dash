# Data Flow and State Management

This document explains how data flows through the application, from user input to API calls to UI updates.

## Table of Contents
1. [Data Flow Overview](#data-flow-overview)
2. [Message Types](#message-types)
3. [State Updates](#state-updates)
4. [API Integration](#api-integration)
5. [Caching Strategy](#caching-strategy)
6. [Error Handling](#error-handling)

---

## Data Flow Overview

### The Elm Architecture Data Flow

```
┌─────────────────────────────────────────────────────────┐
│                    1. User Input                        │
│              (Keyboard, Mouse, Timer)                   │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│                2. Message Created                        │
│         (tea.KeyMsg, tea.MouseMsg, CustomMsg)           │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│              3. Update Function Called                   │
│         model.Update(msg) → (newModel, cmd)             │
│                                                          │
│  • Pattern match on message type                        │
│  • Update state immutably                                │
│  • Return side effects as commands                      │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ├─────────────────┬───────────────┐
                       ▼                 ▼               ▼
              ┌────────────────┐  ┌──────────┐  ┌────────────┐
              │  4a. No Cmd    │  │ 4b. Cmd  │  │ 4c. Batch  │
              │   (just state  │  │ (async   │  │  (multiple │
              │    update)     │  │  task)   │  │   cmds)    │
              └───────┬────────┘  └────┬─────┘  └─────┬──────┘
                      │                │               │
                      │                ▼               │
                      │         ┌─────────────┐        │
                      │         │  Execute    │        │
                      │         │  Command    │        │
                      │         └──────┬──────┘        │
                      │                │               │
                      │                ▼               │
                      │         ┌─────────────┐        │
                      │         │   New Msg   │────────┘
                      │         │  Created    │
                      │         └──────┬──────┘
                      │                │
                      └────────────────┴────────────────┘
                                       │
                                       ▼
                      ┌──────────────────────────────────┐
                      │      5. View Function Called     │
                      │    model.View() → string         │
                      │                                  │
                      │  • Render current state          │
                      │  • Deterministic                 │
                      │  • No side effects               │
                      └────────────┬─────────────────────┘
                                   │
                                   ▼
                      ┌──────────────────────────────────┐
                      │     6. Render to Terminal        │
                      │   (Bubbletea handles this)       │
                      └──────────────────────────────────┘
                                   │
                                   │
                                   ▼
                           Loop back to step 1
```

---

## Message Types

### Built-in Bubbletea Messages

```go
// Keyboard input
type tea.KeyMsg struct {
    Type  tea.KeyType  // KeyEnter, KeyEsc, KeyUp, etc.
    Runes []rune       // Actual characters
    Alt   bool
}

// Mouse input
type tea.MouseMsg struct {
    X, Y   int
    Type   tea.MouseEventType  // MouseLeft, MouseRight, etc.
    Button tea.MouseButton
}

// Window resize
type tea.WindowSizeMsg struct {
    Width  int
    Height int
}

// Batch complete
type tea.BatchMsg struct{}

// Quit
type tea.QuitMsg struct{}
```

### Custom Application Messages

```go
// internal/tui/constants/

// Initialization complete
type InitMsg struct {
    Config *config.Config
    User   *data.User
}

// Error occurred
type ErrMsg struct {
    Err error
}

// Data fetch progress
type ProgressMsg struct {
    Percent float64
    Message string
}

// Task started
type TaskStartedMsg struct {
    Id string
}

// Task finished
type TaskFinishedMsg struct {
    Id     string
    Err    error
    Result interface{}
}
```

### Section-Specific Messages

```go
// internal/tui/components/section/

// Section data fetched
type SectionRowsFetchedMsg struct {
    SectionId int
    Issues    []data.RowData  // Actually PRs or Issues
}

// Section-specific message wrapper
type SectionMsg struct {
    Id          int
    Type        string
    InternalMsg tea.Msg
}
```

---

## State Updates

### Update Flow Example: Refreshing Data

```go
// 1. User presses 'r'
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if key.Matches(msg, m.keys.Refresh) {
            return m.handleRefresh()
        }
    }
}

// 2. Handle refresh
func (m *Model) handleRefresh() (tea.Model, tea.Cmd) {
    section := m.getCurrentSection()

    // Mark as loading
    section.SetIsLoading(true)

    // Create fetch command
    cmd := section.FetchSectionRows()

    // Update model
    m.updateCurrentSection(section)

    return m, cmd
}

// 3. Fetch data (runs in goroutine)
func (s *PRsSection) FetchSectionRows() tea.Cmd {
    return func() tea.Msg {
        // Call GitHub API
        prs, err := data.FetchPullRequests(
            s.GetFilters(),
            s.Config.Limit,
        )

        if err != nil {
            return constants.ErrMsg{Err: err}
        }

        // Return success message
        return section.SectionRowsFetchedMsg{
            SectionId: s.Id,
            Issues:    convertPRsToRows(prs),
        }
    }
}

// 4. Handle fetch result
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case section.SectionRowsFetchedMsg:
        return m.handleSectionRowsFetched(msg)
    }
}

func (m *Model) handleSectionRowsFetched(msg section.SectionRowsFetchedMsg) (tea.Model, tea.Cmd) {
    // Find target section
    section := m.getSectionById(msg.SectionId)

    // Update rows
    section.UpdateRows(msg.Issues)

    // Clear loading
    section.SetIsLoading(false)

    // Update model
    m.updateSection(section)

    return m, nil
}

// 5. View re-renders automatically
func (m *Model) View() string {
    return m.getCurrentSection().View()
}
```

### State Immutability

```go
// Pattern 1: Return new model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    newModel := m  // Copy
    newModel.index++
    return newModel, nil
}

// Pattern 2: Pointer receiver (mutate in place, but conceptually immutable)
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.index++  // Mutate
    return m, nil  // Return pointer
}

// In gh-dash, we use Pattern 2 for performance
// But treat model as immutable conceptually
```

---

## API Integration

### GraphQL Query Flow

```
User requests data
    ↓
Section.FetchSectionRows()
    ↓
Construct GraphQL query
    ↓
Call GitHub API via go-gh
    ↓
Parse response
    ↓
Map to internal data structures
    ↓
Return SectionRowsFetchedMsg
    ↓
Update() handles message
    ↓
Update section rows
    ↓
View() re-renders
```

### Example: Fetching Pull Requests

```go
// internal/data/prapi.go

func FetchPullRequests(
    filters string,
    limit int,
) ([]PullRequestData, error) {
    // 1. Get GraphQL client
    client, err := gh.DefaultGraphQLClient()
    if err != nil {
        return nil, err
    }

    // 2. Define query structure
    var query struct {
        Search struct {
            Nodes []struct {
                PullRequest PullRequestData `graphql:"... on PullRequest"`
            }
            PageInfo PageInfo
        } `graphql:"search(query: $query, type: ISSUE, first: $limit)"`
    }

    // 3. Set variables
    variables := map[string]interface{}{
        "query": graphql.String(filters),
        "limit": graphql.Int(limit),
    }

    // 4. Execute query
    err = client.Query("FetchPRs", &query, variables)
    if err != nil {
        return nil, err
    }

    // 5. Extract data
    prs := make([]PullRequestData, 0)
    for _, node := range query.Search.Nodes {
        prs = append(prs, node.PullRequest)
    }

    return prs, nil
}
```

### API Call Patterns

```go
// Pattern 1: Simple fetch
func fetchData() tea.Cmd {
    return func() tea.Msg {
        data, err := api.Fetch()
        if err != nil {
            return ErrMsg{Err: err}
        }
        return DataFetchedMsg{Data: data}
    }
}

// Pattern 2: With progress updates
func fetchDataWithProgress() tea.Cmd {
    return func() tea.Msg {
        progressChan := make(chan float64)

        go func() {
            api.FetchWithProgress(progressChan)
        }()

        for progress := range progressChan {
            // Send progress updates
            tea.Send(ProgressMsg{Percent: progress})
        }

        return DataFetchedMsg{}
    }
}

// Pattern 3: With pagination
func fetchAllPages() tea.Cmd {
    return func() tea.Msg {
        var allData []Item
        cursor := ""

        for {
            page, nextCursor, err := api.FetchPage(cursor)
            if err != nil {
                return ErrMsg{Err: err}
            }

            allData = append(allData, page...)

            if nextCursor == "" {
                break
            }
            cursor = nextCursor
        }

        return DataFetchedMsg{Data: allData}
    }
}
```

---

## Caching Strategy

### Section-Level Caching

```go
type BaseModel struct {
    // Cached data
    Rows         []RowData
    TotalCount   int
    PageInfo     *data.PageInfo
    LastFetched  time.Time

    // Cache validity
    IsLoading    bool
}

// Check if cache is valid
func (m *BaseModel) IsCacheValid() bool {
    if m.Rows == nil {
        return false
    }

    // Cache expires after N minutes
    expiryDuration := time.Duration(
        m.Ctx.Config.Defaults.RefetchIntervalMinutes,
    ) * time.Minute

    return time.Since(m.LastFetched) < expiryDuration
}

// Use cache or fetch fresh
func (m *BaseModel) GetRows() tea.Cmd {
    if m.IsCacheValid() {
        return nil  // Use cached data
    }

    return m.fetchFreshData()
}
```

### Auto-Refresh Strategy

```go
// Set up auto-refresh
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.initialFetch(),
        m.setupAutoRefresh(),
    )
}

func (m Model) setupAutoRefresh() tea.Cmd {
    interval := time.Duration(
        m.ctx.Config.Defaults.RefetchIntervalMinutes,
    ) * time.Minute

    return tea.Tick(interval, func(t time.Time) tea.Msg {
        return AutoRefreshMsg{Time: t}
    })
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case AutoRefreshMsg:
        // Auto-refresh all sections
        cmds := []tea.Cmd{
            m.refreshAllSections(),
            m.setupAutoRefresh(),  // Schedule next refresh
        }
        return m, tea.Batch(cmds...)
    }
}
```

### Invalidation Strategy

```go
// Invalidate cache on actions that modify data
func (m *Model) mergePR() tea.Cmd {
    return func() tea.Msg {
        err := api.MergePR(prNumber)
        if err != nil {
            return ErrMsg{Err: err}
        }

        // Invalidate cache
        return CacheInvalidatedMsg{
            SectionId: m.Id,
        }
    }
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case CacheInvalidatedMsg:
        section := m.getSectionById(msg.SectionId)
        section.ResetCache()
        return m, section.FetchSectionRows()
    }
}
```

---

## Error Handling

### Error Flow

```
Error occurs
    ↓
Wrapped in ErrMsg
    ↓
Sent to Update()
    ↓
Stored in ProgramContext
    ↓
View() renders error
    ↓
User sees error message
    ↓
Next successful action clears error
```

### Error Message Types

```go
// internal/tui/constants/errMsg.go

type ErrMsg struct {
    Err error
}

func (e ErrMsg) Error() string {
    return e.Err.Error()
}

// Usage in API calls
func fetchData() tea.Cmd {
    return func() tea.Msg {
        data, err := api.Fetch()
        if err != nil {
            return constants.ErrMsg{
                Err: fmt.Errorf("failed to fetch data: %w", err),
            }
        }
        return DataFetchedMsg{Data: data}
    }
}
```

### Error Handling in Update

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case constants.ErrMsg:
        // Store error in context
        m.ctx.Error = msg.Err

        // Log error
        log.Error("Error occurred", "err", msg.Err)

        // Clear loading states
        m.clearLoadingStates()

        return m, nil

    // Other messages clear error
    case tea.KeyMsg:
        m.ctx.Error = nil
        // ... handle key
    }

    return m, nil
}
```

### Error Rendering

```go
func (m Model) View() string {
    var content string

    // Render error if present
    if m.ctx.Error != nil {
        errorView := m.ctx.Styles.Error.Render(
            fmt.Sprintf("Error: %s", m.ctx.Error.Error()),
        )
        content = lipgloss.JoinVertical(
            lipgloss.Left,
            errorView,
            "",
            m.renderMainContent(),
        )
    } else {
        content = m.renderMainContent()
    }

    return content
}
```

### Retry Logic

```go
// Retry with exponential backoff
func fetchWithRetry(maxRetries int) tea.Cmd {
    return func() tea.Msg {
        var lastErr error

        for i := 0; i < maxRetries; i++ {
            data, err := api.Fetch()
            if err == nil {
                return DataFetchedMsg{Data: data}
            }

            lastErr = err

            // Exponential backoff
            backoff := time.Duration(1<<i) * time.Second
            time.Sleep(backoff)
        }

        return ErrMsg{
            Err: fmt.Errorf("failed after %d retries: %w",
                maxRetries, lastErr),
        }
    }
}
```

---

## Task Management

### Task System

```go
// internal/tui/context/context.go

type Task struct {
    Id        string
    StartTime time.Time
    Message   string
}

type ProgramContext struct {
    // ...
    StartTask func(Task) tea.Cmd
}

// Start a task
func (m *Model) startFetch() tea.Cmd {
    task := context.Task{
        Id:      uuid.New().String(),
        Message: "Fetching PRs...",
    }

    return m.ctx.StartTask(task)
}

// Task management in Model
func (m *Model) startTask(task Task) tea.Cmd {
    m.tasks[task.Id] = task
    m.footer.SetRightSection(m.renderRunningTask())
    return m.taskSpinner.Tick
}

func (m *Model) finishTask(taskId string) {
    delete(m.tasks, taskId)
    m.footer.SetRightSection("")
}
```

### Task Flow Example

```
1. User action triggers task
   ↓
2. StartTask() called
   ↓
3. Task added to tasks map
   ↓
4. Spinner starts
   ↓
5. Footer shows "Loading..."
   ↓
6. Async operation executes
   ↓
7. TaskFinishedMsg sent
   ↓
8. Task removed from map
   ↓
9. Spinner stops
   ↓
10. Footer cleared
```

---

## Summary

Data flow in gh-dash follows these principles:

1. **Unidirectional**: Always flows User → Update → Model → View
2. **Message-driven**: All changes via messages
3. **Immutable**: State updates create new state
4. **Async-friendly**: Commands for side effects
5. **Error-safe**: Errors handled gracefully
6. **Cached**: Data cached with TTL
7. **Task-managed**: Long operations tracked

Key patterns:
- **Elm Architecture**: Model-Update-View loop
- **Command Pattern**: Side effects as commands
- **Message Passing**: Components communicate via messages
- **Task System**: Track async operations
- **Error Handling**: Centralized error display
- **Caching**: Reduce API calls

This architecture ensures:
- **Predictability**: Same input → same output
- **Debuggability**: Message log = complete history
- **Testability**: Pure functions easy to test
- **Performance**: Caching reduces API calls
- **Reliability**: Errors don't crash app
