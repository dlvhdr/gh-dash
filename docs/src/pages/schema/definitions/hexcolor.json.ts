export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "color.schema.json",
      title: "Color",
      description:
        "Represents a valid color: either a hex color like `#a3c` or `#aa33cc`, or an ANSI color index from `0` to `255`.",
      type: "string",
      format: "color",
      pattern:
        "^(#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})|([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]))$",
      example: "#aa33cc",
    }),
  );
}
