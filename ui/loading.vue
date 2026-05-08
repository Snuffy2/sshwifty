<!--
Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
Copyright (C) 2026 Snuffy2
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
  <div id="app-loading">
    <div id="app-loading-frame">
      <div v-if="error.length <= 0" id="app-loading-icon"></div>
      <div v-else id="app-loading-error">&times;</div>

      <h1 v-if="error.length <= 0" id="app-loading-title">
        Preparing client application
      </h1>
      <h1 v-else id="app-loading-title" class="error">
        {{ error }}
      </h1>
    </div>
  </div>
</template>

<script>
/**
 * @file loading.vue
 * @description Full-screen loading/error splash shown during application
 * bootstrap. Displays an animated spinner and "Preparing client application"
 * while `error` is empty; switches to an error cross and the error message
 * text once `error` becomes non-empty.
 */
export default {
  props: {
    /**
     * Fatal initialisation error message. An empty string renders the spinner;
     * any non-empty value switches the view to the error state.
     *
     * @type {string}
     */
    error: {
      type: String,
      default: "",
    },
  },
};
</script>
