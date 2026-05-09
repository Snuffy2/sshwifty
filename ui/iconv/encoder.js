// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import * as common from "./common.js";

/**
 * @file iconv/encoder.js
 * @description Charset-aware incremental encoder backed by iconv-lite. Accepts
 * JavaScript strings and emits encoded byte buffers to a caller-supplied
 * output callback for transmission over the socket.
 */

/**
 * Streaming charset encoder.
 *
 * Wraps iconv-lite's low-level encoder for the given `charset`. Encoded byte
 * chunks are delivered to `output`. Errors from both encoding and the output
 * callback are silently swallowed to keep the session alive in the presence of
 * unencodable characters.
 */
export class IconvEncoder {
  /**
   * Creates a new `IconvEncoder`.
   *
   * @param {function(Buffer): void} output - Callback invoked with each encoded
   *   byte chunk produced by the encoder.
   * @param {string} charset - The target charset (e.g. `"UTF-8"`, `"GBK"`).
   *   Must be a value from {@link module:iconv/common.charset}.
   */
  constructor(output, charset) {
    this.out = output;
    this.encoder = common.Iconv.getEncoder(charset);
    this.closed = false;
    return this;
  }

  /**
   * Writes a string into the encoder stream for charset conversion.
   *
   * Encoding errors are silently ignored.
   *
   * @param {string} b - The string to encode.
   * @returns {void}
   */
  write(b) {
    if (this.closed) {
      return;
    }

    try {
      const output = this.encoder.write(b);
      if (output && output.length > 0) {
        this.out(output);
      }
    } catch (e) {
      // Ignore encoding error
    }
  }

  /**
   * Flushes and closes the underlying encode stream.
   *
   * After calling `close`, any subsequent `write` calls will have no effect.
   * Errors during stream termination are silently ignored.
   *
   * @returns {void}
   */
  close() {
    if (this.closed) {
      return;
    }
    this.closed = true;

    try {
      const output = this.encoder.end();
      if (output && output.length > 0) {
        this.out(output);
      }
    } catch (e) {
      // Ignore encoding error
    }
  }
}
