// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import assert from "assert";
import * as fieldBuilder from "./connector_field_builder.js";

describe("Connector field builder", () => {
  it("returns the current suggestion for the field value", () => {
    const field = fieldBuilder.build(1, 0, {
      name: "Host",
      type: "text",
      value: "example.com",
      readonly: false,
      example: "",
      description: "",
      verify: () => null,
      suggestions: () => [],
    });

    assert.deepStrictEqual(field.currentSuggestion(), {
      title: "Input",
      value: "example.com",
      fields: {},
    });
  });
});
