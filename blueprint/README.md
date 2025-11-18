# gh-dash Blueprint Documentation

Welcome to the complete architectural blueprint for gh-dash! This documentation serves as a comprehensive guide to understanding and replicating the technical decisions, design patterns, and implementation strategies used in gh-dash.

## Purpose

This blueprint exists to:

1. **Document the Architecture**: Capture all architectural decisions and their rationale
2. **Enable Replication**: Provide a template for building similar applications
3. **Preserve Knowledge**: Ensure design decisions and patterns are preserved
4. **Facilitate Onboarding**: Help new developers understand the codebase quickly
5. **Guide Development**: Serve as a reference for maintaining consistency

## What is gh-dash?

gh-dash is a **Terminal User Interface (TUI)** application that provides a rich, interactive interface for managing GitHub pull requests, issues, and repository branches directly from the terminal.

**Technology Stack:**
- **Language**: Go 1.24.7+
- **TUI Framework**: Charmbracelet Bubbletea (Elm Architecture)
- **CLI Framework**: Cobra
- **Styling**: Lipgloss
- **API**: GitHub GraphQL via go-gh

## Documentation Structure

This blueprint consists of the following documents:

### Core Architecture
- **[Architecture.md](./Architecture.md)** - Complete architectural overview
  - High-level architecture
  - Module breakdown
  - Application flow
  - State management
  - Concurrency model
  - Performance considerations

### Design and Patterns
- **[Design-Patterns.md](./Design-Patterns.md)** - All design patterns used
  - Elm Architecture
  - Component Pattern
  - Message Passing
  - Factory, Strategy, Adapter patterns
  - Architectural decisions and rationale
  - Anti-patterns avoided

### User Interface
- **[UI-Structure.md](./UI-Structure.md)** - UI organization and layout
  - Overall layout structure
  - View types (PRs, Issues, Repo)
  - Component hierarchy
  - Responsive design
  - Navigation patterns
  - Modal system

- **[Component-Design.md](./Component-Design.md)** - Component architecture guide
  - Component anatomy
  - Component lifecycle
  - Component catalog
  - Creating new components
  - Best practices

- **[Styling.md](./Styling.md)** - Styling and theming system
  - Lipgloss fundamentals
  - Theme system
  - Adaptive colors (light/dark mode)
  - Layout and spacing
  - User customization

### Technical Implementation
- **[Libraries.md](./Libraries.md)** - All libraries and dependencies
  - Core framework libraries
  - UI and styling libraries
  - GitHub integration
  - Configuration and validation
  - Build and development tools

- **[Data-Flow.md](./Data-Flow.md)** - Data flow and state management
  - Elm Architecture data flow
  - Message types
  - State updates
  - API integration
  - Caching strategy
  - Error handling

- **[Configuration.md](./Configuration.md)** - Configuration system
  - Configuration structure
  - Loading and merging
  - Validation
  - Default values
  - User customization

### Visual Documentation
- **[UML-Diagrams.md](./UML-Diagrams.md)** - Visual diagrams
  - Class diagrams
  - Sequence diagrams
  - State diagrams
  - Component diagrams
  - Activity diagrams

### Project Management
- **[Checklist.md](./Checklist.md)** - Task tracking checklist
  - Completed tasks
  - Pending tasks
  - Future enhancements

## Quick Start Guide

### For Understanding the Project

1. Start with **[Architecture.md](./Architecture.md)** for the big picture
2. Review **[Design-Patterns.md](./Design-Patterns.md)** to understand patterns
3. Explore **[UML-Diagrams.md](./UML-Diagrams.md)** for visual understanding
4. Deep dive into specific areas as needed

### For Building Similar Applications

1. Study **[Architecture.md](./Architecture.md)** for architectural decisions
2. Review **[Design-Patterns.md](./Design-Patterns.md)** for pattern rationale
3. Use **[Component-Design.md](./Component-Design.md)** as a template
4. Reference **[Libraries.md](./Libraries.md)** for technology choices
5. Follow **[Styling.md](./Styling.md)** for UI consistency

### For Contributing to gh-dash

1. Read **[Architecture.md](./Architecture.md)** to understand the system
2. Study **[Component-Design.md](./Component-Design.md)** for component patterns
3. Check **[Data-Flow.md](./Data-Flow.md)** for state management
4. Review **[Checklist.md](./Checklist.md)** for current status
5. Follow patterns demonstrated in existing code

## Key Principles

### Architectural Principles

1. **Elm Architecture**: Model-Update-View pattern for predictable state management
2. **Component-Based**: Self-contained, reusable components
3. **Message-Driven**: All communication via typed messages
4. **Immutable State**: State updates create new state
5. **Pure Functions**: Deterministic, testable functions
6. **Composition**: Complex behavior from simple components

### Design Principles

1. **Single Responsibility**: Each component does one thing well
2. **Open/Closed**: Open for extension, closed for modification
3. **Dependency Inversion**: Depend on abstractions, not concretions
4. **Interface Segregation**: Small, focused interfaces
5. **DRY**: Don't Repeat Yourself
6. **KISS**: Keep It Simple, Stupid

### Code Quality Principles

1. **Type Safety**: Leverage Go's type system
2. **Error Handling**: Graceful error handling throughout
3. **Testing**: Unit tests for critical paths
4. **Documentation**: Clear comments and docs
5. **Consistency**: Follow established patterns
6. **Performance**: Optimize where it matters

## Technology Decisions

### Why Bubbletea?

