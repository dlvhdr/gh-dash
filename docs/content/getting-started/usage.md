---
title: Usage
linkTitle: >-
  ![icon:terminal](lucide)&nbsp;Usage
summary: >-
  Get started using `gh-dash`.
weight: 2
---

To use `gh-dash`, follow these steps after you've [installed it][01]:

1. Run

   ```bash
   gh dash
   ```

2. Press ![kbd:`?`]() for help.
3. Run `gh dash --help` for more info:

   ```text
   Usage:
     gh dash [flags]

   Flags:
     -c, --config string   use this configuration file (default lookup: a .gh-dash.yml file if inside a git repo, $GH_DASH_CONFIG env var, or if not set, $XDG_CONFIG_HOME/gh-dash/config.yml)
         --debug           passing this flag will allow writing debug output to debug.log
     -h, --help            help for gh-dash
   ```

## Flags

### `--config`

Specify the path to a configuration file to use for the dashboard. If the configuration file
doesn't exist or is invalid, `gh-dash` returns an error.

```bash
gh dash --config path/to/configuration/file.yml
```

| Aliases |  Type  |                Default                |
| :------ | :----: | :------------------------------------ |
| `-c`    | String | `.gh-dash.yml` file if inside a git repo, `$GH_DASH_CONFIG` env var, or if not set, `$XDG_CONFIG_HOME/gh-dash-config.yml` |

If you don't specify this flag, `gh-dash` uses the default configuration. If the file doesn't exist, gh-dash will create it. The location of the default configuration file depends on your system:

1. If Inside a git repo, `gh-dash` will look for a `.gh-dash.yml` file in the root of the repo.
2. If `$GH_DASH_CONFIG` is a non-empty string, `gh-dash` will use this file for
    its configuration.
3. If `$GH_DASH_CONFIG` isn't set and `$XDG_CONFIG_HOME` is a non-empty string,
    the default path is `$XDG_CONFIG_HOME/gh-dash/config.yml`.
4. If neither `$GH_DASH_CONFIG` or `$XDG_CONFIG_HOME` are set, then:
   - On Linux and macOS systems, the default path is `$HOME/gh-dash/config.yml`.
   - On Windows systems, the default path is `%USERPROFILE%\gh-dash\config.yml`.

For more information about authoring configurations, see [Configuration][02].

### `--debug`

Specify whether `gh-dash` should write logs to the `debug.log` file in the current directory. By
default, `gh-dash` doesn't output debug information.

```bash
gh dash --debug
```

| Aliases |  Type   | Default |
| :------ | :-----: | :------ |
| (None)  | Boolean | `false` |

When you use this flag, `gh-dash` creates the `debug.log` file in the current directory if it doesn't exist. If the file does exist, `gh-dash` appends new log entries to it.

### `--help`

Use this flag to display the help information for `gh-dash` in the terminal. If you specify this
flag, `gh-dash` ignores all other flags. It only displays the help information.

```bash
gh dash --help
```

| Aliases |  Type   | Default |
| :------ | :-----: | :------ |
| `-h`    | Boolean | `false` |

### `--version`

Use this flag to display the version information for `gh-dash` in the terminal. If you specify this
flag with the `--config` or `--debug` flags, `gh-dash` ignores them. It only displays the version
information.

```bash
gh dash --version
```

| Aliases |  Type   | Default |
| :------ | :-----: | :------ |
| `-v`    | Boolean | `false` |

When you use this flag, `gh-dash` emits the following information:

```text
gh-dash version <version>
commit: <commit_sha>
built at: <build_timestamp>
built by: <build_user>
goos: <operating_system>
goarch: <cpu_architecture>
```

- `<version>` is the extension's semantic version without a `v` prefix.
- `<commit_sha>` is the exact commit SHA the extension was built from.
- `<build_timestamp>` is the UTC date and time when the extension was built.
- `<build_user>` is who built the extension. For official releases, this is always `goreleaser`.

For example, the version information for the [v3.7.7 release][03] on Windows with an x64 processor
is:

```text
gh-dash version 3.7.7
commit: 6ce3f89ab0d73dd88e359133699d1cf920f88699
built at: 2023-04-15T08:42:16Z
built by: goreleaser
goos: windows
goarch: amd64
```

## Default Keybindings

When you use `gh-dash`, it displays the dashboard as a terminal UI (TUI). In the TUI, you can use
several commands by pressing key combinations to navigate and interact with the dashboard.

You can press the ![kbd:`?`]() key to toggle the help menu in the UI. The help menu lists the available
keybindings for the current context.

You can press the ![kbd:`q`]() key to quit the dashboard and return to your normal terminal view.

For more information about the keybindings for the dashboard, see [Keybindings][04].

[01]: ./installation.md
[02]: ../configuration/_index.md
[03]: https://github.com/dlvhdr/gh-dash/releases/tag/v3.7.7
[04]: keybindings/_index.md
