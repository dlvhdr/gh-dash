# Projects v2 Section — Design

**Date:** 2026-04-22
**Status:** Draft, ready for implementation (revised after code-grounded review — see §8)
**Author:** Angel Cantu

## 1. Goals, Non-Goals, Scope

### Goal

Add a new top-level **view type** (`ProjectsView`) to gh-dash — alongside the
existing `prs`, `issues`, `notifications`, and `repo` views — that surfaces
GitHub Projects v2. Users configure one or more "project sections" via YAML.
Each section shows a deduplicated list of accessible projects and supports
drilling into project items (issues, PRs, draft items) with custom-field-aware
columns.

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
- **On-disk cache with TTL:** cold-start speed. Invalidated by `r` (refresh)
  and successful mutations.
- **Cross-view cache invalidation:** `r` on the list invalidates cached items
  for any open drill-downs.
- **Session persistence:** remember cursor position per section and the last
  visited project (cursor-only restore; no auto-drill). Auto-drill resume is
  deferred — §6.5.

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
**config → data → TUI section** — and adds two cross-cutting packages
(`persistcache`, `state`) plus changes to the view-routing surface in
`internal/tui/ui.go` and `internal/tui/keys/`.

### ViewType integration surface (the hidden cost)

The existing code models top-level navigation as an enum `ViewType`
(`parser.go:48-81`) with values `prs | issues | notifications | repo`. Adding
a new view is not just a new section type — it requires touching every
routing call site:

