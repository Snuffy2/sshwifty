// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

/**
 * @file Buffered async sender for the ShellPort stream layer.
 *
 * {@link Sender} coalesces outgoing writes into segments up to `maxSegSize`
 * bytes and flushes them either when the internal buffer is full, after a
 * configurable delay (`bufferFlushDelay`), or when `maxBufferedRequests` sends
 * have accumulated — whichever comes first.
 */

import Exception from "./exception.js";
import * as subscribe from "./subscribe.js";

/**
 * Buffered, rate-limited data sender.
 *
 * Outgoing data is queued via {@link Sender#send} and dispatched in batches by
 * the internal {@link Sender#sending} loop. Callers receive a Promise that
 * resolves when their data has been handed to the underlying transport, or
 * rejects on failure.
 */
export class Sender {
  /**
   * constructor
   *
   * @param {function} sender Underlaying sender
   * @param {integer} maxSegSize The size of max data segment
   * @param {integer} bufferFlushDelay Buffer flush delay
   * @param {integer} maxBufferedRequests Buffer flush delay
   *
   */
  constructor(sender, maxSegSize, bufferFlushDelay, maxBufferedRequests) {
    this.sender = sender;
    this.maxSegSize = maxSegSize;
    this.subscribe = new subscribe.Subscribe();
    this.sendingPoc = this.sending();
    this.sendDelay = null;
    this.bufferFlushDelay = bufferFlushDelay;
    this.maxBufferedRequests = maxBufferedRequests;
    this.buffer = new Uint8Array(maxSegSize);
    this.bufferUsed = 0;
    this.bufferedRequests = 0;
    this.closed = false;
  }

  /**
   * Set the send delay of current sender
   *
   * @param {integer} newDelay the new delay
   *
   */
  setDelay(newDelay) {
    this.bufferFlushDelay = newDelay;
  }

  /**
   * Sends data to the this.sender
   *
   * @param {Uint8Array} data to send
   * @param {Array<object>} callbacks to call to return send result
   *
   */
  async sendData(data, callbacks) {
    try {
      await this.sender(data);

      for (let i in callbacks) {
        if (callbacks[i].resolveOnSuccess === false) {
          continue;
        }

        callbacks[i].resolve();
      }
    } catch (e) {
      for (let i in callbacks) {
        callbacks[i].reject(e);
      }
    }
  }

  /**
   * Append data to the end of internal buffer
   *
   * @param {Uint8Array} data data to add
   *
   * @returns {integer} How many bytes of data is added
   *
   */
  appendBuffer(data) {
    const remainSize = this.buffer.length - this.bufferUsed,
      appendLength = data.length > remainSize ? remainSize : data.length;

    this.buffer.set(data.slice(0, appendLength), this.bufferUsed);
    this.bufferUsed += appendLength;

    return appendLength;
  }

  /**
   * Export current buffer and reset it to empty
   *
   * @returns {Uint8Array} Exported buffer
   *
   */
  exportBuffer() {
    const buffer = this.buffer.slice(0, this.bufferUsed);

    this.bufferUsed = 0;
    this.bufferedRequests = 0;

    return buffer;
  }

  /**
   * Sender proc
   *
   */
  async sending() {
    let callbacks = [];

    for (;;) {
      const fetched = await this.subscribe.subscribe();

      // Force flush?
      if (fetched === true || fetched.flush === true) {
        if (this.bufferUsed <= 0) {
          if (fetched.flush === true) {
            fetched.resolve();
          }
          continue;
        }

        await this.sendData(this.exportBuffer(), callbacks);
        callbacks = [];
        if (fetched.flush === true) {
          fetched.resolve();
        }

        continue;
      }

      const callback = {
        resolve: fetched.resolve,
        reject: fetched.reject,
        resolveOnSuccess: true,
      };
      callbacks.push(callback);

      // Add data to buffer and maybe flush when the buffer is full
      let currentSendDataLen = 0;

      while (fetched.data.length > currentSendDataLen) {
        const sentLen = this.appendBuffer(
          fetched.data.slice(currentSendDataLen, fetched.data.length),
        );
        currentSendDataLen += sentLen;
        callback.resolveOnSuccess = currentSendDataLen >= fetched.data.length;

        // Buffer not full, wait for the force flush
        if (this.buffer.length > this.bufferUsed) {
          break;
        }

        await this.sendData(this.exportBuffer(), callbacks);
        callbacks = callback.resolveOnSuccess ? [] : [callback];
      }
    }
  }

  /**
   * Flush buffered data.
   *
   * @returns {Promise<void>} Resolves after pending buffered bytes are sent.
   */
  flush() {
    return new Promise((resolve) => {
      this.subscribe.resolve({
        flush: true,
        resolve,
      });
    });
  }

  /**
   * Close the sender after flushing queued bytes.
   *
   * Pending sends are resolved or rejected by the flush result before the
   * subscription channel is disabled. Calling close more than once is safe.
   *
   * @returns {Promise<void>} Resolves after queued bytes have been flushed and
   *   the sender has been disabled.
   */
  async close() {
    if (this.closed) {
      return;
    }
    this.closed = true;

    if (this.sendDelay !== null) {
      clearTimeout(this.sendDelay);
      this.sendDelay = null;
    }

    await this.flush();

    this.buffered = null;
    this.bufferUsed = 0;
    this.bufferedRequests = 0;

    this.subscribe.reject(new Exception("Sender has been cleared", false));
    this.subscribe.disable();

    this.sendingPoc.catch(() => {});
  }

  /**
   * Send data
   *
   * @param {Uint8Array} data data to send
   *
   * @throws {Exception} when sending has been cancelled
   *
   * @returns {Promise} will be resolved when the data is send and will be
   *          rejected when the data is not
   *
   */
  send(data) {
    if (this.closed) {
      return Promise.reject(new Exception("Sender has been cleared", false));
    }

    if (this.sendDelay !== null) {
      clearTimeout(this.sendDelay);
      this.sendDelay = null;
    }

    const self = this;

    return new Promise((resolve, reject) => {
      self.subscribe.resolve({
        data: data,
        resolve: resolve,
        reject: reject,
      });

      self.bufferedRequests++;

      if (self.bufferedRequests >= self.maxBufferedRequests) {
        self.bufferedRequests = 0;

        self.subscribe.resolve(true);

        return;
      }

      self.sendDelay = setTimeout(() => {
        self.sendDelay = null;
        self.bufferedRequests = 0;

        self.subscribe.resolve(true);
      }, self.bufferFlushDelay);
    });
  }
}
