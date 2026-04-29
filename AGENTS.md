# AGENTS.md

This file provides guidance to AI coding agents (Claude Code, GitHub Copilot, Cursor, Codex, Aider) when working with code in this repository. `CLAUDE.md` is a symlink to this file.

## Project

gh-dash is a `gh` CLI extension that renders a terminal dashboard for GitHub PRs, issues, and notifications. It's a Go TUI built on Charm's Bubble Tea (Elm-style Model/Update/View) with Lip Gloss v2 styling, Glamour for markdown rendering, and Cobra for the CLI entrypoint. Data comes from GitHub's GraphQL API via the `gh` CLI.

## Commands

Tooling is driven by `Taskfile.yaml` (go-task), **not** `make`. Devbox provides the dev shell (`devbox shell`) — all Task commands assume its toolchain.

| Task | Command |
| --- | --- |
| Run locally against your real config | `task` (alias for `task default` → `go run .`) |
| Build & install as `gh` extension | `task install` (installs the local build as `gh dash`) |
| Reinstall the released version | `task install:prod` |
| Run all tests | `task test ./...` (wraps `prism test`) |
| Run a single test | `task test:one -- -run TestName ./path/to/pkg` (wraps `gotip`) |
| Re-run last failing test | `task test:rerun` |
| Lint | `task lint` (golangci-lint, 5 min timeout) |
| Auto-fix lint | `task lint:fix` |
| Format staged Go files | `task fmt` (gofumpt) |
| Debug run with file logging | `task debug` → writes `./debug.log`; tail with `task logs` |
| Warn-only debug | `task debug:warn` |
| Headless Delve debugger | `task dlv` (listens on `127.0.0.1:43000`) |
| Docs dev server | `task docs` (Astro Starlight under `docs/`, pnpm) |
| Docs production build | `task docs-build` |

No git pre-commit hooks are configured — run `task fmt && task lint && task test ./...` before pushing. CI enforces `task lint`, so fixing locally saves a round trip.

## Architecture

Entry flow: `gh-dash.go` → `cmd.Execute()` (`cmd/root.go`, Cobra) → `tui.NewModel()` (`internal/tui/ui.go`) → `tea.NewProgram(model).Run()`.

### Bubble Tea loop
The root `Model` in `internal/tui/ui.go` composes `sidebar`, `prView`, `issueSidebar`, `branchSidebar`, `sections`, `footer`, and `tabs`. `Init()` kicks off async config loading via `tea.Batch(tea.RequestBackgroundColor, m.initScreen)`. `Update(msg)` routes messages to sub-components and merges returned `tea.Cmd`s; `View()` composes their output with Lip Gloss.

### Sections (the extension point)
"Sections" are the units a user configures via YAML (e.g. a PR filter like "My open PRs"). The `Section` interface and `BaseModel` live in `internal/tui/components/section/section.go`. Concrete sections implement it under:

- `internal/tui/components/prssection/`
- `internal/tui/components/issuessection/`
- `internal/tui/components/notificationssection/`
- `internal/tui/components/reposection/`

Each section embeds `BaseModel` (which provides search, pagination, prompt confirmation, and a `table.Model`) and owns a typed row slice (e.g. `[]prrow.Data`). Adding a new section = new package implementing the full interface + YAML plumbing.

### Data layer
`internal/data/` talks to GitHub via the `gh` CLI + `github.com/shurcooL/githubv4` (GraphQL). Key functions: `FetchPullRequests`, `FetchIssues`, `FetchNotifications`. Raw GraphQL structs (`PullRequestData`, `IssueData`) get wrapped into display types (e.g. `prrow.Data` wraps `*data.PullRequestData`) before rendering.

### Config
`internal/config/parser.go` uses **koanf**. Load order (later overrides earlier):
1. `~/.config/gh-dash/config.yml` (or `$XDG_CONFIG_HOME/...`)
2. Repo-local `.gh-dash.yml` if inside a git repo
3. `--config <path>` flag

After loading, keybindings are rebound globally via `keys.Rebind(...)` in `cmd/root.go`. User-defined custom commands are shelled out.

### UI components
Under `internal/tui/components/`:
- `table/` — core table rendering (columns, selection, keybindings)
- `cmp/` — generic floating autocomplete (used by `cmpcontroller` in `prview`/`issueview` for mentions, labels, etc.)
- `search/` — search input (recently refactored to plain `textinput`)
- `prview/`, `issueview/` — sidebar detail panels with input boxes and action lists
- `prompt/` — confirmation dialog
- `footer/`, `tabs/` — chrome

### Theming
Lip Gloss v2 with `compat.AdaptiveColor` for light/dark. Theme type in `internal/tui/theme/theme.go`; styles are accessed via `ctx.Styles.*` where `ctx` is a shared `*context.ProgramContext`.

## Non-obvious patterns

- **Async data via `TaskFinishedMsg`.** Sections return a `tea.Cmd` that performs the fetch in a goroutine and emits a domain `SectionMsg` wrapped in `constants.TaskFinishedMsg`. Before fetching, register a `context.Task` with `ctx.StartTask(task)` so the footer spinner shows progress; the root `Update` calls `context.ClearTask()` when the message comes back. **Do not block the `Update` loop with synchronous fetches.**
- **Shared mutable `ProgramContext`.** `ctx` (screen dims, view, config, theme, styles, `StartTask` callback) is shared by reference across components. Mutate carefully — there's no deep copy.
- **Section interface is wide.** Adding a new section type requires stubbing every method even if you don't need it. Start by copying an existing section package.
- **Keybindings and custom commands** are global post-load. Custom commands shell out — mind `RepoPath` resolution (see recent commit `c8e1dea`) when templating arguments.
- **Preview layout is dynamic.** `ctx.PreviewPosition` / `ctx.DynamicPreviewWidth|Height` are computed at runtime; don't hardcode dimensions.
- **Testing TUI code** uses `internal/tui/testutils/` plus golden-style `testdata/` fixtures under `internal/tui/` and `internal/config/`.

## Conventions

- Commit messages use **Conventional Commits** (release notes are auto-grouped by `feat:`, `fix:`, `docs:`, `deps:` in `.goreleaser.yaml`).
- Release flow is GoReleaser driven from tags (see `.github/workflows/go-release.yml`); do not attempt to cut releases manually.
- Docs live in `docs/` (Astro Starlight, pnpm 10.14.0); the site is completely decoupled from the Go binary.
