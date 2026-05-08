// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

/**
 * @file Command control registry.
 *
 * {@link Controls} maps command-type strings (e.g. `"SSH"`, `"Telnet"`) to
 * their respective control objects. Each control object is expected to expose
 * the interface used by the command's wizard to send data, resize the terminal,
 * and build the live session UI.
 */

import Exception from "./exception.js";

/**
 * Registry that maps command type names to their control interface objects.
 *
 * Populated once at startup with all registered controls; individual commands
 * look up their own control via {@link Controls#get}.
 */
export class Controls {
  /**
   * constructor
   *
   * @param {[]object} controls
   *
   * @throws {Exception} When control type already been defined
   *
   */
  constructor(controls) {
    this.controls = {};

    for (let i in controls) {
      let cType = controls[i].type();

      if (typeof this.controls[cType] === "object") {
        throw new Exception('Control "' + cType + '" already been defined');
      }

      this.controls[cType] = controls[i];
    }
  }

  /**
   * Get a control
   *
   * @param {string} type Type of the control
   *
   * @returns {object} Control object
   *
   * @throws {Exception} When given control type is undefined
   *
   */
  get(type) {
    if (typeof this.controls[type] !== "object") {
      throw new Exception('Control "' + type + '" was undefined');
    }

    return this.controls[type];
  }
}
