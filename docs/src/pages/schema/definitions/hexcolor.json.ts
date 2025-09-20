export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "hexcolor.schema.json",
      title: "Hex Color",
      description: "Represents a valid hex color, like `#a3c` or `#aa33cc`.",
      type: "string",
      format: "hexcolor",
      pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
      example: "#aa33cc",
    }),
  );
}
