export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "prs.schema.json",
      title: "PRs Commands",
      description: "Keybindings for the Pull Request View",
      type: "array",
      items: {
        $ref: "./entry.json",
      },
    }),
  );
}
