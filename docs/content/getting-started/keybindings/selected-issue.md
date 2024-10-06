---
title: Selected Issue
linkTitle: >-
  ![icon:circle-dot](lucide)&nbsp;Selected Issue
weight: 4
summary: >-
  Lists the default keybindings for interacting with an actively selected item
  in the Issues view for the dashboard.
---

## `a` - Assign Issue { #assign-issue }

Press ![kbd:`a`]() to assign one or more users to the issue. When you do, the dashboard opens the
preview pane and displays a new input.

When the unassign input is active, you can specify one or more GitHub usernames to assign to the
issue. By default, the input includes your username if you're not already assigned. If you're
already assigned, the input is empty by default.

To assign more than one user to the issue, specify additional users after one or more whitespace
characters, like a space, tab, or newline. We recommend separating the additional users with a
newline by pressing ![kbd:`Enter`]() after each username.

To submit the list of users to assign to the issue, press ![kbd:`Ctrl`+`d`](). To cancel the
change instead, press ![kbd:`Ctrl`+`c`]() or ![kbd:`Esc`]().

## `A` - Unassign Issue { #unassign-issue }

Press ![kbd:`A`]() to unassign one or more users from the issue. When you do, the dashboard opens
the preview pane and displays a new input.

When the unassign input is active, you can specify one or more GitHub usernames to unassign from
the issue. By default, the input includes all assigned users separated by newlines.

Make sure the list of users to unassign only includes the users you want to unassign before you
submit the list.

To submit the list of users to unassign from the issue, press ![kbd:`Ctrl`+`d`](). To cancel the
change instead, press ![kbd:`Ctrl`+`c`]() or ![kbd:`Esc`]().

## `c` - Comment on Issue { #comment-on-issue }

Press ![kbd:`c`]() to add a comment to the issue. When you do, the dashboard opens a preview pane and
displays a new input.

You can write your comment as GitHub-flavored Markdown in the input.

To submit the comment on the issue, press ![kbd:`Ctrl`+`d`](). To cancel the comment instead, press
![kbd:`Ctrl`+`c`]() or ![kbd:`Esc`]().

## `x` - Close Issue { #close-issue }

Press ![kbd:`x`]() to close the issue. When you do, the dashboard uses the `gh issue close` command
to close the issue.

```alert
---
variant: warning
---
**Prior to v3.10.0:** When you use this command, the dashboard closes the issue immediately and
without prompting for confirmation. Only use this command when you're sure you
want to close the issue.

**Since v3.10.0:** When you use this command, the dashboard displays a confirmation prompt and
closes the issue only after you approve the action.

This command doesn't support closing the issue with a comment. If you want to
add a comment that explains why you're closing the issue, use the
[comment](#c---comment-on-issue) command before or after you use this one.
```

## `X` - Reopen Issue { #reopen-issue }

Press ![kbd:`X`]() to reopen a closed issue. When you do, the dashboard uses the `gh issue reopen`
command to reopen the issue.

```alert
---
variant: warning
---
**Prior to v3.10.0:** When you use this command, the dashboard reopens the issue immediately and
without prompting for confirmation. Only use this command when you're sure you
want to reopen the closed issue.

**Since v3.10.0:** When you use this command, the dashboard displays a confirmation prompt and
reopens the issue only after you approve the action.
```