**Chosen**: Charmbracelet Bubbletea
**Rationale**:
- Proven Elm Architecture pattern
- Excellent developer experience
- Active community and development
- Rich ecosystem (Lipgloss, Bubbles, Glamour)
- Type-safe, functional approach

### Why GraphQL?

**Chosen**: GitHub GraphQL API
**Rationale**:
- Fetch exactly what's needed (no over/under-fetching)
- Single request for complex data
- Type-safe with code generation
- GitHub's preferred API going forward

### Why Go?

**Chosen**: Go programming language
**Rationale**:
- Fast compilation and execution
- Easy cross-platform builds
- Strong typing and tooling
- Excellent for CLIs and TUIs
- Great concurrency primitives

## Replication Guide

To build a similar application:

### 1. Choose Your Framework

**For TUI applications**:
- Use Bubbletea for Elm Architecture
- Use Lipgloss for styling
- Use Bubbles for common components

### 2. Define Your Architecture

- Decide on state management (recommend Elm Architecture)
- Plan component hierarchy
- Define message types
- Establish data flow

### 3. Set Up Configuration

- YAML-based configuration (user-friendly)
- Validation library (go-playground/validator)
- Hierarchical config loading (Koanf)
- Sensible defaults

### 4. Design Your Components

- Follow single responsibility principle
- Make components composable
- Use interfaces for flexibility
- Implement lifecycle methods

### 5. Implement Styling

- Centralize styles
- Support theming
- Handle light/dark mode
- Make it responsive

### 6. Add Data Layer

- Abstract API calls
- Implement caching
- Handle errors gracefully
- Support async operations

### 7. Testing

- Unit test pure functions
- Test component updates
- Test data transformations
- Manual TUI testing

## Common Patterns to Reuse

### 1. Component Pattern

```go
type Model struct {
    // State
    value string
    ctx   *ProgramContext
}

func NewModel(ctx *ProgramContext) Model { }
func (m Model) Init() tea.Cmd { }
func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) { }
func (m Model) View() string { }
```

### 2. Message Passing

```go
type CustomMsg struct {
    Data string
}

func fetchData() tea.Cmd {
    return func() tea.Msg {
        data := fetch()
        return CustomMsg{Data: data}
    }
}
```

### 3. Context Pattern

```go
type ProgramContext struct {
    Config  *Config
    Theme   Theme
    Styles  *Styles
    // ... shared state
}
```

### 4. Theme System

```go
type Theme struct {
    PrimaryText lipgloss.AdaptiveColor
    // ... colors
}

func ParseTheme(cfg *Config) Theme {
    // Load theme from config
}
```

## File Organization

```
your-project/
├── cmd/                  # CLI commands
│   └── root.go
├── internal/             # Application code
│   ├── config/          # Configuration
│   ├── data/            # Data access layer
│   ├── tui/             # Terminal UI
│   │   ├── components/  # UI components
│   │   ├── context/     # Shared context
│   │   ├── keys/        # Key bindings
│   │   └── theme/       # Theming
│   └── utils/           # Utilities
├── blueprint/           # Documentation
├── go.mod               # Dependencies
└── main.go              # Entry point
```

## Learning Path

### Week 1: Fundamentals
- [ ] Read Architecture.md
- [ ] Study Design-Patterns.md
- [ ] Review UML-Diagrams.md
- [ ] Understand Elm Architecture

### Week 2: Implementation
- [ ] Study Component-Design.md
- [ ] Review Data-Flow.md
- [ ] Understand configuration system
- [ ] Build a simple component

### Week 3: Advanced Topics
- [ ] Study styling system
- [ ] Understand caching strategy
- [ ] Review error handling
- [ ] Build a complex feature

### Week 4: Mastery
- [ ] Contribute to codebase
- [ ] Optimize performance
- [ ] Add tests
- [ ] Document changes

## Additional Resources

### Bubbletea Resources
- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Bubbletea Examples](https://github.com/charmbracelet/bubbletea/tree/master/examples)
- [Lipgloss Examples](https://github.com/charmbracelet/lipgloss/tree/master/examples)

### Elm Architecture
- [The Elm Architecture Guide](https://guide.elm-lang.org/architecture/)
- [Elm Architecture Explained](https://dennisreimann.de/articles/elm-architecture-overview.html)

### Go Resources
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## Contributing

When contributing to gh-dash:

1. Read this blueprint thoroughly
2. Follow established patterns
3. Update documentation as needed
4. Add tests for new features
5. Update Checklist.md

## Maintenance

### Keeping Documentation Updated

- Update docs when architecture changes
- Add new patterns to Design-Patterns.md
- Update UML diagrams for major changes
- Keep Checklist.md current

### Review Schedule

- **Monthly**: Review for accuracy
- **Quarterly**: Update for major changes
- **Yearly**: Comprehensive review

## Questions?

If you have questions about:

- **Architecture**: See Architecture.md
- **Patterns**: See Design-Patterns.md
- **Components**: See Component-Design.md
- **Styling**: See Styling.md
- **Data Flow**: See Data-Flow.md
- **Configuration**: See Configuration.md

## Credits

This blueprint documents the architecture of **gh-dash**, created by [dlvhdr](https://github.com/dlvhdr).

The Charmbracelet ecosystem (Bubbletea, Lipgloss, Bubbles) is created and maintained by [Charmbracelet](https://github.com/charmbracelet).

## License

This documentation is part of the gh-dash project and follows the same license.

---

**Last Updated**: 2025-11-18

**Version**: 4.x

**Status**: Complete and actively maintained
