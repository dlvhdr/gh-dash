# Projects v2 Section — Design

**Date:** 2026-04-22
**Status:** Draft, ready for implementation
**Author:** Angel Cantu

## 1. Goals, Non-Goals, Scope

### Goal

Add a fourth top-level section type to gh-dash — alongside PRs, Issues, and
Notifications — that surfaces GitHub Projects v2. Users configure one or more
"project sections" via YAML. Each section shows a deduplicated list of
accessible projects and supports drilling into project items (issues, PRs,
draft items) with custom-field-aware columns.

### In scope (MVP)

- **Discovery:** hybrid. `viewer.projectsV2` is the zero-config default;
  YAML `owners` overrides to explicit users/orgs.
- **List view:** table with `Title | Owner | Status | # Items | # Open | Updated`.
- **Drill-down:** per-project item table with fixed base columns
  (`Title | Type | Repo | Status | Updated`) plus YAML-configured `extraFields`.
- **Mutations:** update an item's built-in `Status` field (one keybind, one
  GraphQL call, optimistic).
- **Sidebar reuse:** selecting an Issue or PR item reuses the existing
  `prView` / `issueView` sidebars.
- **Project-level filters:** `closed`, `titleContains`.
- **On-disk cache with TTL:** cold-start speed. Invalidated by `ctrl+r` and
  successful mutations.
- **Cross-view cache invalidation:** `ctrl+r` on the list invalidates cached
  items for any open drill-downs.
- **Session persistence:** remember cursor position per section and the last
  visited project; expose a `R` keybind to resume the last drill-down.

### Non-goals (deferred)

- Full custom-field editing (assignees, labels, iteration, date, number, text).
  Future SPEC.
- Creating, archiving, or deleting project items from the TUI.
- Respecting GitHub-saved views as an items filter source (`items.view:`).
- Projects classic. It is deprecated; no effort spent on it.
- Item-level YAML filters. Client-side `/` search covers the MVP need.

### Success criteria

1. The user can list Projects v2 from the terminal without leaving gh-dash.
2. The user can drill into a project, see items with triage-relevant fields,
   and move an item between Status values without a web round-trip.
3. With 10+ configured project sections, startup stays responsive because the
   disk cache answers first and the network refresh happens in the background.
4. Closing and reopening the app restores the last-visited section, cursor,
   and project, so multi-project triage resumes mid-flow.

## 2. Architecture & Package Layout

The feature threads the same three layers every other section uses —
**config → data → TUI section** — and adds two cross-cutting packages (cache,
state).

### New and modified packages

```
internal/
├── config/
│   └── parser.go                       (extend: ProjectsSectionConfig + cache/state config)
├── data/
│   ├── projects.go                     (NEW: FetchProjects, FetchProjectItems, UpdateItemStatus)
│   ├── projects_types.go               (NEW: ProjectData, ProjectItemData, field value types)
│   └── cache/                          (NEW: on-disk TTL cache shared by all data/* callers)
│       ├── cache.go
│       └── cache_test.go
├── state/                              (NEW: per-user session state on disk)
│   ├── state.go                        (last section, last cursor, last project)
│   └── state_test.go
└── tui/
    ├── components/
    │   ├── projectsection/             (NEW: list of projects — the "section" itself)
    │   │   ├── projectsection.go       (implements section.Section)
    │   │   ├── projectsection_test.go
    │   │   └── columns.go
    │   ├── projectitemsview/           (NEW: drill-down table for a single project's items)
    │   │   ├── projectitemsview.go
    │   │   ├── columns.go              (fixed base + dynamic extras builder)
    │   │   └── statuscmp.go            (floating autocomplete for Status, reuses cmp/)
    │   └── projectrow/                 (NEW: typed row data for the items table)
    │       └── projectrow.go
    └── ui.go                           (extend: register projectsection, route drill-down,
                                         restore state on startup)
```

### Section lifecycle

