// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import assert from "assert";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

describe("Iconv common", () => {
  it("does not depend on Node stream polyfills", () => {
    const source = readFileSync(
      fileURLToPath(new URL("./common.js", import.meta.url)),
      "utf8",
    );

    assert.equal(source.includes('from "stream"'), false);
    assert.equal(source.includes("enableStreamingAPI"), false);
  });
});
