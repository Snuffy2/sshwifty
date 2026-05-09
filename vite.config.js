// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import vue from "@vitejs/plugin-vue";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vite";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)));
const distDir = path.join(repoRoot, ".tmp", "dist");
const backendTarget = "http://127.0.0.1:8182";
const publicDir = path.join(repoRoot, "ui", "public");

const copiedRootFiles = [
  "README.md",
  "CONFIGURATION.md",
  "DEPENDENCIES.md",
  "LICENSE.md",
];
const generatedHtmlFiles = ["index.html", "error.html"];

const passthroughPublicAssets = new Map([
  ["/sshwifty/assets/site.webmanifest", "site.webmanifest"],
  ["/sshwifty/assets/sshwifty.svg", "sshwifty.svg"],
  ["/sshwifty/assets/robots.txt", "robots.txt"],
]);
const rootCompatibilityAssets = new Map([
  ["/favicon.ico", "favicon.ico"],
  ["/manifest.json", "site.webmanifest"],
  ["/browserconfig.xml", "browserconfig.xml"],
]);
const passthroughContentTypes = new Map([
  ["site.webmanifest", "application/manifest+json"],
  ["browserconfig.xml", "application/xml; charset=utf-8"],
  ["sshwifty.svg", "image/svg+xml"],
  ["favicon.ico", "image/x-icon"],
]);

/**
 * Build a placeholder token for a public asset path.
 *
 * @param {string} publicPath Browser-facing public asset path.
 * @returns {string} Placeholder token that Vite will not rewrite.
 */
function passthroughAssetToken(publicPath) {
  return `__SSHWIFTY_PUBLIC_ASSET__${publicPath}__`;
}

/**
 * Replace Sshwifty public asset paths in a source string.
 *
 * @param {string} source Source text to scan.
 * @param {(publicPath: string) => string} replacer Replacement callback.
 * @returns {string} Source text with public asset paths replaced.
 */
function replacePublicAssetPaths(source, replacer) {
  let updated = source;

  for (const publicPath of passthroughPublicAssets.keys()) {
    updated = updated.replaceAll(publicPath, replacer(publicPath));
  }

  return updated;
}

/**
 * Restore public asset placeholders back to their browser-facing paths.
 *
 * @param {string} source Source text to scan.
 * @returns {string} Source text with passthrough tokens restored.
 */
function restorePublicAssetTokens(source) {
  let updated = source;

  for (const publicPath of passthroughPublicAssets.keys()) {
    updated = updated.replaceAll(passthroughAssetToken(publicPath), publicPath);
  }

  return updated;
}

/**
 * Create a Vite plugin that copies root documentation files into the bundle.
 *
 * @returns {import("vite").Plugin} Vite build plugin.
 */
function copyRootFilesPlugin() {
  return {
    name: "copy-root-files",
    apply: "build",
    closeBundle() {
      fs.mkdirSync(distDir, { recursive: true });
      for (const fileName of copiedRootFiles) {
        fs.copyFileSync(
          path.join(repoRoot, fileName),
          path.join(distDir, fileName),
        );
      }
    },
  };
}

/**
 * Create a Vite plugin that flattens generated HTML output paths.
 *
 * @returns {import("vite").Plugin} Vite build plugin.
 */
function normalizeHtmlOutputsPlugin() {
  return {
    name: "normalize-html-outputs",
    apply: "build",
    closeBundle() {
      const nestedHtmlDir = path.join(distDir, "ui");

      for (const fileName of generatedHtmlFiles) {
        const nestedPath = path.join(nestedHtmlDir, fileName);
        const flatPath = path.join(distDir, fileName);

        if (!fs.existsSync(nestedPath)) {
          continue;
        }

        let html = fs.readFileSync(nestedPath, "utf8");
        html = restorePublicAssetTokens(html);

        fs.writeFileSync(flatPath, html);
        fs.rmSync(nestedPath);
      }

      if (
        fs.existsSync(nestedHtmlDir) &&
        fs.readdirSync(nestedHtmlDir).length === 0
      ) {
        fs.rmdirSync(nestedHtmlDir);
      }
    },
  };
}

/**
 * Create a Vite plugin for Sshwifty's fixed public asset routes.
 *
 * @returns {import("vite").Plugin} Vite plugin for build and dev asset paths.
 */
