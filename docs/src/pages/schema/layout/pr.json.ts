export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "pr.schema.json",
      title: "PR Section Layout",
      description: "Defines the columns a PR section displays in its table.",
      type: "object",
      default: {
        updatedAt: {
          width: 7,
        },
        repo: {
          width: 15,
        },
        author: {
          width: 15,
        },
        assignees: {
          width: 20,
          hidden: true,
        },
        base: {
          width: 15,
          hidden: true,
        },
        lines: {
          width: 16,
        },
      },
      properties: {
        updatedAt: {
          title: "PR Updated At Column",
          description:
            "Defines options for the updated at column in a PR section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
          default: {
            width: 7,
          },
        },
        state: {
          title: "PR State Column",
          description: "Defines options for the state column in a PR section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
        },
        repo: {
          title: "PR Repo Column",
          description: "Defines options for the repo column in a PR section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
          default: {
            width: 15,
          },
        },
        title: {
          title: "PR Title Column",
          description: "Defines options for the title column in a PR section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
        },
        author: {
          title: "PR Author Column",
          description: "Defines options for the author column in a PR section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
          properties: null,
          default: {
            width: 15,
          },
        },
        authorIcon: {
          title: "PR Author Role Icon",
          description:
            "Defines options for the role icon for each PR in a PR section.",
          type: "object",
          properties: {
            hidden: {
              title: "Hide Author Role Icon",
              description:
                "Specify whether the role icon for PR authors should be hidden from view.",
              type: "boolean",
            },
          },
        },
        assignees: {
          title: "PR Assignees Column",
          description:
            "Defines options for the assignees column in a PR section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
          default: {
            width: 20,
            hidden: true,
          },
        },
        base: {
          title: "PR Base Column",
          description: "Defines options for the base column in a PR section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
          default: {
            width: 15,
            hidden: true,
          },
        },
        reviewStatus: {
          title: "PR Review Status Column",
          description:
            "Defines options for the review status column in a PR section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
        },
        ci: {
          title: "PR Continuous Integration Column",
          description: "Defines options for the ci column in a PR section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
        },
        lines: {
          title: "PR Lines Column",
          description: "Defines options for the lines column in a PR section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
          default: {
            width: 16,
          },
        },
      },
    }),
  );
}
