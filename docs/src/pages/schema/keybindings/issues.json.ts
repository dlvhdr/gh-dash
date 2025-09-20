export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "issues.schema.json",
      title: "Issues Commands",
      description: "Keybindings for the Issues View",
      type: "array",
      items: {
        $ref: "./entry.json",
      },
    }),
  );
}
