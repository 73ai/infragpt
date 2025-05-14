import { defineConfig } from "astro/config";
import tailwind from "@astrojs/tailwind";
import sitemap from "@astrojs/sitemap";
import rehypeAutolinkHeadings from "rehype-autolink-headings";
import { autolinkConfig } from "./plugins/rehype-autolink-config";
import rehypeSlug from "rehype-slug";
import alpinejs from "@astrojs/alpinejs";
import AstroPWA from "@vite-pwa/astro";
import icon from "astro-icon";

export default defineConfig({
    site: "https://infragpt.io",
    vite: {
        define: {
            __DATE__: `'${new Date().toISOString()}'`,
        },
    },
    integrations: [
        tailwind(),
        sitemap(),
        alpinejs(),
        AstroPWA({
            mode: "production",
            base: "/",
            scope: "/",
            registerType: "autoUpdate",
            manifest: {
                name: "InfraGPT",
                short_name: "InfraGPT - AI SRE Copilot for the Cloud",
                theme_color: "#ffffff",
            },
            workbox: {
                navigateFallback: "/404",
                globPatterns: ["*.js"],
            },
            devOptions: {
                enabled: false,
                navigateFallbackAllowlist: [/^\/404$/],
                suppressWarnings: true,
            },
        }),
        icon(),
    ],
    markdown: {
        rehypePlugins: [
            rehypeSlug,
            // This adds links to headings
            [rehypeAutolinkHeadings, autolinkConfig],
        ],
    },
    // Removed the experimental section
});