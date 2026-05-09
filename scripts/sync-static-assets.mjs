// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);
const sourceDir = path.join(repoRoot, ".tmp", "dist");
const targetDir = path.join(
  repoRoot,
  "application",
  "controller",
  "static_assets",
);
const targetPlaceholder = ".gitkeep";
const textAssetExtensions = new Set([
  ".css",
  ".html",
  ".js",
  ".json",
  ".md",
  ".svg",
  ".txt",
  ".webmanifest",
  ".xml",
]);

if (!fs.existsSync(sourceDir)) {
  throw new Error(`missing build output directory: ${sourceDir}`);
}

if (!fs.statSync(sourceDir).isDirectory()) {
  throw new Error(`expected a directory: ${sourceDir}`);
}

fs.mkdirSync(targetDir, { recursive: true });

for (const assetPath of fs.readdirSync(targetDir)) {
  if (assetPath === targetPlaceholder) {
    continue;
  }

  fs.rmSync(path.join(targetDir, assetPath), { recursive: true, force: true });
}

fs.cpSync(sourceDir, targetDir, { recursive: true });

for (const assetPath of fs.readdirSync(targetDir, { recursive: true })) {
  const fullPath = path.join(targetDir, assetPath);

  if (!fs.statSync(fullPath).isFile()) {
    continue;
  }

  if (!textAssetExtensions.has(path.extname(fullPath))) {
    continue;
  }

  const content = fs.readFileSync(fullPath, "utf8");
  const normalized = content.replace(/[ \t]+$/gm, "").replace(/\n*$/u, "\n");

  if (normalized !== content) {
    fs.writeFileSync(fullPath, normalized);
  }
}
