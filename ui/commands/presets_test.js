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

  it("returns empty string for id when preset has no id field", () => {
    const preset = new Preset({
      title: "No ID",
      type: "SSH",
      host: "host.example.com:22",
      meta: {},
    });

    assert.strictEqual(preset.id(), "");
  });

  it("emptyPreset returns a preset with empty id", () => {
    const preset = emptyPreset();

    assert.strictEqual(preset.id(), "");
  });

  it("emptyPreset toConfig includes empty id field", () => {
    const preset = emptyPreset();
    const config = preset.toConfig();

    assert.strictEqual(config.id, "");
  });

  it("Presets.toConfig returns empty array for empty preset list", () => {
    const presets = new Presets([]);

    assert.deepStrictEqual(presets.toConfig(), []);
  });

  it("Presets.toConfig preserves preset IDs for multiple presets", () => {
    const presets = new Presets([
      { id: "preset-1", title: "One", type: "SSH", host: "one.home:22", meta: {} },
      { id: "preset-2", title: "Two", type: "Telnet", host: "two.home:23", meta: {} },
    ]);

    const configs = presets.toConfig();

    assert.strictEqual(configs.length, 2);
    assert.strictEqual(configs[0].id, "preset-1");
    assert.strictEqual(configs[1].id, "preset-2");
  });

  it("Preset.toConfig round-trips meta values", () => {
    const preset = new Preset({
      id: "preset-test",
      title: "Test",
      type: "SSH",
      host: "test.home:22",
      tab_color: "#ff0000",
      meta: {
        User: "root",
        Fingerprint: "SHA256:xyz",
        Encoding: "utf-8",
      },
    });

    const config = preset.toConfig();

    assert.strictEqual(config.meta.User, "root");
    assert.strictEqual(config.meta.Fingerprint, "SHA256:xyz");
    assert.strictEqual(config.meta.Encoding, "utf-8");
    assert.strictEqual(config.tab_color, "#ff0000");
  });
});
