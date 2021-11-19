# gh-prs

`gh` cli extension to display a "dashboard" or pull requests by filters you care about.

![demo](https://raw.githubusercontent.com/dlvhdr/gh-prs/main/demo.gif)

Comes with 3 default sections:
* My Pull Requests
* Needs My Review
* Subscribed

## Installation

1. Install gh cli:

```sh
brew install gh
```

2. Install this extension:

```sh
gh extension install dlvhdr/gh-prs
```

## Configuring

Configuration is done in the `sections.yml` file under the extension's directory.

Example `sections.yml` file: 

```yml
- title: My Pull Requests
  repos:
    - dlvhdr/gh-prs
  filters: author:@me
- title: Needs My Review
  repos:
    - dlvhdr/gh-prs
  filters: assignee:@me
- title: Subscribed
  repos:
    - cli/cli
    - charmbracelet/glamour
    - charmbracelet/lipgloss
  filters: -author:@me
```

## Usage

Run
```

### Author
Dolev Hadar dolevc2@gmail.com
