// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

const STORAGE_KEYS = ["sshwifty-knowns", "knowns"];

/**
 * Returns browser localStorage when it is available.
 *
 * @returns {Storage|null} Browser localStorage, or null when access is blocked.
 */
function legacyConnectionHistoryStorage() {
  try {
    if (typeof window === "undefined") {
      return null;
    }

    return window.localStorage;
  } catch (_e) {
    return null;
  }
}

/**
 * Removes browser-local connection history left by older Sshwifty versions.
 *
 * @param {{ removeItem: function(string): void }|null} storage Storage-like object.
 * @returns {void}
 */
export function cleanupLegacyConnectionHistory(storage = null) {
  const targetStorage = storage || legacyConnectionHistoryStorage();

  if (!targetStorage) {
    return;
  }

  for (const key of STORAGE_KEYS) {
    try {
      targetStorage.removeItem(key);
    } catch (_e) {
      // Best-effort cleanup only; blocked storage should not prevent startup.
    }
  }
}
