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
 * Prepare dev-only shell pages so Go embed directives compile before a
 * production asset generation step has populated static_assets. The returned
 * cleanup callback removes only placeholder files created by this call.
 *
 * @param {string} repoRoot Repository root directory.
 * @returns {() => void} Cleanup callback for created placeholder pages.
 */
export function prepareDevStaticAssets(repoRoot) {
  const targetDir = path.join(
    repoRoot,
    "application",
    "controller",
    "static_assets",
  );
  const createdPages = [];

  fs.mkdirSync(targetDir, { recursive: true });

  for (const pageName of ["index.html", "error.html"]) {
    const pagePath = path.join(targetDir, pageName);
    if (fs.existsSync(pagePath)) {
      continue;
    }
    fs.writeFileSync(pagePath, shellPage);
    createdPages.push(pagePath);
  }

  return () => {
    for (const pagePath of createdPages) {
      if (!fs.existsSync(pagePath)) {
        continue;
      }
      if (fs.readFileSync(pagePath, "utf8") !== shellPage) {
        continue;
      }
      fs.unlinkSync(pagePath);
    }
  };
}
