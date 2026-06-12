export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "completions.schema.json",
      title: "Completion Commands",
      description: "Keybindings for the Completions Popup",
      type: "array",
      items: {
        $ref: "./entry.json",
      },
    }),
  );
}
