// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import assert from "assert";
import * as reader from "../stream/reader.js";
import * as address from "./address.js";

describe("Address", () => {
  it("Address Loopback", async () => {
    let addr = new address.Address(address.LOOPBACK, null, 8080),
      buf = addr.buffer();

    let r = new reader.Reader(new reader.Multiple(), (data) => {
      return data;
    });

    r.feed(buf);

    let addr2 = await address.Address.read(r);

    assert.strictEqual(addr2.type(), addr.type());
    assert.deepStrictEqual(addr2.address(), addr.address());
    assert.strictEqual(addr2.port(), addr.port());
  });

  it("Address IPv4", async () => {
    let addr = new address.Address(
        address.IPV4,
        new Uint8Array([127, 0, 0, 1]),
        8080,
      ),
      buf = addr.buffer();

    let r = new reader.Reader(new reader.Multiple(() => {}), (data) => {
      return data;
    });

    r.feed(buf);

    let addr2 = await address.Address.read(r);

    assert.strictEqual(addr2.type(), addr.type());
    assert.deepStrictEqual(addr2.address(), addr.address());
    assert.strictEqual(addr2.port(), addr.port());
  });

  it("Address IPv6", async () => {
    let addr = new address.Address(
        address.IPV6,
        new Uint8Array([0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1]),
        8080,
      ),
      buf = addr.buffer();

    let r = new reader.Reader(new reader.Multiple(() => {}), (data) => {
      return data;
    });

    r.feed(buf);

    let addr2 = await address.Address.read(r);

    assert.strictEqual(addr2.type(), addr.type());
    assert.deepStrictEqual(addr2.address(), addr.address());
    assert.strictEqual(addr2.port(), addr.port());
  });

  it("Address HostName", async () => {
    let addr = new address.Address(
        address.HOSTNAME,
        new Uint8Array([
          "n".charCodeAt(0),
          "i".charCodeAt(0),
          "r".charCodeAt(0),
          "u".charCodeAt(0),
          "i".charCodeAt(0),
          "o".charCodeAt(0),
          "r".charCodeAt(0),
          "g".charCodeAt(0),
          1,
          2,
          3,
        ]),
        8080,
      ),
      buf = addr.buffer();

    let r = new reader.Reader(new reader.Multiple(() => {}), (data) => {
      return data;
    });

    r.feed(buf);

    let addr2 = await address.Address.read(r);

    assert.strictEqual(addr2.type(), addr.type());
    assert.deepStrictEqual(addr2.address(), addr.address());
    assert.strictEqual(addr2.port(), addr.port());
  });
});
