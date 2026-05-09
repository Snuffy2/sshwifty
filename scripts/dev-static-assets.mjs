// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import fs from "node:fs";
import path from "node:path";

const shellPage = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Sshwifty Development Placeholder</title>
  </head>
  <body>
    Vite development server shell placeholder.
  </body>
</html>
`;

/**
 * Ensure dev-only shell pages exist so Go embed directives compile before a
 * production asset generation step has populated static_assets.
 *
 * @param {string} repoRoot Repository root directory.
 */
export function ensureDevStaticAssets(repoRoot) {
  const targetDir = path.join(
    repoRoot,
    "application",
    "controller",
    "static_assets",
  );
  fs.mkdirSync(targetDir, { recursive: true });

  for (const pageName of ["index.html", "error.html"]) {
    const pagePath = path.join(targetDir, pageName);
    if (fs.existsSync(pagePath)) {
      continue;
    }
    fs.writeFileSync(pagePath, shellPage);
  }
}