```
ui.Model
  ├── sections[]                            ← projectsection.Model appended here
  │     └── projectsection.Model
  │           └── drill: *projectitemsview.Model   ← non-nil when user pressed Enter
  ├── sidebar / prView / issueView          ← reused when an item is Issue or PR
  ├── dataCache: *cache.Store               ← shared across data/* functions
  └── userState: *state.Store               ← restored on boot, saved on Stop
```

### Why a nested drill-down rather than a new tab?

Project items are owned by the project. A tab would mean hoisting per-project
state into `ui.Model` and making "back" a tab navigation. A nested drill-down
(parent section + transient child view) is the TUI-native way to model
"looking *into* this thing". It also mirrors the `prompt/` overlay pattern
already in the codebase.

### Interface boundary

`projectsection.Model` satisfies the full `section.Section` interface — the
CLAUDE.md warning about "section interface is wide" is genuine, every method
needs a real or stub implementation. `projectitemsview.Model` is a plain
Bubble Tea sub-model, **not** a `Section`, because it is not user-configurable
and does not appear in the section tabs. Copying the `prview` boundary keeps
the child free of the wide interface it does not need.

## 3. Data Layer & GraphQL

### Public API (package `internal/data`)

```go
// Discovery — owner-scoped, deduped by node ID. Cache-aware.
func FetchProjects(owners []string, filters ProjectFilters) ([]ProjectData, error)

// Drill-down — items plus the field schema for column resolution. Cache-aware.
func FetchProjectItems(projectID string, limit int) (ProjectSchema, []ProjectItemData, error)

// Mutation — only the built-in Status single-select in MVP.
// On success, invalidates the project's cached items.
func UpdateItemStatus(projectID, itemID, statusOptionID string) error
```

### Types

```go
type ProjectData struct {
    ID, Number, Title, URL string
    Owner          string    // "org/login" for display
    Closed         bool
    Public         bool
    ItemsCount     int
    OpenItemsCount int       // computed client-side post-fetch
    UpdatedAt      time.Time
}

type ProjectSchema struct {
    StatusField *StatusFieldDef       // built-in single-select; nil if absent
    ExtraFields map[string]FieldDef   // filtered to section config's extraFields
}

type ProjectItemData struct {
    ID        string
    Type      ItemType       // Issue | PullRequest | DraftIssue
    Title     string
    Repo      string         // "owner/name"; "" for drafts
    URL       string         // "" for drafts
    Content   *ItemContent   // hydrated Issue/PR ref, used by the sidebar
    Fields    map[string]FieldValue
    UpdatedAt time.Time
}
```

### Discovery query

One GraphQL call per configured owner, plus `viewer.projectsV2` when `owners`
is empty. Results are deduped by `ProjectV2.id`.

### Items query shape

```graphql
query ($projectId: ID!, $first: Int!) {
  node(id: $projectId) {
    ... on ProjectV2 {
      fields(first: 50) { nodes { ...FieldSchema } }
      items(first: $first) {
        pageInfo { hasNextPage endCursor }
        nodes {
          id
          type
          content {
            ... on Issue       { ... }
            ... on PullRequest { ... }
            ... on DraftIssue  { title }
          }
          fieldValues(first: 20) { nodes { ...FieldValue } }
          updatedAt
        }
      }
    }
  }
}
```

The `fields` sub-query runs once per drill-down because mapping `fieldValues`
back to column cells requires the schema. `ProjectSchema` is cached per
project for the session.

### On-disk cache

Cache lives at `$XDG_CACHE_HOME/gh-dash/v1/` (`$HOME/.cache/gh-dash/v1/` on
Linux, `$HOME/Library/Caches/gh-dash/v1/` on macOS). File format is JSON; one
file per cache key. Keys:

| Key shape | Content | Default TTL |
| --- | --- | --- |
| `projects/<ownerHash>.json` | `[]ProjectData` per owner | 1h |
| `project-items/<projectID>.json` | `ProjectSchema + []ProjectItemData` | 5m |
| `project-schema/<projectID>.json` | `ProjectSchema` only (used as fallback) | 24h |

