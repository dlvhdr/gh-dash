# gh-dash — Copilot Instructions

`gh-dash` is a `gh` CLI extension that renders a terminal dashboard for GitHub PRs, issues,
and notifications. It's a Go TUI built on Charm's Bubble Tea (Elm-style Model/Update/View),
with data sourced from GitHub's GraphQL API via the `gh` CLI.

Full project context, architecture, conventions, and agent instructions are in
[AGENTS.md](../AGENTS.md).

## Quick Reference

|              |                                                                          |
| ------------ | ------------------------------------------------------------------------ |
| **Stack**    | Go, Bubble Tea, Lip Gloss v2, Glamour, Cobra, koanf                      |
| **Data**     | `gh` CLI + `github.com/shurcooL/githubv4` (GraphQL)                      |
| **Layout**   | `cmd/` (Cobra) → `internal/tui/` (sections, components) ↔ `internal/data/` ↔ `internal/config/` |
| **Run**      | `task` (alias for `go run .`)                                            |
| **Test**     | `task test ./...`                                                        |
| **Lint**     | `task lint` (golangci-lint)                                              |
| **Format**   | `task fmt` (gofumpt)                                                     |

## Architecture Rule

Section components must flow data via `tea.Cmd` → `constants.TaskFinishedMsg` → root
`Update`. **Never block the `Update` loop with synchronous I/O.** Register tasks with
`ctx.StartTask(...)` so the footer spinner reflects progress, and clear them on the returned
message.

## Custom Agents

This repository includes custom Copilot agents for development work:

- **`@dev-loop`** — Orchestrates the full dev workflow: explores → implements → reviews,
  up to 5 iterations until the implementation scores ≥ 8/10
- **`@explorer`** — Read-only codebase audit; maps affected packages, patterns, and reuse
  candidates before any implementation starts
- **`@developer`** — Senior Go engineer agent; implements features/fixes and validates with
  `task lint` + `task test ./...`
- **`@reviewer`** — Outcome-focused code reviewer; scores implementation 1–10 against the
  original problem statement (not just acceptance criteria)

Assign `@dev-loop` to any issue to start an autonomous dev cycle.
