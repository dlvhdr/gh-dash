export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "issue.schema.json",
      title: "Issue Section Layout",
      description:
        "Defines the columns an issue section displays in its table.",
      type: "object",
      default: {
        updatedAt: {
          width: 7,
        },
        repo: {
          width: 15,
        },
        creator: {
          width: 10,
        },
        assignees: {
          width: 20,
          hidden: true,
        },
      },
      properties: {
        updatedAt: {
          title: "Issue Updated At Column",
          description:
            "Defines options for the updated at column in an issue section.",
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
          title: "Issue State Column",
          description:
            "Defines options for the state column in an issue section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
        },
        repo: {
          title: "Issue Repo Column",
          description:
            "Defines options for the repo column in an issue section.",
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
          title: "Issue Title Column",
          description:
            "Defines options for the title column in an issue section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
        },
        creator: {
          title: "Issue Creator Column",
          description:
            "Defines options for the creator column in an issue section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
          default: {
            width: 10,
          },
        },
        creatorIcon: {
          title: "Issue Creator Role Icon",
          description:
            "Defines options for the role icon for each issue in an issue section.",
          type: "object",
          properties: {
            hidden: {
              title: "Hide Creator Icon",
              description:
                "Specify whether the role icon for issue creators should be hidden from view.",
              type: "boolean",
            },
          },
        },
        assignees: {
          title: "Issue Assignees Column",
          description:
            "Defines options for the assignees column in an issue section.",
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
        comments: {
          title: "Issue Comments Column",
          description:
            "Defines options for the comments column in an issue section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
        },
        reactions: {
          title: "Issue Reactions Column",
          description:
            "Defines options for the reactions column in an issue section.",
          type: "object",
          oneOf: [
            {
              $ref: "./options.json",
            },
          ],
        },
      },
    }),
  );
}
