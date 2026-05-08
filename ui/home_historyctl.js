// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

/**
 * @file home_historyctl.js
 * @description Factory for the connection history controller used by the
 * `home` Vue component. Persists known remotes to `localStorage` under the
 * `"sshwifty-knowns"` key and migrates the legacy `"knowns"` key on first run.
 */

import { History } from "./commands/history.js";

/**
 * Builds and returns a {@link History} instance backed by `localStorage`.
 *
 * Performs a one-time migration of the legacy `"knowns"` key to
 * `"sshwifty-knowns"` (see inline TODO). Loads the existing record set, then
 * wires a save callback that serialises the history to JSON and syncs the
 * `connector.knowns` reactive property whenever the history changes.
 *
 * @param {{ connector: { knowns: Array } }} ctx - The home component instance,
 *   used to keep `connector.knowns` in sync with the persisted history.
 * @returns {History} The initialised history instance with up to 64 entries.
 */
export function build(ctx) {
  let rec = [];

  // This renames "knowns" to "sshwifty-knowns"
  // TODO: Remove this after some few years
  try {
    let oldStore = localStorage.getItem("knowns");

    if (oldStore) {
      localStorage.setItem("sshwifty-knowns", oldStore);
      localStorage.removeItem("knowns");
    }
  } catch (e) {
    // Do nothing
  }

  try {
    rec = JSON.parse(localStorage.getItem("sshwifty-knowns"));

    if (!rec) {
      rec = [];
    }
  } catch (e) {
    alert("Unable to load data of Known remotes: " + e);
  }

  return new History(
    rec,
    (h, d) => {
      try {
        localStorage.setItem("sshwifty-knowns", JSON.stringify(d));
        ctx.connector.knowns = h.all();
      } catch (e) {
        alert("Unable to save remote history due to error: " + e);
      }
    },
    64,
  );
}
