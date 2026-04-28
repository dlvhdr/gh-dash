# Project Structure

A map of the repository so you don't have to `fd` around to find things. Paths are relative to the repo root; links resolve in GitHub's UI.

## Top-level layout

```
.
├── gh-dash.go                 # 3-line main; calls cmd.Execute()
├── cmd/                       # Cobra + fang entry points and version ldflags
├── internal/
│   ├── config/                # koanf-based YAML parser + feature flags
│   ├── data/                  # GitHub GraphQL clients (PRs, issues, notifications)
│   ├── git/                   # Local git repo detection (used for .gh-dash.yml lookup)
│   ├── tui/                   # Bubble Tea root Model + ~25 component packages
│   └── utils/                 # Small cross-cutting helpers (string, time, etc.)
├── testdata/                  # GraphQL schema snapshots + fixtures referenced by tests
├── docs/                      # Developer documentation (this directory)
├── .github/workflows/         # CI (build+test, lint, release)
├── .goreleaser.yaml           # Release build matrix + changelog grouping
├── .golangci.yml              # Linter config (referenced by `task lint`)
├── Taskfile.yaml              # go-task recipes (build, test, lint, debug, install)
├── devbox.json                # Pinned dev toolchain (Go, golangci-lint, gofumpt, go-task, ...)
├── devbox.lock                # Devbox lockfile — do not hand-edit
├── go.mod / go.sum            # Go module (currently targets Go 1.25.8)
├── .gh-dash.yml               # Repo-local config used when running `task` from this checkout
├── CONTRIBUTING.md            # Contribution rules + AI-usage policy
├── CLAUDE.md                  # Instruction-style architecture summary for AI tooling
└── README.md                  # User-facing entry point
```

## `cmd/`

| File | Role |
| --- | --- |
| [`cmd/root.go`](../cmd/root.go) | Cobra root command, flag wiring (`--config`, `--debug`, `--cpuprofile`), version ldflag vars (`Version`, `Commit`, `Date`, `BuiltBy`), `bubblezone` global init, final `tea.NewProgram(...).Run()`. |
| [`cmd/sponsors.go`](../cmd/sponsors.go) | Subcommand that prints sponsor info. |

## `internal/config/`

