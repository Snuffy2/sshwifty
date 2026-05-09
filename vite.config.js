// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import vue from "@vitejs/plugin-vue";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vite";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)));
const uiRoot = path.join(repoRoot, "ui");
const distDir = path.join(repoRoot, ".tmp", "dist");
const backendTarget = "http://127.0.0.1:8182";
const publicDir = path.join(uiRoot, "public");
const defaultSourceURL = "https://github.com/Snuffy2/sshwifty";

/**
 * Resolve and validate the source URL embedded into the frontend.
 *
 * @param {NodeJS.ProcessEnv} env Environment variables.
 * @returns {string} HTTPS source URL to expose in the UI.
 */
export function resolveSourceURL(env = process.env) {
  const sourceURL = env.SSHWIFTY_SOURCE_URL ?? defaultSourceURL;

  if (sourceURL.trim() !== sourceURL || sourceURL.length === 0) {
    throw new Error("SSHWIFTY_SOURCE_URL must be a non-empty URL");
  }

  let parsedURL;
  try {
    parsedURL = new URL(sourceURL);
  } catch {
    throw new Error("SSHWIFTY_SOURCE_URL must be a valid URL");
  }

  if (parsedURL.protocol !== "https:") {
    throw new Error("SSHWIFTY_SOURCE_URL must use https:");
  }

  if (parsedURL.username !== "" || parsedURL.password !== "") {
    throw new Error("SSHWIFTY_SOURCE_URL must not include credentials");
  }

  return sourceURL;
}

const sourceURL = resolveSourceURL();

const copiedRootFiles = [
  "README.md",
  "CONFIGURATION.md",
  "DEPENDENCIES.md",
  "LICENSE.md",
];
const fixedPublicAssets = new Map([
  ["/sshwifty/assets/site.webmanifest", "site.webmanifest"],
  ["/sshwifty/assets/sshwifty.svg", "sshwifty.svg"],
  ["/sshwifty/assets/robots.txt", "robots.txt"],
]);
const fixedRootAssets = new Map(
  copiedRootFiles.map((fileName) => [`/sshwifty/assets/${fileName}`, fileName]),
);
const rootCompatibilityAssets = new Map([
  ["/favicon.ico", "favicon.ico"],
  ["/manifest.json", "site.webmanifest"],
  ["/browserconfig.xml", "browserconfig.xml"],
]);
const publicAssetContentTypes = new Map([
  ["site.webmanifest", "application/manifest+json"],
  ["browserconfig.xml", "application/xml; charset=utf-8"],
  ["sshwifty.svg", "image/svg+xml"],
  ["favicon.ico", "image/x-icon"],
]);
const rootAssetContentTypes = new Map([
  [".md", "text/markdown; charset=utf-8"],
]);
const browserEncodingPackages = [
  "/node_modules/buffer/",
  "/node_modules/events/",
  "/node_modules/iconv-lite/",
  "/node_modules/process/",
  "/node_modules/string_decoder/",
];

/**
 * Assign selected dependencies to stable vendor chunks.
 *
 * @param {string} id Rollup module identifier.
 * @returns {string | undefined} Manual chunk name, when applicable.
 */
