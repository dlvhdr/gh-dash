// Outputs: /schema.json
export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "theme.schema.json",
      title: "Theme Options",
      description: "Theme settings for gh-dash",
      type: "object",
      required: [],
      properties: {
        ui: {
          title: "UI Settings",
          type: "object",
          properties: {
            table: {
              title: "Table Settings",
              type: "object",
              properties: {
                sectionsShowCount: {
                  title: "Sections Show Count",
                  description:
                    "Whether the number of results show up next to each section's title in the tab bar.",
                  type: "boolean",
                  default: true,
                },
                showSeparators: {
                  title: "Show Separators",
                  description:
                    "Whether to show the separators between lines in the prs/issues tables.",
                  type: "boolean",
                  default: true,
                },
                compact: {
                  title: "Compact",
                  description:
                    "Whether to show table rows in a compact way or not",
                  type: "boolean",
                  default: false,
                },
              },
            },
          },
        },
        icons: {
          title: "Theme Icons",
          description: "Defines the author-role icons for the dashboard.",
          type: "object",
          properties: {
            newcontributor: {
              title: "New Contributor Role Icon",
              description:
                "Specifies the character to use as the new-contributor-role icon.",
              type: "string",
            },
            contributor: {
              title: "Contributor Role Icon Color",
              description:
                "Specifies the character to use as the contributor-role icon.",
              type: "string",
            },
            collaborator: {
              title: "Collaborator Role Icon Color",
              description:
                "Specifies the character to use as the collaborator-role icon.",
              type: "string",
            },
            member: {
              title: "Member Role Icon Color",
              description:
                "Specifies the character to use as the member-role icon.",
              type: "string",
            },
            owner: {
              title: "Owner Role Icon Color",
              description:
                "Specifies the character to use as the owner-role icon.",
              type: "string",
            },
            unknownrole: {
              title: "Unknown Role Icon Color",
              description:
                "Specifies the character to use as the unknown-role icon.",
              type: "string",
            },
          },
        },
        colors: {
          title: "Theme Colors",
          description:
            "Defines text, background, and border colors for the dashboard.",
          type: "object",
          required: [],
          properties: {
            text: {
              title: "Text Colors",
              description:
                "Defines the foreground (text) colors for the dashboard.",
              type: "object",
              required: [],
              properties: {
                primary: {
                  title: "Primary Text Color",
                  description:
                    "Specifies the color for active text. Must be a valid hex color, like `#a3c` or `#aa33cc`.",
                  type: "string",
                  default: "#ffffff",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
                secondary: {
                  title: "Secondary Text Color",
                  description:
                    "Specifies the color for important text. Must be a valid hex color, like `#a3c` or `#aa33cc`.",
                  type: "string",
                  default: "#c6c6c6",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
                inverted: {
                  title: "Inverted Text Color",
                  description:
                    "Specifies the color for text on an inverted background. Must be a valid hex color, like `#a3c` or `#aa33cc`.",
                  type: "string",
                  default: "#303030",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
                faint: {
                  title: "Faint Text Color",
                  description:
                    "Specifies the color for informational text. Must be a valid hex color, like `#a3c` or `#aa33cc`.",
                  type: "string",
                  default: "#8a8a8a",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
                warning: {
                  title: "Warning Text Color",
                  description:
                    "Specifies the color for warning or error text. Must be a valid hex color, like `#a3c` or `#aa33cc`.",
                  type: "string",
                  default: "#800000",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
                success: {
                  title: "Success Text Color",
                  description:
                    "Specifies the color for success text. Must be a valid hex color, like `#a3c` or `#aa33cc`.",
                  type: "string",
                  default: "#008000",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
              },
            },
            background: {
              title: "Background Colors",
              description: "Defines the background colors for the dashboard.",
              type: "object",
              required: [],
              properties: {
                selected: {
                  title: "Selected Background Color",
                  description:
                    "Defines the background color for selected items. Must be a valid hex color, like `#a3c` or `#aa33cc`.",
                  type: "string",
                  default: "#808080",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
              },
            },
            border: {
              title: "Border Colors",
              description: "Defines the border colors for the dashboard.",
              type: "object",
              required: [],
              properties: {
                primary: {
                  title: "Primary Border Color",
                  description:
                    "Defines the border color for primary elements. Must be a valid hex color, like `#a3c` or `#aa33cc`.",
                  type: "string",
                  default: "#808080",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
                secondary: {
                  title: "Secondary Border Color",
                  description:
                    "Defines the border color for secondary elements. Must be a valid hex color, like `#a3c` or `#aa33cc`.",
                  type: "string",
                  default: "#c0c0c0",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
                faint: {
                  title: "Faint Border Color",
                  description:
                    "Defines the border color between rows in the table. Must be a valid hex color, like `#a3c` or `#aa33cc`.",
                  type: "string",
                  default: "#000000",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
              },
            },
            icon: {
              title: "Icon Colors",
              description: "Defines author-role icon colors for the dashboard.",
              type: "object",
              properties: {
                newcontributor: {
                  title: "New Contributor Role Icon Color",
                  description:
                    "Specifies the icon color for the new-contributor-role icon.",
                  type: "string",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
                contributor: {
                  title: "Contributor Role Icon Color",
                  description:
                    "Specifies the icon color for the contributor-role icon.",
                  type: "string",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
                collaborator: {
                  title: "Collaborator Role Icon Color",
                  description:
                    "Specifies the icon color for the collaborator-role icon.",
                  type: "string",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
                member: {
                  title: "Member Role Icon Color",
                  description:
                    "Specifies the icon color for the member-role icon.",
                  type: "string",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
                owner: {
                  title: "Owner Role Icon Color",
                  description:
                    "Specifies the icon color for the owner-role icon.",
                  type: "string",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
                unknownrole: {
                  title: "Unknown Role Icon Color",
                  description:
                    "Specifies the icon color for the unknown-role icon.",
                  type: "string",
                  pattern: "^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$",
                },
              },
            },
          },
        },
      },
      default: {
        ui: {
          sectionsShowCount: true,
          table: {
            showSeparators: true,
            compact: false,
          },
        },
        colors: {
          text: {
            primary: "#ffffff",
            secondary: "#c6c6c6",
            inverted: "#303030",
            faint: "#8a8a8a",
            warning: "#800000",
            success: "#008000",
          },
          background: {
            selected: "#808080",
          },
          border: {
            primary: "#808080",
            secondary: "#c0c0c0",
            faint: "#000000",
          },
        },
      },
    }),
  );
}
