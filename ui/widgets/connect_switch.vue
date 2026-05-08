<!--
Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
Copyright (C) 2026 Snuffy2
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
  <ul id="connect-switch" class="tab2" :class="{ red: tab === 'known' }">
    <li :class="{ active: tab === 'new' }" @click="switchTab('new')">
      New remote
    </li>

    <li :class="{ active: tab === 'known' }" @click="switchTab('known')">
      Known remotes <span class="label">{{ knownsLength }}</span>
    </li>
  </ul>
</template>

<script>
import "./connect_switch.css";

/**
 * @fileoverview Tab-bar component that switches the connect panel between the
 * "New remote" and "Known remotes" views. Renders as a two-item `<ul>` and
 * shows the count of known remotes as a badge on the second tab.
 *
 * @prop {string} tab          - Currently active tab: `"new"` or `"known"`.
 * @prop {number} knownsLength - Number of known remotes to display in the badge.
 *
 * @emits switch - User clicked a tab. Payload: `{string}` the target tab name.
 */

export default {
  props: {
    tab: {
      type: String,
      default: "new",
    },
    knownsLength: {
      type: Number,
      default: 0,
    },
  },
  emits: ["switch"],
  methods: {
    /**
     * Emits the `switch` event with the target tab name.
     *
     * @param {string} to - The tab to switch to: `"new"` or `"known"`.
     * @emits switch
     * @returns {void}
     */
    switchTab(to) {
      this.$emit("switch", to);
    },
  },
};
</script>
