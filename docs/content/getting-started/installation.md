---
title: Installation
linkTitle: >-
  ![icon:package-open](lucide)&nbsp;Installation
summary: >-
  Get started installing `gh-dash`.
weight: 1
---

## Recommended Steps

1. Install the `gh` CLI - see the [installation instructions][01]

   _Installation requires a minimum version (2.0.0) of the GitHub CLI that supports extensions._

2. Install this extension:

   ```sh
   gh extension install dlvhdr/gh-dash
   ```

3. To get the icons to render properly, you should download and install a [Nerd font][02]. Then,
   select that font as your font for the terminal.

```details { summary="How do I get these exact colors and font?" }

The screenshots in this documentation use [Alacritty][n01] with the
[tokyonight theme][n02] and the [Fira Code][n03] Nerd Font. For the full setup,
see [these dotfiles][n04].

[n01]: https://github.com/alacritty/alacritty
[n02]: https://github.com/folke/tokyonight.nvim
[n03]: https://github.com/ryanoasis/nerd-fonts/tree/master/patched-fonts/FiraCode
[n04]: https://github.com/dlvhdr/dotfiles/blob/main/.config/alacritty/alacritty.yml
```

## Manual Steps

If you want to install this extension **manually**, follow these steps:

1. Clone the repo

   ```bash
   # git
   git clone https://github.com/dlvhdr/gh-dash

   # GitHub CLI
   gh repo clone dlvhdr/gh-dash
   ```

2. `cd` into it

   ```bash
   cd gh-dash
   ```

3. Install it locally

   ```bash
   gh extension install .
   ```

[01]: https://github.com/cli/cli#installation
[02]: https://www.nerdfonts.com/
