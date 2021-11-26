# gh-prs

A `gh` cli extension to display a dashboard with pull requests by filters you care about.

![demo](https://github.com/dlvhdr/gh-prs/blob/8621183574c573e4077360b5027ffea70999b921/demo.gif)

## Installation

Installation requires a minimum version (2.0.0) of the the Github CLI to support extensions.

1. Install the `gh cli` - see the [installation/upgrade instructions](https://github.com/cli/cli#installation)

2. Install this extension:

```sh
gh extension install dlvhdr/gh-prs
```

> if you get an error when execute `gh prs`, do these steps:

1. clone this repo
    ```bash
    # git
    git clone https://github.com/dlvhdr/gh-prs
    
    # github cli
    gh repo clone dlvhdr/gh-prs
    ```

2. cd to the repo
    ```bash
    cd gh-prs
    ```

3. install it locally
    ```bash
    gh extension install .
    ```


## Configuring

Configuration is provided within a `sections.yml` file under the extension's directory. If the configuration file is missing, a prompt to create it will be displayed when running `gh prs`.

Each section is defined by a top level array item and has the following properties:

- title - shown in the TUI
- repos - a list of repos to enumerate
- filters - how the repo's PRs should be filtered - these are plain [github filters](https://docs.github.com/en/search-github/searching-on-github/searching-issues-and-pull-requests)

Example `sections.yml` file:

```yml
- title: My Pull Requests
  repos:
    - dlvhdr/gh-prs
  filters: author:@me
- title: Needs My Review
  repos:
    - dlvhdr/gh-prs
  filters: review-requested:@me
- title: Subscribed
  repos:
    - cli/cli
    - charmbracelet/glamour
    - charmbracelet/lipgloss
  filters: -author:@me
```

## Usage

Run:

```sh
gh prs
```

Then press <kbd>?</kbd> for help.

## Author

Dolev Hadar dolevc2@gmail.com
