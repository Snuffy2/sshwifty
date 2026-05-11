// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { afterEach, describe, expect, test } from "vitest";
import { prepareDevStaticAssets } from "../../scripts/dev-static-assets.mjs";

let testRoot = null;

function createTestRoot() {
  testRoot = fs.mkdtempSync(path.join(os.tmpdir(), "shellport-dev-assets-"));
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

    const cleanup = prepareDevStaticAssets(repoRoot);

    expect(
      fs.readFileSync(path.join(staticAssets, "index.html"), "utf8"),
    ).toContain("Vite development server");
    expect(
      fs.readFileSync(path.join(staticAssets, "error.html"), "utf8"),
    ).toContain("Vite development server");

    cleanup();

    expect(fs.existsSync(path.join(staticAssets, "index.html"))).toBe(false);
    expect(fs.existsSync(path.join(staticAssets, "error.html"))).toBe(false);
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

    const cleanup = prepareDevStaticAssets(repoRoot);

    expect(fs.readFileSync(path.join(staticAssets, "index.html"), "utf8")).toBe(
      "generated index",
    );
    expect(fs.readFileSync(path.join(staticAssets, "error.html"), "utf8")).toBe(
      "generated error",
    );

    cleanup();

    expect(fs.readFileSync(path.join(staticAssets, "index.html"), "utf8")).toBe(
      "generated index",
    );
    expect(fs.readFileSync(path.join(staticAssets, "error.html"), "utf8")).toBe(
      "generated error",
    );
  });

  test("cleanup preserves generated pages written after placeholder creation", () => {
    const repoRoot = createTestRoot();
    const staticAssets = path.join(
      repoRoot,
      "application",
      "controller",
      "static_assets",
    );

    const cleanup = prepareDevStaticAssets(repoRoot);
    fs.writeFileSync(path.join(staticAssets, "index.html"), "generated index");

    cleanup();

    expect(fs.readFileSync(path.join(staticAssets, "index.html"), "utf8")).toBe(
      "generated index",
    );
    expect(fs.existsSync(path.join(staticAssets, "error.html"))).toBe(false);
  });
});
