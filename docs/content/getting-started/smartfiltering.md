---
title: About Smart Filtering
linkTitle: >-
  ![icon:folder-git](lucide)&nbsp;About Smart Filtering
summary: >-
  About the Smart Filtering feature
weight: 3
---

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
