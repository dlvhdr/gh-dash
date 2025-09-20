// Outputs: /schema.json
export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "pr-section.schema.json",
      title: "PR Section Options",
      description: "Defines a section in the dashboard's PRs view.",
      type: "object",
      required: ["title", "filters"],
      properties: {
        title: {
          title: "PR Title",
          description:
            "Defines the section's name as displayed in the tabs for the PRs view.",
          type: "string",
        },
        filters: {
          title: "PR Filters",
          description:
            "Defines the GitHub search filters for the PRs in the section's table.",
          type: "string",
        },
        layout: {
          $ref: "./layout/pr.json",
        },
        limit: {
          title: "PR Fetch Limit",
          type: "integer",
          minimum: 1,
        },
      },
    }),
  );
}
