// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import { execFileSync, spawnSync } from "node:child_process";

/**
 * Resolve the version string to embed in the built binary.
 *
 * @returns {string} Existing environment version, Git description, or dev.
 */
function resolveVersion() {
  if (process.env.SHELLPORT_VERSION) {
    return process.env.SHELLPORT_VERSION;
  }

  try {
    return execFileSync(
      "git",
      ["describe", "--always", "--dirty=*", "--tag"],
      {
        encoding: "utf8",
        stdio: ["ignore", "pipe", "ignore"],
      },
    ).trim();
  } catch {
    return "dev";
  }
}

const version = resolveVersion();
const result = spawnSync(
  "go",
  [
    "build",
    "-ldflags",
    `-s -w -X github.com/Snuffy2/shellport/application.version=${version}`,
  ],
  {
    env: process.env,
    stdio: "inherit",
  },
);

if (result.error) {
  throw result.error;
}

process.exit(result.status ?? 1);
