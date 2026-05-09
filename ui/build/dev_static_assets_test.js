// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { afterEach, describe, expect, test } from "vitest";
import { ensureDevStaticAssets } from "../../scripts/dev-static-assets.mjs";

let testRoot = null;

function createTestRoot() {
  testRoot = fs.mkdtempSync(path.join(os.tmpdir(), "sshwifty-dev-assets-"));
  return testRoot;
}

afterEach(() => {
  if (testRoot === null) {
    return;
  }
  fs.rmSync(testRoot, { recursive: true, force: true });
  testRoot = null;
});

describe("dev static assets", () => {
  test("creates shell pages required by Go embeds on clean checkouts", () => {
    const repoRoot = createTestRoot();
    const staticAssets = path.join(
      repoRoot,
      "application",
      "controller",
      "static_assets",
    );

    ensureDevStaticAssets(repoRoot);

    expect(
      fs.readFileSync(path.join(staticAssets, "index.html"), "utf8"),
    ).toContain("Vite development server");
    expect(
      fs.readFileSync(path.join(staticAssets, "error.html"), "utf8"),
    ).toContain("Vite development server");
  });

  test("does not overwrite generated shell pages", () => {
    const repoRoot = createTestRoot();
    const staticAssets = path.join(
      repoRoot,
      "application",
      "controller",
      "static_assets",
    );
    fs.mkdirSync(staticAssets, { recursive: true });
    fs.writeFileSync(path.join(staticAssets, "index.html"), "generated index");
    fs.writeFileSync(path.join(staticAssets, "error.html"), "generated error");

    ensureDevStaticAssets(repoRoot);

    expect(fs.readFileSync(path.join(staticAssets, "index.html"), "utf8")).toBe(
      "generated index",
    );
    expect(fs.readFileSync(path.join(staticAssets, "error.html"), "utf8")).toBe(
      "generated error",
    );
  });
});
