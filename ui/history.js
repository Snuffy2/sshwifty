// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

/**
 * @file history.js
 * @description Fixed-length ring-buffer of timestamped data samples used to
 * render the traffic and latency sparkline graphs in the status window.
 */

/**
 * Maintains a fixed-length array of `{ data, class }` records that acts as a
 * sliding window: each `update` call drops the oldest entry and appends a new
 * one, keeping the buffer length constant.
 */
export class Records {
  /**
   * Creates a Records ring buffer.
   *
   * @param {Array<{ data: number, class: string }>} data - Pre-allocated array
   *   that determines the buffer capacity. Callers are expected to fill it with
   *   zero-value sentinel entries before construction.
   */
  constructor(data) {
    this.data = data;
  }

  /**
   * Appends a new data point by removing the oldest entry and pushing the new
   * value onto the end, preserving the fixed buffer length.
   *
   * @param {number} newData - The new sample value to record.
   * @returns {void}
   */
  update(newData) {
    this.data.shift();
    this.data.push({ data: newData, class: "" });
  }

  /**
   * Marks all entries in the buffer as expired by setting their `class` to
   * `"expired"`, causing the sparkline renderer to visually dim them.
   *
   * Called when the connection is closed so stale data is distinguished from
   * live data on reconnect.
   *
   * @returns {void}
   */
  expire() {
    for (let i = 0; i < this.data.length; i++) {
      this.data[i].class = "expired";
    }
  }

  /**
   * Returns the underlying data array.
   *
   * The returned reference is the live backing array — mutations will affect
   * the `Records` state. Consumers should treat it as read-only.
   *
   * @returns {Array<{ data: number, class: string }>} The current record buffer.
   */
  get() {
    return this.data;
  }
}
