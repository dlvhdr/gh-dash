---
title: Contributing
linkTitle: >-
  ![icon:heart-handshake](lucide) Contributing
summary: >-
  Explains how folks can contribute to the `gh-dash` project.
weight: 3
platen:
  menu:
    flatten_section: true
---

## Welcome! ![icon:stars](lucide)

Thank you for investing your time in contributing to our project!

In this guide you will get an overview of the contribution workflow from opening an issue, creating
a PR, reviewing, and merging the PR.

## Getting Started

To navigate our codebase with confidence, familiar yourself with:

- [Bubbletea][01] - the TUI framework we're using
- [The Elm architecture][02]
- [charmbracelet/glow][03] - for parsing and presenting markdown

### Code structure

- `ui/` - this is the code that's responsible on rendering the different parts of the TUI
- `data/` - the code that fetches data from GitHub's GraphQL API
- `config/` - code to parse the user's `config.yml` file
- `utils/` - various utilities

## Issues

### Create a new issue

If you spot a problem, first search if an issue already exists.

If a related issue doesn't exist, you can open a new issue using a relevant [issue form][04].

### Solve an issue

Scan through our [existing issues][05] to find one that interests you.

#### Make Changes

1. Fork the repository.

   ```sh
   git clone https://github.com/dlvhdr/gh-dash.git
   ```

   Or, if you have the `gh` cli:

   ```sh
   gh repo clone dlvhdr/gh-dash
   ```

2. Install Go: https://go.dev/
3. Create a working branch and start with your changes!

### Pull Request

When you're finished with the changes, create a pull request.

- Fill the "Ready for review" template so that we can review your PR. This template helps reviewers
  understand your changes as well as the purpose of your pull request.
- Don't forget to [link PR to issue][06] if you are solving one.

### Debugging

- Pass the debug flag: `go run gh-dash.go --debug`
- Write to the log using Go's builtin `log` package
- View the log by running `tail -f debug.log`

```golang
import "log"

// more code...

log.Printf("Some message with a variable %v\n", someVariable)
```

### Your PR is merged!

Congratulations 🎉🎉

<!-- Link Reference Definitions -->
[01]: https://github.com/charmbracelet/bubbletea
[02]: https://guide.elm-lang.org/architecture/
[03]: https://github.com/charmbracelet/glow
[04]: https://github.com/dlvhdr/gh-dash/issues/new/choose
[05]: https://github.com/dlvhdr/gh-dash/issues
[06]: https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue
