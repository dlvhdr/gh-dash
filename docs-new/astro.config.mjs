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
            "configuration/keybindings",
            "configuration/defaults",
            "configuration/issue-section",
            "configuration/pr-section",
            {
              label: "Layout",
              autogenerate: { directory: "configuration/layout" },
            },
            "configuration/schema",
            "configuration/searching",
            "configuration/theme",
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

