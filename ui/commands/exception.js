// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

/**
 * @file Command-layer exception type. A thin Error subclass used consistently
 * across all command modules to signal domain-level failures (bad input,
 * undefined presets, invalid field values, etc.).
 */

/**
 * Command-layer error.
 *
 * Thrown by command builders, field validators, preset parsers, and address
 * parsers to signal invalid state or input. Callers can distinguish it from
 * unexpected runtime errors by checking `instanceof Exception`.
 */
export default class Exception extends Error {
  /**
   * constructor
   *
   * @param {string} message error message
   *
   */
  constructor(message) {
    super(message);
  }
}