function sshwiftyPublicAssetsPlugin() {
  return {
    name: "sshwifty-public-assets",
    enforce: "pre",
    transformIndexHtml: {
      order: "pre",
      handler(html) {
        return replacePublicAssetPaths(html, passthroughAssetToken);
      },
    },
    generateBundle(_, bundle) {
      for (const chunk of Object.values(bundle)) {
        if (chunk.type !== "asset" || typeof chunk.source !== "string") {
          continue;
        }

        chunk.source = restorePublicAssetTokens(chunk.source);
      }
    },
    configureServer(server) {
      const rewriteShellRequest = (req, _res, next) => {
        const requestUrl = req.url ?? "";
        const [requestPath, requestQuery = ""] = requestUrl.split("?", 2);

        if (
          requestPath === "/" ||
          requestPath === "/sshwifty/assets" ||
          requestPath === "/sshwifty/assets/"
        ) {
          req.url =
            "/sshwifty/assets/ui/index.html" +
            (requestQuery.length > 0 ? `?${requestQuery}` : "");
          next();
          return;
        }
        next();
      };

      server.middlewares.use(rewriteShellRequest);
      server.middlewares.use((req, res, next) => {
        const requestUrl = req.url ?? "";
        const [requestPath] = requestUrl.split("?", 1);

        const assetFile =
          passthroughPublicAssets.get(requestPath) ??
          rootCompatibilityAssets.get(requestPath);

        if (!assetFile) {
          next();
          return;
        }

        const filePath = path.join(publicDir, assetFile);
        const contentType =
          passthroughContentTypes.get(assetFile) ?? "text/plain; charset=utf-8";

        res.setHeader("Content-Type", contentType);
        const stream = fs.createReadStream(filePath);
        stream.on("error", (error) => {
          console.error(
            `Failed to stream public asset ${assetFile} from ${filePath}:`,
            error,
          );
          if (!res.headersSent) {
            res.statusCode = error.code === "ENOENT" ? 404 : 500;
          }
          res.end();
        });
        stream.pipe(res);
      });
    },
  };
}

/**
 * Create a Vite plugin that exposes required Node globals in the browser.
 *
 * @returns {import("vite").Plugin} Vite plugin for browser polyfills.
 */
function browserNodePolyfillsPlugin() {
  const virtualModuleId = "virtual:sshwifty-node-globals";
  const resolvedVirtualModuleId = `\0${virtualModuleId}`;

  return {
    name: "sshwifty-browser-node-polyfills",
    resolveId(id) {
      if (id === virtualModuleId) {
        return resolvedVirtualModuleId;
      }

      return null;
    },
    load(id) {
      if (id !== resolvedVirtualModuleId) {
        return null;
      }

      return `
import { Buffer as __Buffer } from "buffer";
import __process from "process";

globalThis.Buffer ??= __Buffer;
globalThis.process ??= __process;
`;
    },
    transformIndexHtml: {
      order: "pre",
      handler() {
        return [
          {
            tag: "script",
            attrs: {
              type: "module",
            },
            children: `import "virtual:sshwifty-node-globals";`,
            injectTo: "head-prepend",
          },
        ];
      },
    },
  };
}

/**
 * Create a Vite plugin that restores passthrough asset paths in dev HTML.
 *
 * @param {string} command Current Vite command.
 * @returns {import("vite").Plugin} Vite HTML transform plugin.
 */
function restoreDevHtmlAssetsPlugin(command) {
  return {
    name: "restore-dev-html-assets",
    transformIndexHtml: {
      order: "post",
      handler(html) {
        if (command === "build") {
          return html;
        }

        return restorePublicAssetTokens(html);
      },
    },
  };
}

export default defineConfig(({ command, mode }) => ({
  base: "/sshwifty/assets/",
  plugins: [
    vue(),
    browserNodePolyfillsPlugin(),
    copyRootFilesPlugin(),
    sshwiftyPublicAssetsPlugin(),
    restoreDevHtmlAssetsPlugin(command),
    normalizeHtmlOutputsPlugin(),
  ],
  publicDir,
  resolve: {
    alias: [
      {
        find: /^~(.*)$/,
        replacement: "$1",
      },
      {
        find: "vue",
        replacement: "vue/dist/vue.esm-bundler.js",
      },
      {
        find: /^stream$/,
        replacement: "stream-browserify",
      },
      {
        find: /^buffer$/,
        replacement: "buffer",
      },
      {
        find: /^events$/,
        replacement: "events",
      },
      {
        find: /^string_decoder$/,
        replacement: "string_decoder",
      },
      {
        find: /^process$/,
        replacement: "process/browser",
      },
    ],
  },
  define: {
    __VUE_OPTIONS_API__: JSON.stringify(true),
    __VUE_PROD_DEVTOOLS__: JSON.stringify(false),
    __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: JSON.stringify(false),
    "process.env.NODE_ENV": JSON.stringify(mode),
  },
  build: {
    target: ["es2020"],
    outDir: distDir,
    emptyOutDir: true,
    sourcemap: command === "serve",
    rollupOptions: {
      input: {
        index: path.join(repoRoot, "ui", "index.html"),
        error: path.join(repoRoot, "ui", "error.html"),
      },
      output: {
        entryFileNames: "[name]-[hash].js",
        chunkFileNames: "chunk-[hash].js",
        assetFileNames: "asset-[hash][extname]",
      },
    },
  },
  server: {
    host: "127.0.0.1",
    port: 5173,
    strictPort: true,
    proxy: {
      "/sshwifty/socket": {
        target: backendTarget,
        ws: true,
      },
    },
  },
  test: {
    include: ["ui/**/*_test.js"],
    globals: true,
    environment: "node",
  },
}));
