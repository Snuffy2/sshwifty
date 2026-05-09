// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import assert from "assert";
import * as sender from "./sender.js";

describe("Sender", () => {
  function generateTestData(size) {
    let d = new Uint8Array(size);

    for (let i = 0; i < d.length; i++) {
      d[i] = i % 256;
    }

    return d;
  }

  /**
   * Wait until the async sender has appended the expected number of bytes.
   *
   * @param {Array<number>} result Mutable result buffer populated by sends.
   * @param {number} expectedLength Number of bytes required before resolving.
   * @returns {Promise<void>} Resolves once the expected bytes are present.
   */
  function waitForResult(result, expectedLength) {
    return new Promise((resolve) => {
      let timer = setInterval(() => {
        if (result.length < expectedLength) {
          return;
        }

        clearInterval(timer);
        timer = null;
        resolve();
      }, 100);
    });
  }

  it("Send", async () => {
    const maxSegSize = 64;
    let result = [];
    let sd = new sender.Sender(
      (rawData) => {
        return new Promise((resolve) => {
          setTimeout(() => {
            for (let i in rawData) {
              result.push(rawData[i]);
            }

            resolve();
          }, 5);
        });
      },
      maxSegSize,
      300,
      3,
    );
    let expected = generateTestData(maxSegSize * 16);

    sd.send(expected);

    let sendCompleted = new Promise((resolve) => {
      let timer = setInterval(() => {
        if (result.length < expected.length) {
          return;
        }

        clearInterval(timer);
        timer = null;
        resolve();
      }, 100);
    });

    await sendCompleted;

    assert.deepStrictEqual(new Uint8Array(result), expected);
  });

  it("Flushes when buffered data reaches segment size", async () => {
    const maxSegSize = 8;
    let result = [];
    let flushCount = 0;
    let sd = new sender.Sender(
      async (rawData) => {
        flushCount++;

        await new Promise((resolve) => {
          setTimeout(() => {
            for (let i in rawData) {
              result.push(rawData[i]);
            }

            resolve();
          }, 5);
        });
      },
      maxSegSize,
      300,
      3,
    );
    let expected = generateTestData(maxSegSize);

    await sd.send(expected);

    assert.strictEqual(flushCount, 1);
    assert.deepStrictEqual(new Uint8Array(result), expected);
  });

  it("flushes when buffered request count reaches limit", async () => {
    const sent = [];
    const sd = new sender.Sender(
      async (rawData) => {
        sent.push(Array.from(rawData));
      },
      8,
      1000,
      3,
    );

    await Promise.all([
      sd.send(Uint8Array.from([1])),
      sd.send(Uint8Array.from([2])),
      sd.send(Uint8Array.from([3])),
    ]);

    assert.deepStrictEqual(sent, [[1, 2, 3]]);
  });

  it("flushes buffered bytes on close", async () => {
    const sent = [];
    const sd = new sender.Sender(
      async (rawData) => {
        sent.push(Array.from(rawData));
      },
      8,
      1000,
      10,
    );
    const pending = sd.send(Uint8Array.from([5, 6]));

    await sd.close();
    await pending;

    assert.deepStrictEqual(sent, [[5, 6]]);
    await assert.rejects(() => sd.send(Uint8Array.from([7])), {
      message: "Sender has been cleared",
      temporary: false,
    });
  });

  it("reports close flush send failures to pending sends", async () => {
    const expectedError = new Error("transport failed");
    const sd = new sender.Sender(
      async () => {
        throw expectedError;
      },
      8,
      1000,
      10,
    );
    const pending = sd.send(Uint8Array.from([5, 6]));

    await sd.close();

    await assert.rejects(pending, expectedError);
  });

  it("waits for every segment before resolving oversized sends", async () => {
    const sent = [];
    let secondSegmentResolved = false;
    let releaseSecondSegment;
    const secondSegment = new Promise((resolve) => {
      releaseSecondSegment = () => {
        secondSegmentResolved = true;
        resolve();
      };
    });
    const sd = new sender.Sender(
      async (rawData) => {
        sent.push(Array.from(rawData));

        if (sent.length === 2) {
          await secondSegment;
        }
      },
      4,
      1000,
      10,
    );
    let resolved = false;
    const pending = sd.send(Uint8Array.from([1, 2, 3, 4, 5, 6]));

    pending.then(() => {
      resolved = true;
    });

    while (sent.length < 2) {
      await new Promise((resolve) => {
        setTimeout(resolve, 0);
      });
    }

    await new Promise((resolve) => {
      setTimeout(resolve, 0);
    });

    assert.strictEqual(resolved, false);
    assert.strictEqual(secondSegmentResolved, false);

    releaseSecondSegment();
    await pending;

    assert.strictEqual(resolved, true);
    assert.deepStrictEqual(sent, [
      [1, 2, 3, 4],
      [5, 6],
    ]);
  });

  it("rejects oversized sends when a later segment fails", async () => {
    const expectedError = new Error("second segment failed");
    let sendCount = 0;
    const sd = new sender.Sender(
      async () => {
        sendCount++;

        if (sendCount === 2) {
          throw expectedError;
        }
      },
      4,
      1000,
      10,
    );

    await assert.rejects(
      sd.send(Uint8Array.from([1, 2, 3, 4, 5, 6])),
      expectedError,
    );
  });

  it("Send (Multiple calls)", async () => {
    const maxSegSize = 64;
    let result = [];
    let sd = new sender.Sender(
      (rawData) => {
        return new Promise((resolve) => {
          setTimeout(() => {
            for (let i in rawData) {
              result.push(rawData[i]);
            }

            resolve();
          }, 10);
        });
      },
      maxSegSize,
      300,
      100,
    );
    let expectedSingle = generateTestData(maxSegSize * 2),
      expectedLen = expectedSingle.length * 16,
      expected = new Uint8Array(expectedLen);

    for (let i = 0; i < expectedLen; i += expectedSingle.length) {
      expected.set(expectedSingle, i);
    }

    for (let i = 0; i < expectedLen; i += expectedSingle.length) {
      setTimeout(() => {
        sd.send(expectedSingle);
      }, 100);
    }

    let sendCompleted = waitForResult(result, expectedLen);

    await sendCompleted;

    assert.deepStrictEqual(new Uint8Array(result), expected);
  });
});