TTLs are overridable per section via YAML (see §6.1). The cache is
opportunistic: a fresh network response always wins and overwrites. The cache
is authoritative only while the network request is in flight, so the list
paints instantly and then reconciles.

### Cache invalidation rules

| Trigger | Invalidates |
| --- | --- |
| `ctrl+r` on project list view | `projects/*` for this section's owners + `project-items/*` for any currently-open drill-down |
| `ctrl+r` on drill-down | `project-items/<projectID>` for the open project |
| Successful status mutation | `project-items/<projectID>` for the containing project |
| Cache file TTL expiry | The expired entry only (lazy, on read) |

### Async pattern

Every call stays inside the existing **`TaskFinishedMsg` pattern**:
`ctx.StartTask(...)` before the goroutine, `tea.Cmd` returns a domain message
wrapped in `constants.TaskFinishedMsg`, the root `Update` calls
`ctx.ClearTask()` on return. No synchronous fetching in the `Update` loop —
`CLAUDE.md` is explicit about this.

### MVP limits

- Projects per owner: **100** initial fetch; paginate only when we hit the
  ceiling (rare at the MVP audience).
- Items per project: **250** initial fetch with a "Load more" affordance.
  Larger projects (2k+ items) stay responsive because we do not block on the
  tail.
- Cache size is bounded by file count, not bytes. No LRU in MVP — `ctrl+r`
  and TTL are the only eviction signals.

## 4. UI — List View & Items Drill-down

Two views, one parent. `projectsection.Model` owns the list and holds an
optional pointer to `projectitemsview.Model`, which becomes non-nil only
after the user drills in. The root `ui.Model` routes messages and renders
based on whether the child is active.

### 4.1 List view (`projectsection.Model`)

Fixed columns:

| Column | Source | Width |
| --- | --- | --- |
| `Title` | `ProjectData.Title` | flex |
| `Owner` | `ProjectData.Owner` | 20 |
| `Status` | `Closed ? "closed" : "open"` | 6 |
| `Items` | `ItemsCount` | 6 |
| `Open` | `OpenItemsCount` | 6 |
| `Updated` | relative (`3h ago`) | 10 |

Embeds `section.BaseModel`, so `/` search, pagination, and confirmation
prompts come for free. The code shape follows `prssection` line for line.

### 4.2 Drill-down (`projectitemsview.Model`)

Entered with `Enter` on a project row. Columns are computed once per
drill-down from `ProjectSchema` plus the section's `extraFields`:

```
Base (always):   Title | Type | Repo | Status | Updated
Extras (config): Priority | Iteration | ...
```

The column builder lives in `projectitemsview/columns.go`. Widths are
computed at build time from observed values, not hardcoded — the existing
`table.Model` already supports this shape.

Rendering rules per field type:

| Field kind | Cell render |
| --- | --- |
| `SingleSelect` (incl. Status) | option name, colored by option color if present |
| `Number` | right-aligned |
| `Date` | `YYYY-MM-DD` |
| `Iteration` | `title (MMM D → MMM D)` |
| `Text` | truncated to column width |
| any unsupported type | `—` (no crash, no hidden row) |

### 4.3 Item activation

When the user presses Enter on an item row:

- `Issue` → populate `ui.issueSidebar` with the hydrated `Content`.
- `PullRequest` → populate `ui.prView`.
- `DraftIssue` → no sidebar; `o` opens the project URL in the browser, since
  drafts have no standalone URL.

Reusing the existing sidebars is the biggest payoff of choosing full
drill-down — detail views are not rebuilt, they consume the same
`IssueData` / `PullRequestData` shapes they already accept.

### 4.4 Status mutation UX (MVP)

Keybind `s` on an item row opens a floating `cmp` autocomplete of the
project's Status options. Enter confirms. The row re-renders optimistically,
then `UpdateItemStatus` fires. On error, the previous status is restored
and the footer flashes a 2-second error toast. The autocomplete populator is
new; the autocomplete widget already exists in `internal/tui/components/cmp/`.

### 4.5 Navigation & keys

