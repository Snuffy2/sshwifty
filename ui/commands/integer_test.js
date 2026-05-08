// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import assert from "assert";
import * as reader from "../stream/reader.js";
import * as integer from "./integer.js";

describe("Integer", () => {
  it("Integer 127", async () => {
    let i = new integer.Integer(127),
      marshalled = i.marshal();

    let r = new reader.Reader(new reader.Multiple(() => {}), (data) => {
      return data;
    });

    assert.strictEqual(marshalled.length, 1);

    r.feed(marshalled);

    let i2 = new integer.Integer(0);

    await i2.unmarshal(r);

    assert.strictEqual(i.value(), i2.value());
  });

  it("Integer MAX", async () => {
    let i = new integer.Integer(integer.MAX),
      marshalled = i.marshal();

    let r = new reader.Reader(new reader.Multiple(() => {}), (data) => {
      return data;
    });

    assert.strictEqual(marshalled.length, 2);

    r.feed(marshalled);

    let i2 = new integer.Integer(0);

    await i2.unmarshal(r);

    assert.strictEqual(i.value(), i2.value());
  });
});
