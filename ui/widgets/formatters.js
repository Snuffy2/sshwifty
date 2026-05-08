// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

/**
 * Formats a bytes-per-second value as a human-readable HTML string.
 *
 * @param {number} n - Raw value in bytes per second.
 * @returns {string} HTML string containing the formatted value and unit.
 */
export function bytePerSecondString(n) {
  const bNames = ["byte/s", "kib/s", "mib/s", "gib/s", "tib/s"];
  let remain = n,
    nUnit = bNames[0];

  for (let i in bNames) {
    nUnit = bNames[i];

    if (remain < 1024) {
      break;
    }

    remain /= 1024;
  }

  return (
    Number(remain.toFixed(2)).toLocaleString() +
    " <span>" +
    nUnit +
    "</span>"
  );
}

/**
 * Formats a millisecond value as a human-readable HTML string.
 *
 * @param {number} n - Latency value in milliseconds.
 * @returns {string} HTML string containing the formatted value and unit, or "??".
 */
export function mSecondString(n) {
  if (n < 0) {
    return "??";
  }

  const bNames = ["ms", "s", "m"];
  let remain = n,
    nUnit = bNames[0];

  for (let i in bNames) {
    nUnit = bNames[i];

    if (remain < 1000) {
      break;
    }

    remain /= 1000;
  }

  return (
    Number(remain.toFixed(2)).toLocaleString() +
    " <span>" +
    nUnit +
    "</span>"
  );
}

/**
 * Converts a key label string into keyboard-icon HTML.
 *
 * @param {string} key - Human-readable key label such as "Ctrl+Alt+Del".
 * @returns {string} HTML string safe for the existing `v-html` usage.
 */
export function specialKeyHTML(key) {
  const head = '<span class="tb-key-icon icon icon-keyboardkey1">',
    tail = "</span>";

  return head + key.split("+").join(tail + "+" + head) + tail;
}
