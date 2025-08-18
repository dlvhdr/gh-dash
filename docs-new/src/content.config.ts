import { defineCollection, z } from "astro:content";
import { docsLoader } from "@astrojs/starlight/loaders";
import { docsSchema } from "@astrojs/starlight/schema";

// 2. Import loader(s)
import { glob } from "astro/loaders";

// 3. Define your collection(s)

const properties = z.record(
  z.string(),
  z.union([
    z.object({
      title: z.string(),
      type: z.string(),
      description: z.string().optional(),
      schematize: z
        .object({
          weight: z.number().optional(),
          details: z.string().optional(),
          format: z.string().optional(),
          default: z
            .object({
              details: z.string().optional(),
              format: z.string().optional(),
              skip_schema_render: z.boolean().optional(),
            })
            .optional(),
        })
        .optional(),
      default: z
        .union([
          z.string(),
          z.record(z.string(), z.unknown()),
          z.number(),
          z.array(z.object({ title: z.string(), filters: z.string() })),
        ])
        .optional(),
      properties: z.record(z.string(), z.unknown()).optional().nullable(),
    }),
    z.object({ $ref: z.string() }),
  ]),
);

const yamlSchemas = defineCollection({
  loader: glob({ pattern: "**/*.yaml", base: "./src/data/schemas" }),
  schema: z.object({
    $schema: z.string(),
    $id: z.string(),
    title: z.string(),
    description: z.string().optional(),
    type: z.string(),
    minimum: z.number().optional(),
    schematize: z
      .object({
        weight: z.number().optional(),
        format: z.string().optional(),
        details: z.string(),
        default: z
          .object({ details: z.string(), format: z.string().optional() })
          .optional(),
      })
      .optional(),
    properties: properties.optional(),
  }),
});

export const collections = {
  docs: defineCollection({ loader: docsLoader(), schema: docsSchema() }),
  yamlSchemas,
};
