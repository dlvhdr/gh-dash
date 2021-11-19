# gh-prs

`gh` cli extension to display a "dashboard" or pull requests by filters you care about.

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
gh prs
```

### Keybindings

 -	Next Pull Request: <kbd>↑</kbd>, <kbd>k</kbd>
 -	Previous Pull Request: <kbd>↓/kbd>, <kbd>j</kbd>
 -	Previous Section: <kbd></kbd>, <kbd>h</kbd>
 -	Next Section: <kbd><kbd>, <kbd>l</kbd>
 -	Page Up: <kbd>Ctrl+u</kbd>
 -	Page Down: <kbd>Ctrl+d</kbd>
 -	Open in GitHub: <kbd>o</kbd>
 -	Help: <kbd>h</kbd>
 -	Quit: <kbd>q</kbd>, <kbd>Ctrl+c</kbd>, <kbd>esc</kbd>
