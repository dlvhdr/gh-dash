# Architecture Overview

## Project Type
**gh-dash** is a Terminal User Interface (TUI) application written in Go that provides a rich interface for managing GitHub pull requests, issues, and repository branches directly from the terminal.

## Technology Stack
- **Language**: Go 1.24.7+
- **Module**: `github.com/dlvhdr/gh-dash/v4`
- **TUI Framework**: Charmbracelet Bubbletea (Elm Architecture)
- **CLI Framework**: Cobra
- **API**: GitHub GraphQL API via cli/go-gh

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     CLI Entry Point                         │
│                   (cmd/root.go)                             │
│  - Flag parsing                                             │
│  - Configuration loading                                    │
│  - Bubbletea program initialization                         │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                    TUI Layer                                │
│                 (internal/tui/)                             │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Model      │  │   Update()   │  │    View()    │     │
│  │ (App State)  │◄─┤  (Logic)     │──┤  (Rendering) │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                             │
│  Components:                                                │
│  • Sections (PRs, Issues, Repo)                            │
│  • Sidebars (Details, Activity, Files)                     │
│  • Tables, Tabs, Search, Footer                            │
└────────────────────┬───────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                  Business Logic Layer                       │
│                                                             │
│  ┌──────────────────┐  ┌──────────────────┐               │
│  │  Data Module     │  │   Git Module     │               │
│  │ (GraphQL API)    │  │ (Local Git Ops)  │               │
│  │ - PRs/Issues     │  │ - Branches       │               │
│  │ - Comments       │  │ - Status         │               │
│  │ - Reviews        │  │ - Remotes        │               │
│  └──────────────────┘  └──────────────────┘               │
│                                                             │
│  ┌──────────────────┐  ┌──────────────────┐               │
│  │  Config Module   │  │  Utils Module    │               │
│  │ - YAML parsing   │  │ - Helpers        │               │
│  │ - Validation     │  │ - Formatters     │               │
│  │ - Keybindings    │  │ - Converters     │               │
│  └──────────────────┘  └──────────────────┘               │
└─────────────────────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                  External Services                          │
│                                                             │
│  • GitHub GraphQL API (via go-gh)                          │
│  • Local Git Repository                                     │
│  • Terminal (Bubbletea rendering engine)                   │
│  • System Clipboard                                         │
│  • System Browser                                           │
└─────────────────────────────────────────────────────────────┘
```

## Application Flow

### 1. Initialization Sequence
```
main()
  └─► cmd.Execute() [Cobra CLI]
       └─► rootCmd.Run()
            ├─► Parse flags (--config, --debug, etc.)
            ├─► Detect git repository context
            ├─► Detect terminal color mode (light/dark)
            ├─► Initialize KeyMap from config
            ├─► Create Bubbletea program
            │    └─► NewModel()
            │         ├─► Initialize empty state
            │         ├─► Create ProgramContext
            │         ├─► Setup task management
            │         └─► Initialize components
            └─► p.Run() [Start event loop]
                 ├─► Model.Init() → initScreen()
                 │    ├─► ParseConfig()
                 │    ├─► FetchUserInfo()
                 │    ├─► Initialize sections
                 │    └─► Setup sidebars
                 ├─► Event Loop
                 │    ├─► Model.Update(msg)
                 │    └─► Model.View()
                 └─► Shutdown
```

### 2. Runtime Architecture: Elm Pattern

The application follows the **Elm Architecture**, a functional reactive programming pattern:

```
┌──────────────────────────────────────────────────────────┐
│                        MODEL                             │
│  (Immutable State - Single Source of Truth)             │
│                                                          │
│  • Current section ID                                    │
│  • PR sections []                                        │
│  • Issue sections []                                     │
│  • Repository section                                    │
│  • Sidebars (PR, Issue, Branch)                         │
│  • Footer, Tabs, Search                                  │
│  • Program Context (shared state)                        │
│  • Running tasks map                                     │
└──────────────────┬───────────────────────────────────────┘
                   │
                   ▼
         ┌─────────────────┐
         │   User Input    │
         │  (Keyboard/     │
         │   Mouse/        │
         │   Resize)       │
         └────────┬────────┘
                  │
                  ▼
┌──────────────────────────────────────────────────────────┐
│                     UPDATE                               │
│  (Pure Function: (Model, Msg) → (Model, Cmd))           │
│                                                          │
│  • Pattern match on message type                        │
│  • Update model immutably                                │
│  • Return new model + side effects (Cmd)                │
│                                                          │
│  Message Types:                                          │
│  - KeyMsg (keyboard input)                              │
│  - WindowSizeMsg (terminal resize)                      │
│  - SectionMsg (section-specific events)                 │
│  - TaskFinishedMsg (async operation completed)          │
│  - SectionRowsFetchedMsg (data loaded)                  │
│  - etc.                                                  │
└──────────────────┬───────────────────────────────────────┘
                   │
                   ▼
         ┌─────────────────┐
         │    COMMANDS     │
         │  (Side Effects) │
         │                 │
         │  • API calls    │
         │  • Git ops      │
         │  • Async tasks  │
         │  • Animations   │
         └────────┬────────┘
                  │
                  │ (Sends new Msg)
                  ▼
