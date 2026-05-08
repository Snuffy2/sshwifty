<!--
Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
Copyright (C) 2026 Snuffy2
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
  <div id="connect-new">
    <ul class="lst1 lst-nostyle">
      <li
        v-for="(connector, ck) in connectors"
        :key="ck"
        @click="select(connector)"
      >
        <div class="lst-wrap">
          <h2 :style="'color: ' + connector.color()">{{ connector.name() }}</h2>
          {{ connector.description() }}
        </div>
      </li>
    </ul>
  </div>
</template>

<script>
import "./connect_new.css";

/**
 * @fileoverview Displays the list of available connector types (SSH, Telnet,
 * etc.) so the user can start a new remote connection. Each connector is
 * rendered as a styled list item with its name and description.
 *
 * @prop {Array} connectors - Available connector descriptors. Each must expose
 *   `color()`, `name()`, and `description()` methods.
 *
 * @emits select - User chose a connector type. Payload: the connector object.
 */

export default {
  props: {
    connectors: {
      type: Array,
      default: () => [],
    },
  },
  methods: {
    /**
     * Emits the `select` event with the chosen connector descriptor.
     *
     * @param {Object} connector - The connector descriptor the user clicked.
     * @emits select
     * @returns {void}
     */
    select(connector) {
      this.$emit("select", connector);
    },
  },
};
</script>
