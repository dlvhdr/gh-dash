# Welcome to `gh-dash` contributing guide ✨

Thank you for investing your time in contributing to our project!

In this guide you will get an overview of the contribution workflow from opening an issue, creating a PR, reviewing, and merging the PR.

## Getting started

To navigate our codebase with confidence, familiar yourself with:

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - the TUI framework we're using
- [The Elm architecture](https://guide.elm-lang.org/architecture/)
- [charmbracelet/glow](https://github.com/charmbracelet/glow) - for parsing and presenting markdown

### Code structure

- `ui/` - this is the code that's responsible on rendering the different parts of the TUI
- `data/` - the code that fetches data from GitHub's GraphQL API
- `config/` - code to parse the user's `config.yml` file
- `utils/` - various utilities

## Issues

### Create a new issue

If you spot a problem, first search if an issue already exists.

If a related issue doesn't exist, you can open a new issue using a relevant [issue form](https://github.com/dlvhdr/gh-dash/issues/new/choose).

### Solve an issue

Scan through our [existing issues](https://github.com/dlvhdr/gh-dash/issues) to find one that interests you.

#### Make Changes

1. Fork the repository.

```sh
git clone https://github.com/dlvhdr/gh-dash.git
```

or if you have the `gh` cli

```sh
gh repo clone dlvhdr/gh-dash
```

2. Install Go: https://go.dev/

3. Create a working branch and start with your changes!

### Pull Request

When you're finished with the changes, create a pull request.

- Fill the "Ready for review" template so that we can review your PR. This template helps reviewers understand your changes as well as the purpose of your pull request.
- Don't forget to [link PR to issue](https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue) if you are solving one.

### Debugging

- Pass the debug flag: `go run gh-dash.go --debug`
- Write to the log by using Go's builtin `log` package
- View the log by running `tail -f debug.log`

```golang
import "log"

// more code...

log.Printf("Some message with a variable %v\n", someVariable)
```

### Running the docs locally

- Check the current Hugo version in the [workflow file](./.github/workflows/hugo.yaml)
- Install correct Hugo Extended version using the [official installation guide](https://gohugo.io/getting-started/installing/)
- Check the Hugo version using `hugo version`
- Go to the `docs/` directory using `cd docs`
- Install the Hugo mods using `hugo mod get`
- Run the Hugo server using `hugo server`

### Your PR is merged!

Congratulations :tada::tada:
