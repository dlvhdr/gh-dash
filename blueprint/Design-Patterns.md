# Design Patterns and Architectural Decisions

This document details all the design patterns used in gh-dash, the reasoning behind their use, and how they're implemented.

## Table of Contents
1. [Elm Architecture Pattern](#1-elm-architecture-pattern)
2. [Component Pattern](#2-component-pattern)
3. [Interface Segregation](#3-interface-segregation)
4. [Message Passing Pattern](#4-message-passing-pattern)
5. [Command Pattern](#5-command-pattern)
6. [Observer Pattern (Subscription Model)](#6-observer-pattern-subscription-model)
7. [Strategy Pattern](#7-strategy-pattern)
8. [Factory Pattern](#8-factory-pattern)
9. [Singleton Pattern (Context)](#9-singleton-pattern-context)
10. [Adapter Pattern](#10-adapter-pattern)
11. [Composition Over Inheritance](#11-composition-over-inheritance)
12. [Immutability Pattern](#12-immutability-pattern)

---

## 1. Elm Architecture Pattern

### What It Is
The Elm Architecture (TEA) is a pattern for building interactive applications with a focus on simplicity, modularity, and maintainability. It consists of three core concepts: Model, Update, and View.

### Why We Use It
- **Predictability**: All state changes happen in one place (Update)
- **Testability**: Pure functions are easy to test
- **Debugging**: Time-travel debugging possible (replay messages)
- **Scalability**: Pattern scales from small to large applications
- **Type Safety**: Compiler-enforced correctness

### How It's Implemented

```go
// internal/tui/ui.go

type Model struct {
    keys          *keys.KeyMap
    sidebar       sidebar.Model
    currSectionId int
    prs           []section.Section
    issues        []section.Section
    ctx           *context.ProgramContext
    // ... more state
}

// Init: Initialize the model
func (m Model) Init() tea.Cmd {
    return m.initScreen
}

// Update: Handle messages and update state
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyMsg(msg)
    case tea.WindowSizeMsg:
        return m.handleWindowSizeMsg(msg)
    // ... more message types
    }
    return m, nil
}

// View: Render the current state
func (m *Model) View() string {
    // Pure function: same model → same output
    return m.renderLayout()
}
```

### Key Principles
1. **Single source of truth**: Model contains all state
2. **Unidirectional data flow**: Model → View → User Input → Update → Model
3. **Pure functions**: View and Update have no side effects
4. **Commands for side effects**: Async operations return `tea.Cmd`

---

## 2. Component Pattern

### What It Is
Breaking down the UI into reusable, self-contained components that manage their own state and behavior.

### Why We Use It
- **Modularity**: Each component is independent
- **Reusability**: Components can be used in multiple places
- **Maintainability**: Changes localized to component
- **Testability**: Components tested in isolation

### How It's Implemented

Every UI component follows this structure:

```go
// internal/tui/components/section/section.go

// Component interface
type Section interface {
    Identifier
    Component
    Table
    Search
    PromptConfirmation
    GetConfig() config.SectionConfig
    UpdateProgramContext(ctx *context.ProgramContext)
    MakeSectionCmd(cmd tea.Cmd) tea.Cmd
    GetPagerContent() string
    // ...
}

type Component interface {
    Update(msg tea.Msg) (Section, tea.Cmd)
    View() string
}

// Base implementation
type BaseModel struct {
    Id          int
    Config      config.SectionConfig
    Ctx         *context.ProgramContext
    Table       table.Model
    SearchBar   search.Model
    // ... component state
}

func NewModel(ctx *context.ProgramContext, options NewSectionOptions) BaseModel {
    // Initialize component with default state
}

func (m *BaseModel) Update(msg tea.Msg) (Section, tea.Cmd) {
    // Handle messages specific to this component
}

func (m *BaseModel) View() string {
    // Render component
}
```

### Component Hierarchy

```
Model (Root)
├── Tabs Component
├── Section Component
│   ├── SearchBar Component
│   ├── Table Component
│   └── PromptConfirmation Component
├── Sidebar Component
│   ├── Activity Component
│   ├── Files Component
│   └── Checks Component
└── Footer Component
```

### Component Communication
- **Parent → Child**: Via parameters (props)
- **Child → Parent**: Via messages
- **Sibling ↔ Sibling**: Via parent (message passing up, then down)

---

## 3. Interface Segregation

### What It Is
Breaking down large interfaces into smaller, more specific ones so clients only depend on methods they use.

### Why We Use It
- **Flexibility**: Components implement only what they need
- **Decoupling**: Reduces dependencies
- **Clarity**: Clear contracts

### How It's Implemented

```go
// internal/tui/components/section/section.go

// Instead of one large interface, we have several small ones

type Identifier interface {
    GetId() int
    GetType() string
}

type Component interface {
    Update(msg tea.Msg) (Section, tea.Cmd)
    View() string
}

type Table interface {
    NumRows() int
    GetCurrRow() data.RowData
    CurrRow() int
    NextRow() int
    PrevRow() int
    // ...
}

type Search interface {
    SetIsSearching(val bool) tea.Cmd
    IsSearchFocused() bool
    ResetFilters()
    GetFilters() string
}

type PromptConfirmation interface {
    SetIsPromptConfirmationShown(val bool) tea.Cmd
    IsPromptConfirmationFocused() bool
    SetPromptConfirmationAction(action string)
    GetPromptConfirmationAction() string
}

// Section interface composes all smaller interfaces
type Section interface {
    Identifier
    Component
    Table
    Search
    PromptConfirmation
    // ... additional methods
}
```

### Benefits
- Components can implement subsets of functionality
- Easy to mock for testing
- Clear separation of concerns

---

## 4. Message Passing Pattern

### What It Is
All communication between components happens via messages (events) rather than direct function calls.

### Why We Use It
- **Decoupling**: Components don't know about each other
- **Async-friendly**: Natural fit for async operations
- **Event sourcing**: Can log/replay all events
- **Testability**: Easy to test message handlers

### How It's Implemented

```go
// internal/tui/components/section/section.go

// Message types
type SectionMsg struct {
    Id          int
    Type        string
    InternalMsg tea.Msg
}

type SectionRowsFetchedMsg struct {
    SectionId int
    Issues    []data.RowData
}

// Creating messages
func (m *BaseModel) MakeSectionCmd(cmd tea.Cmd) tea.Cmd {
    if cmd == nil {
        return nil
    }
    return func() tea.Msg {
        internalMsg := cmd()
        return SectionMsg{
            Id:          m.Id,
            Type:        m.Type,
            InternalMsg: internalMsg,
        }
    }
}

// Handling messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case SectionMsg:
        return m.handleSectionMsg(msg)
    case SectionRowsFetchedMsg:
        return m.handleRowsFetched(msg)
    // ...
    }
}
```

### Message Flow

```
User presses 'r'
    ↓
KeyMsg created
    ↓
Model.Update() receives KeyMsg
    ↓
Section.Update() called with RefreshMsg
    ↓
Section creates SectionMsg wrapping RefreshMsg
    ↓
Message sent back to Model
    ↓
Model routes to appropriate section
    ↓
Section executes refresh, returns TaskStartedMsg
    ↓
Async task completes, sends TaskFinishedMsg
    ↓
Model.Update() receives TaskFinishedMsg
    ↓
Data updated, View() re-renders
```

---

## 5. Command Pattern

### What It Is
Encapsulating actions/operations as objects (in Go, as functions) that can be executed, queued, or passed around.

### Why We Use It
- **Async operations**: Commands can run in background
- **Undo/Redo**: Commands can be reversed (not used currently)
- **Queueing**: Multiple commands can be batched
- **Separation**: Command execution separate from invocation

### How It's Implemented

```go
// Commands are functions that return messages
type Cmd func() Msg

// Example: Fetching PR data
func (m *PRsSection) fetchSectionRows() tea.Cmd {
    return func() tea.Msg {
        // Execute async operation
        prs, err := data.FetchPRs(m.GetFilters())
        if err != nil {
            return constants.ErrMsg{Err: err}
        }
        // Return result as message
        return SectionRowsFetchedMsg{
            SectionId: m.Id,
            Issues:    prs,
        }
    }
}

// Batching commands
func batchCmds(cmds ...tea.Cmd) tea.Cmd {
    return tea.Batch(cmds...)
}

// Usage
return m, tea.Batch(
    m.spinner.Tick,
    m.fetchSectionRows(),
    m.updateFooter(),
)
```

### Command Types
1. **Immediate**: Return a message immediately
2. **Async**: Spawn goroutine, send message when done
3. **Batch**: Execute multiple commands
4. **Tick**: Recurring commands (animations)

---

## 6. Observer Pattern (Subscription Model)

### What It Is
Components subscribe to events/changes and get notified when they occur.

### Why We Use It
- **Reactive**: UI updates automatically when data changes
- **Decoupling**: Publishers don't know about subscribers
- **Animation**: For spinners, timers, etc.

### How It's Implemented

```go
// Subscriptions via Bubbletea
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.spinner.Tick,  // Subscribe to tick events
        // ... other subscriptions
    )
}

// Spinner subscription
type spinner.Model struct {
    Spinner Spinner
}

func (m Model) Tick() tea.Cmd {
    return tea.Tick(time.Second/10, func(t time.Time) tea.Msg {
        return TickMsg{Time: t}
    })
}

// Update handles tick messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case TickMsg:
        // Update spinner frame
        m.spinner, cmd = m.spinner.Update(msg)
        return m, cmd
    }
}
```

### Subscription Use Cases
- **Spinner animations**: Tick every 100ms
- **Auto-refresh**: Poll GitHub API every N minutes
- **Timers**: Countdown timers
- **Window resize**: Terminal size changes

---

## 7. Strategy Pattern

### What It Is
Defining a family of algorithms, encapsulating each one, and making them interchangeable.

### Why We Use It
- **Flexibility**: Swap implementations at runtime
- **Open/Closed**: Open for extension, closed for modification
- **Testability**: Easy to test each strategy

### How It's Implemented

#### Example 1: Section Type Strategies

```go
// Different section types implement same interface
type Section interface {
    Update(msg tea.Msg) (Section, tea.Cmd)
    View() string
    // ...
}

// Concrete strategies
type PRsSection struct {
    BaseModel
}

type IssuesSection struct {
    BaseModel
}

type RepoSection struct {
    BaseModel
}

// Each implements Update and View differently
func (m *PRsSection) Update(msg tea.Msg) (Section, tea.Cmd) {
    // PR-specific logic
}

func (m *IssuesSection) Update(msg tea.Msg) (Section, tea.Cmd) {
    // Issue-specific logic
}
```

#### Example 2: Keybinding Strategies

```go
// Different keybinding strategies per view
type KeyMap struct {
    Universal UniversalKeyMap
    Prs       PrsKeyMap
    Issues    IssuesKeyMap
    Branches  BranchesKeyMap
}

// Update dispatches to appropriate strategy
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch m.ctx.View {
    case config.PRsView:
        return m.handlePRsKeys(msg)
    case config.IssuesView:
        return m.handleIssuesKeys(msg)
    case config.RepoView:
        return m.handleRepoKeys(msg)
    }
}
```

---

## 8. Factory Pattern

### What It Is
Creating objects without specifying exact class, using a factory method.

### Why We Use It
- **Abstraction**: Hide creation complexity
- **Flexibility**: Easy to add new types
- **Centralization**: Creation logic in one place

### How It's Implemented

```go
// internal/tui/components/section/section.go

type NewSectionOptions struct {
    Id       int
    Config   config.SectionConfig
    Ctx      *context.ProgramContext
    Type     string
    Columns  []table.Column
    Singular string
    Plural   string
}

// Factory function
func NewModel(
    ctx *context.ProgramContext,
    options NewSectionOptions,
) BaseModel {
    filters := options.GetConfigFiltersWithCurrentRemoteAdded(ctx)
    m := BaseModel{
        Ctx:          ctx,
        Id:           options.Id,
        Type:         options.Type,
        Config:       options.Config,
        Spinner:      spinner.Model{Spinner: spinner.Dot},
        Columns:      options.Columns,
        SingularForm: options.Singular,
        PluralForm:   options.Plural,
        SearchBar:    search.NewModel(ctx, /* ... */),
        // ... initialize all fields
    }
    m.Table = table.NewModel(/* ... */)
    return m
}

// Creating different section types
func NewPRsSection(ctx *context.ProgramContext, id int, config config.PrsSectionConfig) *PRsSection {
    base := NewModel(ctx, NewSectionOptions{
        Id:       id,
        Type:     "pr",
        Singular: "PR",
        Plural:   "PRs",
        // ...
    })
    return &PRsSection{BaseModel: base}
}
```

---

## 9. Singleton Pattern (Context)

### What It Is
Ensuring a class has only one instance and providing global access to it.

### Why We Use It
- **Shared state**: Configuration, theme, screen size
- **Consistency**: All components see same state
- **Memory efficiency**: One instance shared

### How It's Implemented

```go
// internal/tui/context/context.go

type ProgramContext struct {
    Config            *config.Config
    Theme             theme.Theme
    ScreenWidth       int
    ScreenHeight      int
    MainContentWidth  int
    MainContentHeight int
    View              config.ViewType
    User              *data.User
    Error             error
    StartTask         func(Task) tea.Cmd
    // ... more shared state
}

// Created once in NewModel
func NewModel(location config.Location) Model {
    m := Model{}

    // Single context instance
    m.ctx = &context.ProgramContext{
        RepoPath:   location.RepoPath,
        ConfigFlag: location.ConfigFlag,
        Version:    version,
        StartTask:  func(task context.Task) tea.Cmd {
            // Task management logic
        },
    }

    // All components receive same context pointer
    m.footer = footer.NewModel(m.ctx)
    m.prSidebar = prsidebar.NewModel(m.ctx)
    m.issueSidebar = issuesidebar.NewModel(m.ctx)
    // ...

    return m
}

// Components access context
func (m *BaseModel) UpdateProgramContext(ctx *context.ProgramContext) {
    m.Ctx = ctx  // Update reference
}
```

### Benefits
- All components share configuration
- Theme changes propagate automatically
- Screen resize handled centrally

---

## 10. Adapter Pattern

### What It Is
Converting interface of a class into another interface clients expect.

### Why We Use It
- **Integration**: Connect incompatible interfaces
- **Reusability**: Use existing code with new interfaces
- **Abstraction**: Hide third-party dependencies

### How It's Implemented

#### Example 1: GitHub CLI Adapter

```go
// internal/data/prapi.go

// Adapting GitHub GraphQL API to our data models
func FetchPullRequests(
    filters string,
    limit int,
) ([]PullRequestData, error) {
    // Create GitHub client (third-party)
    client, err := gh.DefaultGraphQLClient()
    if err != nil {
        return nil, err
    }

    // Construct query (GraphQL schema)
    var query struct {
        Search struct {
            Nodes []struct {
                PullRequest PullRequestData `graphql:"... on PullRequest"`
            }
            PageInfo PageInfo
        } `graphql:"search(query: $query, type: ISSUE, first: $limit)"`
    }

    // Execute query
    err = client.Query("FetchPRs", &query, variables)

    // Adapt response to our model
    prs := make([]PullRequestData, 0)
    for _, node := range query.Search.Nodes {
        prs = append(prs, node.PullRequest)
    }

    return prs, nil
}
```

#### Example 2: Git Module Adapter

```go
// internal/git/git.go

// Adapting git-module library to our Repo model
func GetRepo(path string) (*Repo, error) {
    // Use third-party git library
    gitRepo, err := gitm.Open(path)
    if err != nil {
        return nil, err
    }

    // Adapt to our Repo model
    repo := &Repo{
        Repository: gitRepo,
        Origin:     getOrigin(gitRepo),
        Branches:   getBranches(gitRepo),
        // ...
    }

    return repo, nil
}
```

---

## 11. Composition Over Inheritance

### What It Is
Building complex objects by composing simpler ones rather than inheriting from base classes.

### Why We Use It
- **Flexibility**: Change behavior at runtime
- **No fragile base class**: Avoid inheritance pitfalls
- **Go idiomatic**: Go doesn't have inheritance
- **Testing**: Easy to mock composed parts

### How It's Implemented

```go
// internal/tui/components/prssection/prssection.go

// Instead of inheritance, use composition
type PRsSection struct {
    section.BaseModel  // Embed base functionality

    // Composed components
    Rows []PullRequestData
}

// Delegates to embedded BaseModel
func (m *PRsSection) GetId() int {
    return m.BaseModel.GetId()  // Delegate
}

// Override when needed
func (m *PRsSection) Update(msg tea.Msg) (section.Section, tea.Cmd) {
    // Custom logic
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if key.Matches(msg, m.Ctx.Keys.Prs.Approve) {
            return m.approve()  // PR-specific action
        }
    }

    // Fallback to base behavior
    return m.BaseModel.Update(msg)
}

// Compose multiple components
type Model struct {
    sidebar    sidebar.Model      // Composed
    tabs       tabs.Model         // Composed
    footer     footer.Model       // Composed
    prSidebar  prsidebar.Model    // Composed
}
```

### Benefits
- **Modularity**: Each component is independent
- **Reusability**: Components used in multiple places
- **Flexibility**: Easy to swap implementations
- **Testability**: Test components in isolation

---

## 12. Immutability Pattern

### What It Is
Creating objects that cannot be modified after creation. Changes create new objects.

### Why We Use It
- **Thread safety**: No race conditions
- **Predictability**: State doesn't change unexpectedly
- **Time travel**: Can keep history of states
- **Debugging**: Easier to reason about

### How It's Implemented

```go
// Update returns new model, doesn't mutate
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Create new model (in Go, often via pointer receiver)
    // but conceptually immutable

    newModel := m  // Copy

    switch msg := msg.(type) {
    case tea.KeyMsg:
        newModel.currSectionId = (m.currSectionId + 1) % len(m.prs)
    }

    return newModel, nil
}

// Bubbletea enforces immutability via method signatures
// Update must return new model
// View must not mutate model

// Context updates create new context
func (m *Model) updateContext() {
    newCtx := *m.ctx  // Copy
    newCtx.ScreenWidth = width
    newCtx.ScreenHeight = height
    m.ctx = &newCtx
}
```

### Benefits
- **Concurrency**: Safe to read from multiple goroutines
- **Undo/Redo**: Keep old states
- **Testing**: Compare states easily
- **Debugging**: State snapshots for debugging

---

## Architectural Decisions and Their Rationale

### 1. Why Bubbletea (Elm Architecture)?

**Decision**: Use Charmbracelet Bubbletea framework

**Rationale**:
- **Proven pattern**: Elm Architecture battle-tested
- **Go idiomatic**: Fits Go's philosophy (simplicity, composition)
- **Rich ecosystem**: Charmbracelet provides full suite (Lipgloss, Bubbles, Glamour)
- **Active development**: Well-maintained, growing community
- **Testability**: Pure functions, message-driven
- **Developer experience**: Clear patterns, easy to learn

**Alternatives considered**:
- **tview**: More imperative, harder to test
- **termui**: Less maintained, event-driven complexity
- **Raw termbox**: Too low-level, reinvent too much

### 2. Why GraphQL Instead of REST?

**Decision**: Use GitHub GraphQL API

**Rationale**:
- **Efficiency**: Fetch exactly what we need (no over/under-fetching)
- **Single request**: Get PRs + reviews + comments + checks in one query
- **Type safety**: Strong types via code generation
- **Future-proof**: GitHub investing in GraphQL
- **Pagination**: Built-in pagination support

**Trade-offs**:
- **Complexity**: GraphQL queries more complex than REST
- **Learning curve**: Team must learn GraphQL
- **Tooling**: Need GraphQL-specific tools

### 3. Why Configuration via YAML?

**Decision**: Use YAML configuration files

**Rationale**:
- **Human-readable**: Easy to edit
- **Comments**: Users can document their config
- **Industry standard**: Familiar to developers
- **Validation**: Strong validation via go-playground/validator
- **Hierarchy**: XDG standard, repo-local overrides

**Alternatives considered**:
- **TOML**: Less nested, but less familiar
- **JSON**: No comments, less human-friendly
- **Env vars**: Not suitable for complex config

### 4. Why Component-Based Architecture?

**Decision**: Break UI into reusable components

**Rationale**:
- **Maintainability**: Change one component without affecting others
- **Reusability**: Table, Search, Footer used in multiple places
- **Testing**: Test components in isolation
- **Parallel development**: Team can work on different components
- **Clear boundaries**: Each component has defined responsibility

**Trade-offs**:
- **Boilerplate**: More files, more interfaces
- **Learning curve**: Understand component lifecycle

### 5. Why Context for Shared State?

**Decision**: Use ProgramContext singleton

**Rationale**:
- **Centralized**: One place for shared state
- **Type safe**: Compile-time checks
- **Performance**: No prop drilling through 10 layers
- **Consistency**: All components see same theme, config

**Trade-offs**:
- **Global state**: Can be abused
- **Testing**: Need to mock context

---

## Anti-Patterns Avoided

### 1. God Object
**Avoided by**: Splitting Model into components, each with specific responsibility

### 2. Callback Hell
**Avoided by**: Message passing instead of callbacks

### 3. Tight Coupling
**Avoided by**: Interfaces, dependency injection, message passing

### 4. Premature Optimization
**Avoided by**: Profile first, optimize after

### 5. Magic Numbers
**Avoided by**: Constants in constants/ directory

### 6. Shared Mutable State
**Avoided by**: Immutability, message passing for communication

---

## Summary

gh-dash leverages these design patterns to create a:

- **Maintainable** codebase (clear patterns, separation of concerns)
- **Testable** application (pure functions, interfaces, message-driven)
- **Extensible** architecture (easy to add sections, components, features)
- **Performant** TUI (async operations, efficient rendering)
- **Reliable** tool (type safety, error handling, validation)

The combination of **Elm Architecture**, **Component Pattern**, and **Message Passing** creates a robust foundation that scales from simple displays to complex interactive UIs.

Every pattern choice has clear rationale, balancing:
- **Complexity** vs **Simplicity**
- **Flexibility** vs **Structure**
- **Performance** vs **Maintainability**

These patterns are not dogmatic - they're practical tools that emerged from real-world needs and continue to evolve as the application grows.
