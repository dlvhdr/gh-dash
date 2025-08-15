// Outputs: /schema.json
export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "gh-dash.schema.yaml",
      title: "Dashboard Configuration",
      description: "Settings for the GitHub Dashboard.",
      type: "object",
      properties: {
        prSections: {
          title: "Pull Request Sections",
          description: "Define sections for the dashboard's PR view.",
          schematize: {
            weight: 1,
            details:
              "The `prSections` setting defines one or more sections to display in the dashboard's PRs\nview as tabs. Each section needs a title, which is displayed as the tab name for the\nsection, and a [GitHub search filter]. The dashboard queries GitHub with the search filter\nto populate the list of PRs to display for that section.\n\nFor more information about defining a PR section, see [sref:PR Section Options].\n\n[GitHub search filter]: https://docs.github.com/en/search-github/searching-on-github/searching-issues-and-pull-requests\n[sref:PR Section Options]: pr-section\n",
            default: {
              details:
                "By default, the PRs view on the dashboard has three sections:\n\n- The ![styled:`My Pull Requests`]() section fetches open PRs authored by you.\n- The ![styled:`Needs My Review`]() section fetches open PRs where you are a requested\n  reviewer.\n- The ![styled:`Involved`]() section fetches open PRs authored by someone else and that\n  meet at least one of the following criteria:\n\n  - The PR is assigned to you.\n  - The PR's body, a comment on the PR, or one of the PR's commit messages mentions you.\n  - You commented on the PR.\n",
              format: "yaml",
            },
          },
          type: "array",
          items: {
            $ref: "./pr-section.yaml",
          },
          default: [
            {
              title: "My Pull Requests",
              filters: "is:open author:@me",
            },
            {
              title: "Needs My Review",
              filters: "is:open review-requested:@me",
            },
            {
              title: "Involved",
              filters: "is:open involves:@me -author:@me",
            },
          ],
        },
        issuesSections: {
          title: "Issue Sections",
          description: "Define sections for the dashboard's Issues view.",
          schematize: {
            weight: 2,
            details:
              "The `issueSections` setting defines one or more sections to display in the dashboard's\nIssues view as tabs. Each section needs a title, which is displayed as the tab name for the\nsection, and a [GitHub search filter]. The dashboard queries GitHub with the search filter\nto populate the list of issues to display for that section.\n\nFor more information about defining an issue section, see [sref:Issue Section Options].\n\n[GitHub search filter]: https://docs.github.com/en/search-github/searching-on-github/searching-issues-and-pull-requests\n[sref:Issue Section Options]: issue-section\n",
            default: {
              details:
                "By default, the Issues view on the dashboard has three sections:\n\n- The ![styled:`My Issues`]() section fetches open issues authored by you.\n- The ![styled:`Assigned`]() section fetches open issues assigned to you.\n- The ![styled:`Involved`]() section fetches open issues authored by someone else and that\n  meet at least one of the following criteria:\n\n  - The issue is assigned to you.\n  - The issue's body or a comment on the issue mentions you.\n  - You commented on the issue.\n",
            },
            format: "yaml",
          },
          type: "array",
          items: {
            $ref: "./issue-section.yaml",
          },
          default: [
            {
              title: "My Issues",
              filters: "is:open author:@me",
            },
            {
              title: "Assigned",
              filters: "is:open assignee:@me",
            },
            {
              title: "Involved",
              filters: "is:open involves:@me -author:@me",
            },
          ],
        },
        defaults: {
          $ref: "./defaults.yaml",
          schematize: {
            weight: 3,
          },
        },
        repoPaths: {
          title: "Repo Path Map",
          description:
            "Key-value pairs that match repositories to local file paths.",
          schematize: {
            weight: 4,
            details:
              "You can use the `repoPaths` setting to map repository names (as keys) to local paths\n(as values). This map coonfigures where to locate repositories when checking out PRs.\n\nThe mappings can be exact matches, like the full name or path of a repository, or wildcard\nmatches using the `owner` and a partial path.\n\nAn exact match for the full repository name to a full path takes priority over a matching\nwildcard. Wildcard keys must match to a wildcard path.\n\nThe `RepoName` and `RepoPath` keybinding arguments are fully expanded when sent to the\ncommand.\n",
            skip_schema_render: true,
            example_format: "yaml",
          },
          type: "object",
          examples: [
            {
              schematize: {
                title: "Matching Repositories",
                details:
                  "In this example, the first key is defined with a wildcard (`*`) as `dlvhdr/*`, so the\ndashboard will use this entry to resolve the path for any repo in the `dlvhdr`\nnamespace, like [`dlvhdr/jb`] or [`dlvhdr/harbor`]. Note that the value for this key\nalso uses a wildcard. If the key specifies a wildcard, the value must specify one too.\n\nIf a repository in the `dlvhdr` namespace has been cloned into the `~/code/repos`\nfolder, the dashboard will be able to checkout PRs for that repository locally.\n\nIf the repository isn't found, the [checkout command] raises an error.\n\nThe second key defines an exact mapping between the `dlvhdr/gh-dash` repository and the\n`~/code/gh-dash` folder. Because this is a non-wildcard mapping, the dashboard will use\nthis value to resolve the repository path for `dlvhdr/gh-dash` even though there's also\nan entry for `dlvhdr/*`.\n",
              },
              "dlvhdr/*": "~/code/repos/*",
              "dlvhdr/gh-dash": "~/code/gh-dash",
            },
          ],
          patternProperties: {
            "\\*$": {
              type: "string",
              pattern: "\\*$",
              title: "With a Wildcard",
              description:
                "If the repo name (key) includes an asterisk, the path (value) must too.",
              schematize: {
                href: "ends-with-wildcard",
                no_pattern_in_heading: true,
                weight: 1,
                details:
                  "If the key for an `repoPath` entry ends with a wildcard (`*`), its value must also have\na wildcard. If a key ends with a wildcard but the value doesn't, `gh-dash` won't be\nable to correctly map repositories to folders.\n",
              },
            },
            "^[^\\*]+$": {
              type: "string",
              pattern: "^[^\\*]+$",
              title: "Without a Wildcard",
              description:
                "If the repo name (key) doesn't include an asterisk, the path (value) can't either.",
              schematize: {
                href: "no-asterisks",
                no_pattern_in_heading: true,
                weight: 2,
                details:
                  "If the key for an `repoPath` entry doesn't have a wildcard (`*`), its value must not\nhave a wildcard. If a key ends without a wildcard but the value does, `gh-dash` won't\nbe able to correctly map repositories to folders.\n",
              },
            },
          },
        },
        keybindings: {
          title: "Keybindings",
          description: "Define keybindings to run shell commands.",
          schematize: {
            details:
              "Define your own custom keybindings to run shell commands using [Go templates]. You can define\nyour keybindings for the PRs and Issues views separately.\n",
            skip_schema_render: true,
            example_format: "yaml",
            weight: 5,
          },
          type: "object",
          properties: {
            prs: {
              $ref: "./keybindings/prs.yaml",
              schematize: {
                weight: 1,
              },
            },
            issues: {
              $ref: "./keybindings/issues.yaml",
              schematize: {
                weight: 2,
              },
            },
          },
          examples: [
            {
              schematize: {
                title: "Pin an Issue",
                details:
                  "This example binds <kbd>P</kbd> in the Issues view of the dashboard to the\n`gh issue pin` command to pin the selected issue in the repository.\n",
              },
              issues: [
                {
                  key: "P",
                  command:
                    "gh issue pin {{ .IssueNumber }} --repo {{ .RepoName }}",
                },
              ],
            },
            {
              schematize: {
                title: "Review PRs",
                details:
                  "This example binds <kbd>c`]() and ![kbd:`v</kbd> in the PRs view of the dashboard to\ncheckout the selected PR and open the repository in Neovim and VS Code respectively.\n\nThe Neovim keybinding opens the editor in another `tmux` window.\n",
              },
              prs: [
                {
                  key: "c",
                  command:
                    "tmux new-window -c {{.RepoPath}} '\n  gh pr checkout {{.PrNumber}} &&\n  nvim -c \":DiffviewOpen master...{{.HeadRefName}}\"\n'\n",
                },
                {
                  key: "v",
                  command:
                    "cd {{.RepoPath}} && code . && gh pr checkout {{.PrNumber}}\n",
                },
              ],
            },
          ],
        },
        theme: {
          $ref: "./theme.yaml",
          schematize: {
            weight: 6,
          },
        },
        pager: {
          title: "Pager",
          description: "Specify the pager settings to use in the dashboard.",
          type: "object",
          schematize: {
            skip_schema_render: true,
            weight: 7,
          },
          properties: {
            diff: {
              title: "Diff Pager",
              description: "Specifies the pager to use when diffing.",
              type: "string",
              anyOf: [
                {
                  type: "string",
                },
                {
                  enum: ["less", "delta"],
                },
              ],
              default: "less",
            },
          },
        },
        showAuthorIcons: {
          title: "Show Author Role Icons",
          description:
            "Specifies whether to show author-role icons in the dashboard.\nSet this value to `false` to hide the author-role icons.\nSee the [Theme Icons](theme#icons) section and\n[Icon Colors](theme#colors.icons) section for\nconfiguration options you can set to change the author-role icons and author-role icon colors.\n",
          type: "boolean",
          schematize: {
            weight: 8,
          },
        },
        smartFilteringAtLaunch: {
          title: "Smart Filtering At Launch",
          description:
            "Set this to `false` to disable [Smart Filtering](/getting-started/smartfiltering) at `gh-dash` launch.\n",
          type: "boolean",
          schematize: {
            weight: 8,
          },
        },
      },
    }),
  );
}
