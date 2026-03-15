# Contributing To `gh-dash`

Thank you for investing your time in contributing to our project!

In this guide you will get an overview of the contribution workflow from opening an issue, creating a PR, reviewing, and merging the PR.

## The Critical Rule

- The most important rule: you must understand your code. If you can't explain what your changes do and how they interact with the greater system without the aid of AI tools, do not contribute to this project.
- The second most important rule: when you submit a PR you must be willing to address comments and maintain this code. Dot not submit drive-by PRs that solve your own issue without the willingness to iterate on it. Keep these in your own fork.
- Using AI to write code is fine. You can gain understanding by interrogating an agent with access to the codebase until you grasp all edge cases and effects of your changes. What's not fine is submitting agent-generated slop without that understanding. Be sure to read the [AI Usage Policy](AI_POLICY.md).

## AI Usage

The project has strict rules for AI usage. Please see the [AI Usage Policy](AI_POLICY.md). This is very important.

## Quick Guide

### I Have an Idea for a Feature

Like bug reports, first search through both issues and discussions and try to find if your feature has already been requested. Otherwise, open a discussion in the ["Feature Requests, Ideas"](https://github.com/dlvhdr/gh-dash/issues/new?template=feature_request.md) category.

### I've Implemented a Feature

- If there is an issue for the feature, open a pull request straight away.
- If there is no issue, open a discussion and link to your branch.
- If you want to live dangerously, open a pull request and hope for the best.

### I Have a Question Which Is Neither a Bug Report nor a Feature Request

Open a [Q&A discussion](https://github.com/dlvhdr/gh-dash/discussions/categories/q-a), or join our [Discord Server](https://discord.gg/SXNXp9NctV) and ask away in the #help forum channel.

## Working on the Code

### Installing Required Tooling

Our project uses [Devbox](https://github.com/jetpack-io/devbox) to manage its development environment.

Using Devbox will get your dev environment up and running easily and make sure we're all using the same tools with the same versions.

- Clone this repo

```sh
git clone git@github.com:dlvhdr/gh-dash.git && cd gh-dash
```

- Install `devbox`

```sh
curl -fsSL https://get.jetpack.io/devbox | bash
```

- Start the `devbox` shell and run the setup (will take a while on first time)

```sh
devbox shell
```

_This will create a shell where all required tools are installed._

- _(Optional)_ Set up `direnv` so `devbox shell` runs automatically
  - [direnv](https://www.jetify.com/devbox/docs/ide_configuration/direnv/) is a tool that allows setting unique environment variables per directory in your filesystem.
    - Install `direnv` with: `brew install direnv`
    - Add the following line at the end of the `~/.bashrc` file: `eval "$(direnv hook bash)"`
      - See [direnv's installation instructions](https://direnv.net/docs/hook.html) for other shells.
    - Enable `direnv` by running `direnv allow`
- _(Optional)_ Install the VSCode Extension
  - Follow [this guide](https://www.jetify.com/devbox/docs/ide_configuration/vscode/) to set up VSCode to automatically run `devbox shell`.

#### Troubleshooting

- delete the `.devbox` directory at the project's root

### Navigating the Codebase

To navigate our codebase with confidence, familiarize yourself with:

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - the TUI framework we're using
- [The Elm architecture](https://guide.elm-lang.org/architecture/)
- [charmbracelet/glow](https://github.com/charmbracelet/glow) - for parsing and presenting Markdown

#### Code Structure

- `ui/` - this is the code that's responsible for rendering the different parts of the TUI
- `data/` - the code that fetches data from GitHub's GraphQL API
- `config/` - code to parse the user's `config.yml` file
- `utils/` - various utilities

### Debugging

- Write to the log by using Charm's `log` package
- Tail the log by running `task logs`
- Run `dash` in debug mode with `task debug` in another terminal window / pane

```go
import "charm.land/log/v2"

// more code...

log.Debug("some message", "someVariable", someVariable)
```

### Running the Docs Locally

- Run the docs site by running `task docs`

* Go to `localhost:4321` to view them
