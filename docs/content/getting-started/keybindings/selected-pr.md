---
title: Selected PR
linkTitle: >-
  ![icon:git-pull-request](lucide)&nbsp;Selected PR
weight: 5
summary: >-
  Lists the default keybindings for interacting with an actively selected item
  in the PRs view for the dashboard.
---

## `a` - Assign PR { #assign-pr }

Press ![kbd:`a`]() to assign one or more users to the PR. When you do, the dashboard opens the
preview pane and displays a new input.

When the unassign input is active, you can specify one or more GitHub usernames to assign to the
PR. By default, the input includes your username if you're not already assigned. If you're already
assigned, the input is empty by default.

To assign more than one user to the PR, specify additional users after one or more whitespace
characters, like a space, tab, or newline. We recommend separating the additional users with a
newline by pressing ![kbd:`Enter`]() after each username.

To submit the list of users to assign to the PR, press ![kbd:`Ctrl`+`d`](). To cancel the
change instead, press ![kbd:`Ctrl`+`c`]() or ![kbd:`Esc`]().

## `A` - Unassign PR { #unassign-pr }

Press ![kbd:`A`]() to unassign one or more users from the PR. When you do, the dashboard opens the
preview pane and displays a new input.

When the unassign input is active, you can specify one or more GitHub usernames to unassign from
the PR. By default, the input includes all assigned users separated by newlines.

Make sure the list of users to unassign only includes the users you want to unassign before you
submit the list.

To submit the list of users to unassign from the PR, press ![kbd:`Ctrl`+`d`](). To cancel the
change instead, press ![kbd:`Ctrl`+`c`]() or ![kbd:`Esc`]().

## `c` - Comment on PR { #comment-on-pr }

Press ![kbd:`c`]() to add a comment to the PR. When you do, the dashboard opens a preview pane and
displays a new input.

You can write your comment as GitHub-flavored Markdown in the input.

To submit the comment on the PR, press ![kbd:`Ctrl`+`d`](). To cancel the comment instead, press
![kbd:`Ctrl`+`c`]() or ![kbd:`Esc`]().

## `C` - Checkout PR { #checkout-pr }

Press ![kbd:`C`]() to checkout the PR locally. The dashboard checks for the `repoPaths` key in your
configuration to find the repository on your local filesystem.

The dashboard errors if you haven't defined `repoPaths` in your configuration or if the dashboard
can't determine where the repository for this PR is located using that setting.

If the dashboard is able to locate the repository for the PR on your local filesystem, it uses the
`gh pr checkout` command to checkout the PR locally.

## `d` - View PR Diff { #view-pr-diff }

Press ![kbd:`d`]() to display the PR's diff in the terminal. The dashboard uses the `pager.diff`
setting in your configuration, which defaults to `less`, to display the diff.

The dashboard view is replaced by PR's change diff displayed with the configured pager. When you
exit the pager, the view returns to the dashboard.

```alert
---
variant: warning
---
There's a known bug when using this command on Windows. When you do, the diff
is sent to the terminal but the dashboard doesn't wait for you it to exit.

Instead, the diff is displayed in your terminal output without paging when you
exit the dashboard.
```

## `m` - Merge PR { #merge-pr }

Press ![kbd:`m`]() to merge the PR. When you do, the dashboard uses the `gh pr merge` command to
merge the PR.

```alert
---
variant: danger
---
**Prior to v3.10.0:** When you use this command, the dashboard merges the PR immediately and without
prompting for confirmation. Only use this command when you're sure you want to
merge the PR.

**Since v3.10.0:** When you use this command, the dashboard displays a confirmation prompt and
merges the PR only after you approve the action.
```

## `u` - Update PR { #update-pr}

Press ![kbd:`u`]() to update the PR branch. When you do, the dashboard uses the
`gh pr update-branch` command to update the PR. This command updates the branch with a merge commit.

## `v` - Approve PR { #approve-pr}

Press ![kbd:`v`]() to approve the PR. When you do, the dashboard uses the
`gh pr review --approve` command to approve the PR. This will prompt you to add an optional comment to the approval.

## `w` - Watch PR checks { #watch-pr-checks}

Press ![kbd:`w`]() to watch the PR check and get a desktop notification if they succeed or fail. When you do, the dashboard uses the
`gh pr checks --watch` command to watch the PR checks.

## `W` - Mark PR as Ready for Review { #mark-pr-as-ready-for-review}

Press ![kbd:`W`]() to mark the PR as ready for review. When you do, the dashboard uses the
`gh pr ready` command to convert the PR from draft status to ready for review.

## `x` - Close PR { #close-pr }

Press ![kbd:`x`]() to close the PR. When you do, the dashboard uses the `gh pr close` command to
close the PR.

```alert
---
variant: warning
---
**Prior to v3.10.0:** When you use this command, the dashboard closes the PR immediately and without
prompting for confirmation. Only use this command when you're sure you want to
close the PR.

**Since v3.10.0:** When you use this command, the dashboard displays a confirmation prompt and
closes the PR only after you approve the action.

This command doesn't support closing the PR with a comment. If you want to add
a comment that explains why you're closing the PR, use the
[comment](#c---comment-on-pr) command before or after you use this one.
```

## `X` - Reopen PR { #reopen-pr }

Press ![kbd:`X`]() to reopen a closed PR. When you do, the dashboard uses the `gh pr reopen`
command to reopen the PR.

```alert
---
variant: warning
---
**Prior to v3.10.0:** When you use this command, the dashboard reopens the PR immediately and without
prompting for confirmation. Only use this command when you're sure you want to
reopen the closed PR.

**Since v3.10.0:** When you use this command, the dashboard displays a confirmation prompt and
reopens the PR only after you approve the action.
```
