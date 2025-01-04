# gh-dash

✨ A GitHub (`gh`) CLI extension to display a dashboard with **pull requests** and **issues** by filters you care about.

<a href="https://github.com/charmbracelet/bubbletea/releases"><img src="https://img.shields.io/github/release/dlvhdr/gh-dash.svg" alt="Latest Release"></a>

<img src="https://user-images.githubusercontent.com/6196971/198704107-6775a0ba-669d-418b-9ae9-59228aaa84d1.gif" />

## ✨ Features

- 🌅 fully configurable - define sections using GitHub filters
- 🔍 search for both prs and issues
- 📝 customize columns with `hidden`, `width` and `grow` props
- ⚡️ act on prs and issues with checkout, comment, open, merge, diff, etc...
- ⌨️ set custom actions with new keybindings
- 💅 use custom themes
- 🔭 view details about a pr/issue with a detailed sidebar
- 🪟 write multiple configuration files to easily switch between completely different dashboards
- ♻️ set an interval for auto refreshing the dashboard

## 📃 Docs

See the docs site at [dlvhdr.github.io/gh-dash](https://dlvhdr.github.io/gh-dash) to get started,
or just skim this README.

## 📦 Installation

1. Install the `gh` CLI - see the [installation](https://github.com/cli/cli#installation)

   _Installation requires a minimum version (2.0.0) of the GitHub CLI that supports extensions._

2. Install this extension:

   ```sh
   gh extension install dlvhdr/gh-dash
   ```

3. To get the icons to render properly you should download and install a Nerd font from https://www.nerdfonts.com/.
   Then, select that font as your font for the terminal.

<details>
   <summary>Installing Manually</summary>

> If you want to install this extension **manually**, follow these steps:

1. Clone the repo

   ```shell
   # git
   git clone https://github.com/dlvhdr/gh-dash
   ```

   ```shell
   # GitHub CLI
   gh repo clone dlvhdr/gh-dash
   ```

2. Cd into it

   ```bash
   cd gh-dash
   ```

3. Build it

   ```bash
   go build
   ```

4. Install it locally
   ```bash
   gh extension install .
   ```
   </details>

<details>
    <summary>Updating from an older version</summary>

```bash
gh extension upgrade dlvhdr/gh-dash
```
</details>

<details>
   <summary>How do I get these exact colors and font?</summary>

> I'm using [Alacritty](https://github.com/alacritty/alacritty) with the [tokyonight theme](https://github.com/folke/tokyonight.nvim) and the [Fira Code](https://github.com/ryanoasis/nerd-fonts/tree/master/patched-fonts/FiraCode) Nerd Font.
> For my full setup check out [my dotfiles](https://github.com/dlvhdr/dotfiles/blob/main/.config/alacritty/alacritty.yml).


</details>

## ⚡️ Usage

Run

```sh
gh dash
```

Then press <kbd>?</kbd> for help.

Run `gh dash --help` for more info:

```
Usage:
  gh dash [flags]

Flags:
  -c, --config string   use this configuration file 
                        (default lookup:
                          1. a .gh-dash.yml file if inside a git repo
                          2. $GH_DASH_CONFIG env var
                          3. $XDG_CONFIG_HOME/gh-dash/config.yml
                        )
      --debug           passing this flag will allow writing debug output to debug.log
  -h, --help            help for gh-dash
```

## ⚙️ Configuring

A section is defined by a:

- title - shown in the TUI
- filters - how the repo's PRs should be filtered - these are plain [GitHub filters](https://docs.github.com/en/search-github/searching-on-github/searching-issues-and-pull-requests)

All configuration is provided within a `config.yml` file under the extension's directory (either a `.gh-dash.yml` file if inside a repo, `$XDG_CONFIG_HOME/gh-dash` or `~/.config/gh-dash/` or your OS config dir) or `$GH_DASH_CONFIG`.

An example `config.yml` file contains:

```yml
prSections:
  - title: My Pull Requests
    filters: is:open author:@me
    layout:
      author:
        hidden: true
        # width: <number of columns>
        # grow: <bool> this will make the column grow in size
  - title: Needs My Review
    filters: is:open review-requested:@me
  - title: Subscribed
    filters: is:open -author:@me repo:cli/cli repo:dlvhdr/gh-dash
    limit: 50 # optional limit of rows fetched for this section
issuesSections:
  - title: Created
    filters: is:open author:@me
  - title: Assigned
    filters: is:open assignee:@me
  - title: Subscribed
    filters: is:open -author:@me repo:microsoft/vscode repo:dlvhdr/gh-dash
defaults:
  layout:
    prs:
      repo:
        grow: true,
        width: 10
        hidden: false
    # issues: same structure as prs
  prsLimit: 20 # global limit
  issuesLimit: 20 # global limit
  preview:
    open: true # whether to have the preview pane open by default
    width: 60 # width in columns
  refetchIntervalMinutes: 30 # will re-fetch all sections every 30 minutes
repoPaths: # configure where to locate repos when checking out PRs
  :owner/:repo: ~/src/github.com/:owner/:repo # template if you always clone GitHub repos in a consistent location
  dlvhdr/*: ~/code/repos/dlvhdr/* # will match dlvhdr/repo-name to ~/code/repos/dlvhdr/repo-name
  dlvhdr/gh-dash: ~/code/gh-dash # will not match wildcard and map to specified path
keybindings: # optional, define custom keybindings - see more info below
theme: # optional, see more info below
pager:
  diff: less # or delta for example
confirmQuit: false # show prompt on quit or not
```

### 🗃 Running with a different config file

You can run `gh dash --config <path-to-file>` to run `gh-dash` against another config file.

This lets you easily define multiple dashboards with different sections.<br>
It can be useful if you want to have a 🧳 work and 👩‍💻 personal dashboards, or if you want to view multiple dashboards at the same time.

### ⌨️ Keybindings

You can:

1. Override the builtin commands keybindings
2. Define your own custom keybindings to run bash commands using [Go Templates](https://pkg.go.dev/text/template).

#### Overriding builtin commands keybindings

To override the "checkout" keybinding you can include this in your `config.yml` file:

```yaml
keybindings:
  prs:
    - key: O
      builtin: checkout
```

The list of available builtin commands are:

1. `universal`: up, down, firstLine, lastLine, togglePreview, openGithub, refresh, refreshAll, pageDown, pageUp, nextSection, prevSection, search, copyurl, copyNumber, help, quit
2. `prs`: approve, assign, unassign, comment, diff, checkout, close, ready, reopen, merge, update, watchChecks, viewIssues
3. `Issues`: assign, unassign, comment, close, reopen, viewPrs

To unbind the "esc" keybinding you can include this in your `config.yml` file:

```yaml
keybindings:
  prs:
    - key: esc
      builtin:
```

#### Defining custom keybindings

This is available for both PRs and Issues.
For PRs, the available arguments are:

| Argument      | Description                                                                     |
| ------------- | ------------------------------------------------------------------------------- |
| `RepoName`    | The full name of the repo (e.g. `dlvhdr/gh-dash`)                               |
| `RepoPath`    | The path to the Repo, using the `config.yml` `repoPaths` key to get the mapping |
| `PrNumber`    | The PR number                                                                   |
| `HeadRefName` | The PR's remote branch name                                                     |
| `BaseRefName` | The PR's base branch name                                                       |

For Issues, the available arguments are:

| Argument      | Description                                                                     |
| ------------- | ------------------------------------------------------------------------------- |
| `RepoName`    | The full name of the repo (e.g. `dlvhdr/gh-dash`)                               |
| `RepoPath`    | The path to the Repo, using the `config.yml` `repoPaths` key to get the mapping |
| `IssueNumber` | The Issue number                                                                |

**Examples**

1. To review a PR with either Neovim or VSCode include the following in your `config.yml` file:

```yaml
repoPaths:
  dlvhdr/gh-dash: ~/code/gh-dash

keybindings:
  prs:
    - key: c
      command: >
        tmux new-window -c {{.RepoPath}} '
          gh pr checkout {{.PrNumber}} &&
          nvim -c ":DiffviewOpen master...{{.HeadRefName}}"
        '
    - key: v
      command: >
        cd {{.RepoPath}} &&
        code . &&
        gh pr checkout {{.PrNumber}}
```

2. To pin an issue include the following in your `config.yml` file:

```yaml
keybindings:
  issues:
    - key: P
      command: gh issue pin {{.IssueNumber}} --repo {{.RepoName}}
```

### 🚥 Repo Path Matching

Repo name to path mappings can be exact match (full name, full path) or wildcard matched using the `owner` and partial path.

An exact match for the full repo name to a full path takes priority over a matching wildcard, and wildcard matches must match to a wildcard path.

An `:owner/:repo` template can be specified as a generic fallback.

```yaml
repoPaths:
  :owner/:repo: ~/src/github.com/:owner/:repo # template if you always clone GitHub repos in a consistent location
  dlvhdr/*: ~/code/repos/dlvhdr/* # will match dlvhdr/repo-name to ~/code/repos/dlvhdr/repo-name
  dlvhdr/gh-dash: ~/code/gh-dash # will not match wildcard and map to specified path
```

The `RepoName` and `RepoPath` keybinding arguments are fully expanded when sent to the command.

### 💅 Custom Themes

To override the default set of terminal colors and instead create your own color scheme, you can define one in your `config.yml` file.
If you choose to go this route, you need to specify _all_ of the following keys as colors in hex format (`#RRGGBB`), otherwise validation will fail.

```yaml
theme:
  ui:
    sectionsShowCount: true
    table:
      showSeparator: true
  colors:
    text:
      primary: "#E2E1ED"
      secondary: "#666CA6"
      inverted: "#242347"
      faint: "#3E4057"
      warning: "#F23D5C"
      success: "#3DF294"
      error: "#D20F39"
    background:
      selected: "#39386B"
    border:
      primary: "#383B5B"
      secondary: "#39386B"
      faint: "#2B2B40"
```

### 🪟 Layout

You can customize each section's layout as well as the global layout.

For example, to hide the `author` column for **all** PR sections, include the following in your `config.yml`.

```
defaults:
  layout:
    prs:
      author:
        hidden: true
```

- For `prs` the column names are: `updatedAt, repo, author, title, reviewStatus, state, ci, lines, assignees, base`.
- For `issues` the column names are: `updatedAt, state, repo, title, creator, assignees, comments, reactions`.
- The available properties to control are: `grow` (false, true), `width` (number of cells), and `hidden` (false, true).

## Author

Dolev Hadar dolevc2@gmail.com
