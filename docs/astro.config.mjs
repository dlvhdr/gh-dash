// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import tailwindcss from "@tailwindcss/vite";
import astroBrokenLinksChecker from "astro-broken-links-checker";

// https://astro.build/config
export default defineConfig({
  site: "https://gh-dash.dev",
  integrations: [
    astroBrokenLinksChecker({
      logFilePath: "broken-links.log", // Optional: specify the log file path
      checkExternalLinks: false, // Optional: check external links (currently, caching to disk is not supported, and it is slow )
    }),
    starlight({
      title: "DASH",
      customCss: ["./src/styles/custom.css", "./src/fonts/font-face.css"],
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
            "configuration/schema",
            "configuration/defaults",
            "configuration/searching",
            "configuration/pr-section",
            "configuration/issue-section",
            "configuration/repo-paths",
            "configuration/keybindings",
            "configuration/theme",
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
      ],
    }),
  ],
  vite: {
    plugins: [tailwindcss()],
  },
});
