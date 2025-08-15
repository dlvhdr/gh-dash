// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

// https://astro.build/config
export default defineConfig({
  site: "https://dlvhdr.github.io/gh-dash/",
  integrations: [
    starlight({
      title: "Dash",
      customCss: ["./src/styles/custom.css"],
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/dlvhdr/gh-dash",
        },
      ],
      sidebar: [
        {
          label: "Start Here",
          autogenerate: { directory: "getting-started" },
        },
        {
          label: "Configuration",
          autogenerate: { directory: "configuration" },
        },
        { slug: "contributing" },
      ],
    }),
  ],
});
