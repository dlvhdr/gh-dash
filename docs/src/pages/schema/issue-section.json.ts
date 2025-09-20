export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "issue-section.schema.json",
      title: "Issue Section Options",
      description: "Defines a section in the dashboard's Issues view.",
      type: "object",
      required: ["title", "filters"],
      properties: {
        title: {
          title: "Issue Title",
          description:
            "Defines the section's name as displayed in the tabs for the issues view.",
          type: "string",
        },
        filters: {
          title: "Issue Filters",
          description:
            "Defines the GitHub search filters for the issues in the section's table.",
          type: "string",
        },
        layout: { $ref: "./layout/issue.json", schematize: { weight: 3 } },
        limit: {
          title: "Issue Fetch Limit",
          type: "integer",
          minimum: 1,
        },
      },
    }),
  );
}
