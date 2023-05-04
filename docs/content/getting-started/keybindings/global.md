---
title: Global
linkTitle: >-
  ![icon:globe](lucide) Global
weight: 1
summary: >-
  Lists the default keybindings for controlling the dashboard globally.
---

This section lists the default keybindings for controlling the dashboard globally. These are
available in both the PRs and Issues views.

## `?` - Toggle Help { #toggle-help }

Press ![kbd:`?`]() to toggle the help menu in the UI. The help menu lists the available
keybindings for the current context.

## `/` - Search { #search }

Press ![kbd:`/`]() to focus on the dashboard's search input box. When you move the focus to the
search input box you can edit the current section's GitHub search criteria. To refresh the section
with your updated query, press ![kbd:`enter`](). After the dashboard updates, focus is returned to
the active section.

Any changes you make to the search query for a section aren't persistent. If you close the
dashboard and reopen it, the dashboard displays the sections with the queries defined in your
[configuration file](../../configuration/_index.md). To make persistent changes to your sections
or add a new section, update your configuration.

## `r` - Refresh Current Section { #refresh-current-section }

Press ![kbd:`r`]() to refresh the current section's work items. When you do, the dashboard reruns
the defined query for the section and displays the returned work items.

## `R` - Refresh All Sections { #refresh-all-sections }

Press ![kbd:`R`]() to refresh every section in the dashboard's current view. When you do, the
dashboard reruns the defined query for every section and displays the returned work items for the
current section. When you navigate to another section, it displays the updated work items for that
section.

## `s` - Switch View { #switch-view }

Press the ![kbd:`s`]() key to switch the dashboard from the PRs view to the Issues view or the
Issues view to the PRs view. The first time you switch to a view in your dashboard, the dashboard
runs the defined query for every section in that view.

## `q` - Quit { #quit }

Press the ![kbd:`q`]() key to quit the dashboard and return to your normal terminal view.