Koanf-based YAML loading with validation. The parser entrypoint is `ConfigParser.ParseConfig(location Location)` in [`parser.go`](../internal/config/parser.go). Understanding the load order is covered in [architecture.md](./architecture.md#config-layer).

Files:
- [`parser.go`](../internal/config/parser.go) — the whole schema lives here: load order, XDG resolution, validation, plus every YAML struct (`Config`, `SectionConfig`, `PrsSectionConfig`, `IssuesSectionConfig`, `NotificationsSectionConfig`, `Keybindings`, `Theme`, …).
- [`feature_flags.go`](../internal/config/feature_flags.go) — `FF_*` env-driven feature flags (e.g. `FF_REPO_VIEW`, `FF_MOCK_DATA`).
- [`utils.go`](../internal/config/utils.go) — small helpers consumed by the parser.
- [`parser_test.go`](../internal/config/parser_test.go) + `testdata/` — golden-fixture tests for config parsing.

> The schema is intentionally flat — everything (sections, keybindings, theme) is defined in `parser.go`. There are no `pr_section.go` / `keybindings.go` / `theme.go` split files.

## `internal/data/`

GitHub fetchers. Shells out through the `gh` CLI's auth + `shurcooL/githubv4` for GraphQL type safety. Core surface:

- `FetchPullRequests` → `[]PullRequestData`
- `FetchIssues` → `[]IssueData`
- `FetchNotifications` → `[]NotificationData`
- Supporting types: `PageInfo`, `RowData` (interface), various per-entity structs that are later wrapped into display-layer types under `internal/tui/components/*row/`.

## `internal/git/`

Thin wrapper around local git operations — just enough to detect the current repo from the working directory so the config parser can find a repo-local `.gh-dash.yml`.

## `internal/tui/`

The bulk of the codebase. Structure:

```
internal/tui/
├── ui.go                     # Root Model, NewModel, initScreen, Update/View routing
├── common/                   # Shared UI constants (search height, dimensions)
├── constants/                # Logo, app-wide strings, TaskFinishedMsg wrapper
├── context/                  # *ProgramContext (shared state), task + task-queue types
├── keys/                     # KeyMap + Rebind (applies user overrides globally)
├── markdown/                 # Glamour-based renderer for PR/issue bodies
├── theme/                    # Theme struct + default palettes
├── testutils/                # Golden-file helpers, fixture loaders for tests
└── components/               # Visual components (see below)
```

### `internal/tui/components/`

25 packages. Grouped by role:

| Group | Packages |
| --- | --- |
| **Sections** (YAML-configurable panes) | [`prssection`](../internal/tui/components/prssection/), [`issuessection`](../internal/tui/components/issuessection/), [`notificationssection`](../internal/tui/components/notificationssection/), [`reposection`](../internal/tui/components/reposection/), [`section`](../internal/tui/components/section/) (base model + `Section` interface) |
| **Rows** (typed row wrappers for the table) | [`prrow`](../internal/tui/components/prrow/), [`issuerow`](../internal/tui/components/issuerow/), [`notificationrow`](../internal/tui/components/notificationrow/) |
| **Detail panels** (sidebar/preview for a selected item) | [`prview`](../internal/tui/components/prview/), [`issueview`](../internal/tui/components/issueview/), [`notificationview`](../internal/tui/components/notificationview/), [`sidebar`](../internal/tui/components/sidebar/), [`branchsidebar`](../internal/tui/components/branchsidebar/) |
| **Primitive UI** | [`table`](../internal/tui/components/table/), [`search`](../internal/tui/components/search/), [`inputbox`](../internal/tui/components/inputbox/), [`prompt`](../internal/tui/components/prompt/), [`listviewport`](../internal/tui/components/listviewport/), [`carousel`](../internal/tui/components/carousel/) |
| **Autocomplete** | [`cmp`](../internal/tui/components/cmp/) (floating menu), [`cmpcontroller`](../internal/tui/components/cmpcontroller/) (wires `cmp` into PR/issue view inputs for mentions, labels, etc.) |
| **Chrome** | [`footer`](../internal/tui/components/footer/), [`tabs`](../internal/tui/components/tabs/) |
| **Misc** | [`branch`](../internal/tui/components/branch/), [`tasks`](../internal/tui/components/tasks/), [`common`](../internal/tui/components/common/) |

## `internal/utils/`

Misc helpers that don't belong to a single domain (string pointers, time formatting, custom template funcs for user commands via `sprout`).

## `testdata/`

Holds the `api.github.com.graphql-schema.json` snapshot used for GraphQL type assertions plus any fixture files referenced by tests. Component-level golden files live under the component package itself (e.g. `internal/config/testdata/`).

## `.github/workflows/`

| Workflow | Trigger | Purpose |
| --- | --- | --- |
| [`build-and-test.yaml`](../.github/workflows/build-and-test.yaml) | PR (skips `docs/**`) | `go build ./...` + `go test ./...`; auto-approves & auto-merges Dependabot PRs. |
| [`lint.yml`](../.github/workflows/lint.yml) | PR | `golangci-lint` (action v9). |
| [`go-release.yml`](../.github/workflows/go-release.yml) | Git tag push (`*`) | Runs GoReleaser in `--draft` mode with Discord + Bluesky announce secrets. |
| [`dependabot-sync.yml`](../.github/workflows/dependabot-sync.yml) | Scheduled | Keeps dependabot PRs rebased. |

## Tooling config at a glance

- `.golangci.yml` — linter rule set (loaded by `task lint`).
- `.goreleaser.yaml` — release matrix (see [build-and-deployment.md](./build-and-deployment.md#release-pipeline)).
- `devbox.json` — pinned toolchain (see [development.md](./development.md#dev-environment)).
- `Taskfile.yaml` — every day-to-day command. Run `task --list` inside `devbox shell` for the full menu.
