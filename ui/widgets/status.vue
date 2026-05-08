<!--
Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
Copyright (C) 2026 Snuffy2
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
  <window
    id="conn-status"
    flash-class="home-window-display"
    :display="display"
    @display="$emit('display', $event)"
  >
    <h1 class="window-title">Connection status</h1>

    <div id="conn-status-info">
      {{ status.description }}
    </div>

    <div id="conn-status-delay" class="conn-status-chart">
      <div class="counters">
        <div class="counter">
          <div class="name">Delay</div>
          <div class="value" v-html="mSecondString(status.delay)"></div>
        </div>
      </div>

      <div class="chart">
        <chart
          id="conn-status-delay-chart"
          :width="480"
          :height="50"
          type="Bar"
          :enabled="display"
          :values="status.delayHistory"
        >
          <defs>
            <linearGradient
              id="conn-status-delay-chart-background"
              gradientUnits="userSpaceOnUse"
              x1="0"
              y1="0"
              x2="0"
              y2="100%"
            >
              <stop stop-color="var(--color-start)" offset="0%" />
              <stop stop-color="var(--color-stop)" offset="100%" />
            </linearGradient>
          </defs>
        </chart>
      </div>
    </div>

    <div id="conn-status-traffic" class="conn-status-chart">
      <div class="counters">
        <div class="counter">
          <div class="name">Inbound</div>
          <div class="value" v-html="bytePerSecondString(status.inbound)"></div>
        </div>

        <div class="counter">
          <div class="name">Outbound</div>
          <div
            class="value"
            v-html="bytePerSecondString(status.outbound)"
          ></div>
        </div>
      </div>

      <div class="chart">
        <chart
          id="conn-status-traffic-chart-in"
          :width="480"
          :height="25"
          type="Bar"
          :max="inoutBoundMax"
          :enabled="display"
          :values="status.inboundHistory"
          @max="inboundMaxColUpdated"
        >
          <defs>
            <linearGradient
              id="conn-status-traffic-chart-in-background"
              gradientUnits="userSpaceOnUse"
              x1="0"
              y1="0"
              x2="0"
              y2="100%"
            >
              <stop stop-color="var(--color-start)" offset="0%" />
              <stop stop-color="var(--color-stop)" offset="100%" />
            </linearGradient>
          </defs>
        </chart>
      </div>

      <div class="chart">
        <chart
          id="conn-status-traffic-chart-out"
          :width="480"
          :height="25"
          type="UpsideDownBar"
          :max="inoutBoundMax"
          :enabled="display"
          :values="status.outboundHistory"
          @max="outboundMaxColUpdated"
        >
          <defs>
            <linearGradient
              id="conn-status-traffic-chart-out-background"
              gradientUnits="userSpaceOnUse"
              x1="0"
              y1="0"
              x2="0"
              y2="100%"
            >
              <stop stop-color="var(--color-start)" offset="0%" />
              <stop stop-color="var(--color-stop)" offset="100%" />
            </linearGradient>
          </defs>
        </chart>
      </div>
    </div>
  </window>
</template>

<script>
/* eslint vue/attribute-hyphenation: 0 */

import "./status.css";

import Window from "./window.vue";
import Chart from "./chart.vue";
import { bytePerSecondString, mSecondString } from "./formatters.js";

/**
 * @fileoverview Connection status overlay widget. Displays latency and
 * inbound/outbound traffic as real-time bar charts. The two traffic charts
 * share a common y-axis scale so in/out bars are visually comparable — the
 * `inoutBoundMax` computed via `inboundMaxColUpdated` and `outboundMaxColUpdated`
 * is fed back to both charts as the `max` prop.
 *
 * @prop {boolean} display - Controls overlay visibility.
 * @prop {Object}  status  - Live connection metrics object with fields:
 *   `description`, `delay`, `delayHistory`, `inbound`, `inboundHistory`,
 *   `outbound`, `outboundHistory`.
 *
 * @emits display - Forwarded from the window widget; payload: `{boolean}`.
 */

export default {
  components: {
    window: Window,
    chart: Chart,
  },
  props: {
    display: {
      type: Boolean,
      default: false,
    },
    status: {
      type: Object,
      default: () => {
        return {
          description: "",
          delay: 0,
          delayHistory: [],
          inbound: 0,
          inboundHistory: [],
          outbound: 0,
          outboundHistory: [],
        };
      },
    },
  },
  /**
   * @returns {{inoutBoundMax: number, inboundMax: number, outboundMax: number}}
   *   `inboundMax` and `outboundMax` track the latest data maxima from each traffic chart.
   *   `inoutBoundMax` is the larger of the two and is fed back to both charts as `max`
   *   so they share a common y-axis scale.
   */
  data() {
    return {
      inoutBoundMax: 0,
      inboundMax: 0,
      outboundMax: 0,
    };
  },
  methods: {
    /**
     * Formats a bytes-per-second value for display.
     *
     * @param {number} n - Raw value in bytes per second.
     * @returns {string} HTML string containing the formatted value and unit.
     */
    bytePerSecondString(n) {
      return bytePerSecondString(n);
    },
    /**
     * Formats a millisecond value for display.
     *
     * @param {number} n - Latency value in milliseconds.
     * @returns {string} HTML string containing the formatted value and unit, or "??".
     */
    mSecondString(n) {
      return mSecondString(n);
    },
    /**
     * Updates the tracked inbound maximum and recomputes the shared y-axis maximum.
     * Called when the inbound traffic chart emits a `max` event.
     *
     * @param {number} d - The new inbound data maximum observed by the chart.
     * @returns {void}
     */
    inboundMaxColUpdated(d) {
      this.inboundMax = d;

      this.inoutBoundMax =
        this.inboundMax > this.outboundMax ? this.inboundMax : this.outboundMax;
    },
    /**
     * Updates the tracked outbound maximum and recomputes the shared y-axis maximum.
     * Called when the outbound traffic chart emits a `max` event.
     *
     * @param {number} d - The new outbound data maximum observed by the chart.
     * @returns {void}
     */
    outboundMaxColUpdated(d) {
      this.outboundMax = d;

      this.inoutBoundMax =
        this.inboundMax > this.outboundMax ? this.inboundMax : this.outboundMax;
    },
  },
};
</script>
