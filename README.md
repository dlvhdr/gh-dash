<br />
<p align="center">
  <a  class="underline: none;" href="https://gh-dash.dev">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="./docs/public/logo.png">
      <img alt="Text changing depending on mode. Light: 'So light!' Dark: 'So dark!'" width="600" src="./docs/public/logo-light.png">
    </picture>
  </a>
</p>

<p align="center">
    <a href="https://gh-dash.dev" target="_blank">→ https://gh-dash.dev ←</a>
</p>
<p align="center">
  A rich terminal UI for GitHub that doesn't break your flow.
  <br />
  <br />
  <a href="https://github.com/dlvhdr/gh-dash/releases"><img src="https://img.shields.io/github/release/dlvhdr/gh-dash.svg" alt="Latest Release"></a>
  <a href="https://discord.gg/SXNXp9NctV"><img src="https://img.shields.io/discord/1413193703476035755?label=discord" alt="Discord"/></a>
  <a href="https://github.com/sponsors/dlvhdr"><img src=https://img.shields.io/github/sponsors/dlvhdr?logo=githubsponsors&color=EA4AAA /></a>
</p>

<br />

<img src="./docs/src/assets/overview.gif" />

## 📃 Docs

Installation instructions, configuration options etc. can be found at the docs
site [https://gh-dash.dev](https://gh-dash.dev).

## ⚡️ Usage

Run

```sh
gh dash
```

Then press <kbd>?</kbd> for help.

Run `gh dash --help` for more info.

## ❤️ Donating

If you enjoy dash and want to help, consider supporting the project with a
donation at [https://github.com/sponsors/dlvhdr](https://github.com/sponsors/dlvhdr).

## 🙏 Contributing

See the contributing guide at [https://www.gh-dash.dev/contributing](https://www.gh-dash.dev/contributing/).

## 🛞 Under the hood

gh-dash uses:

- [bubbletea](https://github.com/charmbracelet/bubbletea) for the TUI
- [lipgloss](https://github.com/charmbracelet/lipgloss) for the styling
- [glamour](https://github.com/charmbracelet/glamour) for rendering markdown
- [vhs](https://github.com/charmbracelet/vhs) for generating the GIF
- [cobra](https://github.com/spf13/cobra) for the CLI
- [gh](https://github.com/cli/cli) for the GitHub functionality
- [delta](https://github.com/dandavison/delta) for viewing PR diffs

## Author

Dolev Hadar dolevc2@gmail.com