| Key | List view | Drill-down |
| --- | --- | --- |
| `Enter` | drill into project | activate item (open sidebar) |
| `Esc` / `b` | — | back to list |
| `o` | open project in browser | open item in browser |
| `s` | — | change Status |
| `ctrl+r` | refetch projects + invalidate items caches for any open drill-downs | refetch items |
| `/` | search projects | search items (client-side over `Title`) |
| `R` | resume last-visited drill-down (if one exists in state) | — |

## 5. State Machine — Async, Errors, Refresh

Bubble Tea's `Update` loop is synchronous, so every fetch and mutation
threads through `tea.Cmd` → goroutine → `TaskFinishedMsg`. Here is how each
user action maps to that pipeline.

### 5.1 Message flow

```
User action             tea.Cmd                        Goroutine does                    Returns
-----------------------------------------------------------------------------------------------------
open section            fetchProjectsCmd               cache.Get + FetchProjects         ProjectsFetchedMsg
Enter on project        fetchProjectItemsCmd           cache.Get + FetchProjectItems     ProjectItemsFetchedMsg
ctrl+r on list          fetchProjectsCmd (force)       invalidate + FetchProjects        ProjectsFetchedMsg
ctrl+r on drill-down    fetchProjectItemsCmd (force)   invalidate + FetchProjectItems    ProjectItemsFetchedMsg
s → option              updateStatusCmd (optimistic)   UpdateItemStatus                  StatusUpdatedMsg | StatusUpdateErrMsg
Load more               fetchProjectItemsCmd(after)    FetchProjectItems(+cursor)        ProjectItemsAppendMsg
R (resume)              (synchronous)                  restore drill from state          (immediate)
```

Every domain message is wrapped in `constants.TaskFinishedMsg` so the root
`Update` can `ctx.ClearTask(...)` uniformly. `ctx.StartTask(...)` is called
before the goroutine fires so the footer spinner is visible from t=0.

### 5.2 Cache-first fetch pattern

```
fetchProjectsCmd(owners, filters):
  1. entry, hit := cache.Get("projects/<ownerHash>")
  2. if hit && !forceRefresh && !expired(entry): return ProjectsFetchedMsg{data: entry, stale: false}
  3. if hit && !forceRefresh && expired(entry):  emit ProjectsFetchedMsg{data: entry, stale: true}  (paint stale first)
  4. go FetchProjects(...) → on success: cache.Put(...); emit ProjectsFetchedMsg{data: fresh, stale: false}
```

The stale-then-fresh pattern means the list paints in milliseconds on a warm
cache; the network request is background. `stale: true` can be surfaced as a
footer hint in a future iteration; the MVP just repaints on fresh arrival.

### 5.3 Optimistic status update

```go
// On 's' → option selected:
// 1. Capture previousStatus := item.Fields["Status"]
// 2. Mutate item.Fields["Status"] in place, repaint.
// 3. Return updateStatusCmd(...) → goroutine.
// 4. On StatusUpdatedMsg: invalidate cache for this project's items; no-op visually.
// 5. On StatusUpdateErrMsg: restore previousStatus; flash error in footer.
```

Optimistic because a round-trip to GitHub is 300–800ms, and waiting feels
sluggish during rapid triage.

### 5.4 Error surfaces

| Error | Where it shows | Recovery |
| --- | --- | --- |
| Discovery fetch failed (whole section) | Section body: `Failed to load projects — press ctrl+r` | Manual retry |
| One owner failed, others fine | Warning in footer with the owner's name; other owners still render | Per-owner isolation |
| Items fetch failed | Drill-down body with retry affordance | `ctrl+r` |
| Status mutation failed | 2s footer flash, optimistic revert | Manual retry |
| Project has no Status field | `s` is a no-op with a footer hint; drill-down still works | User edits in web UI |
| Cache file corrupt | Log at `warn`, delete file, refetch | Self-heals |
| State file corrupt | Log at `warn`, reset to defaults; do not crash | Self-heals |

No silent failures — each path logs via `log/slog` and surfaces something
user-visible.

### 5.5 Concurrency & cancellation

