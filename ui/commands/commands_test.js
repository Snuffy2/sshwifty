// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import assert from "assert";
import { describe, it } from "vitest";
import * as command from "./commands.js";

describe("Command prompts", () => {
  it("exposes secondary prompt actions", () => {
    const prompt = command.prompt(
      "Title",
      "Message",
      "Continue",
      () => {},
      () => {},
      [],
      [
        {
          text: "Save",
          respond() {
            return "saved";
          },
        },
      ],
    );

    assert.deepStrictEqual(prompt.data().actions, [
      {
        text: "Save",
        respond: prompt.data().actions[0].respond,
      },
    ]);
    assert.strictEqual(prompt.data().actions[0].respond(), "saved");
  });

  it("forwards fingerprint saver callbacks to interactive command wizards", () => {
    const saveFingerprint = () => {};
    let receivedSaveFingerprint = null;
    const commands = new command.Commands([
      {
        id() {
          return 0;
        },
        name() {
          return "Fake";
        },
        description() {
          return "Fake command";
        },
        color() {
          return "#000";
        },
        wizard(
          _info,
          _preset,
          _session,
          _kept,
          _streams,
          _subs,
          _controls,
          _history,
          saver,
        ) {
          receivedSaveFingerprint = saver;
          return {
            run() {},
            started() {
              return false;
            },
            control() {
              return {
                ui() {
                  return "Fake";
                },
              };
            },
            close() {},
          };
        },
        execute() {},
        launch() {},
        launcher() {},
        represet(preset) {
          return preset;
        },
      },
    ]);

    commands
      .all()[0]
      .wizard(null, null, null, null, null, null, () => {}, saveFingerprint);

    assert.strictEqual(receivedSaveFingerprint, saveFingerprint);
  });
});
