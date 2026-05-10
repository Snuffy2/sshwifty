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
});
