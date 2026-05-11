// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

const STORAGE_KEYS = ["sshwifty-knowns", "knowns"];

/**
 * Removes browser-local connection history left by older Sshwifty versions.
 *
 * @param {{ removeItem: function(string): void }} storage Storage-like object.
 * @returns {void}
 */
export function cleanupLegacyConnectionHistory(storage = window.localStorage) {
  for (const key of STORAGE_KEYS) {
    try {
      storage.removeItem(key);
    } catch (_e) {
      // Best-effort cleanup only; blocked storage should not prevent startup.
    }
  }
}