- **Stale responses:** each fetch tags its message with a monotonic `reqID`.
  `Update` drops messages whose `reqID` is older than the current in-flight
  one. Prevents flicker on rapid `ctrl+r` or quick project-switching.
- **No in-flight HTTP cancellation:** the existing `data/` helpers do not
  accept a `context.Context`. Matches PR/Issue section behavior; not
  worth expanding scope here.
- **Single in-flight per scope:** at most one list fetch and one items fetch
  per open drill-down. Pressing `ctrl+r` during a pending fetch is a no-op
  with a footer hint.

### 5.6 Session state persistence

State file: `$XDG_STATE_HOME/gh-dash/v1/state.json` (Linux) or
`$HOME/Library/Application Support/gh-dash/v1/state.json` (macOS).

```json
{
  "lastSectionID": "projects-my-work",
  "perSection": {
    "projects-my-work": {
      "cursor": 3,
      "lastProjectID": "PVT_abc123",
      "lastVisitedAt": "2026-04-22T10:12:31Z"
    }
  }
}
```

- On boot, `ui.NewModel` reads the file and restores the active section,
  cursor index, and "last project" hint.
- `R` on the list view drills into `lastProjectID` if it still exists in the
  current list; otherwise a footer hint says it's gone.
- Writes happen on cursor change (debounced 500ms) and on drill-in/out.
  On app exit, a final flush is attempted in a `defer` — best-effort only.

### 5.7 Empty states

| Situation | Render |
| --- | --- |
| Zero configured owners, zero viewer projects | `No projects. Configure owners in .gh-dash.yml` |
| Owner resolved but has zero projects | Section header with an empty table |
| Project has zero items | Drill-down header + `No items yet — o to open in browser` |

## 6. Config Schema, Testing, Rollout

### 6.1 YAML schema (MVP)

```yaml
# ~/.config/gh-dash/config.yml or repo-local .gh-dash.yml

projectsSections:
  - title: "My Work"
    owners: []                # empty => falls back to viewer.projectsV2
    filters:
      closed: false           # optional; default: any
      titleContains: ""       # optional; substring match
    extraFields:              # custom fields to surface as columns
      - Priority
      - Iteration
    limit: 100                # projects per owner; default 100
    cache:
      projectsTTL: 1h         # default 1h
      itemsTTL: 5m            # default 5m

  - title: "Org Roadmaps"
    owners: [my-org, platform-team]
    filters: { closed: false }
    extraFields: [Priority]

# Global cache and state toggles (optional)
cache:
  enabled: true               # default true; false disables on-disk cache entirely
  dir: ""                     # override default XDG path
state:
  enabled: true               # default true; controls session persistence
```

### Parsing

Extend `internal/config/parser.go`. The top-level `Config` struct gains
`ProjectsSections []ProjectsSectionConfig`, `Cache CacheConfig`, and
`State StateConfig`. Validation: unknown keys error loudly (koanf does not
silently drop — good). `TTL` fields use `time.ParseDuration`.

### 6.2 Testing strategy

| Layer | What we test | How |
| --- | --- | --- |
| `data/projects.go` | Query construction, response → `ProjectData`/`ProjectItemData` mapping, dedupe by node ID | Table-driven tests with canned GraphQL JSON under `internal/data/testdata/projects/` |
| `data/projects.go` mutation | Payload shape for `updateProjectV2ItemFieldValue` | Canned request assertion |
| `data/cache` | Put/Get round-trip, TTL expiry, invalidation, corrupt-file recovery | Pure Go, `t.TempDir()` |
| `state` | Read/write round-trip, missing-file initialization, corrupt-file recovery | Pure Go, `t.TempDir()` |
| `projectitemsview/columns.go` | Column builder for `(ProjectSchema, extraFields)` — base + extras, right order, missing-schema fallback | Pure-func unit tests |
| `projectsection` / `projectitemsview` | Message routing, `reqID` staleness drop, optimistic revert on error, cache invalidation on mutation | `internal/tui/testutils/` + golden files |
| Config | YAML → struct, error on unknown keys, viewer-fallback, TTL parsing | Fixtures under `internal/config/testdata/projects/` |

