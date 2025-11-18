# Project Checklist

This checklist tracks completed features, pending tasks, and future enhancements for gh-dash.

## Documentation Status

### Completed ✓
- [x] Architecture documentation
- [x] Design patterns documentation
- [x] UI structure documentation
- [x] Component design guide
- [x] Libraries and dependencies catalog
- [x] Styling and theming guide
- [x] UML diagrams
- [x] Data flow documentation
- [x] Configuration system documentation
- [x] Blueprint README
- [x] Project checklist

## Core Features Status

### Completed ✓

#### Pull Requests View
- [x] List pull requests from GitHub
- [x] Multiple configurable sections
- [x] Filter by author, reviewer, repository
- [x] Search/filter functionality
- [x] PR details sidebar
- [x] View PR description
- [x] View PR activity (comments, reviews)
- [x] View changed files
- [x] View CI/CD check status
- [x] Approve PR
- [x] Merge PR
- [x] Close/reopen PR
- [x] Mark as ready for review
- [x] Update PR (sync with base)
- [x] Comment on PR
- [x] Open PR in browser
- [x] Copy PR number/URL

#### Issues View
- [x] List issues from GitHub
- [x] Multiple configurable sections
- [x] Filter by assignee, author, labels
- [x] Search/filter functionality
- [x] Issue details sidebar
- [x] View issue description
- [x] View issue comments
- [x] View labels and assignees
- [x] Close/reopen issue
- [x] Comment on issue
- [x] Assign issue
- [x] Add labels
- [x] Open issue in browser
- [x] Copy issue number/URL

#### Repository/Branch View
- [x] List local branches
- [x] Show commits ahead/behind
- [x] Show last commit info
- [x] Checkout branch
- [x] Delete branch
- [x] Create PR from branch
- [x] Branch details sidebar

#### Configuration System
- [x] YAML-based configuration
- [x] Hierarchical config loading (global + local)
- [x] Config validation
- [x] Default configuration
- [x] Custom sections
- [x] Custom keybindings
- [x] Theme customization
- [x] Layout customization
- [x] Repository-local configs

#### UI/UX Features
- [x] Responsive terminal UI
- [x] Keyboard navigation
- [x] Mouse support
- [x] Tab navigation between sections
- [x] Sidebar toggle
- [x] Search mode
- [x] Confirmation prompts
- [x] Loading indicators
- [x] Error messages
- [x] Help overlay
- [x] Light/dark mode support
- [x] Adaptive colors
- [x] Custom icons
- [x] Nerd font support

#### Performance
- [x] Async data fetching
- [x] Non-blocking UI
- [x] Caching with TTL
- [x] Auto-refresh
- [x] Pagination support
- [x] Virtual scrolling (table)

#### Developer Experience
- [x] Debug mode
- [x] Logging system
- [x] CPU profiling support
- [x] Task runner (Taskfile)
- [x] Linting (golangci-lint)
- [x] Code formatting (gofumpt)
- [x] Testing framework

#### Build and Release
- [x] Cross-platform builds
- [x] Multi-architecture support
- [x] Automated releases (Goreleaser)
- [x] GitHub Actions CI/CD
- [x] Checksums
- [x] Release notes

## Pending Tasks

### High Priority
- [ ] Add test coverage for core components
- [ ] Add integration tests
- [ ] Performance profiling and optimization
- [ ] Memory usage optimization
- [ ] Improve error messages
- [ ] Add retry logic for network failures

### Medium Priority
- [ ] Add PR review request functionality
- [ ] Add issue milestone management
- [ ] Add PR/Issue assignment from TUI
- [ ] Add label creation/editing
- [ ] Export data to CSV/JSON
- [ ] Add custom filters with saved presets

### Low Priority
- [ ] Add statistics/metrics view
- [ ] Add timeline view for PR/Issue activity
- [ ] Add diff viewer within TUI
- [ ] Add commit history view
- [ ] Add GitHub Actions workflow trigger
- [ ] Add notification system preferences

## Future Enhancements

### Features to Consider
- [ ] Multi-repository view
- [ ] Organization-wide dashboard
- [ ] PR review workflow
- [ ] Code review inline comments
- [ ] GitHub Projects integration
- [ ] GitHub Discussions integration
- [ ] Pull request templates
- [ ] Issue templates
- [ ] Automated PR descriptions
- [ ] AI-assisted PR reviews
- [ ] Webhook support for real-time updates
- [ ] Desktop notifications (macOS, Linux, Windows)
- [ ] Sound notifications
- [ ] Custom themes (import/export)
- [ ] Plugin system
- [ ] Scripting support
- [ ] GitLab support
- [ ] Bitbucket support

