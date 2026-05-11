<!--
Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
Copyright (C) 2026 Snuffy2
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
  <div id="connect-known-list">
    <div v-if="presetCount <= 0" id="connect-known-list-empty">
      No presets available
    </div>
    <div v-else>
      <div id="connect-known-list-presets">
        <h3>Presets</h3>

        <ul class="hlst lstcl2">
          <li
            v-for="(preset, pk) in presets"
            :key="pk"
            :class="{ disabled: presetDisabled(preset) }"
          >
            <div class="lst-wrap" @click="selectPreset(preset)">
              <div class="labels">
                <span
                  class="type"
                  :style="'background-color: ' + preset.command.color()"
                >
                  {{ preset.command.name() }}
                </span>
              </div>

              <h4 :title="preset.preset.title()">
                {{ preset.preset.title() }}
              </h4>
            </div>
          </li>
        </ul>

        <div v-if="restrictedToPresets" id="connect-known-list-presets-alert">
          The operator has restricted the outgoing connections. You can only
          connect to remotes from the pre-defined presets.
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import "./connect_known.css";

/**
 * @fileoverview Lists server-defined presets in the connection picker.
 *
 * Preset entries can be disabled when `restrictedToPresets` is true and the
 * preset lacks a host.
 *
 * @prop {Array}    presets             - Server-defined preset connections.
 * @prop {boolean}  restrictedToPresets - When true, only fully-specified presets are selectable.
 *
 * @emits select-preset  - User chose a preset. Payload: preset object.
 */

export default {
  props: {
    presets: {
      type: Array,
      default: () => [],
    },
    restrictedToPresets: {
      type: Boolean,
      default: () => false,
    },
  },
  emits: ["select-preset"],
  computed: {
    /**
     * Returns the number of renderable presets.
     *
     * @returns {number} Preset count, or zero when presets is not an array.
     */
    presetCount() {
      return Array.isArray(this.presets) ? this.presets.length : 0;
    },
  },
  methods: {
    /**
     * Returns whether a preset should be rendered as non-interactive.
     *
     * A preset is disabled when `restrictedToPresets` is true and the preset
     * does not specify a host (i.e. requires the user to fill in the address).
     *
     * @param {Object} preset - The preset descriptor.
     * @returns {boolean} True if the preset should be disabled.
     */
    presetDisabled(preset) {
      if (!this.restrictedToPresets || preset.preset.host().length > 0) {
        return false;
      }

      return true;
    },
    /**
     * Emits `select-preset` with the chosen preset.
     * No-op while busy or if the preset is disabled.
     *
     * @param {Object} preset - The preset descriptor chosen by the user.
     * @emits select-preset
     * @returns {void}
     */
    selectPreset(preset) {
      if (this.presetDisabled(preset)) {
        return;
      }

      this.$emit("select-preset", preset);
    },
  },
};
</script>