Goal: every new public function in `data/`, `cache/`, `state/`, and the
column builder has direct tests. TUI tests cover state transitions, not
pixel layout.

### 6.3 Rollout plan (commit sequence)

Each row is one PR; ordered so the tree compiles and tests pass at every
commit. Conventional Commits — `.goreleaser.yaml` groups release notes by
prefix.

| # | PR | Commit prefix | Depends on |
| --- | --- | --- | --- |
| 1 | Config schema (`projectsSections`, `cache`, `state`) + parser + fixtures | `feat(config):` | — |
| 2 | `internal/data/cache` TTL store + tests | `feat(data):` | 1 |
| 3 | `internal/state` session store + tests | `feat(state):` | 1 |
| 4 | Data layer: `FetchProjects`, types, dedupe, cache integration, tests | `feat(data):` | 2 |
| 5 | Data layer: `FetchProjectItems`, `ProjectSchema`, cache integration, tests | `feat(data):` | 4 |
| 6 | TUI: `projectsection.Model` list view (no drill-down yet), state-driven cursor restore | `feat(tui):` | 3, 4 |
| 7 | TUI: `projectitemsview.Model` drill-down, base columns, state-driven last-project + `R` resume | `feat(tui):` | 5, 6 |
| 8 | TUI: extra fields from YAML flow into drill-down columns | `feat(tui):` | 7 |
| 9 | TUI: cross-view invalidation on `ctrl+r` from list | `feat(tui):` | 7 |
| 10 | Mutation: `UpdateItemStatus` + `s` keybind + optimistic revert + cache invalidate | `feat(tui):` | 5, 7 |
| 11 | Docs: new section type in `docs/` | `docs:` | all |

Each PR is shippable in isolation behind `projectsSections`. Absence of the
config key means zero behavior change, so the feature is opt-in without a
feature flag.

### 6.4 Observability

- Each GraphQL call logs at `debug` with `owner`, `project_id`,
  `items_fetched`, `elapsed_ms`, and `cache_hit` (bool) — consistent with
  existing `FetchPullRequests` logging.
- Cache invalidation events log at `debug` with trigger reason.
- Mutation failures log at `warn` with the error class (auth, not-found,
  validation).
- State read/write failures log at `warn` but never raise (self-healing).
- No metrics emission. gh-dash is a CLI, not a daemon.

### 6.5 Planned follow-ups (post-MVP)

These are confirmed directions, not ideas in limbo:

1. Full custom-field editing (assignees, labels, iteration, date, number,
   text). Its own SPEC, because it needs a type-aware field-editor widget.
2. `items.view:` support — respect a GitHub-saved project view as the
   server-side filter and sort.
3. Cache LRU eviction once long-running users start accumulating entries.
4. Stale-banner affordance in the list header ("cached 3m ago — fetching…").
5. Per-repo invalidation hooks when notifications suggest project changes.

## 7. Decision Log

| # | Decision | Rationale |
| --- | --- | --- |
| 1 | Projects v2 only (not classic) | Classic is deprecated; investing in it is sunk cost. |
| 2 | Hybrid discovery (viewer + YAML owners) | Zero-config happy path, explicit override for multi-org users. |
| 3 | Full drill-down into items | Triage-from-terminal is the value prop; list-only is half the feature. |
| 4 | Hybrid columns (fixed base + YAML `extraFields`) | Predictable layout across projects, still surfaces custom workflows. |
| 5 | MVP: Status mutation only; full edits deferred | Status is the one field touched constantly; rest needs a type-aware editor (future SPEC). |
| 6 | Project-level filters only in MVP | Item-level filters evolve into a mini query language — do not invent prematurely. |
| 7 | On-disk cache in MVP | Cold-start speed matters with 10+ sections; retrofitting a cache is more painful than building it upfront. |
| 8 | Session state persistence in MVP | The feature's point is resumable triage across sessions. Without state, cold-starts lose context. |
