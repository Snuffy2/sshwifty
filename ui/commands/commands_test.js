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

  it("returns empty actions array when no actions parameter is provided", () => {
    const prompt = command.prompt(
      "Title",
      "Message",
      "Continue",
      () => {},
      () => {},
      [],
    );

    assert.deepStrictEqual(prompt.data().actions, []);
  });

  it("preserves multiple secondary actions", () => {
    const action1 = { text: "Save", respond: () => "save" };
    const action2 = { text: "Delete", respond: () => "delete" };

    const prompt = command.prompt(
      "Title",
      "Message",
      "Continue",
      () => {},
      () => {},
      [],
      [action1, action2],
    );

    const actions = prompt.data().actions;
    assert.strictEqual(actions.length, 2);
    assert.strictEqual(actions[0].text, "Save");
    assert.strictEqual(actions[1].text, "Delete");
    assert.strictEqual(actions[0].respond(), "save");
    assert.strictEqual(actions[1].respond(), "delete");
  });

  it("preserves the correct actionText in a prompt", () => {
    const prompt = command.prompt(
      "My Title",
      "My Message",
      "Submit Now",
      () => {},
      () => {},
      [],
    );

    assert.strictEqual(prompt.data().actionText, "Submit Now");
  });
});
