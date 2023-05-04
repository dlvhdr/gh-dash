---
title: Configuration
weight: 2
summary: >-
  Learn how to configure your terminal dashboard for GitHub with gh-dash.
platen:
  menu:
    flatten_section: true
cascade:
  platen:
    title_as_heading: false
---

`gh-dash` has extensive configuration options.

You can use the default configuration file or use the [`--config`][01] flag to specify an alternate
configuration.

If you don't specify the `--config` flag, `gh-dash` uses the default configuration. If the default
configuration file doesn't already exist, `gh-dash` creates it. The location of the default
configuration file depends on your system:

1. If `$XDG_CONFIG_HOME` is a non-empty string, the default path is
   `$XDG_CONFIG_HOME/gh-dash/config.yml`.
1. If `$XDG_CONFIG_HOME` isn't set, then:
   - On Linux and macOS systems, the default path is `$HOME/gh-dash/config.yml`.
   - On Windows systems, the default path is `%USERPROFILE%\gh-dash\config.yml`.

After `gh-dash` creates the default configuration, you can edit it.

## Options

The configuration for `gh-dash` is schematized. The pages in this section list the configuration
options, their defaults, and how you can use them.

```section
```

## Using the Schema in VS Code

The `gh-dash` configuration schema is published here:

[`https://dlvdhr.github.io/gh-dash/configuration/gh-dash/schema.json`][02]

You can get edit-time feedback, validation, and IntelliSense for your configurations in VS Code by
following these steps:

1. Install [Red Hat's YAML extension for VS Code][03].
1. Open your `gh-dash` configuration file.
1. Add the following line to the top of your configuration file:

   ```yaml
   # yaml-language-server: $schema=https://dlvdhr.github.io/gh-dash/configuration/gh-dash/schema.json
   ```

1. Instead of adding a comment to your configuration file, you could create the
   `.vscode/settings.json` file in your `gh-dash` configuration folder and add this setting:

   ```json
   {
       "yaml.schemas": {
           "https://dlvdhr.github.io/gh-dash/configuration/gh-dash/schema.json": "*.yml"
       }
   }
   ```

With the directive comment at the top of your configuration or the VS Code settings file, you can
then open your configurations and edit them with support for validation. When you hover on an
option in your configuration file, you'll get a brief synopsis of the option and a link to its
documentation on this site.

## Examples

These examples show a few ways you might configure your dashboard.

### Theming example

The color palette in this example is inspired by the [Monokai Pro Spectrum Filter][01] palette.

```yaml
theme:
  colors:
    text:
      primary: "#F7F1FF"
      secondary: "#5AD4E6"
      inverted: "#F7F1FF"
      faint: "#3E4057"
      warning: "#FC618D"
      success: "#7BD88F"
    background:
      selected: "#535155"
    border:
      primary: "#948AE3"
      secondary: "#7BD88F"
      faint: "#3E4057"
```

[01]: ../getting-started/usage.md#--config
[02]: /configuration/gh-dash/schema.json
[03]: https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml