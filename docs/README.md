# gh-dash — Developer Documentation

`gh-dash` is a `gh` CLI extension that renders a terminal dashboard for GitHub PRs, issues, and notifications. It's a Go TUI built on [Charm's Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm-style Model/Update/View), [Lip Gloss v2](https://github.com/charmbracelet/lipgloss) for styling, [Glamour](https://github.com/charmbracelet/glamour) for markdown rendering, and [Cobra](https://github.com/spf13/cobra) for the CLI entrypoint. All GitHub data flows through the `gh` CLI's GraphQL bridge.

This directory holds **developer / maintainer** reference documentation. For **end-user** docs (installation, configuration YAML, keybindings, themes) see [gh-dash.dev](https://gh-dash.dev).

## Start here

| Doc | Read this if you want to… |
| --- | --- |
| [architecture.md](./architecture.md) | Understand the Bubble Tea loop, the `Section` extension point, and how data flows from GitHub to the screen. |
| [project-structure.md](./project-structure.md) | Locate a package or figure out where a piece of code lives. |
| [build-and-deployment.md](./build-and-deployment.md) | Build locally, run tests & linters, or trace the GoReleaser tag-driven release pipeline. |
| [development.md](./development.md) | Set up a dev environment, add a new section, add a custom command, or understand async fetch patterns. |

## Related files

- [`../CONTRIBUTING.md`](../CONTRIBUTING.md) — contribution rules, AI-usage policy, devbox setup walk-through.
- [`../CLAUDE.md`](../CLAUDE.md) — instruction-style architecture summary consumed by AI tooling (overlaps with `architecture.md` but is written as guidance rather than reference).
- [`../Taskfile.yaml`](../Taskfile.yaml) — canonical list of dev tasks (`task --list` from inside `devbox shell`).
- [`../.goreleaser.yaml`](../.goreleaser.yaml) — release build matrix and changelog grouping.

## Conventions used across these docs

- Mermaid diagrams render natively on GitHub — no build step required.
- File paths are relative to the repo root and link to source so GitHub makes them clickable.
- Commands assume you are inside `devbox shell` unless noted otherwise.
- When docs mention "the root model" they mean the `Model` struct defined in [`../internal/tui/ui.go`](../internal/tui/ui.go).
