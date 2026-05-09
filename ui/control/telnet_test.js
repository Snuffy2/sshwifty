// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import assert from "assert";
import * as reader from "../stream/reader.js";
import * as telnet from "./telnet.js";

describe("Telnet Control", () => {
  function buildControl(sent) {
    const events = {};
    const control = new telnet.Telnet({
      get() {
        return {
          forget() {},
          hex() {
            return "#000000";
          },
        };
      },
    }).build({
      charset: "utf-8",
      close() {},
      events: {
        place(name, callback) {
          events[name] = callback;
        },
      },
      send(data) {
        sent.push(Array.from(data));
      },
      tabColor: "",
    });

    return { control, events };
  }

  it("encodes terminal resize dimensions", async () => {
    const sent = [];
    const { control, events } = buildControl(sent);

    control.resize({ rows: 40, cols: 120 });
    await events.inband(
      new reader.Buffer(Uint8Array.from([0xff, 0xfd, 0x1f]), () => {}),
    );
    await events.completed();

    assert.deepStrictEqual(sent, [
      [0xff, 0xfb, 0x1f],
      [0xff, 0xfa, 0x1f, 0x00, 0x78, 0x00, 0x28, 0xff, 0xf0],
    ]);
  });

  it("encodes initial NAWS accept dimensions", async () => {
    const sent = [];
    const { control, events } = buildControl(sent);

    control.windowDim = { rows: 40, cols: 120 };
    await events.inband(
      new reader.Buffer(Uint8Array.from([0xff, 0xfd, 0x1f]), () => {}),
    );
    await events.completed();

    assert.deepStrictEqual(sent, [
      [0xff, 0xfb, 0x1f, 0xff, 0xfa, 0x1f, 0x00, 0x78, 0x00, 0x28, 0xff, 0xf0],
    ]);
  });
});
