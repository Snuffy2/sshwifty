// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import assert from "assert";
import { describe, it } from "vitest";
import { Preset, Presets, emptyPreset } from "./presets.js";

describe("Presets", () => {
  it("preserves preset IDs and exports writable preset data", () => {
    const preset = new Preset({
      id: "preset-atlantis",
      title: "Atlantis",
      type: "SSH",
      host: "atlantis.home:22",
      tab_color: "#123456",
      meta: {
        User: "pi",
        Fingerprint: "SHA256:old",
      },
    });

    assert.strictEqual(preset.id(), "preset-atlantis");
    assert.deepStrictEqual(preset.toConfig(), {
      id: "preset-atlantis",
      title: "Atlantis",
      type: "SSH",
      host: "atlantis.home:22",
      tab_color: "#123456",
      meta: {
        User: "pi",
        Fingerprint: "SHA256:old",
      },
    });
  });

  it("exports all presets as config objects", () => {
    const presets = new Presets([
      {
        id: "preset-a",
        title: "A",
        type: "SSH",
        host: "a.home:22",
        meta: {},
      },
    ]);

    assert.deepStrictEqual(presets.toConfig(), [
      {
        id: "preset-a",
        title: "A",
        type: "SSH",
        host: "a.home:22",
        tab_color: "",
        meta: {},
      },
    ]);
  });

  it("emptyPreset returns a preset with an empty ID", () => {
    const preset = emptyPreset();

    assert.strictEqual(preset.id(), "");
  });

  it("emptyPreset toConfig includes id field", () => {
    const preset = emptyPreset();
    const config = preset.toConfig();

    assert.ok(Object.prototype.hasOwnProperty.call(config, "id"));
    assert.strictEqual(config.id, "");
  });

  it("Preset with no ID returns empty string from id()", () => {
    const preset = new Preset({
      id: "",
      title: "No ID",
      type: "SSH",
      host: "host.home:22",
      meta: {},
    });

    assert.strictEqual(preset.id(), "");
  });

  it("Presets.toConfig returns empty array for empty collection", () => {
    const presets = new Presets([]);

    assert.deepStrictEqual(presets.toConfig(), []);
  });

  it("Presets.toConfig includes all preset IDs", () => {
    const presets = new Presets([
      { id: "preset-a", title: "A", type: "SSH", host: "a.home:22", meta: {} },
      { id: "preset-b", title: "B", type: "SSH", host: "b.home:22", meta: {} },
    ]);

    const configs = presets.toConfig();
    assert.strictEqual(configs.length, 2);
    assert.strictEqual(configs[0].id, "preset-a");
    assert.strictEqual(configs[1].id, "preset-b");
  });
});
