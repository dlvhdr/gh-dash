// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import tailwindcss from "@tailwindcss/vite";

// https://astro.build/config
export default defineConfig({
  site: "https://dlvhdr.github.io/gh-dash/",
  integrations: [
    starlight({
      title: "DASH",
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
          items: [
            "getting-started",
            "getting-started/usage",
            {
              label: "Keybindings",
              autogenerate: { directory: "getting-started/keybindings" },
            },
          ],
        },
        {
          label: "Configuration",
          items: [
            "configuration",
            "configuration/examples",
            "configuration/defaults",
            "configuration/pr-section",
            "configuration/issue-section",
            {
              label: "Layout",
              autogenerate: { directory: "configuration/layout" },
            },
            "configuration/keybindings",
            "configuration/searching",
            "configuration/theme",
            "configuration/schema",
          ],
        },
        { slug: "contributing" },
      ],
    }),
  ],
  vite: {
    plugins: [tailwindcss()],
  },
});
