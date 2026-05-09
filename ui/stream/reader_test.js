// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import assert from "assert";
import * as reader from "./reader.js";

describe("Reader", () => {
  it("Buffer preserves unread bytes after partial export", async () => {
    const depleted = [];
    const buf = new reader.Buffer(Uint8Array.from([0, 1, 2, 3]), () => {
      depleted.push(true);
    });

    let ex = buf.export(2);

    assert.strictEqual(ex.length, 2);
    assert.deepStrictEqual(ex, Uint8Array.from([0, 1]));
    assert.strictEqual(buf.remains(), 2);

    ex = buf.export(2);

    assert.strictEqual(ex.length, 2);
    assert.deepStrictEqual(ex, Uint8Array.from([2, 3]));
    assert.strictEqual(buf.remains(), 0);
    assert.strictEqual(depleted.length, 1);
  });

  it("Buffer", async () => {
    let buf = new reader.Buffer(
      new Uint8Array([0, 1, 2, 3, 4, 5, 6, 7]),
      () => {},
    );

    let ex = buf.export(1);

    assert.strictEqual(ex.length, 1);
    assert.strictEqual(ex[0], 0);
    assert.strictEqual(buf.remains(), 7);

    ex = await reader.readCompletely(buf);

    assert.strictEqual(ex.length, 7);
    assert.deepStrictEqual(ex, new Uint8Array([1, 2, 3, 4, 5, 6, 7]));
    assert.strictEqual(buf.remains(), 0);
  });

  it("Multiple closeWithReason rejects pending and next reads with reason", async () => {
    const reason = "multiple closed for test";
    const multiple = new reader.Multiple(() => {});
    const pending = multiple.export(1);

    multiple.closeWithReason(reason);

    await assert.rejects(pending, (error) => {
      assert.strictEqual(error.message, reason);
      assert.strictEqual(error.temporary, false);
      return true;
    });

    await assert.rejects(multiple.export(1), (error) => {
      assert.strictEqual(error.message, reason);
      assert.strictEqual(error.temporary, false);
      return true;
    });
  });

  it("readUntil repeats delimiter search before exporting buffered data", async () => {
    const buffered = Uint8Array.from([0xff, 0xfd, 0x1f]);
    let indexChecks = 0;
    const raceReader = {
      async buffered() {
        return buffered.length;
      },
      async export(n) {
        return buffered.slice(0, n);
      },
      async indexOf(byteData) {
        indexChecks++;

        if (indexChecks === 1) {
          return -1;
        }

        return buffered.indexOf(byteData);
      },
    };

    assert.deepStrictEqual(await reader.readUntil(raceReader, 0xff), {
      data: Uint8Array.from([0xff]),
      found: true,
    });
    assert.strictEqual(indexChecks, 2);
  });

  it("Reader", async () => {
    const maxTests = 3;
    let IntvCount = 0,
      r = new reader.Reader(new reader.Multiple(() => {}), (data) => {
        return data;
      }),
      expected = [
        0, 1, 2, 3, 4, 5, 6, 7, 0, 1, 2, 3, 4, 5, 6, 7, 0, 1, 2, 3, 4, 5, 6, 7,
      ],
      feedIntv = setInterval(() => {
        r.feed(Uint8Array.from(expected.slice(0, 8)));

        IntvCount++;

        if (IntvCount < maxTests) {
          return;
        }

        clearInterval(feedIntv);
      }, 300);

    let result = [];

    while (result.length < expected.length) {
      result.push((await r.export(1))[0]);
    }

    assert.deepStrictEqual(result, expected);
  });

  it("readOne", async () => {
    let r = new reader.Reader(new reader.Multiple(() => {}), (data) => {
      return data;
    });

    setTimeout(() => {
      r.feed(Uint8Array.from([0, 1, 2, 3, 4, 5, 7]));
    }, 100);

    let rr = await reader.readOne(r);

    assert.deepStrictEqual(rr, Uint8Array.from([0]));

    rr = await reader.readOne(r);

    assert.deepStrictEqual(rr, Uint8Array.from([1]));
  });

  it("readN", async () => {
    let r = new reader.Reader(new reader.Multiple(() => {}), (data) => {
      return data;
    });

    setTimeout(() => {
      r.feed(Uint8Array.from([0, 1, 2, 3, 4, 5, 7]));
    }, 100);

    let rr = await reader.readN(r, 3);

    assert.deepStrictEqual(rr, Uint8Array.from([0, 1, 2]));

    rr = await reader.readN(r, 3);

    assert.deepStrictEqual(rr, Uint8Array.from([3, 4, 5]));
  });

  it("Limited", async () => {
    const maxTests = 3;
    let IntvCount = 0,
      r = new reader.Reader(new reader.Multiple(() => {}), (data) => {
        return data;
      }),
      expected = [0, 1, 2, 3, 4, 5, 6, 7, 0, 1],
      limited = new reader.Limited(r, 10),
      feedIntv = setInterval(() => {
        r.feed(Uint8Array.from(expected.slice(0, 8)));

        IntvCount++;

        if (IntvCount < maxTests) {
          return;
        }

        clearInterval(feedIntv);
      }, 300);

    let result = [];

    while (!limited.completed()) {
      result.push((await limited.export(1))[0]);
    }

    assert.strictEqual(limited.completed(), true);
    assert.deepStrictEqual(result, expected);
  });

  it("readCompletely", async () => {
    const maxTests = 3;
    let IntvCount = 0,
      r = new reader.Reader(new reader.Multiple(() => {}), (data) => {
        return data;
      }),
      expected = [0, 1, 2, 3, 4, 5, 6, 7, 0, 1],
      limited = new reader.Limited(r, 10),
      feedIntv = setInterval(() => {
        r.feed(Uint8Array.from(expected.slice(0, 8)));

        IntvCount++;

        if (IntvCount < maxTests) {
          return;
        }

        clearInterval(feedIntv);
      }, 300);

    let result = await reader.readCompletely(limited);

    assert.strictEqual(limited.completed(), true);
    assert.deepStrictEqual(result, Uint8Array.from(expected));
  });

  it("readUntil", async () => {
    const maxTests = 3;
    let IntvCount = 0,
      r = new reader.Reader(new reader.Multiple(() => {}), (data) => {
        return data;
      }),
      sample = [0, 1, 2, 3, 4, 5, 6, 7, 0, 1],
      expected1 = new Uint8Array([0, 1, 2, 3, 4, 5, 6, 7]),
      expected2 = new Uint8Array([0, 1]),
      limited = new reader.Limited(r, 10),
      feedIntv = setInterval(() => {
        r.feed(Uint8Array.from(sample));

        IntvCount++;

        if (IntvCount < maxTests) {
          return;
        }

        clearInterval(feedIntv);
      }, 300);

    let result = await reader.readUntil(limited, 7);

    assert.strictEqual(limited.completed(), false);
    assert.deepStrictEqual(result.data, expected1);
    assert.deepStrictEqual(result.found, true);

    result = await reader.readUntil(limited, 7);

    assert.strictEqual(limited.completed(), true);
    assert.deepStrictEqual(result.data, expected2);
    assert.deepStrictEqual(result.found, false);
  });
});