┌──────────────────────────────────────────────────────────┐
│                      VIEW                                │
│  (Pure Function: Model → String)                        │
│                                                          │
│  • Render current state to terminal output              │
│  • No side effects, deterministic                       │
│  • Uses Lipgloss for styling                            │
│  • Composed from component views                        │
│                                                          │
│  Layout:                                                 │
│  ┌────────────────────────────────┐                     │
│  │ Tabs (PRs | Issues | Repo)     │                     │
│  ├─────────────────┬──────────────┤                     │
│  │                 │              │                     │
│  │  Section        │   Sidebar    │                     │
│  │  (Table/List)   │   (Details)  │                     │
│  │                 │              │                     │
│  ├─────────────────┴──────────────┤                     │
│  │ Footer (Help/Status)           │                     │
│  └────────────────────────────────┘                     │
└──────────────────────────────────────────────────────────┘
                   │
                   │ (Loops back)
                   ▼
                 MODEL
```

## Module Breakdown

### 1. cmd/ - CLI Commands
**Purpose**: Command-line interface entry points

**Files**:
- `root.go`: Root Cobra command, flag definitions, initialization
- `sponsors.go`: Sponsors command

**Responsibilities**:
- Parse command-line flags
- Initialize logging/profiling
- Load configuration
- Create and run TUI program
- Handle early exit scenarios

### 2. internal/config/ - Configuration Management
**Purpose**: Configuration parsing, validation, and management

**Files**:
- `parser.go`: Main configuration parser using Koanf
- `feature_flags.go`: Feature flag management
- `utils.go`: Configuration utilities

**Responsibilities**:
- Load config from YAML files (with fallback hierarchy)
- Validate configuration schema
- Provide default configuration
- Merge global and local configs
- Handle keybinding customization

### 3. internal/data/ - Data Access Layer
**Purpose**: GitHub API interaction via GraphQL

**Files**:
- `prapi.go`: Pull request queries
- `issueapi.go`: Issue queries
- `commonapi.go`: Common queries (user, version)
- `*.go`: Data models (PullRequestData, IssueData, etc.)

**Responsibilities**:
- GraphQL query construction
- API communication
- Data model definition
- Pagination handling
- Check runs/status fetching

### 4. internal/git/ - Git Integration
**Purpose**: Local git repository operations

**Files**:
- `git.go`: Git operations wrapper

**Responsibilities**:
- Get current repository info
- List branches with metadata
- Track uncommitted changes
- Calculate commits ahead/behind
- Manage remotes

### 5. internal/tui/ - Terminal User Interface
**Purpose**: The entire UI rendering and interaction logic

**Structure**:
```
internal/tui/
├── ui.go              # Main Model, Update, View
├── components/        # Reusable UI components
│   ├── section/       # Base section (abstract)
│   ├── prssection/    # PR section implementation
│   ├── issuessection/ # Issue section implementation
│   ├── reposection/   # Repository/branches section
│   ├── prsidebar/     # PR details sidebar
│   ├── issuesidebar/  # Issue details sidebar
│   ├── branchsidebar/ # Branch sidebar
│   ├── table/         # Table component
│   ├── tabs/          # Tab navigation
│   ├── footer/        # Footer/status bar
│   ├── search/        # Search box
│   └── ...            # Other components
├── context/           # Shared program context
├── keys/              # Keyboard bindings
├── theme/             # Theming and colors
├── constants/         # Constants, icons
└── markdown/          # Markdown rendering
```

**Responsibilities**:
- Event handling (keyboard, mouse, resize)
- State management
- Component orchestration
- Rendering pipeline
- Task management

### 6. internal/utils/ - Utility Functions
**Purpose**: Shared helper functions

**Responsibilities**:
- String formatting
- Time utilities
- Type conversions
- Pointer helpers

## State Management

### ProgramContext
A centralized context object passed to all components containing:

```go
type ProgramContext struct {
    // Configuration
    Config         *config.Config
    ConfigFlag     string
    RepoPath       string

    // Screen dimensions
    ScreenWidth    int
    ScreenHeight   int
    MainContentWidth  int
    MainContentHeight int

    // Theme and styles
    Theme          theme.Theme
    Styles         *context.Styles

    // Current view state
    View           config.ViewType  // PRs, Issues, or Repo
    User           *data.User
    Error          error

    // Task management
    StartTask      func(Task) tea.Cmd

    // Metadata
    Version        string
}
```

This context is:
- **Immutable reference**: Components receive a pointer but shouldn't mutate it
- **Centralized**: Single source of truth for shared state
- **Passed down**: Parent components pass it to children
- **Updated centrally**: Only the main Model updates it

### Task System

Asynchronous operations (API calls, git operations) are managed via a task system:

```
User Action → StartTask() → Spinner visible → Background goroutine
                                                       ↓
                                              Task executes
                                                       ↓
                                        Sends TaskFinishedMsg
                                                       ↓
                              Update() receives msg → Update model → Spinner cleared
