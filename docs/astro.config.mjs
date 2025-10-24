// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import tailwindcss from "@tailwindcss/vite";
import astroBrokenLinksChecker from "astro-broken-links-checker";

import node from "@astrojs/node";

const ogUrl = new URL("og.png", "https://gh-dash.dev/").href;
const ogImageAlt = "DASH Through Your GitHub";

// https://astro.build/config
export default defineConfig({
  site: "https://gh-dash.dev",
  integrations: [
    astroBrokenLinksChecker({
      logFilePath: "broken-links.log",
      checkExternalLinks: false,
    }),
    starlight({
      title: "DASH",
      favicon: "/favicon.png",
      customCss: ["./src/styles/custom.css", "./src/fonts/font-face.css"],
      head: [
        {
          tag: "meta",
          attrs: { property: "og:image", content: ogUrl },
        },
        {
          tag: "meta",
          attrs: { property: "og:image:alt", content: ogImageAlt },
        },
        {
          tag: "meta",
          attrs: {
            name: "description",
            content:
              "DASH - a rich terminal UI for GitHub that doesn't break your flow",
          },
        },
      ],
      components: {
        Header: "./src/components/Header.astro",
        PageTitle: "./src/components/Title.astro",
      },
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/dlvhdr/gh-dash",
        },
      ],
      sidebar: [
        {
          label: "Getting Started",
          items: [
            "getting-started",
            "getting-started/usage",
            "getting-started/updating",
          ],
        },
        {
          label: "Keybindings",
          collapsed: true,
          autogenerate: { directory: "getting-started/keybindings" },
        },
        {
          label: "Configuration",
          collapsed: true,
          items: [
            "configuration",
            "configuration/schema",
            "configuration/defaults",
            "configuration/searching",
            "configuration/pr-section",
            "configuration/issue-section",
            "configuration/repo-paths",
            "configuration/keybindings",
            "configuration/theme",
            "configuration/reusing",
            "configuration/examples",
            {
              label: "Layout",
              items: [
                "configuration/layout/options",
                "configuration/layout/pr",
                "configuration/layout/issue",
              ],
            },
          ],
        },
        { slug: "contributing" },
        { slug: "donating" },
        {
          label: "Insiders 🌟",
          items: [
            "insiders",
            {
              label: "ENHANCE",
              collapsed: true,
              items: [
                "insiders/enhance/getting-started",
                "insiders/enhance/usage",
                "insiders/enhance/keybindings",
                "insiders/enhance/dash-integration",
                "insiders/enhance/theme",
              ],
            },
          ],
        },
      ],
    }),
  ],

  vite: {
    plugins: [tailwindcss()],
  },
  adapter: node({
    mode: "standalone",
  }),
});
