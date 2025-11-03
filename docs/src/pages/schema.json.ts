export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "gh-dash.schema.json",
      title: "Dashboard Configuration",
      description: "Settings for the GitHub Dashboard.",
      type: "object",
      properties: {
        prSections: {
          title: "Pull Request Sections",
          description: "Define sections for the dashboard's PR view.",
          type: "array",
          items: {
            $ref: "./schema/pr-section.json",
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
          type: "array",
          items: {
            $ref: "./schema/issue-section.json",
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
          $ref: "./schema/defaults.json",
        },
        repoPaths: {
          title: "Repo Path Map",
          description:
            "Key-value pairs that match repositories to local file paths.",
          type: "object",
          examples: [
            {
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
            },
            "^[^\\*]+$": {
              type: "string",
              pattern: "^[^\\*]+$",
              title: "Without a Wildcard",
              description:
                "If the repo name (key) doesn't include an asterisk, the path (value) can't either.",
            },
          },
        },
        keybindings: {
          title: "Keybindings",
          description: "Define keybindings to run shell commands.",
          type: "object",
          properties: {
            prs: {
              $ref: "./schema/keybindings/prs.json",
            },
            issues: {
              $ref: "./schema/keybindings/issues.json",
            },
          },
          examples: [
            {
              issues: [
                {
                  key: "P",
                  command:
                    "gh issue pin {{ .IssueNumber }} --repo {{ .RepoName }}",
                },
              ],
            },
            {
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
          $ref: "./schema/theme.json",
        },
        pager: {
          title: "Pager",
          description: "Specify the pager settings to use in the dashboard.",
          type: "object",
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
        },
        smartFilteringAtLaunch: {
          title: "Smart Filtering At Launch",
          description:
            "Set this to `false` to disable [Smart Filtering](/getting-started/smartfiltering) at `gh-dash` launch.\n",
          type: "boolean",
        },
        confirmQuit: {
          title: "Confirm Quit",
          description:
            "Specifies whether the user needs to confirm when quitting `gh-dash`",
          type: "boolean",
          default: "false",
        },
      },
    }),
  );
}
