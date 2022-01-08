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

3. To get the icons to render properly you should download and install a Nerd font from https://www.nerdfonts.com/.
Then, select that font as your font for the terminal.

<details>
    <summary><strong>Installing Manually</strong></summary>

> if you want to install this extension **manually**, do these steps:

1. clone repo
    ```bash
    # git
    git clone https://github.com/dlvhdr/gh-prs
    
    # github cli
    gh repo clone dlvhdr/gh-prs
    ```

2. cd to it
    ```bash
    cd gh-prs
    ```

3. install it locally
    ```bash
    gh extension install .
    ```
</details>

## Configuring

Configuration is provided within a `config.yml` file under the extension's directory.

The default `config.yml` file contains:

```yml
prSections:
  - title: My Pull Requests
    filters: is:open author:@me
  - title: Needs My Review
    filters: is:open review-requested:@me
  - title: Subscribed
    filters: is:open -author:@me repo:cli/cli repo:dlvhdr/gh-prs`
defaults:
  preview:
    open: true
    width: 60
```

Adding PR sections is as easy as adding to the list of `prSections` where the properties are:

- title - shown in the TUI
- filters - how the repo's PRs should be filtered - these are plain [github filters](https://docs.github.com/en/search-github/searching-on-github/searching-issues-and-pull-requests)

## Usage

Run:

```sh
gh prs
```

Then press <kbd>?</kbd> for help.

## Author

Dolev Hadar dolevc2@gmail.com
