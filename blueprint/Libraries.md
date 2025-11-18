# Libraries and Dependencies

This document catalogs all libraries used in gh-dash, their purpose, version, and why they were chosen.

## Table of Contents
1. [Core Framework Libraries](#core-framework-libraries)
2. [UI and Styling Libraries](#ui-and-styling-libraries)
3. [GitHub Integration](#github-integration)
4. [Configuration and Validation](#configuration-and-validation)
5. [Git Integration](#git-integration)
6. [Utilities](#utilities)
7. [Build and Development Tools](#build-and-development-tools)
8. [Testing Libraries](#testing-libraries)

---

## Core Framework Libraries

### 1. Bubbletea
- **Package**: `github.com/charmbracelet/bubbletea v1.3.5`
- **Purpose**: Terminal UI framework based on Elm Architecture
- **Why chosen**:
  - Proven pattern (Elm Architecture)
  - Active development and community
  - Excellent developer experience
  - Pure functional approach
  - Built-in support for commands and subscriptions

**Usage**:
```go
import tea "github.com/charmbracelet/bubbletea"

type Model struct { }

func (m Model) Init() tea.Cmd { }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { }
func (m Model) View() string { }
```

**Key Features Used**:
- Model-Update-View pattern
- Message passing
- Command system for async operations
- Window resize handling
- Keyboard/mouse input
- Batch commands

---

### 2. Cobra
- **Package**: `github.com/spf13/cobra v1.9.1`
- **Purpose**: CLI framework for building command-line applications
- **Why chosen**:
  - Industry standard for Go CLIs
  - Rich flag parsing
  - Subcommand support
  - Help generation
  - Shell completion

**Usage**:
```go
import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
    Use:   "gh-dash",
    Short: "GitHub dashboard in your terminal",
    Run: func(cmd *cobra.Command, args []string) {
        // Run TUI
    },
}

rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file")
```

**Key Features Used**:
- Root command definition
- Flag parsing (--config, --debug, --cpuprofile)
- Help text generation
- Error handling

---

## UI and Styling Libraries

### 3. Lipgloss
- **Package**: `github.com/charmbracelet/lipgloss v1.1.1`
- **Purpose**: Terminal styling and layout library
- **Why chosen**:
  - Declarative styling (like CSS)
  - Adaptive colors (light/dark mode)
  - Layout primitives (Join, Place, Align)
  - Border styles
  - Part of Charmbracelet ecosystem

**Usage**:
```go
import "github.com/charmbracelet/lipgloss"

var style = lipgloss.NewStyle().
    Foreground(lipgloss.Color("205")).
    Background(lipgloss.Color("235")).
    Padding(1, 2).
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("63"))

rendered := style.Render("Hello, World!")
```

**Key Features Used**:
- Color management (AdaptiveColor)
- Borders and padding
- Alignment (Place, JoinVertical, JoinHorizontal)
- Width/height calculations
- Text truncation and wrapping

---

### 4. Bubbles
- **Package**: `github.com/charmbracelet/bubbles v0.21.0`
- **Purpose**: Reusable Bubbletea components
- **Why chosen**:
  - Pre-built common components
  - Consistent API
  - Battle-tested
  - Part of Charmbracelet ecosystem

**Components Used**:
- **textinput**: Text input fields (search, prompts)
- **spinner**: Loading indicators
- **key**: Keyboard binding helpers

**Usage**:
```go
import "github.com/charmbracelet/bubbles/textinput"

input := textinput.New()
input.Placeholder = "Search..."
input.Focus()

// In Update
input, cmd = input.Update(msg)

// In View
view := input.View()
```

---

### 5. Glamour
- **Package**: `github.com/charmbracelet/glamour v0.10.0`
- **Purpose**: Markdown rendering in the terminal
- **Why chosen**:
  - Beautiful markdown rendering
  - Syntax highlighting
  - Configurable styles
  - Part of Charmbracelet ecosystem

**Usage**:
```go
import "github.com/charmbracelet/glamour"

renderer, _ := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
    glamour.WithWordWrap(width),
)
rendered, _ := renderer.Render(markdown)
```

**Used For**:
- Rendering PR/Issue descriptions
- Rendering comments
- Displaying markdown in sidebars

---

### 6. Bubblezone
- **Package**: `github.com/lrstanley/bubblezone v1.0.0`
- **Purpose**: Mouse support for Bubbletea
- **Why chosen**:
  - Easy mouse region detection
  - Click handling
  - Hover support

**Usage**:
```go
import zone "github.com/lrstanley/bubblezone"

// Mark clickable area
content := zone.Mark("button-id", "Click me")

// In Update, handle clicks
if zone.Get("button-id").InBounds(msg) {
    // Button was clicked
}
```

**Used For**:
- Clickable table rows
- Clickable buttons
- Mouse navigation

---

### 7. Charmbracelet Log
- **Package**: `github.com/charmbracelet/log v0.4.2`
- **Purpose**: Structured logging library
- **Why chosen**:
  - Structured logging
  - Pretty output
  - Log levels
  - Part of Charmbracelet ecosystem

**Usage**:
```go
import "github.com/charmbracelet/log"

log.Debug("Fetching PRs", "filters", filters, "limit", limit)
log.Info("Started task", "id", taskId)
log.Error("Failed to fetch", "err", err)
```

**Log Levels**:
- Debug: Detailed info for debugging
- Info: General informational messages
- Warn: Warning messages
- Error: Error messages

---

## GitHub Integration

### 8. go-gh
- **Package**: `github.com/cli/go-gh/v2 v2.12.1`
- **Purpose**: Official GitHub CLI library
- **Why chosen**:
  - Official GitHub library
  - Authentication handling (uses gh CLI auth)
  - GraphQL client
  - Repository utilities

**Usage**:
```go
import (
    gh "github.com/cli/go-gh/v2/pkg/api"
    "github.com/cli/go-gh/v2/pkg/repository"
)

// GraphQL client
client, err := gh.DefaultGraphQLClient()

// Get current repo
repo, err := repository.Current()
```

**Key Features Used**:
- GraphQL client creation
- Authentication (via gh CLI)
- Repository detection
- API endpoint access

---

### 9. shurcooL-graphql
- **Package**: `github.com/cli/shurcooL-graphql v0.0.4`
- **Purpose**: GraphQL client for Go
- **Why chosen**:
  - Type-safe queries
  - Auto-generated queries from structs
  - Pagination support
  - Used by gh CLI

**Usage**:
```go
import graphql "github.com/cli/shurcooL-graphql"

var query struct {
    Repository struct {
        PullRequests struct {
            Nodes []PullRequest
        } `graphql:"pullRequests(first: $limit)"`
    } `graphql:"repository(owner: $owner, name: $name)"`
}

variables := map[string]interface{}{
    "owner": graphql.String(owner),
    "name":  graphql.String(name),
    "limit": graphql.Int(limit),
}

err := client.Query("FetchPRs", &query, variables)
```

---

### 10. githubv4
- **Package**: `github.com/shurcooL/githubv4 v0.0.0-20240727222349-48295856cce7`
- **Purpose**: GitHub GraphQL API v4 client
- **Why chosen**:
  - Official GitHub v4 API support
  - Type definitions for GitHub types
  - Mutation support

---

### 11. gh-checks
- **Package**: `github.com/dlvhdr/x/gh-checks v0.0.0-20251114174027-320f1169b8e8`
- **Purpose**: GitHub check runs/status utilities
- **Why chosen**:
  - Simplifies check status parsing
  - Type definitions for check states

---

## Configuration and Validation

### 12. Koanf
- **Package**: `github.com/knadh/koanf/v2 v2.3.0`
- **Purpose**: Configuration management library
- **Why chosen**:
  - Multiple config sources (file, env, flags)
  - Layered configs (merge global + local)
  - Type-safe unmarshaling
  - Validation support

**Related Packages**:
- `github.com/knadh/koanf/parsers/yaml v1.1.0` - YAML parsing
- `github.com/knadh/koanf/providers/file v1.2.0` - File provider
- `github.com/knadh/koanf/maps v0.1.2` - Map utilities

**Usage**:
```go
import (
    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/parsers/yaml"
    "github.com/knadh/koanf/providers/file"
)

k := koanf.New(".")
k.Load(file.Provider(configPath), yaml.Parser())

var config Config
k.Unmarshal("", &config)
```

**Key Features Used**:
- YAML parsing
- Config merging (global + local)
- Nested config support
- Type unmarshaling

---

### 13. Validator
- **Package**: `github.com/go-playground/validator/v10 v10.27.0`
- **Purpose**: Struct validation
- **Why chosen**:
  - Declarative validation via struct tags
  - Rich validation rules
  - Custom validators
  - Industry standard

**Usage**:
```go
import "github.com/go-playground/validator/v10"

type Config struct {
    Port   int    `validate:"required,min=1,max=65535"`
    Email  string `validate:"required,email"`
    Width  *int   `validate:"omitempty,gt=0"`
}

validate := validator.New()
err := validate.Struct(config)
```

**Validations Used**:
- `required`: Field must be present
- `omitempty`: Skip if empty
- `gt=0`: Greater than zero
- `hexcolor`: Valid hex color
- `email`: Valid email

---

## Git Integration

### 14. git-module
- **Package**: `github.com/aymanbagabas/git-module v1.8.4`
- **Purpose**: Git operations library
- **Why chosen**:
  - Pure Go git implementation
  - No system git dependency
  - Repository inspection
  - Branch management

**Usage**:
```go
import gitm "github.com/aymanbagabas/git-module"

repo, err := gitm.Open(repoPath)

// Get branches
branches, err := repo.Branches()

// Get status
status, err := repo.Status()
```

**Key Features Used**:
- Open repository
- List branches
- Get commit info
- Track changes
- Remote management

---

## Utilities

### 15. clipboard
- **Package**: `github.com/atotto/clipboard v0.1.4`
- **Purpose**: Cross-platform clipboard access
- **Why chosen**:
  - Simple API
  - Cross-platform (Linux, Mac, Windows)
  - No external dependencies

**Usage**:
```go
import "github.com/atotto/clipboard"

// Copy to clipboard
clipboard.WriteAll("text to copy")

// Read from clipboard
text, err := clipboard.ReadAll()
```

**Used For**:
- Copy PR/Issue number
- Copy PR/Issue URL

---

### 16. browser
- **Package**: `github.com/cli/browser v1.3.0`
- **Purpose**: Open URLs in browser
- **Why chosen**:
  - Cross-platform browser opening
  - Detects default browser
  - Part of gh CLI ecosystem

**Usage**:
```go
import "github.com/cli/browser"

browser.OpenURL("https://github.com/owner/repo/pull/123")
```

---

### 17. beeep
- **Package**: `github.com/gen2brain/beeep v0.11.1`
- **Purpose**: System notifications
- **Why chosen**:
  - Cross-platform notifications
  - Simple API
  - Desktop integration

**Usage**:
```go
import "github.com/gen2brain/beeep"

beeep.Notify("gh-dash", "PR checks passed!", "")
```

**Used For**:
- Notify when PR checks complete
- Alert on important events

---

### 18. termenv
- **Package**: `github.com/muesli/termenv v0.16.0`
- **Purpose**: Terminal capability detection
- **Why chosen**:
  - Detect color support
  - Light/dark mode detection
  - Terminal feature detection

**Usage**:
```go
import "github.com/muesli/termenv"

// Detect light/dark mode
hasDarkBackground := termenv.HasDarkBackground()

// Get color profile
profile := termenv.ColorProfile()
```

---

### 19. go-sprout
- **Package**: `github.com/go-sprout/sprout v1.0.1`
- **Purpose**: Template function library
- **Why chosen**:
  - Rich template functions
  - Time manipulation
  - String utilities

**Usage**:
```go
import "github.com/go-sprout/sprout"

handler := sprout.New()
funcs := handler.Build()

tmpl, _ := template.New("search").Funcs(funcs).Parse(searchValue)
```

**Used For**:
- Template functions in search filters
- Date/time formatting

---

## Build and Development Tools

### 20. Goreleaser
- **Config**: `.goreleaser.yaml`
- **Purpose**: Automated release building
- **Features**:
  - Cross-platform builds
  - Multi-architecture support
  - GitHub release creation
  - Checksum generation

**Platforms**:
- Linux (amd64, arm64, arm, 386)
- macOS (amd64, arm64)
- Windows (amd64, arm64, 386)
- FreeBSD (amd64, arm64, 386)
- Android (arm64)

---

### 21. golangci-lint
- **Package**: `github.com/golangci/golangci-lint`
- **Purpose**: Linting and static analysis
- **Linters Used**:
  - `bodyclose`: Check HTTP body closes
  - `goprintffuncname`: Check Printf function names
  - `misspell`: Spell checking
  - `nolintlint`: Check nolint directives
  - `rowserrcheck`: Check SQL row errors
  - `sqlclosecheck`: Check SQL statement closes
  - `staticcheck`: Static analysis
  - `tparallel`: Check t.Parallel usage
  - `whitespace`: Check whitespace

---

### 22. Task (go-task)
- **Tool**: Task runner
- **Config**: `Taskfile.yaml`
- **Purpose**: Build automation
- **Tasks**:
  - `task`: Run application
  - `task debug`: Run with debug logging
  - `task test`: Run tests
  - `task lint`: Run linters
  - `task fmt`: Format code
  - `task docs-dev`: Run docs server

---

### 23. gofumpt
- **Tool**: Stricter gofmt
- **Purpose**: Code formatting
- **Why chosen**:
  - Stricter than gofmt
  - Consistent formatting
  - Removes ambiguity

---

### 24. nerdfix
- **Tool**: Nerd font fixer
- **Purpose**: Fix nerd font characters
- **Usage**:
```bash
task check-nerd-font  # Check for issues
task fix-nerd-font    # Fix nerd fonts
```

---

## Testing Libraries

### 25. testify
- **Package**: `github.com/stretchr/testify v1.10.0`
- **Purpose**: Testing toolkit
- **Why chosen**:
  - Rich assertions
  - Mocking support
  - Suite support

**Usage**:
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
    assert.Equal(t, 123, value)
    assert.Contains(t, str, "expected")
    assert.NoError(t, err)
}
```

---

### 26. teatest
- **Package**: `github.com/charmbracelet/x/exp/teatest v0.0.0`
- **Purpose**: Testing Bubbletea applications
- **Why chosen**:
  - Official Bubbletea testing utilities
  - Simulate messages
  - Test Update/View

**Usage**:
```go
import "github.com/charmbracelet/x/exp/teatest"

tm := teatest.NewTestModel(t, model)
tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
```

---

## Dependency Management

### go.mod
```go
module github.com/dlvhdr/gh-dash/v4

go 1.24.7

// Direct dependencies
require (
    github.com/charmbracelet/bubbletea v1.3.5
    github.com/charmbracelet/lipgloss v1.1.1
    github.com/charmbracelet/glamour v0.10.0
    github.com/spf13/cobra v1.9.1
    github.com/cli/go-gh/v2 v2.12.1
    // ... more
)

// Indirect dependencies (40+)
require (
    // Transitive dependencies
)
```

### Versioning Strategy
- **Semantic versioning**: Major.Minor.Patch
- **v4 module**: Current major version
- **Controlled updates**: Dependencies updated deliberately
- **Dependabot**: Automated dependency updates

---

## Why These Libraries?

### Common Themes

1. **Charmbracelet Ecosystem**: Cohesive, well-maintained, active community
2. **Type Safety**: Prefer type-safe over runtime checks
3. **Pure Go**: No C dependencies (except system libs)
4. **Cross-platform**: Works on Linux, Mac, Windows
5. **Active Development**: Libraries still maintained
6. **Community**: Popular, well-documented

### Trade-offs

#### Chosen: Bubbletea
- **Pro**: Elm pattern, testable, composable
- **Con**: Steeper learning curve than imperative

#### Chosen: GraphQL
- **Pro**: Efficient, type-safe, single request
- **Con**: More complex than REST

#### Chosen: Koanf
- **Pro**: Flexible, layered configs
- **Con**: More complex than viper

---

## Summary

gh-dash uses **30+ libraries** across these categories:

| Category | Libraries | Purpose |
|----------|-----------|---------|
| **Framework** | Bubbletea, Cobra | TUI, CLI |
| **UI/Styling** | Lipgloss, Bubbles, Glamour | Rendering, components |
| **GitHub** | go-gh, graphql, githubv4 | API integration |
| **Config** | Koanf, Validator | Configuration |
| **Git** | git-module | Local git operations |
| **Utils** | clipboard, browser, beeep | System integration |
| **Build** | Goreleaser, golangci-lint | CI/CD |
| **Test** | testify, teatest | Testing |

All libraries chosen for:
- **Reliability**: Battle-tested, maintained
- **Compatibility**: Cross-platform, pure Go
- **Integration**: Work well together
- **Philosophy**: Align with project values

When adding new libraries, consider:
1. Is it maintained?
2. Does it fit the ecosystem?
3. Is it cross-platform?
4. Does it add significant value?
5. Are there lighter alternatives?
