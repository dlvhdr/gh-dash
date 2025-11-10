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
  <a href="https://www.jetify.com/devbox/docs/contributor-quickstart/" alt="Built with Devbox"><img src="https://www.jetify.com/img/devbox/shield_galaxy.svg" /></a>
</p>

<br />

<img src="./docs/src/assets/overview.gif" />

## 🌟 Features

> [!NOTE]
> If you like quickly navigating with your keyboard, seeing the PRs and issues you need and you <strong>love the terminal</strong> - <code>DASH</code> is for you! 🫵🏽

- User-defined, per-repo, PRs & issues sections
- Overridable vim-style keyboard hotkeys
- Custom actions to perform your specific workflow needs
- Everything you can do on GitHub - diff, comment, checkout, push, update etc.
- Control every setting with a YAML config file

## 📃 Docs

`DASH` has an extensive docs site at [gh-dash.dev/getting-started](https://gh-dash.dev/getting-started).

## ❤️ Donating

If you enjoy `DASH` and want to help, consider supporting the project with a
donation at the [sponsors page](https://github.com/sponsors/dlvhdr).

## 👥 Discord

Have questions? Join our [Discord community](https://discord.gg/SXNXp9NctV)!

## 🙏 Contributing

See the contribution guide at [https://www.gh-dash.dev/contributing](https://www.gh-dash.dev/contributing/).

## 🛞 Under the hood

`DASH` uses:

- [bubbletea](https://github.com/charmbracelet/bubbletea) for the TUI
- [lipgloss](https://github.com/charmbracelet/lipgloss) for the styling
- [glamour](https://github.com/charmbracelet/glamour) for rendering markdown
- [vhs](https://github.com/charmbracelet/vhs) for generating the GIF
- [cobra](https://github.com/spf13/cobra) for the CLI
- [gh](https://github.com/cli/cli) for the GitHub functionality
- [delta](https://github.com/dandavison/delta) for viewing PR diffs

## Author

Dolev Hadar [@dlvhdr](https://github.com/dlvhdr).
