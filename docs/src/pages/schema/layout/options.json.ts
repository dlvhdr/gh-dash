export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "layout.options.schema.json",
      title: "Valid Layout Options",
      type: "object",
      properties: {
        grow: {
          title: "Grow Column",
          description:
            "Select whether the column should grow to fill available space.",
          type: "boolean",
        },
        width: {
          title: "Column Width",
          description: "Select the column's width by cell count.",
          type: "integer",
          minimum: 0,
        },
        hidden: {
          title: "Hide Column",
          description: "Select whether the column should be hidden from view.",
          type: "boolean",
        },
      },
    }),
  );
}
