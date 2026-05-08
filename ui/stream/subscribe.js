// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

/**
 * @file Single-slot async pub/sub primitive used throughout the stream layer.
 *
 * {@link Subscribe} acts as a rendezvous point: a producer calls
 * {@link Subscribe#resolve} or {@link Subscribe#reject}, and a consumer awaits
 * {@link Subscribe#subscribe}. Pending events are queued when no consumer is
 * waiting.
 */

import Exception from "./exception.js";

/** @private @type {number} Pending-queue entry type for a rejection. */
const typeReject = 0;
/** @private @type {number} Pending-queue entry type for a resolution. */
const typeResolve = 1;

/**
 * Asynchronous single-consumer pub/sub channel.
 *
 * Producers call {@link Subscribe#resolve} or {@link Subscribe#reject} to push
 * values. The consumer calls {@link Subscribe#subscribe} to receive the next
 * value, blocking until one is available. When no more values will be produced,
 * the channel can be shut down with {@link Subscribe#disable}.
 */
export class Subscribe {
  /**
   * constructor
   *
   */
  constructor() {
    this.res = null;
    this.rej = null;
    this.pending = [];
    this.disabled = null;
  }

  /**
   * Returns the number of queued producer events in `this.pending` plus one
   * when `this.res` or `this.rej` is non-null for an active consumer waiter.
   *
   * @returns {number} Count from `this.pending.length` plus any active waiter.
   */
  pendings() {
    return (
      this.pending.length + (this.rej !== null || this.res !== null ? 1 : 0)
    );
  }

  /**
   * Resolve the subscribe waiter
   *
   * @param {any} d Resolve data which will be send to the subscriber
   */
  resolve(d) {
    if (this.res !== null) {
      this.res(d);

      return;
    }

    this.pending.push([typeResolve, d]);
  }

  /**
   * Reject the subscribe waiter
   *
   * @param {any} e Error message that will be send to the subscriber
   *
   */
  reject(e) {
    if (this.rej !== null) {
      this.rej(e);

      return;
    }

    this.pending.push([typeReject, e]);
  }

  /**
   * Waiting and receive subscribe data
   *
   * @returns {Promise<any>} Data receiver
   *
   */
  subscribe() {
    if (this.pending.length > 0) {
      let p = this.pending.shift();

      switch (p[0]) {
        case typeReject:
          throw p[1];

        case typeResolve:
          return p[1];

        default:
          throw new Exception("Unknown pending type", false);
      }
    }

    if (this.disabled) {
      throw new Exception(this.disabled, false);
    }

    let self = this;

    return new Promise((resolve, reject) => {
      self.res = (d) => {
        self.res = null;
        self.rej = null;

        resolve(d);
      };

      self.rej = (e) => {
        self.res = null;
        self.rej = null;

        reject(e);
      };
    });
  }

  /**
   * Disable current subscriber when all internal data is readed
   *
   * @param {string} reason Reason of the disable
   *
   */
  disable(reason) {
    this.disabled = reason;
  }
}
