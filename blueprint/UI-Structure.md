# UI Structure and Layout

This document describes how the user interface is structured, organized, and rendered in gh-dash.

## Table of Contents
1. [Overall Layout](#overall-layout)
2. [View Types](#view-types)
3. [Screen Layout](#screen-layout)
4. [Component Hierarchy](#component-hierarchy)
5. [Responsive Design](#responsive-design)
6. [Navigation Patterns](#navigation-patterns)
7. [Modal and Overlay System](#modal-and-overlay-system)

---

## Overall Layout

gh-dash uses a multi-pane layout that adapts based on the current view and user configuration.

### Main Layout Structure

```
┌────────────────────────────────────────────────────────────┐
│ TABS: PRs | Issues | Repo                  (if applicable) │
├──────────────────────────────┬─────────────────────────────┤
│                              │                             │
│                              │                             │
│      MAIN CONTENT            │       SIDEBAR               │
│      (Section)               │       (Details/Preview)     │
│                              │                             │
│  • Table of items            │  • Item details             │
│  • Search bar                │  • Activity                 │
│  • Current selection         │  • Files changed            │
│                              │  • CI/CD checks             │
│                              │                             │
├──────────────────────────────┴─────────────────────────────┤
│ FOOTER: Help • Status • Running tasks                      │
└────────────────────────────────────────────────────────────┘
```

### Layout Dimensions

```go
// internal/tui/context/context.go

type ProgramContext struct {
    ScreenWidth       int  // Total terminal width
    ScreenHeight      int  // Total terminal height
    MainContentWidth  int  // Width available for content
    MainContentHeight int  // Height available for content
}

// Calculations (from ui.go)
func (m *Model) calculateDimensions() {
    // Full screen
    m.ctx.ScreenWidth = termWidth
    m.ctx.ScreenHeight = termHeight

    // Account for sidebar
    sidebarWidth := 0
    if m.sidebar.IsOpen() {
        sidebarWidth = m.ctx.Config.Defaults.Preview.Width
    }

    // Main content area
    m.ctx.MainContentWidth = termWidth - sidebarWidth
    m.ctx.MainContentHeight = termHeight - footerHeight - tabsHeight

    // Pass to components
    m.updateComponentDimensions()
}
```

---

## View Types

gh-dash has three main views, each with a distinct layout and purpose.

### 1. PRs View (`config.PRsView`)

**Purpose**: Browse and manage pull requests

**Layout**:
```
┌────────────────────────────────────────────────────────────┐
│ [PRs] | Issues | Repo                                      │
├──────────────────────────────┬─────────────────────────────┤
│ Search: is:open author:@me   │                             │
├──────────────────────────────┤                             │
│                              │                             │
│ # │ Updated │ Repo │ Title   │    PR DETAILS               │
│ ─────────────────────────────│                             │
│ 1 │ 2h ago  │ api  │ Fix bug │  📝 Description             │
│►2 │ 1d ago  │ web  │ Feature │  💬 Activity                │
│ 3 │ 3d ago  │ cli  │ Update  │  📁 Files changed (3)       │
│                              │  ✓ Checks                   │
│ My PRs (3)                   │                             │
│                              │                             │
├──────────────────────────────┴─────────────────────────────┤
│ ?: help • r: refresh • /: search • enter: open             │
└────────────────────────────────────────────────────────────┘
```

**Features**:
- Multiple sections (configurable via config)
- Tab between sections
- Toggle preview sidebar
- Search/filter PRs
- Keyboard shortcuts for PR actions

### 2. Issues View (`config.IssuesView`)

**Purpose**: Browse and manage issues

**Layout**:
```
┌────────────────────────────────────────────────────────────┐
│ PRs | [Issues] | Repo                                      │
├──────────────────────────────┬─────────────────────────────┤
│ Search: is:open assignee:@me │                             │
├──────────────────────────────┤                             │
│                              │                             │
│ # │ Created │ Repo │ Title   │    ISSUE DETAILS            │
│ ─────────────────────────────│                             │
│►1 │ 2h ago  │ api  │ Bug     │  📝 Description             │
│ 2 │ 5h ago  │ web  │ Feature │  💬 Comments (5)            │
│ 3 │ 1d ago  │ cli  │ Question│  🏷  Labels                 │
│                              │  👥 Assignees               │
│ My Issues (3)                │                             │
│                              │                             │
├──────────────────────────────┴─────────────────────────────┤
│ ?: help • r: refresh • /: search • enter: open             │
└────────────────────────────────────────────────────────────┘
```

**Features**:
- Multiple sections (configurable)
- Toggle preview sidebar
- Search/filter issues
- Keyboard shortcuts for issue actions

### 3. Repo View (`config.RepoView`)

**Purpose**: Manage local repository branches

**Layout**:
```
┌────────────────────────────────────────────────────────────┐
│ PRs | Issues | [Repo]                                      │
├──────────────────────────────┬─────────────────────────────┤
│ Repository: dlvhdr/gh-dash   │                             │
├──────────────────────────────┤                             │
│                              │                             │
│ Branch    │ Updated │ Status │    BRANCH DETAILS           │
│ ──────────────────────────────│                             │
│►main      │ 2h ago  │ ↑5 ↓2  │  📊 Commits ahead: 5        │
│ feature/x │ 1d ago  │ ↑0 ↓10 │  📊 Commits behind: 2       │
│ hotfix/y  │ 3d ago  │ ↑1 ↓0  │  📅 Last commit: 2h ago     │
│                              │  💬 "Fix critical bug"      │
│ Branches (3)                 │  🌿 Origin: main            │
│                              │                             │
├──────────────────────────────┴─────────────────────────────┤
│ ?: help • c: checkout • d: delete • p: create PR           │
└────────────────────────────────────────────────────────────┘
```

**Features**:
- List local branches
- See commits ahead/behind
- Checkout, delete, create PR from branch
- View branch details

---

## Screen Layout

### Layout Components

#### 1. Tabs Component

```go
// internal/tui/components/tabs/tabs.go

type Model struct {
    tabs        []string      // ["PRs", "Issues", "Repo"]
    activeTab   int           // Current tab index
    ctx         *context.ProgramContext
}

func (m Model) View() string {
    // Render tabs with active highlighting
    var tabs []string
    for i, tab := range m.tabs {
        if i == m.activeTab {
            tabs = append(tabs, activeTabStyle.Render(tab))
        } else {
            tabs = append(tabs, inactiveTabStyle.Render(tab))
        }
    }
    return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}
```

**Position**: Top of screen (if not RepoView)
**Height**: 2 lines
**Content**: View names with active indicator

#### 2. Main Content (Section)

```go
// internal/tui/components/section/section.go

func (m *BaseModel) View() string {
    search := m.SearchBar.View(m.Ctx)
    return m.Ctx.Styles.Section.ContainerStyle.Render(
        lipgloss.JoinVertical(
            lipgloss.Left,
            search,           // Search bar
            m.GetMainContent(), // Table or empty state
        ),
    )
}
```

**Position**: Left side of screen
**Width**: `ScreenWidth - SidebarWidth`
**Content**:
- Search bar (1-2 lines)
- Table with rows
- Section title and count

#### 3. Sidebar

```go
// internal/tui/components/sidebar/sidebar.go

type Model struct {
    isOpen       bool
    content      string
    ctx          *context.ProgramContext
}

func (m Model) View() string {
    if !m.isOpen {
        return ""
    }
    return m.ctx.Styles.Sidebar.ContainerStyle.Render(m.content)
}
```

**Position**: Right side of screen
**Width**: Configurable (default 50 columns)
**Content**: Depends on view (PR sidebar, Issue sidebar, Branch sidebar)

#### 4. Footer

```go
// internal/tui/components/footer/footer.go

type Model struct {
    leftSection   string  // Help/keybindings
    rightSection  string  // Status/running tasks
    ctx           *context.ProgramContext
}

func (m Model) View() string {
    left := m.Ctx.Styles.Footer.LeftStyle.Render(m.leftSection)
    right := m.Ctx.Styles.Footer.RightStyle.Render(m.rightSection)

    // Pad middle with spaces
    gap := m.ctx.ScreenWidth - lipgloss.Width(left) - lipgloss.Width(right)
    return lipgloss.JoinHorizontal(
        lipgloss.Top,
        left,
        strings.Repeat(" ", gap),
        right,
    )
}
```

**Position**: Bottom of screen
**Height**: 1-2 lines
**Content**:
- Left: Help text, keybindings
- Right: Running tasks, status

### Rendering Pipeline

```
Model.View()
    ↓
┌───────────────────────┐
│ Calculate dimensions  │
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│ Render tabs (if not   │
│ RepoView)             │
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│ Render current        │
│ section.View()        │
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│ Render sidebar.View() │
│ (if open)             │
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│ Join section +        │
│ sidebar horizontally  │
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│ Render error (if any) │
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│ Render footer.View()  │
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│ Join all vertically   │
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│ zone.Scan() for mouse │
└──────────┬────────────┘
           ↓
    Terminal Output
```

---

## Component Hierarchy

### Visual Hierarchy

```
Model (Root)
│
├─ Tabs
│  └─ Tab items (PRs, Issues, Repo)
│
├─ Current Section
│  ├─ SearchBar
│  │  └─ InputBox
│  │
│  ├─ Table
│  │  ├─ Header row
│  │  ├─ Data rows
│  │  └─ ListViewPort (scrolling)
│  │
│  └─ PromptConfirmation (modal)
│     └─ InputBox
│
├─ Sidebar (conditional)
│  │
│  ├─ PRSidebar (if PRsView)
│  │  ├─ Activity tab
│  │  ├─ Files tab
│  │  └─ Checks tab
│  │
│  ├─ IssueSidebar (if IssuesView)
│  │  ├─ Activity tab
│  │  └─ Labels/Assignees
│  │
│  └─ BranchSidebar (if RepoView)
│     └─ Branch details
│
└─ Footer
   ├─ Left section (help)
   └─ Right section (status)
```

### Component Communication

```
        Parent (Model)
           │
           │ props (ctx, config)
           ↓
┌──────────┴──────────┐
│                     │
Child A          Child B
(Section)        (Sidebar)
   │                 │
   │ SectionMsg      │ SidebarMsg
   ↓                 ↓
        Parent
    (routes messages)
```

---

## Responsive Design

### Terminal Resize Handling

```go
// internal/tui/ui.go

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Update dimensions
        m.ctx.ScreenWidth = msg.Width
        m.ctx.ScreenHeight = msg.Height

        // Recalculate layout
        m.recalculateDimensions()

        // Update all components
        m.updateComponentDimensions()

        // Force re-render
        return m, nil
    }
}

func (m *Model) updateComponentDimensions() {
    // Update sections
    for i := range m.prs {
        m.prs[i].UpdateProgramContext(m.ctx)
    }

    // Update sidebar
    m.sidebar.UpdateProgramContext(m.ctx)

    // Update footer
    m.footer.UpdateProgramContext(m.ctx)

    // Update tabs
    m.tabs.UpdateProgramContext(m.ctx)
}
```

### Adaptive Sizing

Components adapt to available space:

```go
// internal/tui/components/section/section.go

func (m *BaseModel) GetDimensions() constants.Dimensions {
    return constants.Dimensions{
        Width:  max(0, m.Ctx.MainContentWidth - padding),
        Height: max(0, m.Ctx.MainContentHeight - searchHeight),
    }
}

// Table uses available space
func (m *table.Model) View() string {
    // Fit rows to available height
    visibleRows := m.calculateVisibleRows()

    // Truncate columns to fit width
    m.fitColumnsToWidth()

    return m.render()
}
```

### Sidebar Toggle

```go
// Toggle sidebar visibility
func (m *Model) togglePreview() (tea.Model, tea.Cmd) {
    m.sidebar.isOpen = !m.sidebar.isOpen

    // Recalculate main content width
    if m.sidebar.isOpen {
        m.ctx.MainContentWidth = m.ctx.ScreenWidth - sidebarWidth
    } else {
        m.ctx.MainContentWidth = m.ctx.ScreenWidth
    }

    // Update components
    m.updateComponentDimensions()

    return m, nil
}
```

---

## Navigation Patterns

### Keyboard Navigation

```
┌─────────────────────────────────────────────────┐
│ VERTICAL NAVIGATION (within section)           │
├─────────────────────────────────────────────────┤
│ j / ↓      Move down one row                    │
│ k / ↑      Move up one row                      │
│ g          Jump to first row                    │
│ G          Jump to last row                     │
│ Ctrl+d     Page down                            │
│ Ctrl+u     Page up                              │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│ HORIZONTAL NAVIGATION (between sections)       │
├─────────────────────────────────────────────────┤
│ Tab        Next section                         │
│ Shift+Tab  Previous section                     │
│ 1-9        Jump to section N                    │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│ VIEW NAVIGATION (between views)                │
├─────────────────────────────────────────────────┤
│ Ctrl+p     Switch to PRs view                   │
│ Ctrl+i     Switch to Issues view                │
│ Ctrl+r     Switch to Repo view                  │
└─────────────────────────────────────────────────┘
```

### Focus Management

```go
// Focus states
type FocusState int

const (
    FocusTable FocusState = iota
    FocusSearch
    FocusPrompt
    FocusSidebar
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Route input based on focus
    if m.section.IsSearchFocused() {
        return m.section.SearchBar.Update(msg)
    }

    if m.section.IsPromptConfirmationFocused() {
        return m.section.PromptConfirmationBox.Update(msg)
    }

    // Default: table has focus
    return m.section.Table.Update(msg)
}
```

---

## Modal and Overlay System

### Modal Types

#### 1. Search Modal

```go
// Activated with '/'
func (m *Model) activateSearch() (tea.Model, tea.Cmd) {
    cmd := m.currentSection().SetIsSearching(true)
    return m, cmd
}

// Layout
┌──────────────────────────────────────┐
│ Search: [is:open author:@me_______] │  ← Focus here
└──────────────────────────────────────┘
```

#### 2. Prompt Confirmation

```go
// For actions that need confirmation
func (m *Model) showConfirmation(action string) (tea.Model, tea.Cmd) {
    m.currentSection().SetPromptConfirmationAction(action)
    cmd := m.currentSection().SetIsPromptConfirmationShown(true)
    return m, cmd
}

// Layout
┌──────────────────────────────────────────────┐
│                                              │
│  Are you sure you want to merge this PR?    │
│  (Y/n) [_]                                   │
│                                              │
└──────────────────────────────────────────────┘
```

#### 3. Input Prompt

```go
// For user input (e.g., branch name)
func (m *Model) promptInput(prompt string) (tea.Model, tea.Cmd) {
    m.currentSection().PromptConfirmationBox.SetPrompt(prompt)
    cmd := m.currentSection().SetIsPromptConfirmationShown(true)
    return m, cmd
}

// Layout
┌──────────────────────────────────────────────┐
│                                              │
│  Enter branch name: [feature/________]       │
│                                              │
└──────────────────────────────────────────────┘
```

### Overlay Rendering

```go
// Overlays render on top of main content
func (m *Model) View() string {
    // Render base layout
    base := m.renderBaseLayout()

    // Render modal if active
    if m.currentSection().IsPromptConfirmationFocused() {
        modal := m.currentSection().GetPromptConfirmation()
        return m.overlayModal(base, modal)
    }

    return base
}

func (m *Model) overlayModal(base, modal string) string {
    // Center modal on screen
    return lipgloss.Place(
        m.ctx.ScreenWidth,
        m.ctx.ScreenHeight,
        lipgloss.Center,
        lipgloss.Center,
        modal,
        lipgloss.WithWhitespaceChars("█"),
        lipgloss.WithWhitespaceForeground(lipgloss.Color("236")),
    )
}
```

---

## State-Driven UI

The entire UI is a pure function of the model state:

```
State → View

If state doesn't change, view doesn't change
If state changes, view automatically updates
```

### Example: Loading States

```go
// State
type BaseModel struct {
    IsLoading bool
    Rows      []Row
}

// View adapts to state
func (m *BaseModel) View() string {
    if m.IsLoading {
        return m.renderLoadingState()
    }

    if len(m.Rows) == 0 {
        return m.renderEmptyState()
    }

    return m.renderRows()
}
```

### Example: Sidebar State

```go
// State
type Model struct {
    sidebar sidebar.Model
}

// View includes sidebar only if open
func (m *Model) View() string {
    main := m.currentSection().View()

    if m.sidebar.IsOpen() {
        sidebar := m.sidebar.View()
        return lipgloss.JoinHorizontal(lipgloss.Top, main, sidebar)
    }

    return main
}
```

---

## Summary

The UI structure of gh-dash is:

1. **Hierarchical**: Clear parent-child component relationships
2. **Responsive**: Adapts to terminal size changes
3. **Modal-friendly**: Overlays for user input
4. **Keyboard-first**: All actions accessible via keyboard
5. **State-driven**: UI automatically reflects model state
6. **Composable**: Components combine to create complex layouts

This structure makes it easy to:
- Add new views
- Modify layouts
- Test components
- Maintain consistency
- Handle edge cases (small terminals, etc.)