| Touch-point | File:line | What changes |
| --- | --- | --- |
| Enum value | `internal/config/parser.go:77-81` | Add `ProjectsView ViewType = "projects"` |
| Enum unmarshal | `internal/config/parser.go:58` | Reject unknown values (today's impl silently accepts) |
| Default-view decision | `internal/config/parser.go` (view resolution) | Document fallback when no projects section configured |
| Config merge | `internal/config/parser.go:~637` | Ensure `projectsSections`, `cache`, `state` survive the manual merge |
| Section registry | `internal/tui/ui.go:48` (Model fields) | Add `projects []section.Section` slice |
| Fetch-all loop | `internal/tui/ui.go` `fetchAllViewSections` | Include projects in fan-out |
| View getter | `internal/tui/ui.go:1476` `getCurrentViewSections` | Add case |
| View setter | `internal/tui/ui.go:1508` `setCurrentViewSections` | Add case; stop the silent fall-through to issues |
| View switcher | `internal/tui/ui.go:1583` `switchSelectedView` | Add projects to rotation |
| Keymap | `internal/tui/keys/keys.go` `CreateKeyMapForView` | Projects-specific keymap |
| Keybinding rebind | `cmd/root.go` | Include projects keymap in post-load rebind |
| Tab renderer | `internal/tui/components/tabs/` | Register the new view |

This fan-out is **PR #0** in the revised rollout (§6.3) — scaffolding the view
without any data behind it, so every subsequent PR lands into a routed view.

### New and modified packages

```
internal/
├── config/
│   └── parser.go                       (extend: ViewType, ProjectsSectionConfig,
│                                        CacheConfig, StateConfig; enforce strict unknown-key check)
├── data/
│   ├── projects.go                     (NEW: FetchProjects, FetchProjectItems,
│   │                                    UpdateItemStatus, FetchIssue — follows prapi.go shape,
│   │                                    honors FF_MOCK_DATA feature flag)
│   └── projects_types.go               (NEW: ProjectData, ProjectItemData, field value types)
├── persistcache/                       (NEW — was "data/cache/"; renamed to avoid collision
│   │                                    with the existing in-memory otter cache at
│   │                                    internal/data/cache.go)
│   ├── persistcache.go                 (TTL on-disk store; injected into data/projects.go)
│   └── persistcache_test.go
├── state/                              (NEW: per-user session state on disk)
│   ├── state.go                        (last section, last cursor, last project ID)
│   └── state_test.go
└── tui/
    ├── components/
    │   ├── projectsection/             (NEW: list of projects — the "section" itself)
    │   │   ├── projectsection.go       (implements section.Section)
    │   │   ├── projectsection_test.go
    │   │   └── columns.go
    │   ├── projectitemsview/           (NEW: drill-down table for a single project's items)
    │   │   ├── projectitemsview.go
    │   │   ├── columns.go              (fixed base + dynamic extras builder; resolves extras
    │   │   │                            by field ID after name lookup)
    │   │   └── statuspicker.go         (selection-mode wrapper over cmp.Model; NOT the
    │   │                                text-editing cmpcontroller)
    │   └── projectrow/                 (NEW: typed row data for the items table)
    │       └── projectrow.go
    ├── keys/
    │   └── projectKeys.go              (NEW: projects-specific keymap)
    └── ui.go                           (extend: register projectsection, route drill-down,
                                         restore state on startup)
```

### Cache / state / config path resolution

All three use a single centralized helper (new: `internal/xdgpath`) built on
Go's `os.UserConfigDir`, `os.UserCacheDir`, and `os.UserHomeDir`. This
normalizes cross-platform paths (Linux XDG, macOS `~/Library/...`, Windows
`%LocalAppData%`) and keeps existing `XDG_CONFIG_HOME` overrides working
(`parser.go:579` already honors it — we preserve that behavior).

### Section lifecycle

```
ui.Model
  ├── projects[]                            ← projectsection.Model appended here
  │     └── projectsection.Model
  │           └── drill: *projectitemsview.Model   ← non-nil when user pressed Enter
  ├── sidebar / prView / issueView          ← reused when an item is Issue or PR
  │                                            (requires hydration — see §4.3)
  ├── persistCache: *persistcache.Store     ← injected into data/projects.go fetchers
  └── userState: *state.Store               ← restored on boot, saved on Stop
```

The `persistCache` is explicitly injected into fetcher calls (§3 API) rather
than being a hidden package-global. This matches the package-level GraphQL
client pattern `prapi.go:493` uses and makes tests easy to wire with a
`t.TempDir()` backing store.

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
// Discovery — owner-scoped, deduped by node ID. Cache-aware via injected store.
func FetchProjects(
    cache *persistcache.Store,
    owners []OwnerRef,                // typed: {Kind: OwnerOrg|OwnerUser, Login string}
    filters ProjectFilters,
) ([]ProjectData, error)

// Drill-down — items plus field schema. Supports pagination.
// `after` empty ⇒ first page; otherwise resumes from cursor.
func FetchProjectItems(
    cache *persistcache.Store,
    projectID, after string,
    limit int,
) (ProjectSchema, []ProjectItemData, PageInfo, error)

// Mutation — built-in Status single-select in MVP.
// Non-nil optionID ⇒ updateProjectV2ItemFieldValue.
// Nil optionID ⇒ clearProjectV2ItemFieldValue.
// On success, invalidates project-items/<projectID> in the cache.
func UpdateItemStatus(
    cache *persistcache.Store,
    projectID, itemID, statusFieldID string,
    optionID *string,
) error

// Hydration — fetches full IssueData by URL for sidebar reuse when a project
// item is an Issue. Follows the existing FetchPullRequest(url) pattern.
// PR items don't need this at the call site because prview calls
// data.FetchPullRequest internally (see prview.go:576).
func FetchIssue(issueURL string) (EnrichedIssueData, error)
```

All four honor the existing `FF_MOCK_DATA` feature flag (`prapi.go:496`) so
tests can run without network access, consistent with PR-side fetch paths.

### Types

```go
type OwnerKind int
const (
    OwnerOrg OwnerKind = iota
    OwnerUser
)

type OwnerRef struct {
    Kind  OwnerKind
    Login string
}

type ProjectData struct {
    ID, Number, Title, URL string
    Owner               OwnerRef  // typed; Login for display, Kind drives GraphQL route
    Closed              bool
    Public              bool
    ItemsCount          int       // from items.totalCount — accurate regardless of fetch limit
    OpenItemsCountLoaded int      // computed from currently-fetched page; "Loaded" suffix
                                  // makes the approximation explicit (renamed from
                                  // OpenItemsCount — see §7 decision 9)
    UpdatedAt           time.Time
}

type ProjectSchema struct {
    // StatusField is the built-in single-select used for mutation.
    // nil if the project has no field named "Status" (MVP only wires this one).
    StatusField *StatusFieldDef

    // ExtraFields are resolved by YAML name → field ID once at schema fetch time.
    // Keyed by field ID (NOT name) — same name across projects may map to
    // different field IDs with different types. Duplicates within one project
    // log a warn and first match wins.
    ExtraFields map[string]FieldDef

    // ExtraFieldOrder preserves the YAML-declared order for column rendering.
    ExtraFieldOrder []string   // field IDs in config order
}

type ProjectItemData struct {
    ID        string
    Type      ItemType       // Issue | PullRequest | DraftIssue
    Title     string
    Repo      string         // "owner/name"; "" for drafts
    URL       string         // "" for drafts — drafts have no standalone URL
    Fields    map[string]FieldValue  // keyed by field ID
    UpdatedAt time.Time
}

type PageInfo struct {
    HasNextPage bool
    EndCursor   string
}
```

### Why extra fields are keyed by ID, not name

If two projects both have a field named "Priority" but one is a
single-select and the other is a number field, name-keyed rendering would
silently render the wrong cell. Resolving name → ID once per drill-down
(at schema fetch time) makes the mapping explicit and type-safe at render
time. The YAML stays ergonomic (users type `extraFields: [Priority]`); the
resolution step is per-project internal bookkeeping.

### Discovery query

One GraphQL call per configured `OwnerRef`, plus `viewer.projectsV2` when
`owners` is empty. The query branches on `Kind`:

- `OwnerOrg` → `organization(login: $login) { projectsV2(first: $limit) { ... } }`
- `OwnerUser` → `user(login: $login) { projectsV2(first: $limit) { ... } }`

Results across owners are deduped by `ProjectV2.id`.

### Items query shape

```graphql
query ($projectId: ID!, $first: Int!, $after: String) {
  node(id: $projectId) {
    ... on ProjectV2 {
      fields(first: 50) {
        nodes { ...FieldSchema }
        pageInfo { hasNextPage endCursor }
      }
      items(first: $first, after: $after) {
        totalCount              # authoritative ItemsCount, independent of page size
        pageInfo { hasNextPage endCursor }
        nodes {
          id
          type
          content {
            ... on Issue       { ... }
            ... on PullRequest { ... }
            ... on DraftIssue  { title }
          }
          fieldValues(first: 20) {
            nodes { ...FieldValue }
            pageInfo { hasNextPage endCursor }
          }
          updatedAt
        }
      }
    }
  }
}
```

The `fields` sub-query runs once per drill-down because mapping `fieldValues`
back to column cells requires the schema. `ProjectSchema` is cached in the
`project-schema/<projectID>` entry (24h TTL).

### GraphQL ceiling handling

Three nested collections have hard `first:` limits. Each needs an explicit
truncation policy:

| Collection | Limit | Policy if truncated |
| --- | --- | --- |
| `fields(first:50)` | 50 | If `pageInfo.hasNextPage`, log `warn` with project ID; render with what we got. Re-paginate in a follow-up fetch before MVP ships if any real project trips this. |
| `items(first:$first,after:$after)` | 250 default | Cursor-paginated via `after`. "Load more" affordance on the drill-down surfaces additional pages. |
| `fieldValues(first:20)` per item | 20 | Same warn-log policy. In practice projects rarely have >20 fields with a non-null value per item. |

`ItemsCount` uses `items.totalCount` directly (accurate even if the current
page is truncated). `OpenItemsCountLoaded` is only computed over fetched
items — the field name makes the approximation explicit. A future iteration
can extend it to a server-side count via a separate filtered query.

### On-disk cache (`internal/persistcache`)

The cache lives under `os.UserCacheDir() + "/gh-dash/v1/"`, resolved via the
shared `internal/xdgpath` helper:

- Linux: `$XDG_CACHE_HOME/gh-dash/v1/` or `$HOME/.cache/gh-dash/v1/`
- macOS: `$HOME/Library/Caches/gh-dash/v1/`
- Windows: `%LocalAppData%/gh-dash/v1/`

File format is JSON; one file per cache key. Keys:

| Key shape | Content | Default TTL |
| --- | --- | --- |
| `projects/<ownerHash>.json` | `[]ProjectData` per owner | 1h |
| `project-items/<projectID>/<pageCursor>.json` | `[]ProjectItemData` for that page | 5m |
| `project-items/<projectID>/_meta.json` | `ProjectSchema + PageInfo chain` | 5m |
| `project-schema/<projectID>.json` | `ProjectSchema` only (fallback when items cache evicted) | 24h |

Paginated items cache layout: each page is stored under its cursor. The
`_meta.json` sidecar tracks which cursor chain belongs to "first page" vs
appended pages, so a "Load more" hit can stitch pages deterministically.

TTLs are overridable per section via YAML (see §6.1). The cache is
opportunistic: a fresh network response always wins and overwrites. It is
authoritative only while the network request is in flight, so the list
paints instantly and then reconciles.

**Naming rationale:** the package is called `persistcache` (not `cache`) to
avoid colliding with the pre-existing in-memory cache at
`internal/data/cache.go`. See §7 decision 10 for the rationale.

### Cache invalidation rules

| Trigger | Invalidates |
| --- | --- |
| `r` on project list view | `projects/*` for this section's owners + `project-items/<projectID>/*` for any currently-open drill-down |
| `r` on drill-down | `project-items/<projectID>/*` for the open project (all pages + meta) |
| Successful status mutation | `project-items/<projectID>/*` for the containing project |
| Cache file TTL expiry | The expired entry only (lazy, on read) |
| Corrupt cache file | Delete and refetch; log `warn` (self-healing) |

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
| `Owner` | `ProjectData.Owner.Login` (kind-prefixed: `org:acme`, `user:angel`) | 20 |
| `Status` | `Closed ? "closed" : "open"` | 6 |
| `Items` | `ItemsCount` (from `items.totalCount`) | 6 |
| `Open` | `OpenItemsCountLoaded` (marked `~` when paginated) | 6 |
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

### 4.3 Item activation (sidebar hydration)

When the user presses Enter on an item row, we **hydrate** full data for the
sidebar from its URL — project item GraphQL fragments are a subset of what
the existing sidebars need, so feeding them raw would leave fields blank or
trigger broken follow-up fetches.

| Item type | Activation path |
| --- | --- |
| `Issue` | `data.FetchIssue(url)` → populate `ui.issueSidebar` with full `EnrichedIssueData`. `FetchIssue` is a new helper mirroring `FetchPullRequest(url)` at `prapi.go:550`. |
| `PullRequest` | Pass URL to `prView`; `prview.go:576` already calls `data.FetchPullRequest(url)` internally — no new hydration call site needed. |
| `DraftIssue` | **No sidebar.** Drafts have no standalone URL, number, or repo. `o` on a draft opens the **project** URL in the browser. The item row renders with a `(draft)` badge in the `Type` column. |

Reusing the sidebars is still the right call — it avoids building a third
detail view — but the SPEC was wrong to call it "feeding the same shapes".
A round-trip for full hydration is required and is async (wrapped in the
usual `TaskFinishedMsg` pattern).

### 4.4 Status mutation UX (MVP)

Keybind `S` (shift-s) on an item row opens a **selection-mode** picker
populated with the project's Status options. Enter confirms, Esc cancels.
The row re-renders optimistically, then `UpdateItemStatus(projectID, itemID,
statusFieldID, optionID)` fires. On error, the previous status is restored
and the footer flashes a 2-second error toast.

**Why `S` and not `s`:** lowercase `s` already switches between PRs and
Issues views in `prKeys.go:108` / `issueKeys.go:58`. Adding a project-view
override is fine in isolation but invites muscle-memory collisions — shifted
variant keeps the two semantically distinct.

**Why a new picker and not `cmpcontroller`:** `cmp.Model` (generic floating
widget) is currently wrapped by `cmpcontroller` for *text editing* — the
controller drives a text input and filters completions as the user types.
Status mutation is *selection*, not editing. We add a thin wrapper
`statuspicker.go` that reuses `cmp.Model` directly but in "select-only" mode:
no text input, arrow keys navigate, Enter confirms.

### 4.5 Navigation & keys

All bindings verified non-colliding against `internal/tui/keys/keys.go`,
`prKeys.go`, `issueKeys.go`.

| Key | List view | Drill-down | Notes |
| --- | --- | --- | --- |
| `Enter` | drill into project | activate item (hydrate + open sidebar) | |
| `Esc` / `b` | — | back to list | `b` matches existing sidebar back convention |
| `o` | open project in browser | open item in browser (or project URL for drafts) | `keys.go:165` existing |
| `S` | — | change Status (shift-s) | Avoids `s`=switch-view collision |
| `r` | refetch projects + invalidate items caches for any open drill-downs | refetch items | Uses existing refresh binding at `keys.go:169` |
| `R` | refresh all views | — | Existing binding at `keys.go:173`, unchanged |
| `/` | search projects | search items (client-side over `Title`) | |
| `L` | — | load more items (next page) | Shift-l; no collision |

Resume-last-drill-down is deferred to §6.5 — cursor restore is what actually
ships in MVP.

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
r on list               fetchProjectsCmd (force)       invalidate + FetchProjects        ProjectsFetchedMsg
r on drill-down         fetchProjectItemsCmd (force)   invalidate + FetchProjectItems    ProjectItemsFetchedMsg
S → option              updateStatusCmd (optimistic)   UpdateItemStatus                  StatusUpdatedMsg | StatusUpdateErrMsg
L (load more)           fetchProjectItemsCmd(after)    FetchProjectItems(+cursor)        ProjectItemsAppendMsg
Enter on Issue item     hydrateIssueCmd                FetchIssue(url)                   IssueHydratedMsg → opens sidebar
Enter on PR item        (none — prview hydrates)       FetchPullRequest(url) in prview   existing flow
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

- **Stale responses:** reuse the existing `LastFetchTaskId` pattern
  (`prssection.go:464`). Each fetch tags its returned message with the
  section's current task ID; `Update` drops messages whose task ID doesn't
  match the latest. Prevents flicker on rapid `r` or quick project-switching.
  **Do not invent a parallel `reqID` concept** — the pattern is already
  codified and tested.
- **No in-flight HTTP cancellation:** the existing `data/` helpers do not
  accept a `context.Context`. Matches PR/Issue section behavior; not
  worth expanding scope here.
- **Single in-flight per scope:** at most one list fetch and one items fetch
  per open drill-down. Pressing `r` during a pending fetch is a no-op
  with a footer hint.

### 5.6 Session state persistence

State file path is resolved via `internal/xdgpath` using Go's
`os.UserConfigDir`:

- Linux: `$XDG_CONFIG_HOME/gh-dash/state/v1/state.json` (or
  `$HOME/.config/gh-dash/state/v1/state.json`)
- macOS: `$HOME/Library/Application Support/gh-dash/state/v1/state.json`
- Windows: `%AppData%/gh-dash/state/v1/state.json`

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

- On boot, `ui.NewModel` reads the file and restores the active section and
  cursor index. `lastProjectID` is stored but **not used to auto-drill in
  MVP** — it seeds the §6.5 follow-up item ("resume last drill") without
  building it now.
- Writes happen on cursor change (debounced 500ms) and on drill-in/out.
  On app exit, a final flush is attempted in a `defer` — best-effort only.
- Corrupt state file: log `warn`, reset to defaults, never crash
  (self-healing).

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
    owners:                   # prefix-shorthand form to disambiguate
      - org:my-org            #   user vs organization (GitHub GraphQL
      - org:platform-team     #   uses separate query roots for each)
      - user:angelcantugr
    filters: { closed: false }
    extraFields: [Priority]

# Global cache and state toggles (optional)
cache:
  enabled: true               # default true; false disables on-disk cache entirely
  dir: ""                     # override default OS cache dir
state:
  enabled: true               # default true; controls session persistence
```

**Owner syntax.** Each entry in `owners` uses `<kind>:<login>` where `kind`
is one of `org` or `user`. Bare logins (`my-org` without prefix) are
rejected at parse time with a clear error — disambiguation is mandatory
because `organization(login:)` and `user(login:)` are different GraphQL
query roots and a guess-and-fallback strategy masks config errors as
404s. This is a small ergonomic cost that pays for itself the first time
someone mistypes a login.

### Parsing

Extend `internal/config/parser.go`:

1. Top-level `Config` struct gains `ProjectsSections []ProjectsSectionConfig`,
   `Cache CacheConfig`, `State StateConfig`.
2. **Explicit strict validation** is added (the existing `StrictMerge: true`
   at `parser.go:31` does not enforce unknown-key rejection — the earlier
   claim that "koanf does not silently drop" was wrong). Add a
   `validateKnownKeys(rawMap, allowedKeys)` pass before `UnmarshalWithConf`
   that errors on unknown top-level keys and unknown `projectsSections[*]`
   fields.
3. `ViewType.UnmarshalJSON` at `parser.go:58` is extended to reject unknown
   values (today's implementation silently accepts anything). Add
   `ProjectsView ViewType = "projects"` to the allowlist.
4. Manual merge at `parser.go:~637` must include `projectsSections`, `cache`,
   `state`, and new keybindings. Today's merge preserves only PR/Issues/
   Notifications sections; naive addition would drop the new keys.
5. `TTL` fields use `time.ParseDuration`.
6. Owner strings are parsed into `OwnerRef` at load time; malformed entries
   error with the entry's location.

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
| 0 | Core: add `ProjectsView` to `ViewType`, wire into `ui.go` routing (`getCurrentViewSections`, `setCurrentViewSections`, `switchSelectedView`, `fetchAllViewSections`), `keys.CreateKeyMapForView`, tabs renderer. Also fix `ViewType.UnmarshalJSON` to reject unknown values. No data yet. | `feat(core):` | — |
| 1 | `internal/xdgpath` centralized path helper + tests | `feat(core):` | — |
| 2 | Config schema (`projectsSections`, `cache`, `state`) + parser + strict unknown-key validation + merge updates + fixtures | `feat(config):` | 0, 1 |
| 3 | `internal/persistcache` TTL store + tests (pure-Go, `t.TempDir()`) | `feat(persistcache):` | 1 |
| 4 | `internal/state` session store + tests | `feat(state):` | 1 |
| 5 | Data layer: `FetchProjects`, types (`OwnerRef`, `ProjectData`), owner-routing, dedupe, persistcache integration, tests | `feat(data):` | 2, 3 |
| 6 | Data layer: `FetchProjectItems` (cursor-paginated), `ProjectSchema` with field-ID resolution, persistcache integration, tests | `feat(data):` | 5 |
| 7 | Data layer: `FetchIssue(url)` helper for sidebar hydration, tests | `feat(data):` | 5 |
| 8 | TUI: `projectsection.Model` list view (no drill-down), state-driven cursor restore, `projectKeys.go` keymap | `feat(tui):` | 0, 4, 5 |
| 9 | TUI: `projectitemsview.Model` drill-down, base columns, hydration-on-activate for Issues/PRs, draft no-sidebar behavior, `L` load-more | `feat(tui):` | 6, 7, 8 |
| 10 | TUI: extra fields (YAML → field-ID → dynamic columns) | `feat(tui):` | 9 |
| 11 | TUI: cross-view invalidation on `r` from list | `feat(tui):` | 9 |
| 12 | Mutation: `UpdateItemStatus` + `statuspicker.go` selection-mode wrapper + `S` keybind + optimistic revert + cache invalidate | `feat(tui):` | 6, 9 |
| 13 | Docs: new view type + section type in `docs/` | `docs:` | all |

Each PR is shippable in isolation behind `projectsSections`. Absence of the
config key means zero behavior change, so the feature is opt-in without a
feature flag. **PR #0 is explicitly empty of data** — it only scaffolds the
view so every subsequent PR has a routed home.

### 6.4 Observability

- Each GraphQL call logs at `debug` with `owner_kind`, `owner_login`,
  `project_id`, `items_fetched`, `after_cursor`, `elapsed_ms`, and
  `cache_hit` (bool) — consistent with existing `FetchPullRequests` logging
  plus the new project-specific keys.
- Field pagination truncation logs at `warn` with `project_id` and which
  collection truncated (`fields`, `fieldValues`).
- Cache invalidation events log at `debug` with trigger reason and key pattern.
- Mutations log at `debug` on success (`field_id`, `item_id`, `new_value`)
  and `warn` on failure (error class: auth, not-found, validation).
- State read/write failures log at `warn` but never raise (self-healing).
- Owner-parse errors log the exact YAML path and expected format.
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
6. **Resume-last-drill-down** via a keybind — state is already persisted in
   §5.6, only the UX needs adding (pick a non-colliding key, handle
   project-gone case).
7. Rate-limit awareness on rapid `r` — back off or batch if GitHub GraphQL
   point budget is getting hot.
8. Accurate `OpenItemsCount` via a separate server-side filtered `items`
   query, replacing the `Loaded` approximation.

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
| 8 | Session state persistence in MVP (cursor-only; resume-drill deferred) | Cursor restore is the first-order need; auto-drill is second-order and needs more UX thought. |
| 9 | Rename `ItemsCount`/`OpenItemsCount` to surface fetch-limit truth | `ItemsCount` uses `items.totalCount` (exact); `OpenItemsCountLoaded` acknowledges it's an approximation over fetched items. Honest naming beats silent inaccuracy. |
| 10 | New on-disk cache package is `persistcache`, not `data/cache` | `internal/data/cache.go` already exists (otter in-memory cache). Collision-free naming now avoids confusion forever. |
| 11 | Owners use `<kind>:<login>` shorthand, not bare logins | `organization(login:)` and `user(login:)` are separate GraphQL roots. Guess-and-fallback masks config typos as 404s; explicit prefix fails loud. |
| 12 | Extra fields keyed by field ID, not name | Same field name across projects can map to different IDs and types. Name-keyed rendering would silently render the wrong cell. |
| 13 | Status mutation on `S` (shift-s) not `s` | Lowercase `s` already switches between PRs and Issues. Shifted variant avoids muscle-memory collision while staying mnemonic. |
| 14 | Refresh on `r` (existing), not `ctrl+r` | Reuse codified binding at `keys.go:169`; no reason to invent a parallel. |
| 15 | Reuse `LastFetchTaskId` pattern for stale-response drop | Already in `prssection.go:464`; a second mechanism (`reqID`) would fragment the pattern. |
| 16 | Selection-mode `statuspicker.go` instead of extending `cmpcontroller` | `cmpcontroller` is a text-editing wrapper; Status picking is selection-only. Adapting the generic widget is cleaner than repurposing the text wrapper. |

## 8. Revision History

### 2026-04-22 — v1 → v2 (after code-grounded second-opinion review)

A Codex ReAct review surfaced several claims in v1 that were false when
checked against actual code. The SPEC was revised in-place with the
following substantive changes:

**Corrected mental-model errors:**

- v1 called this the "fourth" section type — actually fifth (`ViewType` already
  has `prs/issues/notifications/repo`). Integration surface expanded from "a
  new section" to "a new view type with ~10 routing touch-points"
  (`ui.go:48,1476,1508,1583` + keys + tabs + config merge).
- v1 claimed koanf "does not silently drop" unknown keys. It does. Explicit
  `validateKnownKeys` pass added (§6.1).
- v1 said sidebar reuse "consumes the same shapes" — false; `prview.go:576`
  hydrates via `FetchPullRequest(url)`. §4.3 rewritten to describe hydration
  explicitly; drafts excluded from sidebar entirely.
- v1's `UpdateItemStatus(projectID, itemID, optionID)` was unimplementable —
  GitHub's mutation requires `fieldID` too, and clearing uses a different
  mutation. §3 signature fixed; §7 decision 5 clarified.

**Renamed to avoid collisions:**

- Proposed `internal/data/cache/` package collided with the existing
  `internal/data/cache.go` (otter in-memory cache). Renamed to
  `internal/persistcache`.
- Keybindings `ctrl+r`, `s`, `R` all collided with existing bindings
  (`keys.go:165-174`, `prKeys.go:108`, `issueKeys.go:58`). Rebound per §4.5
  decision log.

**Added missing pieces:**

- Owner disambiguation with `<kind>:<login>` shorthand (§6.1, decision 11).
- Cursor-paginated `FetchProjectItems` with `PageInfo` return and
  pagination-aware cache layout (§3, §3.On-disk cache).
- Field-ID resolution for `extraFields` with warn-on-duplicate behavior
  (§3 types, decision 12).
- `FetchIssue(url)` hydration helper (§3 API, new PR #7).
- Cross-platform path resolution via `internal/xdgpath` using
  `os.UserCacheDir`/`os.UserConfigDir` (§2, §5.6).
- GraphQL ceiling handling with explicit truncation logs (§3).
- `ProjectsView` plumbing elevated to PR #0 so later PRs have a routed home
  (§2, §6.3).

**Deferred out of MVP:**

- `R`-keybind resume into the last drill-down — state is persisted,
  UX isn't wired (§5.6, §6.5 #6).

**Deliberately kept:**

- Reuse of `LastFetchTaskId` pattern rather than inventing `reqID` (§5.5,
  decision 15).
