// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

/**
 * @file Length-prefixed string encoding for the ShellPort command protocol.
 *
 * {@link String} wraps a raw `Uint8Array` and can serialize itself as a
 * variable-length integer length prefix followed by the string bytes
 * ({@link String#buffer}), or decode itself from a stream reader
 * ({@link String.read}). The {@link truncate} helper is provided for capping
 * display strings in the UI.
 */

import * as reader from "../stream/reader.js";
import * as integer from "./integer.js";

/**
 * Length-prefixed binary string for the command protocol.
 *
 * Wire format: `[Integer length][...bytes]`. The static {@link String.read}
 * factory decodes from a stream reader; {@link String#buffer} encodes for
 * sending.
 */
export class String {
  /**
   * Read String from given reader
   *
   * @param {reader.Reader} rd Source reader
   *
   * @returns {String} readed string
   *
   */
  static async read(rd) {
    let l = new integer.Integer(0);

    await l.unmarshal(rd);

    return new String(await reader.readN(rd, l.value()));
  }

  /**
   * constructor
   *
   * @param {Uint8Array} str String data
   */
  constructor(str) {
    this.str = str;
  }

  /**
   * Return the string
   *
   * @returns {Uint8Array} String data
   *
   */
  data() {
    return this.str;
  }

  /**
   * Return serialized String as array
   *
   * @returns {Uint8Array} serialized String
   *
   */
  buffer() {
    let lBytes = new integer.Integer(this.str.length).marshal(),
      buf = new Uint8Array(lBytes.length + this.str.length);

    buf.set(lBytes, 0);
    buf.set(this.str, lBytes.length);

    return buf;
  }
}

/**
 * Truncates a string to the maximum length
 *
 * @param {string} str Source string
 * @param {integer} maxLength Max length
 * @param {string} exceed Text appends the string if it was truncated
 *
 * @returns {string} truncated String
 *
 */
export function truncate(str, maxLength, exceed) {
  if (str.length <= maxLength) {
    return str;
  }
  return str.substring(0, maxLength) + exceed;
}