### Technical Improvements
- [ ] Increase test coverage to 80%+
- [ ] Add benchmark tests
- [ ] Add performance regression tests
- [ ] Improve documentation coverage
- [ ] Add API documentation
- [ ] Add architecture decision records (ADRs)
- [ ] Improve error handling
- [ ] Add telemetry (opt-in)
- [ ] Add crash reporting (opt-in)
- [ ] Improve accessibility
- [ ] Add internationalization (i18n)
- [ ] Optimize bundle size
- [ ] Reduce memory footprint
- [ ] Improve startup time

### Documentation Improvements
- [ ] Add video tutorials
- [ ] Add interactive examples
- [ ] Add cookbook with recipes
- [ ] Add troubleshooting guide
- [ ] Add FAQ
- [ ] Add migration guides
- [ ] Improve API documentation
- [ ] Add contributing guide
- [ ] Add code of conduct
- [ ] Add security policy

## Known Issues

### Critical
- None currently

### High Priority
- [ ] Occasional panic on rapid terminal resize
- [ ] Large PR diffs can be slow to render
- [ ] Memory leak with long-running sessions (needs investigation)

### Medium Priority
- [ ] Mouse scroll sometimes stops working
- [ ] Search highlighting can be inconsistent
- [ ] Theme changes require restart

### Low Priority
- [ ] Some nerd font icons don't render in all terminals
- [ ] Very long PR titles truncate awkwardly
- [ ] Footer can overlap content on small terminals

## Testing Status

### Test Coverage
- **Overall Coverage**: ~40% (needs improvement)
- **Core Model**: 60%
- **Components**: 30%
- **Config**: 70%
- **Data Layer**: 20%

### Test Types
- [x] Unit tests (partial)
- [ ] Integration tests (missing)
- [ ] E2E tests (missing)
- [ ] Performance tests (missing)
- [ ] Regression tests (missing)

## Release History

### v4.x (Current)
- Module path: `github.com/dlvhdr/gh-dash/v4`
- Go version: 1.24.7+
- Latest features: Branch view, improved sidebar, better theming

### Previous Versions
- v3.x: Major refactoring
- v2.x: Added issues view
- v1.x: Initial release

## Metrics and Goals

### Current Metrics
- **Stars**: Tracked on GitHub
- **Downloads**: Via GitHub releases
- **Contributors**: Open source project
- **Issues**: Tracked on GitHub
- **PRs**: Tracked on GitHub

### Goals for Next Release
- [ ] 80% test coverage
- [ ] 10% performance improvement
- [ ] 50% reduction in memory usage
- [ ] 5+ new features
- [ ] 20+ bug fixes
- [ ] Improved documentation

## Community

### Contributing
- [ ] Add CONTRIBUTING.md
- [ ] Add issue templates
- [ ] Add PR templates
- [ ] Add code owners
- [ ] Add style guide

### Communication
- [ ] Discord server
- [ ] Discussions on GitHub
- [ ] Twitter/social media

## Maintenance

### Regular Tasks
- [ ] Weekly: Review new issues
- [ ] Weekly: Review new PRs
- [ ] Monthly: Update dependencies
- [ ] Monthly: Review and update docs
- [ ] Quarterly: Major version planning
- [ ] Yearly: Comprehensive refactoring review

### Dependency Updates
- [ ] Bubbletea: Keep updated
- [ ] Lipgloss: Keep updated
- [ ] go-gh: Monitor for breaking changes
- [ ] Other deps: Monthly review

## Notes

### Architecture Decisions
- Elm Architecture chosen for predictability and testability
- Bubbletea chosen for excellent DX and active community
- GraphQL chosen for efficient data fetching
- Go chosen for performance and cross-platform support

### Design Decisions
- Component-based architecture for modularity
- Message passing for decoupling
- Immutable state for predictability
- Pure functions for testability
- Composition over inheritance

### Future Considerations
- Consider adding plugin system for extensibility
- Consider adding scripting support (Lua/Starlark)
- Consider multi-repository support
- Consider supporting other Git platforms
- Consider mobile companion app

---

**Last Updated**: 2025-11-18

**Maintainer**: gh-dash team

**Status**: Actively developed and maintained
