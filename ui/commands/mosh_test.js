// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import assert from "assert";
import * as reader from "../stream/reader.js";
import * as header from "../stream/header.js";
import * as address from "./address.js";
import * as command from "./commands.js";
import * as mosh from "./mosh.js";
import * as presets from "./presets.js";
import * as strings from "./string.js";

describe("Mosh Command", () => {
  const callbacks = {
    "initialization.failed"() {},
    initialized() {},
    "hook.before_connected"() {},
    "connect.failed"() {},
    "connect.succeed"() {},
    "connect.fingerprint"() {},
    "connect.credential"() {},
    "@stdout"() {},
    close() {},
    "@completed"() {},
  };

  /**
   * Build a reader around a single buffer.
   *
   * @param {Uint8Array} buffer Source bytes.
   * @returns {reader.Reader} Reader containing the buffer.
   */
  function makeReader(buffer) {
    let rd = new reader.Reader(new reader.Multiple(() => {}), (data) => data);

    rd.feed(buffer);

    return rd;
  }

  it("uses the Mosh command id", () => {
    assert.strictEqual(new mosh.Command().id(), 0x02);
  });

  it("uses a protocol color distinct from SSH green", () => {
    assert.strictEqual(new mosh.Command().color(), "#c73");
  });

  it("keeps launcher compatibility and omits the default Mosh Server", () => {
    const cmd = new mosh.Command();

    assert.strictEqual(
      cmd.launcher({
        user: "alice",
        host: "example.com:22",
        authentication: "Password",
        charset: "utf-8",
        moshServer: "mosh-server",
      }),
      "alice@example.com:22|Password|utf-8",
    );
    assert.strictEqual(
      cmd.launcher({
        user: "alice",
        host: "example.com:22",
        authentication: "Password",
        charset: "utf-8",
        moshServer: "",
      }),
      "alice@example.com:22|Password|utf-8",
    );
  });

  it("parses legacy and custom launcher formats", () => {
    let commandHandler = null;
    const controls = {
      get(type) {
        assert.strictEqual(type, "Mosh");

        return {};
      },
    };
    const streams = {
      request(id, callback) {
        assert.strictEqual(id, 0x02);

        commandHandler = callback({ close() {} });
      },
    };
    const customMoshServer = "/opt/mosh/bin/mosh-server";
    const cmd = new mosh.Command();

    cmd
      .launch(
        null,
        "alice@example.com:22|Password",
        streams,
        null,
        controls,
        null,
      )
      .stepInitialPrompt();
    assert.strictEqual(commandHandler.config.charset, "utf-8");
    assert.strictEqual(commandHandler.config.moshServer, "mosh-server");

    cmd
      .launch(
        null,
        "alice@example.com:22|Password|iso-8859-1",
        streams,
        null,
        controls,
        null,
      )
      .stepInitialPrompt();
    assert.strictEqual(commandHandler.config.charset, "utf-8");
    assert.strictEqual(commandHandler.config.moshServer, "mosh-server");

    cmd
      .launch(
        null,
        cmd.launcher({
          user: "alice",
          host: "example.com:22",
          authentication: "Password",
          charset: "utf-8",
          moshServer: customMoshServer,
        }),
        streams,
        null,
        controls,
        null,
      )
      .stepInitialPrompt();
    assert.strictEqual(commandHandler.config.moshServer, customMoshServer);
  });

  it("validates the Mosh Server field", () => {
    const wizard = new mosh.Command().wizard(
      null,
      presets.emptyPreset(),
      {},
      [],
      null,
      null,
      {
        get(type) {
          assert.strictEqual(type, "Mosh");

          return {};
        },
      },
      null,
    );
    const field = wizard
      .stepInitialPrompt()
      .data()
      .inputs.find((input) => input.name === "Mosh Server");

    assert.ok(field);
    assert.strictEqual(field.value, "mosh-server");
    assert.throws(() => field.verify(""), /Mosh Server must be specified/);
    assert.throws(
      () => field.verify("/usr/local/bin/mosh-server --flag"),
      /without arguments/,
    );
    assert.strictEqual(
      field.verify("/usr/local/bin/mosh-server"),
      "Will start /usr/local/bin/mosh-server",
    );
  });

  it("maps unsupported proxy initialization failures to a clear message", () => {
    let commandHandler = null;
    let resolvedStep = null;
    const initialHeader = new header.InitialStream(0, 0);

    new mosh.Command()
      .execute(
        null,
        {
          user: "alice",
          host: "example.com:22",
          authentication: "Password",
          charset: "utf-8",
        },
        {},
        [],
        {
          request(id, callback) {
            assert.strictEqual(id, 0x02);

            commandHandler = callback({ close() {} });
          },
        },
        {
          resolve(step) {
            resolvedStep = step;
          },
        },
        {
          get(type) {
            assert.strictEqual(type, "Mosh");

            return {};
          },
        },
        null,
      )
      .stepInitialPrompt();

    initialHeader.set(0x02, 0x04, false);
    commandHandler.initialize(initialHeader);

    assert.strictEqual(resolvedStep.type(), command.NEXT_DONE);
    assert.strictEqual(resolvedStep.data().success, false);
    assert.strictEqual(resolvedStep.data().errorTitle, "Request failed");
    assert.strictEqual(
      resolvedStep.data().errorMessage,
      "Mosh does not support SOCKS5 proxying in this version because its session uses UDP",
    );
  });

  it("includes the configured Mosh Server in the initial payload", async () => {
    let sent = null;
    const commandHandler = new mosh.Mosh(
      null,
      {
        user: new TextEncoder().encode("alice"),
        host: address.parseHostPort("example.com:22", 22),
        auth: 0x01,
        charset: "utf-8",
        credential: "",
        moshServer: "/usr/local/bin/mosh-server",
      },
      callbacks,
    );

    commandHandler.run({
      send(data) {
        sent = data;
      },
    });

    const rd = makeReader(sent);

    assert.deepStrictEqual(
      (await strings.String.read(rd)).data(),
      new TextEncoder().encode("alice"),
    );
    assert.strictEqual((await address.Address.read(rd)).port(), 22);
    assert.strictEqual((await reader.readOne(rd))[0], 0x01);
    assert.deepStrictEqual(
      (await strings.String.read(rd)).data(),
      new TextEncoder().encode("/usr/local/bin/mosh-server"),
    );
    assert.deepStrictEqual(
      (await strings.String.read(rd)).data(),
      new TextEncoder().encode(""),
    );
  });

  it("defaults the Mosh Server in the initial payload", async () => {
    let sent = null;
    const commandHandler = new mosh.Mosh(
      null,
      {
        user: new TextEncoder().encode("alice"),
        host: address.parseHostPort("example.com:22", 22),
        auth: 0x01,
        charset: "utf-8",
        credential: "",
      },
      callbacks,
    );

    commandHandler.run({
      send(data) {
        sent = data;
      },
    });

    const rd = makeReader(sent);

    await strings.String.read(rd);
    await address.Address.read(rd);
    await reader.readOne(rd);

    assert.deepStrictEqual(
      (await strings.String.read(rd)).data(),
      new TextEncoder().encode("mosh-server"),
    );
    assert.deepStrictEqual(
      (await strings.String.read(rd)).data(),
      new TextEncoder().encode(""),
    );
  });
});
