// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, test } from "vitest";
import { resolveDevAssetRoute } from "../../vite.config.js";

const testDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(testDir, "../..");
const viteConfigPath = path.join(repoRoot, "vite.config.js");
const indexHtmlPath = path.join(repoRoot, "ui", "index.html");
const errorHtmlPath = path.join(repoRoot, "ui", "error.html");
const legacyHelperNames = [
  ["passthrough", "AssetToken"],
  ["replace", "PublicAssetPaths"],
  ["restore", "PublicAssetTokens"],
  ["normalize", "HtmlOutputsPlugin"],
  ["restore", "DevHtmlAssetsPlugin"],
].map((parts) => parts.join(""));
const legacyPolyfillNames = [
  ["stream", "-browserify"],
  ["virtual:", "sshwifty-node-globals"],
  ["browser", "NodePolyfillsPlugin"],
].map((parts) => parts.join(""));

function readSource(filePath) {
  return fs.readFileSync(filePath, "utf8");
}

describe("vite config cleanup guards", () => {
  test("vite.config.js drops legacy helper names", () => {
    const viteConfig = readSource(viteConfigPath);

    for (const helperName of legacyHelperNames) {
      expect(viteConfig).not.toContain(helperName);
    }
  });

  test("vite.config.js keeps the simplified uiRoot and publicDir wiring", () => {
    const viteConfig = readSource(viteConfigPath);

    expect(viteConfig).toContain("uiRoot");
    expect(viteConfig).toContain("root: uiRoot");
    expect(viteConfig).toMatch(/^\s*publicDir,\s*$/m);
  });

  test("vite.config.js removes the node polyfill asset plumbing", () => {
    const viteConfig = readSource(viteConfigPath);

    for (const polyfillName of legacyPolyfillNames) {
      expect(viteConfig).not.toContain(polyfillName);
    }
  });

  test("ui/index.html uses BASE_URL asset links and local scripts", () => {
    const indexHtml = readSource(indexHtmlPath);

    expect(indexHtml).not.toContain('href="/sshwifty/assets/');
    expect(indexHtml).not.toContain('src="/ui/');
    expect(indexHtml).toContain("%BASE_URL%sshwifty.svg");
    expect(indexHtml).toContain("%BASE_URL%site.webmanifest");
    expect(indexHtml).toContain("%BASE_URL%DEPENDENCIES.md");
    expect(indexHtml).toContain("%BASE_URL%README.md");
    expect(indexHtml).toContain("%BASE_URL%LICENSE.md");
    expect(indexHtml).toContain('src="node-globals.js"');
    expect(indexHtml).toContain('src="app.js"');
  });

  test("ui/error.html uses local stylesheet paths", () => {
    const errorHtml = readSource(errorHtmlPath);

    expect(errorHtml).not.toContain('href="/ui/');
    expect(errorHtml).toContain('href="app.css"');
    expect(errorHtml).toContain('href="common.css"');
    expect(errorHtml).toContain('href="landing.css"');
  });

  test("development asset routes include root documentation links", () => {
    const readmeAsset = resolveDevAssetRoute("/sshwifty/assets/README.md");
    const licenseAsset = resolveDevAssetRoute("/sshwifty/assets/LICENSE.md");
    const manifestAsset = resolveDevAssetRoute(
      "/sshwifty/assets/site.webmanifest",
    );
    const faviconAsset = resolveDevAssetRoute("/favicon.ico");

    expect(readmeAsset).toEqual({
      filePath: path.join(repoRoot, "README.md"),
      contentType: "text/markdown; charset=utf-8",
    });
    expect(licenseAsset).toEqual({
      filePath: path.join(repoRoot, "LICENSE.md"),
      contentType: "text/markdown; charset=utf-8",
    });
    expect(manifestAsset).toEqual({
      filePath: path.join(repoRoot, "ui", "public", "site.webmanifest"),
      contentType: "application/manifest+json",
    });
    expect(faviconAsset).toEqual({
      filePath: path.join(repoRoot, "ui", "public", "favicon.ico"),
      contentType: "image/x-icon",
    });
    expect(resolveDevAssetRoute("/sshwifty/assets/missing.md")).toBeNull();
  });
});
