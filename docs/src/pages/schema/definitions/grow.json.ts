export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "grow.schema.json",
      title: "Grow Column",
      description:
        "Select whether the column should grow to fill available space.",
      type: "boolean",
    }),
  );
}
