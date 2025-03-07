---
title: Searching
linkTitle: >-
  ![icon:search](lucide)&nbsp;Searching
summary: >-
  How to search and filter issues and prs
weight: 3
---

# Searching

Searching for prs and issues is done by defining sections and their filters.
Search filters are defined using [GitHub search filters][01].

For example, this section, shows open PRs authored by anyone but me, which were updated in the last 2 weeks.

```yaml
prsSections:
  - title: Review
    filter: >-
      is:open
      -author:@me
      updated>={{ nowModify "-2w" }}
```

Note: don't specify `is:pr` for this setting. The dashboard always adds that filter for PR
sections.

You can define any combination of search filters. To make it easier to read and maintain
your filters, we recommend using the `>-` syntax after the `filter` key and writing one
filter per line.

For more information about writing filters for searching GitHub, see [Searching issues and pull requests][02].

## Search Templates

In addition to GitHub's filters, gh-dash adds templating functions.

### `nowModify`

The `nowModify` function helps you calculate relative dates in the ISO-8601 format (which is what GitHub expects).

Given the date today is 2025-02-02, a search filter of `updated>={{ nowModify "-1mo" }}` will output `updated>=2025-01-02`.

A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as `1w`, `-8h` or `1mo2w`.

The available units are:

- [Go's builtin durations](https://pkg.go.dev/time#ParseDuration) - `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`
- Additionally:
  - `d`/`D` for days
  - `w`/`W` for weeks
  - `M`/`mo` for months
  - `y`/`Y` for years

## Smart Filtering

By default, if the directory you launch `gh-dash` from is a clone of a remote GitHub repo (or if you
have the `GH_REPO` environment variable set to a particular remote), then for any of your PR
sections and issue sections with `filters` values in your [configuration](/configuration) that don’t
have an explicit `repo:` field, `gh-dash` adds a `repo:<RepoName>` field to the search-bar value for
them (where _`<RepoName>`_ is the name of the remote repo).

That is, `gh-dash` further filters those sections down to only the PRs/issues for the GitHub
repo name specified in your `GH_REPO` environment variable — or else the repo name of the remote
tracked by the clone directory from which `gh-dash` launched.

For that, `gh-dash` first checks and uses the repo name in the `GH_REPO` environment variable (if
you have that set). If `gh-dash` doesn’t find that, then it next checks for the value of the remote
repo name tracked by the clone directory from which you launched `gh-dash` — by looking through all
GitHub remotes configured for that clone in the following order:

1. `upstream`
2. `github`
3. `origin`

…and, otherwise, if `gh-dash` finds no remotes with any of those names, then it uses the repo name
for the first remote in the output that `git remote` shows.

To disable Smart Filtering at launch, set [`smartFilteringAtLaunch`](/configuration/gh-dash/#smartfilteringatlaunch)
to `false` in your [configuration](/configuration).

To toggle Smart Filtering on or off for the current section you’re currently viewing, either use the
`t` key — or else use whatever custom keybinding you have set for the `togglesearch` builtin in the
`keybindings` section of your [configuration](/configuration).

[01]: https://docs.github.com/en/search-github/searching-on-github/searching-issues-and-pull-requests
[02]: https://docs.github.com/en/search-github/getting-started-with-searching-on-github/understanding-the-search-syntax
