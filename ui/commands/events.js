// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

/**
 * @file Typed event bus used by SSH and Telnet command instances.
 *
 * {@link Events} validates that all required event handlers are registered at
 * construction time. Events prefixed with `@` are "placeholders" — they must
 * be filled in later via {@link Events#place} before they can be fired.
 */

import Exception from "./exception.js";

/**
 * Strict typed event emitter for command lifecycle events.
 *
 * At construction, every event name in `events` must have a corresponding
 * function in `callbacks`. Names prefixed with `@` are deferred placeholders
 * that must be filled with {@link Events#place} before {@link Events#fire}
 * is called for them.
 */
export class Events {
  /**
   * constructor
   *
   * @param {[]string} events required events
   * @param {object} callbacks Callbacks
   *
   * @throws {Exception} When event handler is not registered
   *
   */
  constructor(events, callbacks) {
    this.events = {};
    this.placeHolders = {};

    for (let i in events) {
      if (typeof callbacks[events[i]] !== "function") {
        throw new Exception(
          'Unknown event type for "' +
            events[i] +
            '". Expecting "function" got "' +
            typeof callbacks[events[i]] +
            '" instead.',
        );
      }

      let name = events[i];

      if (name.indexOf("@") === 0) {
        name = name.substring(1);

        this.placeHolders[name] = null;
      }

      this.events[name] = callbacks[events[i]];
    }
  }

  /**
   * Place callbacks to pending placeholder events
   *
   * @param {string} type Event Type
   * @param {function} callback Callback function
   */
  place(type, callback) {
    if (this.placeHolders[type] !== null) {
      throw new Exception(
        'Event type "' +
          type +
          '" cannot be appended. It maybe ' +
          "unregistered or already been acquired",
      );
    }

    if (typeof callback !== "function") {
      throw new Exception(
        'Unknown event type for "' +
          type +
          '". Expecting "function" got "' +
          typeof callback +
          '" instead.',
      );
    }

    delete this.placeHolders[type];

    this.events[type] = callback;
  }

  /**
   * Fire an event
   *
   * @param {string} type Event type
   * @param  {...any} data Event data
   *
   * @returns {any} The result of the event handler
   *
   * @throws {Exception} When event type is not registered
   *
   */
  fire(type, ...data) {
    if (!this.events[type] && this.placeHolders[type] !== null) {
      throw new Exception("Unknown event type: " + type);
    }

    return this.events[type](...data);
  }
}
