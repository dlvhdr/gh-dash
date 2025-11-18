# UML Diagrams

This document contains UML diagrams showing the structure, relationships, and interactions in gh-dash using Mermaid syntax.

## Table of Contents
1. [Class Diagrams](#class-diagrams)
2. [Sequence Diagrams](#sequence-diagrams)
3. [State Diagrams](#state-diagrams)
4. [Component Diagrams](#component-diagrams)
5. [Activity Diagrams](#activity-diagrams)

---

## Class Diagrams

### 1. Core Model Structure

```mermaid
classDiagram
    class Model {
        -keys KeyMap
        -sidebar sidebar.Model
        -prSidebar prsidebar.Model
        -issueSidebar issuesidebar.Model
        -branchSidebar branchsidebar.Model
        -currSectionId int
        -footer footer.Model
        -repo Section
        -prs []Section
        -issues []Section
        -tabs tabs.Model
        -ctx *ProgramContext
        -taskSpinner spinner.Model
        -tasks map[string]Task
        +Init() tea.Cmd
        +Update(tea.Msg) (tea.Model, tea.Cmd)
        +View() string
    }

    class ProgramContext {
        +Config *config.Config
        +Theme theme.Theme
        +Styles *Styles
        +ScreenWidth int
        +ScreenHeight int
        +MainContentWidth int
        +MainContentHeight int
        +View ViewType
        +User *data.User
        +Error error
        +StartTask func(Task) tea.Cmd
    }

    class Section {
        <<interface>>
        +GetId() int
        +GetType() string
        +Update(tea.Msg) (Section, tea.Cmd)
        +View() string
        +NumRows() int
        +CurrRow() int
        +BuildRows() []Row
    }

    class BaseModel {
        +Id int
        +Config SectionConfig
        +Ctx *ProgramContext
        +Table table.Model
        +SearchBar search.Model
        +IsSearching bool
        +TotalCount int
        +Update(tea.Msg) (Section, tea.Cmd)
        +View() string
    }

    class PRsSection {
        +BaseModel
        +Rows []PullRequestData
        +BuildRows() []Row
        +approve() tea.Cmd
        +merge() tea.Cmd
    }

    class IssuesSection {
        +BaseModel
        +Rows []IssueData
        +BuildRows() []Row
        +close() tea.Cmd
        +reopen() tea.Cmd
    }

    class RepoSection {
        +BaseModel
        +Rows []Branch
        +BuildRows() []Row
        +checkout() tea.Cmd
        +deleteBranch() tea.Cmd
    }

    Model --> ProgramContext : contains
    Model --> Section : manages multiple
    Section <|.. BaseModel : implements
    BaseModel <|-- PRsSection : extends
    BaseModel <|-- IssuesSection : extends
    BaseModel <|-- RepoSection : extends
```

### 2. Configuration System

```mermaid
classDiagram
    class Config {
        +PRSections []PrsSectionConfig
        +IssuesSections []IssuesSectionConfig
        +Repo RepoConfig
        +Defaults Defaults
        +Keybindings Keybindings
        +Theme *ThemeConfig
        +ConfirmQuit bool
        +ShowAuthorIcons bool
    }

    class PrsSectionConfig {
        +Title string
        +Filters string
        +Limit *int
        +Layout PrsLayoutConfig
    }

    class Defaults {
        +Preview PreviewConfig
        +PrsLimit int
        +IssuesLimit int
        +View ViewType
        +Layout LayoutConfig
        +RefetchIntervalMinutes int
    }

    class ThemeConfig {
        +Ui UIThemeConfig
        +Colors *ColorThemeConfig
        +Icons *IconThemeConfig
    }

    class Keybindings {
        +Universal []Keybinding
        +Issues []Keybinding
        +Prs []Keybinding
        +Branches []Keybinding
    }

    class ConfigParser {
        -k *koanf.Koanf
        +ParseConfig(Location) (Config, error)
        +getDefaultConfig() Config
        +mergeConfigs(string, string) (Config, error)
    }

    Config --> PrsSectionConfig : contains
    Config --> Defaults : contains
    Config --> ThemeConfig : contains
    Config --> Keybindings : contains
    ConfigParser ..> Config : creates
```

### 3. Data Layer

```mermaid
classDiagram
    class PullRequestData {
        +Number int
        +Title string
        +Body string
        +Author User
        +State string
        +ReviewDecision string
        +Assignees Assignees
        +Comments Comments
        +Reviews Reviews
        +Files ChangedFiles
        +IsDraft bool
    }

    class IssueData {
        +Number int
        +Title string
        +Body string
        +Author User
        +State string
        +Assignees Assignees
        +Comments Comments
        +Labels Labels
    }

    class Repository {
        +Name string
        +Owner string
        +Url string
    }

    class User {
        +Login string
        +Name string
    }

    class Repo {
        +Origin string
        +Remotes []string
        +Branches []Branch
        +HeadBranchName string
        +Status NameStatus
    }

    class Branch {
        +Name string
        +LastUpdatedAt *time.Time
        +CommitsAhead int
        +CommitsBehind int
        +IsCheckedOut bool
    }

    PullRequestData --> Repository : belongs to
    PullRequestData --> User : has author
    IssueData --> Repository : belongs to
    IssueData --> User : has author
    Repo --> Branch : contains many
```

### 4. Component Hierarchy

```mermaid
classDiagram
    class Component {
        <<interface>>
        +Update(tea.Msg) (Component, tea.Cmd)
        +View() string
    }

    class Table {
        +Columns []Column
        +Rows []Row
        +currItem int
        +viewport ListViewPort
        +NextItem() int
        +PrevItem() int
        +Update(tea.Msg) tea.Cmd
        +View() string
    }

    class SearchBar {
        +textInput textinput.Model
        +prefix string
        +Focus()
        +Blur()
        +Value() string
        +Update(tea.Msg) tea.Cmd
        +View() string
    }

    class Sidebar {
        +isOpen bool
        +content string
        +Toggle()
        +SetContent(string)
        +View() string
    }

    class Footer {
        +leftSection string
        +rightSection string
        +SetLeftSection(string)
        +SetRightSection(string)
        +View() string
    }

    class Tabs {
        +tabs []string
        +activeTab int
        +NextTab()
        +PrevTab()
        +View() string
    }

    Component <|.. Table : implements
    Component <|.. SearchBar : implements
    Component <|.. Sidebar : implements
    Component <|.. Footer : implements
    Component <|.. Tabs : implements
```

---

## Sequence Diagrams

### 1. Application Startup

```mermaid
sequenceDiagram
    participant User
    participant Main
    participant Cobra
    participant Config
    participant Bubbletea
    participant Model

    User->>Main: run gh-dash
    Main->>Cobra: Execute()
    Cobra->>Cobra: Parse flags
    Cobra->>Config: ParseConfig()
    Config-->>Cobra: config
    Cobra->>Bubbletea: NewProgram(model)
    Bubbletea->>Model: Init()
    Model->>Model: initScreen()
    Model->>Config: Load sections
    Model->>Data: FetchUserInfo()
    Data-->>Model: user info
    Model-->>Bubbletea: initial command
    Bubbletea->>Bubbletea: Start event loop
    Bubbletea-->>User: Display TUI
```

### 2. User Presses Key

```mermaid
sequenceDiagram
    participant User
    participant Bubbletea
    participant Model
    participant Section
    participant Data

    User->>Bubbletea: Press 'r' (refresh)
    Bubbletea->>Model: Update(KeyMsg)
    Model->>Model: handleKeyMsg()
    Model->>Section: Update(RefreshMsg)
    Section->>Section: StartTask()
    Section->>Data: FetchPRs()
    Data-->>Section: TaskFinishedMsg
    Section->>Section: Update(TaskFinishedMsg)
    Section->>Section: Update rows
    Section-->>Model: (newSection, cmd)
    Model-->>Bubbletea: (newModel, cmd)
    Bubbletea->>Model: View()
    Model->>Section: View()
    Section-->>Model: rendered content
    Model-->>Bubbletea: full view
    Bubbletea-->>User: Display updated UI
```

### 3. Data Fetching Flow

```mermaid
sequenceDiagram
    participant Section
    participant Context
    participant Task
    participant GitHub
    participant Update

    Section->>Context: StartTask(fetchTask)
    Context->>Task: Execute in goroutine
    Task->>GitHub: GraphQL query
    GitHub-->>Task: PR data
    Task->>Task: Process data
    Task->>Update: Send DataFetchedMsg
    Update->>Section: Update(DataFetchedMsg)
    Section->>Section: Update rows
    Section->>Section: Clear spinner
    Section-->>Update: new model
```

### 4. Window Resize

```mermaid
sequenceDiagram
    participant Terminal
    participant Bubbletea
    participant Model
    participant Section
    participant Sidebar
    participant Footer

    Terminal->>Bubbletea: Terminal resize
    Bubbletea->>Model: Update(WindowSizeMsg)
    Model->>Model: Update dimensions
    Model->>Model: Update ProgramContext
    Model->>Section: UpdateProgramContext()
    Section->>Section: Recalculate dimensions
    Model->>Sidebar: UpdateProgramContext()
    Sidebar->>Sidebar: Recalculate dimensions
    Model->>Footer: UpdateProgramContext()
    Footer->>Footer: Recalculate dimensions
    Model-->>Bubbletea: (newModel, nil)
    Bubbletea->>Model: View()
    Model-->>Bubbletea: Re-rendered with new dimensions
```

### 5. PR Approval Flow

```mermaid
sequenceDiagram
    participant User
    participant Model
    participant Section
    participant GitHub
    participant Notification

    User->>Model: Press 'a' (approve)
    Model->>Section: Update(KeyMsg)
    Section->>Section: Show confirmation prompt
    Section-->>User: "Approve this PR? (Y/n)"
    User->>Section: Press 'Y'
    Section->>GitHub: Mutation: ApprovePR
    GitHub-->>Section: Success
    Section->>Notification: Show notification
    Notification-->>User: "PR approved!"
    Section->>Section: Refresh PR data
    Section-->>Model: Updated section
```

---

## State Diagrams

### 1. Section State Machine

```mermaid
stateDiagram-v2
    [*] --> Loading
    Loading --> Empty : No data
    Loading --> Populated : Has data
    Loading --> Error : Fetch failed

    Populated --> Searching : Press '/'
    Searching --> Populated : Submit/Cancel

    Populated --> PromptShown : Action requires confirmation
    PromptShown --> Populated : Confirm/Cancel

    Populated --> Loading : Refresh
    Empty --> Loading : Refresh
    Error --> Loading : Retry

    Populated --> [*]
    Empty --> [*]
    Error --> [*]
```

### 2. Sidebar State Machine

```mermaid
stateDiagram-v2
    [*] --> Closed
    Closed --> Opening : Toggle sidebar
    Opening --> Open : Transition complete
    Open --> Closing : Toggle sidebar
    Closing --> Closed : Transition complete

    Open --> ChangingTab : Select tab
    ChangingTab --> Open : Tab changed

    Open --> LoadingContent : Fetch details
    LoadingContent --> Open : Content loaded

    Closed --> [*]
    Open --> [*]
```

### 3. Task Execution State

```mermaid
stateDiagram-v2
    [*] --> Idle
    Idle --> Running : StartTask()
    Running --> Success : Task completes
    Running --> Failed : Task errors

    Success --> Idle : Result processed
    Failed --> Idle : Error handled

    Running --> Cancelled : User cancels
    Cancelled --> Idle : Cleanup

    Idle --> [*]
```

---

## Component Diagrams

### 1. High-Level Architecture

```mermaid
graph TB
    subgraph "CLI Layer"
        CMD[Cobra Commands]
    end

    subgraph "TUI Layer"
        MODEL[Model]
        SECTIONS[Sections]
        SIDEBARS[Sidebars]
        COMPONENTS[Components]
    end

    subgraph "Business Logic"
        CONFIG[Config]
        DATA[Data/API]
        GIT[Git]
    end

    subgraph "External"
        GITHUB[GitHub API]
        GITREPO[Git Repository]
        TERMINAL[Terminal]
    end

    CMD --> MODEL
    MODEL --> SECTIONS
    MODEL --> SIDEBARS
    SECTIONS --> COMPONENTS
    MODEL --> CONFIG
    SECTIONS --> DATA
    SECTIONS --> GIT
    DATA --> GITHUB
    GIT --> GITREPO
    MODEL --> TERMINAL
```

### 2. Data Flow Architecture

```mermaid
graph LR
    subgraph "Input"
        KEYBOARD[Keyboard Input]
        MOUSE[Mouse Input]
        RESIZE[Window Resize]
    end

    subgraph "Processing"
        UPDATE[Update Function]
        COMMANDS[Commands]
    end

    subgraph "State"
        MODEL[Model State]
    end

    subgraph "Rendering"
        VIEW[View Function]
        LIPGLOSS[Lipgloss]
    end

    subgraph "Output"
        TERMINAL[Terminal Display]
    end

    KEYBOARD --> UPDATE
    MOUSE --> UPDATE
    RESIZE --> UPDATE
    UPDATE --> MODEL
    UPDATE --> COMMANDS
    COMMANDS --> UPDATE
    MODEL --> VIEW
    VIEW --> LIPGLOSS
    LIPGLOSS --> TERMINAL
```

---

## Activity Diagrams

### 1. PR Merge Flow

```mermaid
flowchart TD
    START([User presses 'm'])
    --> CHECK_SELECTED{PR selected?}

    CHECK_SELECTED -->|No| ERROR_NO_PR[Show error]
    CHECK_SELECTED -->|Yes| CHECK_MERGEABLE{PR mergeable?}

    CHECK_MERGEABLE -->|No| ERROR_NOT_MERGEABLE[Show error]
    CHECK_MERGEABLE -->|Yes| SHOW_PROMPT[Show confirmation]

    SHOW_PROMPT --> USER_INPUT{User confirms?}

    USER_INPUT -->|No| CANCEL[Cancel]
    USER_INPUT -->|Yes| START_TASK[Start merge task]

    START_TASK --> SHOW_SPINNER[Show spinner]
    SHOW_SPINNER --> CALL_API[Call GitHub API]

    CALL_API --> API_RESPONSE{Success?}

    API_RESPONSE -->|No| SHOW_ERROR[Show error message]
    API_RESPONSE -->|Yes| REFRESH_DATA[Refresh PR data]

    REFRESH_DATA --> NOTIFY[Show notification]
    NOTIFY --> END([Done])

    ERROR_NO_PR --> END
    ERROR_NOT_MERGEABLE --> END
    CANCEL --> END
    SHOW_ERROR --> END
```

### 2. Configuration Loading

```mermaid
flowchart TD
    START([Application starts])
    --> PARSE_FLAGS[Parse command-line flags]

    PARSE_FLAGS --> CHECK_FLAG{--config flag?}

    CHECK_FLAG -->|Yes| LOAD_PROVIDED[Load provided config]
    CHECK_FLAG -->|No| CHECK_REPO{In git repo?}

    CHECK_REPO -->|Yes| CHECK_LOCAL{.gh-dash.yml exists?}
    CHECK_REPO -->|No| LOAD_GLOBAL[Load global config]

    CHECK_LOCAL -->|Yes| MERGE_CONFIGS[Merge global + local]
    CHECK_LOCAL -->|No| LOAD_GLOBAL

    LOAD_PROVIDED --> VALIDATE
    MERGE_CONFIGS --> VALIDATE
    LOAD_GLOBAL --> VALIDATE[Validate config]

    VALIDATE --> VALID{Valid?}

    VALID -->|Yes| APPLY[Apply config]
    VALID -->|No| ERROR[Show error + example]

    APPLY --> END([Config loaded])
    ERROR --> EXIT([Exit])
```

### 3. Search/Filter Flow

```mermaid
flowchart TD
    START([User presses '/'])
    --> SHOW_SEARCH[Show search bar]

    SHOW_SEARCH --> FOCUS[Focus input]
    FOCUS --> WAIT_INPUT[Wait for user input]

    WAIT_INPUT --> USER_ACTION{User action?}

    USER_ACTION -->|Types| UPDATE_VALUE[Update search value]
    USER_ACTION -->|Enter| SUBMIT_SEARCH
    USER_ACTION -->|Esc| CANCEL

    UPDATE_VALUE --> WAIT_INPUT

    SUBMIT_SEARCH --> BLUR_INPUT[Blur input]
    BLUR_INPUT --> START_FETCH[Start fetch task]

    START_FETCH --> FETCH_DATA[Fetch filtered data]
    FETCH_DATA --> UPDATE_ROWS[Update table rows]
    UPDATE_ROWS --> END([Search complete])

    CANCEL --> RESET[Reset to original value]
    RESET --> END
```

---

## Summary

These diagrams illustrate:

1. **Class Diagrams**: Structure and relationships between classes
2. **Sequence Diagrams**: Interaction flows over time
3. **State Diagrams**: State transitions of components
4. **Component Diagrams**: High-level architecture
5. **Activity Diagrams**: Workflow processes

Key patterns visible:
- **Elm Architecture**: Update → Model → View cycle
- **Message Passing**: Components communicate via messages
- **Async Tasks**: Background operations with callbacks
- **Layered Architecture**: CLI → TUI → Business Logic → External

Use these diagrams to:
- Understand system structure
- Plan new features
- Debug issues
- Onboard new developers
- Document architecture decisions
