// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

/**
 * @file Stream-layer exception type. Extends the built-in Error with a
 * `temporary` flag so callers can distinguish transient failures (which may
 * be retried) from permanent ones (which require connection teardown).
 */

/**
 * Stream-layer error that carries a temporality flag.
 *
 * Callers should inspect {@link Exception#temporary} to decide whether to
 * retry the operation or abort the connection.
 */
export default class Exception extends Error {
  /**
   * constructor
   *
   * @param {string} message error message
   * @param {boolean} temporary whether or not the error is temporary
   *
   */
  constructor(message, temporary) {
    super(message);

    this.temporary = temporary;
  }
}