function vendorChunkName(id) {
  if (!id.includes("/node_modules/")) {
    return undefined;
  }

  if (id.includes("/node_modules/@xterm/")) {
    return "vendor-xterm";
  }

  if (id.includes("/node_modules/vue/")) {
    return "vendor-vue";
  }

  if (browserEncodingPackages.some((packagePath) => id.includes(packagePath))) {
    return "vendor-encoding";
  }

  return "vendor";
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
    /**
     * Copy root documentation files after Vite writes the build bundle.
     */
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
 * Resolve a development server fixed asset route to a source file.
 *
 * @param {string} requestPath Browser request path without query string.
 * @returns {{ filePath: string, contentType: string } | null} Route target.
 */
export function resolveDevAssetRoute(requestPath) {
  const rootFile = fixedRootAssets.get(requestPath);
  if (rootFile) {
    const filePath = path.join(repoRoot, rootFile);
    const contentType =
      rootAssetContentTypes.get(path.extname(rootFile)) ??
      "text/plain; charset=utf-8";

    return { filePath, contentType };
  }

  const publicFile =
    fixedPublicAssets.get(requestPath) ??
    rootCompatibilityAssets.get(requestPath);

  if (!publicFile) {
    return null;
  }

  const filePath = path.join(publicDir, publicFile);
  const contentType =
    publicAssetContentTypes.get(publicFile) ?? "text/plain; charset=utf-8";

  return { filePath, contentType };
}

/**
 * Rewrite shell module script paths for the development URL namespace.
 *
 * The source HTML keeps local script paths so Vite can resolve entrypoints
 * during production builds. In development, the shell is served from
 * `/sshwifty/assets/`, so those relative paths must point to Vite's source
 * module URLs instead of resolving beside the served shell URL.
 *
 * @param {string} html Transformed development shell HTML.
 * @returns {string} HTML with development script paths rewritten.
 */
export function rewriteDevShellScriptPaths(html) {
  return html
    .replaceAll('src="node-globals.js"', 'src="/node-globals.js"')
    .replaceAll('src="app.js"', 'src="/app.js"');
}

/**
 * Normalize development shell URLs after Vite transforms HTML placeholders.
 *
 * @param {string} html Transformed development shell HTML.
 * @returns {string} HTML with fixed public asset URLs restored.
 */
function normalizeDevShellAssetPaths(html) {
  return html.replaceAll(
    "/sshwifty/assets/sshwifty/assets/",
    "/sshwifty/assets/",
  );
}

/**
 * Render the development shell through Vite's HTML transform pipeline.
 *
 * @param {import("vite").ViteDevServer} server Vite dev server.
 * @param {string} requestQuery Request query string without leading question mark.
 * @param {string | undefined} sourceHtml Optional source HTML for tests.
 * @returns {Promise<string>} Transformed development shell HTML.
 */
export async function renderDevShellHtml(server, requestQuery, sourceHtml) {
  const html =
    sourceHtml ?? fs.readFileSync(path.join(uiRoot, "index.html"), "utf8");
  const transformedHtml = await server.transformIndexHtml(
    "/index.html" + (requestQuery.length > 0 ? `?${requestQuery}` : ""),
    rewriteDevShellScriptPaths(html),
  );

  return normalizeDevShellAssetPaths(transformedHtml);
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
    /**
     * Add development middleware for Sshwifty shell and fixed public assets.
     *
     * @param {import("vite").ViteDevServer} server Vite dev server.
     */
    configureServer(server) {
      /**
       * Rewrite shell entry requests to the Vite-served index document.
       *
       * @param {import("node:http").IncomingMessage} req HTTP request.
       * @param {import("node:http").ServerResponse} _res HTTP response.
       * @param {() => void} next Next middleware callback.
       */
      const rewriteShellRequest = async (req, res, next) => {
        const requestUrl = req.url ?? "";
        const [requestPath, requestQuery = ""] = requestUrl.split("?", 2);

        if (
          requestPath === "/" ||
          requestPath === "/sshwifty/assets" ||
          requestPath === "/sshwifty/assets/"
        ) {
          try {
            const transformedHtml = await renderDevShellHtml(
              server,
              requestQuery,
            );

            res.setHeader("Content-Type", "text/html; charset=utf-8");
            res.end(transformedHtml);
          } catch (error) {
            next(error);
          }
          return;
        }
        next();
      };

      server.middlewares.use(rewriteShellRequest);
      server.middlewares.use(
        /**
         * Serve selected public assets from fixed compatibility routes.
         *
         * @param {import("node:http").IncomingMessage} req HTTP request.
         * @param {import("node:http").ServerResponse} res HTTP response.
         * @param {() => void} next Next middleware callback.
         */
        (req, res, next) => {
          const requestUrl = req.url ?? "";
          const [requestPath] = requestUrl.split("?", 1);

          const asset = resolveDevAssetRoute(requestPath);
          if (!asset) {
            next();
            return;
          }

          res.setHeader("Content-Type", asset.contentType);
          const stream = fs.createReadStream(asset.filePath);
          stream.on(
            "error",
            /**
             * Convert file stream failures into HTTP errors.
             *
             * @param {NodeJS.ErrnoException} error Stream error.
             */
            (error) => {
              console.error(
                `Failed to stream dev asset ${requestPath} from ${asset.filePath}:`,
                error,
              );
              if (!res.headersSent) {
                res.statusCode = error.code === "ENOENT" ? 404 : 500;
              }
              res.end();
            },
          );
          stream.pipe(res);
        },
      );
    },
  };
}

export default defineConfig(
  /**
   * Build the Vite configuration for the current command and mode.
   *
   * @param {{ command: string, mode: string }} env Vite config environment.
   * @returns {import("vite").UserConfig} Vite configuration.
   */
  ({ command, mode }) => ({
    base: "/sshwifty/assets/",
    root: uiRoot,
    plugins: [vue(), copyRootFilesPlugin(), sshwiftyPublicAssetsPlugin()],
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
      __SSHWIFTY_SOURCE_URL__: JSON.stringify(sourceURL),
      "process.env.NODE_ENV": JSON.stringify(mode),
    },
    build: {
      target: ["es2020"],
      outDir: distDir,
      emptyOutDir: true,
      sourcemap: command === "serve",
      rollupOptions: {
        input: {
          index: path.join(uiRoot, "index.html"),
          error: path.join(uiRoot, "error.html"),
        },
        output: {
          manualChunks: vendorChunkName,
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
      root: repoRoot,
      include: ["ui/**/*_test.js"],
      globals: true,
      environment: "node",
    },
  }),
);
