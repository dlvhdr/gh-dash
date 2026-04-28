# Development Guide

Task-oriented recipes for contributing to `gh-dash`. Pairs with:

- [architecture.md](./architecture.md) — conceptual model.
- [project-structure.md](./project-structure.md) — where code lives.
- [build-and-deployment.md](./build-and-deployment.md) — build/test/release commands.
- [`../CONTRIBUTING.md`](../CONTRIBUTING.md) — PR-process and AI-usage policy (read first).

## Dev environment

The repo standardises on [devbox](https://www.jetify.com/devbox) to keep every contributor on the same toolchain.

```sh
# One-time
curl -fsSL https://get.jetpack.io/devbox | bash

# Every time you start work
cd gh-dash
devbox shell
```

### What `devbox shell` gives you

Pinned via [`devbox.json`](../devbox.json):

| Tool | Version | Purpose |
| --- | --- | --- |
| `go` | 1.23 | Local compiler (note the caveat below) |
| `golangci-lint` | 2.10.1 | Linter (`task lint`) |
| `gofumpt` | 0.8.0 | Formatter (`task fmt`) |
| `go-task` | 3.44.1 | `task` CLI — every recipe in [`Taskfile.yaml`](../Taskfile.yaml) |
| `nerdfix` | 0.4.2 | Validates Nerd Font glyphs in source |
| `fd` | 10.2.0 | Used by `task check-nerd-font` |
| `gh` | latest | GitHub CLI — the auth path `gh-dash` uses at runtime |
| `git` | latest | — |

Plus two Go tools installed automatically by the shell init hook via `go install`:
- `prism` ([go.dalton.dog/prism](https://go.dalton.dog/prism)) — test runner wrapper used by `task test`.
- `gotip` ([github.com/lusingander/gotip](https://github.com/lusingander/gotip)) — used by `task test:one` / `task test:rerun`.

### ⚠️ Go version caveat

There is a deliberate split between the two pinned Go versions:

| File | Version | Scope |
| --- | --- | --- |
| [`devbox.json`](../devbox.json) | `1.23` | Local `devbox shell` |
| [`go.mod`](../go.mod) | `go 1.25.8` | Module target |
| [`build-and-test.yaml`](../.github/workflows/build-and-test.yaml) | `go-version-file: ./go.mod` | CI reads from `go.mod`, so CI uses 1.25.x |

That means CI builds against 1.25 while your local `devbox shell` uses 1.23. If a newer-Go-only feature is used, CI will be green but `task build` locally will fail. If that bites you, update `devbox.json` rather than downgrading `go.mod`.

### Optional quality-of-life

- **direnv** — drop an `.envrc` running `eval "$(devbox generate direnv --print-envrc)"` and `devbox shell` activates automatically on `cd`. The repo already ships an `.envrc` stub.
- **VS Code Devbox extension** — auto-activates the shell in the integrated terminal.

## Running locally

```sh
task                 # go run . (uses your real config)
task debug           # writes debug.log; tail with `task logs`
task install         # build & install as `gh dash` from local source
task install:prod    # reinstall the upstream release
```

`task debug` prints a banner of tildes so each new run is easy to find when tailing `debug.log`.

## Commit & PR conventions

- **Conventional Commits**. The release changelog in [`.goreleaser.yaml`](../.goreleaser.yaml) groups by prefix: `feat:`, `fix:`, `docs:`, `deps:`, everything else falls into "Other work". See [build-and-deployment.md](./build-and-deployment.md#changelog-grouping).
- Commits prefixed `test:` and `chore` are **excluded** from release notes entirely.
- Pre-push: run `task fmt && task lint && task test ./...`. CI enforces the latter two.
- Dependabot PRs auto-merge once CI passes (configured in [`build-and-test.yaml`](../.github/workflows/build-and-test.yaml)).

## Recipes

### How to add a new section

A "section family" = a new kind of row (beyond PRs / issues / notifications / repo). A "new section instance of an existing family" is just YAML and doesn't need code.

For a brand-new family:

1. **Copy an existing section package** as a template. [`prssection/`](../internal/tui/components/prssection/) is the most mature — duplicate it to `internal/tui/components/<name>section/`.
2. **Implement the full [`Section`](../internal/tui/components/section/section.go#L143) interface**. You embed [`section.BaseModel`](../internal/tui/components/section/section.go#L30) for free pagination/search/prompt/table behaviour, but the rest is yours:
   - `Identifier` — `GetId()`, `GetType()`
   - `Component` — `Update(msg tea.Msg) (Section, tea.Cmd)`, `View() string`
   - `Table` — `NumRows`, `GetCurrRow`, paging helpers, `BuildRows`, `ResetRows`, `GetIsLoading` / `SetIsLoading`
   - `Search` — `SetIsSearching`, `IsSearchFocused`, `ResetFilters`, `GetFilters`, `ResetPageInfo`
   - `PromptConfirmation` — confirmation-dialog plumbing
   - Plus the seven standalone methods (`GetConfig`, `UpdateProgramContext`, `MakeSectionCmd`, `GetPagerContent`, `GetItemSingularForm`, `GetItemPluralForm`, `GetTotalCount`).
3. **Add a row wrapper** under `internal/tui/components/<name>row/` that wraps the raw data type into a `data.RowData` with UI-specific fields.
4. **Extend the root model** in [`internal/tui/ui.go`](../internal/tui/ui.go#L48) — the root currently holds `prs []section.Section`, `issues []section.Section`, `notifications []section.Section`, and a single `repo section.Section`. Your new family needs its own field and routing in `Update`/`View`.
5. **Wire a YAML schema** in [`internal/config/parser.go`](../internal/config/parser.go). All section schemas live in this single file — copy the `PrsSectionConfig` / `IssuesSectionConfig` types and add a matching field to the top-level `Config` struct (`PRSections`, `IssuesSections`, etc.).
6. **Add a fetcher** under [`internal/data/`](../internal/data/) — use `shurcooL/githubv4` for typed GraphQL. Return the raw struct; the row wrapper lives in the TUI layer.
7. **Use the async pattern**. Never block `Update`. See below.

Why is this so involved? Because `Section` is deliberately wide — it's the one contract that routes keys, renders a table, handles search, and owns a confirmation prompt. Copying `prssection` gives you a working skeleton in minutes.

### How to use the async fetch pattern

Every fetch goes through `ctx.StartTask` + a `tea.Cmd`:

```go
func (m *MySection) fetch() tea.Cmd {
    taskId := fmt.Sprintf("fetch-%d", m.GetId())
    task := context.Task{
        Id:           taskId,
        FinishedText: "Loaded",
        State:        context.TaskStart,
        Error:        nil,
    }
    startCmd := m.Ctx.StartTask(task)   // returns the spinner tick; spinner now visible

    fetchCmd := func() tea.Msg {
        rows, err := data.FetchMyThings(...)   // may take seconds; runs in a goroutine
        if err != nil {
            return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
        }
        return constants.TaskFinishedMsg{
            TaskId:      taskId,
            SectionId:   m.GetId(),
            SectionType: m.GetType(),
            Msg:         section.SectionRowsFetchedMsg{SectionId: m.GetId(), Issues: rows},
        }
    }

    return tea.Batch(startCmd, m.MakeSectionCmd(fetchCmd))
}
```

Key rules:
- **Register before you fetch.** `ctx.StartTask` installs the task in `m.tasks` on the root model so the footer spinner shows. Forgetting it = silent fetch.
- **Wrap with `MakeSectionCmd`.** This envelopes the inner message in `section.SectionMsg` so the root can route it back to your section.
- **Return `TaskFinishedMsg`.** The root clears the task (hides spinner) when it arrives. Errors live on the same message.

### How to add a custom command

User-defined commands in YAML are templated with `text/template` (augmented with [sprout](https://github.com/go-sprout/sprout) functions) and shelled out via `exec.Command`. Two sharp edges:

1. **`RepoPath` resolution.** The template var `{{.RepoPath}}` is only populated if `gh-dash` was started from inside a git repo (via `git.GetRepoInPwd()` in [`cmd/root.go`](../cmd/root.go)). Commands that assume a repo path need to fall back gracefully.
2. **Quoting.** Arguments go through the shell — use the template helpers to quote when values may contain spaces.

See recent commits around commands / templating in `git log --grep "custom command"` for the pattern.

### How to add a key binding

1. Add a new entry to the appropriate `KeyMap` in [`internal/tui/keys/`](../internal/tui/keys/).
2. Add a user-overridable field to the `Keybindings` struct in [`internal/config/parser.go`](../internal/config/parser.go) (search for `type Keybindings struct`).
3. `keys.Rebind(...)` applies the user's overrides globally after config load — no per-component wiring needed.
4. Add a case to the relevant `Update` switch on the component that owns the binding.

### How to test

- **Unit tests** — standard Go. `task test ./path/to/pkg` or `task test:one -- -run TestName ./pkg`.
- **Golden files** — many TUI components compare rendered output against `testdata/*.golden` files. Regenerate with the update flag documented at the top of the test file (usually `-update`).
- **Config parsing** — fixtures under [`internal/config/testdata/`](../internal/config/testdata/) exercise every valid and invalid YAML shape. Add fixtures, don't invent inline YAML in tests.
- **Rendering without GitHub** — set `FF_MOCK_DATA=1` when running the binary to substitute canned fixtures.

## Debugging

- **Log from anywhere**:
  ```go
  import "charm.land/log/v2"
  log.Debug("merging PR", "id", prId, "base", base)
  ```
  With `task debug` running, `task logs` in a second pane shows output live.
- **Delve** (`task dlv`) starts a headless Delve server on `127.0.0.1:43000`. Attach your editor's debugger of choice.
- **CPU profile**: add `--cpuprofile=cpu.out`, reproduce the slow path, then `go tool pprof cpu.out`.
- **Weird Nerd Font glyphs** rendering as boxes? `task check-nerd-font`. Auto-fix with `task fix-nerd-font`.

## Tips for navigating the codebase

- Grep for `Section` usage in [`internal/tui/ui.go`](../internal/tui/ui.go) to see how the root wires sections in.
- Follow a message: pick any `...Msg` type in [`internal/tui/constants/`](../internal/tui/constants/) or a component package, then grep for its constructor to find the dispatch site and the handler.
- The `*context.ProgramContext` pointer is shared state — when a field "appears" in a deep component (e.g. `PreviewPosition`), grep where it's mutated (usually in the root `Update` or a resize handler).
