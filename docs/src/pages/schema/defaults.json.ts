export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "defaults.schema.json",
      title: "Default Options",
      description: "Define options that can be overridden per-section.",
      type: "object",
      default: {
        preview: {
          open: true,
          width: 0.45,
        },
        prsLimit: 20,
        prApproveComment: "LGTM",
        issuesLimit: 20,
        view: "prs",
        refetchIntervalMinutes: 30,
      },
      properties: {
        layout: {
          title: "Layout Options",
          description:
            "Defines the layout for the work item tables in the dashboard.",
          type: "object",
          properties: {
            prs: {
              $ref: "./layout/pr.json",
            },
            issues: {
              $ref: "./layout/issue.json",
            },
          },
        },
        prsLimit: {
          title: "PR Fetch Limit",
          description:
            "Global limit on the number of PRs fetched for the dashboard",
          type: "integer",
          minimum: 1,
          default: 20,
        },
        issuesLimit: {
          title: "Issue Fetch Limit",
          description:
            "Global limit on the number of issues fetched for the dashboard",
          type: "integer",
          minimum: 1,
          default: 20,
        },
        preview: {
          title: "Preview Pane",
          description: "Defaults for the preview pane",
          type: "object",
          properties: {
            open: {
              title: "Open on Load",
              description:
                "Whether to have the preview pane open by default when the dashboard loads.",
              type: "boolean",
              default: true,
            },
            width: {
              title: "Preview Pane Width",
              description:
                "Specifies the width of the preview pane. Numbers between 0 and 1 represent size relative to overall terminal window size (e.g 0.4 is 40%), numbers >=1 represent size in columns.",
              type: "number",
              minimum: 0,
              default: 0.45,
            },
          },
        },
        refetchIntervalMinutes: {
          title: "Refetch Interval in Minutes",
          description:
            "Specifies how often to refetch PRs and Issues in minutes.",
          type: "integer",
          minimum: 0,
          default: 30,
        },
        dateFormat: {
          title: "Date format",
          description: "Specifies how dates are formatted.",
          type: "integer",
          minimum: 1,
          default: 30,
        },
        view: {
          title: "Default View",
          description:
            "Specifies whether the dashboard should display the PRs or Issues view on load.",
          type: "string",
          enum: ["issues", "prs"],
          default: "prs",
        },
        prApproveComment: {
          title: "PR Approve Comment",
          description: "The default comment prefilled when approving a PR.",
          type: "string",
          default: "LGTM",
        },
      },
    }),
  );
}