```

**Benefits**:
- Non-blocking UI
- Visual feedback
- Error handling
- Completion tracking

## Data Flow

### Fetching PRs/Issues

```
1. User presses 'r' (refresh)
   ↓
2. Update() creates SectionMsg
   ↓
3. Section.Update() receives msg
   ↓
4. Section calls ctx.StartTask()
   ↓
5. Background goroutine:
   - Constructs GraphQL query
   - Calls GitHub API
   - Parses response
   - Sends SectionRowsFetchedMsg
   ↓
6. Section.Update() receives data msg
   ↓
7. Section updates rows, table re-renders
   ↓
8. View() displays updated data
```

### User Interaction Flow

```
Keyboard Input
   ↓
Update() receives tea.KeyMsg
   ↓
Match on key binding
   ↓
┌─────────────┬─────────────┬─────────────┐
│ Navigation  │  Action     │   Modal     │
├─────────────┼─────────────┼─────────────┤
│ Up/Down     │ 'a' Approve │ '/' Search  │
│ NextSection │ 'm' Merge   │ Prompt      │
│ PageUp/Down │ 'c' Comment │ Confirm     │
└─────────────┴─────────────┴─────────────┘
   ↓              ↓              ↓
Update           Start         Set modal
cursor           async task    visible
   ↓              ↓              ↓
View()          Spinner        View() shows
re-renders      animates       modal overlay
```

## Component Lifecycle

Every component follows the Bubbletea lifecycle:

```
1. NewModel(...) → Create initial state
   ↓
2. Init() → Return initial command (tea.Cmd)
   ↓
3. Update(msg) → Handle messages, return (model, cmd)
   ↓  ↑
   │  │ (Event loop)
   │  │
   └──┘
   ↓
4. View() → Render to string (called after each Update)
```

## Rendering Pipeline

```
Model.View()
   ↓
┌──────────────────────┐
│ Render Tabs          │ (if not RepoView)
└──────────┬───────────┘
           ▼
┌──────────────────────┐
│ Render Section       │ (current section's View())
│  + Sidebar           │ (joined horizontally)
└──────────┬───────────┘
           ▼
┌──────────────────────┐
│ Render Error         │ (if present)
└──────────┬───────────┘
           ▼
┌──────────────────────┐
│ Render Footer        │
└──────────┬───────────┘
           ▼
┌──────────────────────┐
│ Zone.Scan()          │ (for mouse support)
└──────────┬───────────┘
           ▼
    Terminal Output
```

## Concurrency Model

- **Main thread**: Bubbletea event loop (single-threaded)
- **Goroutines**: Spawned for async tasks (API, git)
- **Communication**: Via `tea.Cmd` and message passing
- **Thread safety**: Messages sent to main loop via channels

**No shared mutable state** between goroutines - all communication via messages.

## Error Handling

```
Error occurs → Error returned → Wrapped in msg → Update() receives
                                                        ↓
                                                 Store in ctx.Error
                                                        ↓
                                                  View() renders error
                                                        ↓
                                                 User sees error message
```

Errors are:
- Stored in ProgramContext
- Displayed in UI
- Cleared on next successful action
- Logged to debug.log (if enabled)

## Performance Considerations

1. **Lazy loading**: Only fetch visible data
2. **Pagination**: Load more as needed
3. **Caching**: Section data cached until refresh
4. **Efficient rendering**: Lipgloss optimizes terminal output
5. **Async operations**: Non-blocking UI
6. **Virtual scrolling**: Table component handles large datasets

## Extension Points

The architecture makes it easy to:

1. **Add new sections**: Implement `Section` interface
2. **Add new views**: Add to `ViewType` enum, create sections
3. **Add keybindings**: Update keys/ and config
4. **Add API data**: Extend GraphQL queries in data/
5. **Add components**: Create in components/, follow Elm pattern
6. **Add themes**: Extend theme configuration

## Security Considerations

- Uses official GitHub CLI authentication (`go-gh`)
- No credentials stored by gh-dash
- API tokens managed by `gh` CLI
- TLS for all API communication
- Input validation via `go-playground/validator`

## Deployment

- **Single binary**: Cross-compiled for multiple platforms
- **No dependencies**: Static binary (CGO_ENABLED=0)
- **Goreleaser**: Automated multi-platform builds
- **GitHub Releases**: Versioned releases with checksums

## Summary

gh-dash is architected as a **layered, component-based TUI application** following the **Elm Architecture** pattern. It separates concerns cleanly between:

- **Presentation** (TUI components)
- **Business logic** (Data, Git modules)
- **Configuration** (YAML-based, validated)
- **External services** (GitHub API, Git)

This architecture provides:
- **Maintainability**: Clear separation of concerns
- **Testability**: Pure functions, mockable interfaces
- **Extensibility**: Easy to add features
- **Performance**: Async operations, efficient rendering
- **User experience**: Responsive, non-blocking UI
