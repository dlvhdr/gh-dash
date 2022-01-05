# Welcome to `gh-prs` contributing guide ✨

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

If a related issue doesn't exist, you can open a new issue using a relevant [issue form](https://github.com/dlvhdr/gh-prs/issues/new/choose).

### Solve an issue

Scan through our [existing issues](https://github.com/dlvhdr/gh-prs/issues) to find one that interests you.

#### Make Changes

1. Fork the repository.

```sh
git clone https://github.com/dlvhdr/gh-prs.git
```
or if you have the `gh` cli
```sh
gh repo clone dlvhdr/gh-prs
```

2. Install Go: https://go.dev/

3. Create a working branch and start with your changes!

### Pull Request

When you're finished with the changes, create a pull request.

- Fill the "Ready for review" template so that we can review your PR. This template helps reviewers understand your changes as well as the purpose of your pull request.
- Don't forget to [link PR to issue](https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue) if you are solving one.

### Your PR is merged!

Congratulations :tada::tada:
