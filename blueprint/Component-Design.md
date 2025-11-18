# Component Design Guide

This document provides detailed guidance on how components are designed, structured, and implemented in gh-dash. Use this as a blueprint for creating new components or understanding existing ones.

## Table of Contents
1. [Component Architecture](#component-architecture)
2. [Component Lifecycle](#component-lifecycle)
3. [Component Catalog](#component-catalog)
4. [Creating New Components](#creating-new-components)
5. [Component Best Practices](#component-best-practices)

---

## Component Architecture

### What is a Component?

A component in gh-dash is a self-contained UI element that:
- Manages its own local state
- Receives external state via ProgramContext
- Responds to messages (events)
- Renders itself to a string
- Follows the Elm Architecture pattern

### Component Interface

```go
// Minimal component interface
type Component interface {
    Update(msg tea.Msg) (Component, tea.Cmd)
    View() string
}

// Most components also implement
type LifecycleComponent interface {
    Component
    Init() tea.Cmd
}

// And may implement
type ResizableComponent interface {
    Component
    UpdateProgramContext(ctx *context.ProgramContext)
}
```

### Component Anatomy

```go
// internal/tui/components/{component}/component.go

package component

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

// 1. MODEL: Component state
type Model struct {
    // Internal state
    selectedIndex int
    items         []string
    isActive      bool

    // External dependencies
    ctx *context.ProgramContext

    // Composed components (if any)
    subComponent SubComponent
}

// 2. CONSTRUCTOR: Create with initial state
func NewModel(ctx *context.ProgramContext) Model {
    return Model{
        selectedIndex: 0,
        items:         []string{},
        isActive:      false,
        ctx:           ctx,
    }
}

// 3. INIT: Return initial command
func (m Model) Init() tea.Cmd {
    // Return command for initial setup
    // Or nil if no initialization needed
    return nil
}

// 4. UPDATE: Handle messages
func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyMsg(msg)
    case CustomMsg:
        return m.handleCustomMsg(msg)
    }
    return *m, nil
}

// 5. VIEW: Render to string
func (m Model) View() string {
    // Render based on current state
    return m.renderContent()
}

// 6. HELPERS: Private methods
func (m Model) renderContent() string {
    // Rendering logic
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
    // Key handling logic
}
```

---

## Component Lifecycle

### 1. Creation Phase

```go
// Component is created
model := NewModel(ctx)

// Initial command returned
cmd := model.Init()

// Command executed (if not nil)
if cmd != nil {
    msg := cmd()  // Execute command
    // Message will be sent to Update
}
```

### 2. Update Phase (Event Loop)

```
Message arrives
    ↓
Update(msg) called
    ↓
Match message type
    ↓
Update state
    ↓
Return (newModel, cmd)
    ↓
View() called
    ↓
Render to terminal
    ↓
Wait for next message
```

### 3. Resize/Context Update

```go
// Terminal resized or context changed
newCtx := getUpdatedContext()

// Component receives update
component.UpdateProgramContext(newCtx)

// Component recalculates dimensions
component.recalculate()

// View() will use new dimensions
```

### 4. Cleanup Phase

```
User quits
    ↓
No explicit cleanup needed
    ↓
Bubbletea handles shutdown
    ↓
Terminal restored
```

---

## Component Catalog

### 1. Section Component (Base)

**Location**: `internal/tui/components/section/`

**Purpose**: Abstract base for PR/Issue/Repo sections

**State**:
```go
type BaseModel struct {
    Id                    int
    Config                config.SectionConfig
    Ctx                   *context.ProgramContext
    Table                 table.Model
    SearchBar             search.Model
    PromptConfirmationBox prompt.Model
    IsSearching           bool
    TotalCount            int
    // ...
}
```

**Key Methods**:
- `GetDimensions()` - Calculate available space
- `SetIsSearching(bool)` - Toggle search mode
- `ResetFilters()` - Reset to default filters
- `BuildRows()` - Convert data to table rows

**Usage**:
```go
section := section.NewModel(ctx, section.NewSectionOptions{
    Id:       1,
    Type:     "pr",
    Config:   config.PRSections[0],
    Columns:  columns,
    Singular: "PR",
    Plural:   "PRs",
})
```

---

### 2. Table Component

**Location**: `internal/tui/components/table/`

**Purpose**: Display tabular data with scrolling

**State**:
```go
type Model struct {
    Columns      []Column
    Rows         []Row
    currItem     int           // Current selection
    viewport     ListViewPort  // Scrolling
    dimensions   Dimensions
    // ...
}
```

**Key Features**:
- Column configuration (width, hidden, alignment)
- Virtual scrolling (only render visible rows)
- Selection tracking
- Keyboard navigation
- Empty/loading states

**Usage**:
```go
table := table.NewModel(
    ctx,
    dimensions,
    lastUpdated,
    createdAt,
    columns,
    rows,
    "PR",
    emptyStateMsg,
    loadingMsg,
    compact,
)
```

**Column Definition**:
```go
type Column struct {
    Title   string
    Width   *int  // nil = auto
    Hidden  *bool
    Grow    *bool // Expand to fill space
    Render  func(data RowData) string
}

// Example
columns := []table.Column{
    {
        Title: "Title",
        Grow:  utils.BoolPtr(true),
        Render: func(data RowData) string {
            return data.(*PullRequestData).Title
        },
    },
    {
        Title: "Updated",
        Width: utils.IntPtr(10),
        Render: func(data RowData) string {
            return formatTime(data.(*PullRequestData).UpdatedAt)
        },
    },
}
```

---

### 3. Sidebar Components

#### PR Sidebar

**Location**: `internal/tui/components/prsidebar/`

**Purpose**: Display PR details, activity, files, checks

**State**:
```go
type Model struct {
    pr              *data.PullRequestData
    activeTab       int  // 0: Activity, 1: Files, 2: Checks
    activitySection Activity
    filesSection    Files
    checksSection   Checks
    ctx             *context.ProgramContext
}
```

**Tabs**:
1. **Activity**: Comments, reviews, timeline
2. **Files**: Changed files with diff stats
3. **Checks**: CI/CD status

#### Issue Sidebar

**Location**: `internal/tui/components/issuesidebar/`

**Purpose**: Display issue details, comments, labels

**State**:
```go
type Model struct {
    issue           *data.IssueData
    activeTab       int  // 0: Activity, 1: Labels
    activitySection Activity
    ctx             *context.ProgramContext
}
```

#### Branch Sidebar

**Location**: `internal/tui/components/branchsidebar/`

**Purpose**: Display branch details, commit info

**State**:
```go
type Model struct {
    branch *git.Branch
    ctx    *context.ProgramContext
}
```

---

### 4. Search Component

**Location**: `internal/tui/components/search/`

**Purpose**: Filter search input

**State**:
```go
type Model struct {
    textInput   textinput.Model  // Bubbles text input
    prefix      string           // "is:pr", "is:issue"
    ctx         *context.ProgramContext
}
```

**Key Features**:
- Prefix display (e.g., "is:pr")
- Input validation
- Submit on Enter
- Escape to cancel

**Usage**:
```go
search := search.NewModel(ctx, search.SearchOptions{
    Prefix:       "is:open",
    InitialValue: "author:@me",
})

// Focus for input
search.Focus()

// Get value
filters := search.Value()
```

---

### 5. Footer Component

**Location**: `internal/tui/components/footer/`

**Purpose**: Display help, status, running tasks

**State**:
```go
type Model struct {
    leftSection  string  // Help text
    rightSection string  // Status/tasks
    ctx          *context.ProgramContext
}
```

**Layout**:
```
┌──────────────────────────────────────────────────┐
│ ?: help • r: refresh │         Running task... ✓ │
└──────────────────────────────────────────────────┘
```

**Key Methods**:
- `SetLeftSection(text)` - Update help text
- `SetRightSection(text)` - Update status
- `RenderRunningTasks(tasks)` - Show active tasks

---

### 6. Tabs Component

**Location**: `internal/tui/components/tabs/`

**Purpose**: View switcher (PRs, Issues, Repo)

**State**:
```go
type Model struct {
    tabs      []string
    activeTab int
    ctx       *context.ProgramContext
}
```

**Layout**:
```
┌──────────────────────────────────┐
│ [PRs] | Issues | Repo            │
└──────────────────────────────────┘
```

**Key Features**:
- Active tab highlighting
- Keyboard navigation (Ctrl+p, Ctrl+i, Ctrl+r)
- Dynamic tab visibility

---

### 7. Prompt Component

**Location**: `internal/tui/components/prompt/`

**Purpose**: User confirmation/input dialogs

**State**:
```go
type Model struct {
    textInput textinput.Model
    prompt    string
    ctx       *context.ProgramContext
}
```

**Usage Scenarios**:
- Confirmation: "Are you sure? (Y/n)"
- Input: "Enter branch name: ___"
- Text entry: "Comment: ___"

**Key Methods**:
- `SetPrompt(text)` - Set prompt text
- `Focus()` - Activate for input
- `Value()` - Get entered value

---

### 8. Spinner Component

**Location**: Uses Bubbles `spinner.Model`

**Purpose**: Loading indicator

**Usage**:
```go
spinner := spinner.Model{
    Spinner: spinner.Dot,
}

// In Update
case spinner.TickMsg:
    m.spinner, cmd = m.spinner.Update(msg)
    return m, cmd

// In View
if m.isLoading {
    return m.spinner.View() + " Loading..."
}
```

---

## Creating New Components

### Step-by-Step Guide

#### 1. Create Component Directory

```bash
mkdir -p internal/tui/components/mycomponent
cd internal/tui/components/mycomponent
```

#### 2. Define Model

```go
// mycomponent.go
package mycomponent

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type Model struct {
    // Component state
    value string
    items []string
    index int

    // External dependencies
    ctx *context.ProgramContext
}
```

#### 3. Implement Constructor

```go
type Options struct {
    InitialValue string
    Items        []string
}

func NewModel(ctx *context.ProgramContext, opts Options) Model {
    return Model{
        value: opts.InitialValue,
        items: opts.Items,
        index: 0,
        ctx:   ctx,
    }
}
```

#### 4. Implement Init

```go
func (m Model) Init() tea.Cmd {
    // Return initial command
    // E.g., fetch data, start animation, etc.
    return m.fetchData()
}

func (m Model) fetchData() tea.Cmd {
    return func() tea.Msg {
        // Fetch data asynchronously
        data := loadData()
        return DataFetchedMsg{Data: data}
    }
}
```

#### 5. Implement Update

```go
type DataFetchedMsg struct {
    Data []string
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyMsg(msg)

    case DataFetchedMsg:
        m.items = msg.Data
        return *m, nil

    case tea.WindowSizeMsg:
        return m.handleResize(msg)
    }

    return *m, nil
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch msg.String() {
    case "j", "down":
        m.index = min(m.index+1, len(m.items)-1)
    case "k", "up":
        m.index = max(m.index-1, 0)
    case "enter":
        return m.selectItem()
    }
    return *m, nil
}
```

#### 6. Implement View

```go
func (m Model) View() string {
    if len(m.items) == 0 {
        return m.renderEmptyState()
    }

    return m.renderItems()
}

func (m Model) renderEmptyState() string {
    return m.ctx.Styles.EmptyState.Render("No items")
}

func (m Model) renderItems() string {
    var items []string
    for i, item := range m.items {
        if i == m.index {
            items = append(items, m.ctx.Styles.Selected.Render("> " + item))
        } else {
            items = append(items, "  " + item)
        }
    }
    return strings.Join(items, "\n")
}
```

#### 7. Add Helper Methods

```go
func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
    m.ctx = ctx
    // Recalculate dimensions if needed
}

func (m Model) GetSelectedItem() string {
    if m.index >= 0 && m.index < len(m.items) {
        return m.items[m.index]
    }
    return ""
}

func (m *Model) Reset() {
    m.index = 0
    m.value = ""
}
```

#### 8. Add Tests

```go
// mycomponent_test.go
package mycomponent

import (
    "testing"
    tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
    ctx := &context.ProgramContext{}
    opts := Options{Items: []string{"a", "b", "c"}}

    m := NewModel(ctx, opts)

    if len(m.items) != 3 {
        t.Errorf("Expected 3 items, got %d", len(m.items))
    }
}

func TestUpdate_KeyDown(t *testing.T) {
    m := NewModel(ctx, Options{Items: []string{"a", "b"}})

    newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})

    if newModel.index != 1 {
        t.Errorf("Expected index 1, got %d", newModel.index)
    }
}
```

#### 9. Integrate into Parent

```go
// In parent component (e.g., ui.go)
import "github.com/dlvhdr/gh-dash/v4/internal/tui/components/mycomponent"

type Model struct {
    // ...
    myComponent mycomponent.Model
}

func NewModel() Model {
    // ...
    m.myComponent = mycomponent.NewModel(m.ctx, mycomponent.Options{
        Items: []string{"item1", "item2"},
    })
    return m
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Update component
    newComponent, cmd := m.myComponent.Update(msg)
    m.myComponent = newComponent
    return m, cmd
}

func (m *Model) View() string {
    // Include component view
    return m.myComponent.View()
}
```

---

## Component Best Practices

### 1. Single Responsibility

Each component should do one thing well.

**Good**:
```go
// SearchBar only handles search input
type SearchBar struct {
    input textinput.Model
}

// Table only handles displaying rows
type Table struct {
    rows []Row
}
```

**Bad**:
```go
// Component doing too much
type Mega struct {
    search  string
    rows    []Row
    sidebar Sidebar
    footer  Footer
    // Too many responsibilities!
}
```

### 2. Immutability

Treat state as immutable. Update returns new model.

**Good**:
```go
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    newModel := m  // Copy
    newModel.index++
    return newModel, nil
}
```

**Bad**:
```go
func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    m.index++  // Mutating in place (acceptable but less idiomatic)
    return *m, nil
}
```

### 3. Pure View Functions

View should be deterministic - same state → same output.

**Good**:
```go
func (m Model) View() string {
    // Only reads m, no mutations
    return fmt.Sprintf("Index: %d", m.index)
}
```

**Bad**:
```go
func (m *Model) View() string {
    m.renderCount++  // Side effect! Don't do this
    return fmt.Sprintf("Rendered %d times", m.renderCount)
}
```

### 4. Composition Over Inheritance

Build complex components from simple ones.

**Good**:
```go
type ComplexComponent struct {
    header Header
    body   Body
    footer Footer
}

func (m ComplexComponent) View() string {
    return lipgloss.JoinVertical(
        lipgloss.Left,
        m.header.View(),
        m.body.View(),
        m.footer.View(),
    )
}
```

### 5. Context for Shared State

Use ProgramContext for theme, config, dimensions.

**Good**:
```go
type Model struct {
    ctx *context.ProgramContext
}

func (m Model) View() string {
    style := m.ctx.Styles.Primary
    return style.Render("Text")
}
```

**Bad**:
```go
type Model struct {
    theme       Theme
    config      Config
    screenWidth int
    // Duplicating shared state
}
```

### 6. Message Types for Communication

Use custom message types for component communication.

**Good**:
```go
type ItemSelectedMsg struct {
    Index int
    Item  string
}

func (m Model) selectItem() tea.Cmd {
    return func() tea.Msg {
        return ItemSelectedMsg{
            Index: m.index,
            Item:  m.items[m.index],
        }
    }
}
```

**Bad**:
```go
// Using global variables
var selectedItem string

func (m Model) selectItem() tea.Cmd {
    selectedItem = m.items[m.index]  // Bad!
    return nil
}
```

### 7. Error Handling

Handle errors gracefully, show to user.

**Good**:
```go
func (m Model) fetchData() tea.Cmd {
    return func() tea.Msg {
        data, err := api.Fetch()
        if err != nil {
            return constants.ErrMsg{Err: err}
        }
        return DataFetchedMsg{Data: data}
    }
}

// In Update
case constants.ErrMsg:
    m.error = msg.Err
    return m, nil

// In View
if m.error != nil {
    return errorStyle.Render(m.error.Error())
}
```

### 8. Testability

Write components to be easily testable.

```go
// Testable: no external dependencies in constructor
func NewModel(ctx *context.ProgramContext) Model {
    return Model{ctx: ctx}
}

// Test
func TestComponent(t *testing.T) {
    ctx := &context.ProgramContext{
        ScreenWidth: 100,
    }
    m := NewModel(ctx)

    // Test Update
    newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})

    // Test View
    output := newM.View()
    assert.Contains(t, output, "expected text")
}
```

### 9. Documentation

Document public methods and complex logic.

```go
// Model represents a searchable list component.
// It handles keyboard navigation, filtering, and selection.
type Model struct {
    // items contains all available items
    items []string

    // filteredItems contains items matching current filter
    filteredItems []string

    // index is the currently selected item index
    index int
}

// NewModel creates a new Model with the given items.
// The initial filter is empty, showing all items.
func NewModel(items []string) Model {
    // ...
}

// Filter updates the filtered items based on the query.
// Returns a command that will trigger a re-render.
func (m *Model) Filter(query string) tea.Cmd {
    // ...
}
```

### 10. Performance Considerations

#### Virtual Scrolling
```go
// Only render visible rows
func (m Model) View() string {
    start := m.viewport.YOffset
    end := start + m.viewport.Height
    visibleRows := m.rows[start:end]

    return m.renderRows(visibleRows)
}
```

#### Memoization
```go
// Cache expensive renders
type Model struct {
    cachedView  string
    cacheValid  bool
}

func (m Model) View() string {
    if m.cacheValid {
        return m.cachedView
    }

    view := m.expensiveRender()
    m.cachedView = view
    m.cacheValid = true
    return view
}

// Invalidate cache on update
func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    m.cacheValid = false
    // ... handle update
}
```

---

## Summary

Components in gh-dash are:

1. **Self-contained**: Own state, own rendering
2. **Composable**: Build complex UIs from simple parts
3. **Testable**: Pure functions, clear interfaces
4. **Reusable**: Generic, configurable
5. **Type-safe**: Compile-time checks
6. **Elm-based**: Model-Update-View pattern

When creating components:
- Follow the lifecycle (Init → Update → View)
- Keep them focused (single responsibility)
- Make them composable (embed smaller components)
- Handle errors gracefully
- Write tests
- Document public APIs

This architecture makes gh-dash maintainable, extensible, and robust.
