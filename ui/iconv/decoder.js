// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import * as common from "./common.js";

// NativeDecoder is removed because it rely on `subscribe.Subscribe` which
// currently has bad implementation
// export class NativeDecoder {
//   constructor(output, charset) {
//     let self = this;
//     return (async (output, charset) => {
//       let startSubs = new subscribe.Subscribe();
//       self.source = new ReadableStream({
//         start(controller) {
//           startSubs.resolve(controller);
//         },
//       });
//       self.ctl = await startSubs.subscribe();
//       self.source
//         .pipeThrough(new TextDecoderStream(charset, {}))
//         .pipeTo(new WritableStream({ write: output }));
//       return self;
//     })(output, charset);
//   }

//   write(b) {
//     return this.ctl.enqueue(b);
//   }

//   close() {
//     return this.ctl.close();
//   }
// }

/**
 * @file iconv/decoder.js
 * @description Charset-aware incremental decoder backed by iconv-lite.
 * Consumes raw byte arrays and emits decoded strings to a caller-supplied
 * output callback.
 */

/**
 * Streaming charset decoder.
 *
 * Wraps iconv-lite's low-level decoder for the given `charset`. Decoded string
 * chunks are delivered to `output`. Errors from both decoding and the output
 * callback are silently swallowed to keep the session alive in the presence of
 * malformed data.
 */
export class IconvDecoder {
  /**
   * Creates a new `IconvDecoder`.
   *
   * @param {function(string): void} output - Callback invoked with each decoded
   *   string chunk produced by the decoder.
   * @param {string} charset - The source charset (e.g. `"UTF-8"`, `"Shift-JIS"`).
   *   Must be a value from {@link module:iconv/common.charset}.
   */
  constructor(output, charset) {
    this.out = output;
    this.decoder = common.Iconv.getDecoder(charset);
    this.closed = false;
    return this;
  }

  /**
   * Writes a raw byte buffer into the decoder for charset conversion.
   *
   * The decoded string output is delivered synchronously to the `output`
   * callback during this call. Decoding errors (e.g. invalid byte sequences)
   * are silently ignored.
   *
   * @param {Uint8Array} b - Raw bytes encoded in the session charset.
   * @returns {void}
   */
  write(b) {
    if (this.closed) {
      return;
    }

    try {
      const output = this.decoder.write(b);
      if (output && output.length > 0) {
        this.out(output);
      }
    } catch (e) {
      // Ignore decoding error
    }
  }

  /**
   * Flushes and closes the underlying decoder.
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
      const output = this.decoder.end();
      if (output && output.length > 0) {
        this.out(output);
      }
    } catch (e) {
      // Ignore decoding error
    }
  }
}
