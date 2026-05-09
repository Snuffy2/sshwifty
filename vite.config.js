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

function passthroughAssetToken(publicPath) {
  return `__SSHWIFTY_PUBLIC_ASSET__${publicPath}__`;
}

function replacePublicAssetPaths(source, replacer) {
  let updated = source;

  for (const publicPath of passthroughPublicAssets.keys()) {
    updated = updated.replaceAll(publicPath, replacer(publicPath));
  }

  return updated;
}

function restorePublicAssetTokens(source) {
  let updated = source;

  for (const publicPath of passthroughPublicAssets.keys()) {
    updated = updated.replaceAll(
      passthroughAssetToken(publicPath),
      publicPath,
    );
  }

  return updated;
}

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

      if (fs.existsSync(nestedHtmlDir) && fs.readdirSync(nestedHtmlDir).length === 0) {
        fs.rmdirSync(nestedHtmlDir);
      }
    },
  };
}

function sshwiftyPublicAssetsPlugin(command) {
  return {
    name: "sshwifty-public-assets",
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

      server.middlewares.stack.unshift({
        route: "",
        handle: rewriteShellRequest,
      });

      server.middlewares.use((req, res, next) => {
        const requestUrl = req.url ?? "";
        const [requestPath] = requestUrl.split("?", 1);

        const assetFile = passthroughPublicAssets.get(requestPath);

        if (!assetFile) {
          next();
          return;
        }

        const filePath = path.join(publicDir, assetFile);
        const contentType =
          assetFile === "site.webmanifest"
            ? "application/manifest+json"
            : assetFile.endsWith(".svg")
              ? "image/svg+xml"
              : "text/plain; charset=utf-8";

        res.setHeader("Content-Type", contentType);
        fs.createReadStream(filePath).pipe(res);
      });
    },
  };
}

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
    copyRootFilesPlugin(),
    sshwiftyPublicAssetsPlugin(command),
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
    ],
  },
  define: {
    __VUE_OPTIONS_API__: JSON.stringify(true),
    __VUE_PROD_DEVTOOLS__: JSON.stringify(false),
    __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: JSON.stringify(false),
    "process.env.NODE_ENV": JSON.stringify(mode),
  },
  build: {
    outDir: distDir,
    emptyOutDir: true,
    sourcemap: true,
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
