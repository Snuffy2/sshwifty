// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import assert from "node:assert";
import {
  bytePerSecondString,
  mSecondString,
  specialKeyHTML,
} from "./formatters.js";

describe("widget formatters", function() {
  it("formats bytes per second using binary units", function() {
    assert.strictEqual(bytePerSecondString(512), "512 <span>byte/s</span>");
    assert.strictEqual(bytePerSecondString(1536), "1.5 <span>kib/s</span>");
    assert.strictEqual(
      bytePerSecondString(1024 * 1024 * 2),
      "2 <span>mib/s</span>",
    );
  });

  it("formats millisecond values using time units", function() {
    assert.strictEqual(mSecondString(-1), "??");
    assert.strictEqual(mSecondString(42), "42 <span>ms</span>");
    assert.strictEqual(mSecondString(1500), "1.5 <span>s</span>");
  });

  it("formats special toolbar keys as keyboard icon HTML", function() {
    assert.strictEqual(
      specialKeyHTML("Ctrl+Alt+Del"),
      '<span class="tb-key-icon icon icon-keyboardkey1">Ctrl</span>+' +
        '<span class="tb-key-icon icon icon-keyboardkey1">Alt</span>+' +
        '<span class="tb-key-icon icon icon-keyboardkey1">Del</span>',
    );
  });
});
