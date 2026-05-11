// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import path from "node:path";
import { describe, expect, test } from "vitest";
import { cleanupLegacyConnectionHistory } from "./legacy_connection_history_cleanup.js";

const repoRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);

function readProjectFile(relativePath) {
  return readFileSync(path.join(repoRoot, relativePath), "utf8");
}

describe("connected-before removal", () => {
  test("connect UI no longer exposes known-remotes history", () => {
    const checkedFiles = [
      "ui/home.vue",
      "ui/widgets/connect.vue",
      "ui/widgets/connect_known.vue",
    ];

    for (const checkedFile of checkedFiles) {
      const source = readProjectFile(checkedFile);

      expect(source, checkedFile).not.toContain("Connected before");
      expect(source, checkedFile).not.toContain("knowns");
      expect(source, checkedFile).not.toContain("known-select");
      expect(source, checkedFile).not.toContain("known-remove");
      expect(source, checkedFile).not.toContain("known-clear-session");
      expect(source, checkedFile).not.toContain("known-remotes");
      expect(source, checkedFile).not.toContain("sshwifty-knowns");
    }
  });

  test("removes legacy connection history storage keys", () => {
    const removedKeys = [];
    const storage = {
      removeItem(key) {
        removedKeys.push(key);
      },
    };

    cleanupLegacyConnectionHistory(storage);

    expect(removedKeys).toEqual(["sshwifty-knowns", "knowns"]);
  });
});
