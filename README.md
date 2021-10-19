# gh-prs

`gh` cli extension to display a "dashboard" of prs by filters you care about.

![demo](https://raw.githubusercontent.com/dlvhdr/gh-prs/main/demo.gif)

Supports 3 sections:
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

## Usage

```
gh prs [repos...]

OPTIONS
    -h, --help             display help
    -r, --repo repo-name   specify a repo to fetch PRs for

EXAMPLE
  gh prs -r dlvhdr/gh-prs -r cli/cli
```

### Author
Dolev Hadar dolevc2@gmail.com
