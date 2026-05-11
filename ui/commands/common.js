// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

/**
 * @file Shared utilities for the command layer: address parsing (IPv4, IPv6,
 * hostname), string/binary conversions, and character-set helpers used by the
 * SSH and Telnet command modules.
 */

import Exception from "./exception.js";
import * as iconv from "../iconv/common.js";

export const MAX_HOOK_OUTPUT_LEN = 128;
export const HOOK_OUTPUT_STR_ELLIPSIS = "...";
export const charsetPresets = iconv.charset;

const numCharators = {
  0: true,
  1: true,
  2: true,
  3: true,
  4: true,
  5: true,
  6: true,
  7: true,
  8: true,
  9: true,
};

const hexCharators = {
  0: true,
  1: true,
  2: true,
  3: true,
  4: true,
  5: true,
  6: true,
  7: true,
  8: true,
  9: true,
  a: true,
  b: true,
  c: true,
  d: true,
  e: true,
  f: true,
};

/**
 * Test whether or not given string is all number
 *
 * @param {string} d Input data
 *
 * @returns {boolean} Return true if given string is all number, false otherwise
 *
 */
export function isNumber(d) {
  for (let i = 0; i < d.length; i++) {
    if (!numCharators[d[i]]) {
      return false;
    }
  }

  return true;
}

/**
 * Test whether or not given string is all hex
 *
 * @param {string} d Input data
 *
 * @returns {boolean} Return true if given string is all hex, false otherwise
 *
 */
export function isHex(d) {
  let dd = d.toLowerCase();

  for (let i = 0; i < dd.length; i++) {
    if (!hexCharators[dd[i]]) {
      return false;
    }
  }

  return true;
}

/**
 * Test whether or not given string is a valid hostname as far as the ShellPort
 * client consider. This function will return true if the string contains only
 * printable charactors (Latin-1 printable range with a few gaps excluded).
 *
 * @private
 * @param {string} d Input data
 *
 * @returns {boolean} `true` when all characters are in the accepted printable
 *   range, `false` if any control or disallowed character is found.
 *
 */
function isHostname(d) {
  for (let i = 0; i < d.length; i++) {
    const dChar = d.charCodeAt(i);

    if (dChar >= 32 && dChar <= 126) {
      continue;
    }

    if (dChar === 128) {
      continue;
    }

    if (dChar >= 130 && dChar <= 140) {
      continue;
    }

    if (dChar === 142) {
      continue;
    }

    if (dChar >= 145 && dChar <= 156) {
      continue;
    }

    if (dChar >= 158 && dChar <= 159) {
      continue;
    }

    if (dChar >= 161 && dChar <= 255) {
      continue;
    }

    return false;
  }

  return true;
}

/**
 * Parse IPv4 address
 *
 * @param {string} d IP address
 *
 * @returns {Uint8Array} Parsed IPv4 Address
 *
 * @throws {Exception} When the given ip address was not an IPv4 addr
 *
 */
export function parseIPv4(d) {
  const addrSeg = 4;

  let s = d.split(".");

  if (s.length != addrSeg) {
    throw new Exception("Invalid address");
  }

  let r = new Uint8Array(addrSeg);

  for (let i in s) {
    if (!isNumber(s[i])) {
      throw new Exception("Invalid address");
    }

    let ii = parseInt(s[i], 10); // Only support dec

    if (isNaN(ii)) {
      throw new Exception("Invalid address");
    }

    if (ii > 0xff) {
      throw new Exception("Invalid address");
    }

    r[i] = ii;
  }

  return r;
}

/**
 * Parse IPv6 address. ::ffff: notation is NOT supported
 *
 * @param {string} d IP address
 *
 * @returns {Uint8Array} Parsed IPv6 Address
 *
 * @throws {Exception} When the given ip address was not an IPv6 addr
 *
 */
export function parseIPv6(d) {
  const addrSeg = 8;
  let s = d.split(":");

  if (s.length > addrSeg || s.length <= 1) {
    throw new Exception("Invalid address");
  }

  if (s[0].charAt(0) === "[") {
    s[0] = s[0].substring(1, s[0].length);
    let end = s.length - 1;
    if (s[end].charAt(s[end].length - 1) !== "]") {
      throw new Exception("Invalid address");
    }
    s[end] = s[end].substring(0, s[end].length - 1);
  }

  let r = new Uint8Array(addrSeg * 2),
    rIndexShift = 0;
  for (let i = 0; i < s.length; i++) {
    if (s[i].length <= 0) {
      rIndexShift = addrSeg - s.length;
      continue;
    }
    if (!isHex(s[i])) {
      throw new Exception("Invalid address");
    }
    let ii = parseInt(s[i], 16); // Only support hex
    if (isNaN(ii)) {
      throw new Exception("Invalid address");
    }
    if (ii > 0xffff) {
      throw new Exception("Invalid address");
    }
    let j = (rIndexShift + i) * 2;
    r[j] = ii >> 8;
    r[j + 1] = ii & 0xff;
  }

  return r;
}

/**
 * Convert string into a {Uint8Array}
 *
 * @param {string} d Input
 *
 * @returns {Uint8Array} Output
 *
 */
export function strToUint8Array(d) {
  let r = new Uint8Array(d.length);

  for (let i = 0, j = d.length; i < j; i++) {
    r[i] = d.charCodeAt(i);
  }

  return r;
}

/**
 * Convert a binary string into a {Uint8Array}.
 *
 * Each character is truncated to the low byte so browser-native conversion
 * matches the legacy Buffer binary encoding behavior.
 *
 * @param {string} d Input binary string.
 *
 * @returns {Uint8Array} Byte array containing one byte per input character.
 *
 */
export function strToBinary(d) {
  let result = new Uint8Array(d.length);

  for (let i = 0, j = d.length; i < j; i++) {
    result[i] = d.charCodeAt(i) & 0xff;
  }

  return result;
}

const hostnameVerifier = new RegExp("^([0-9A-Za-z_.-]+)$");

/**
 * Parse hostname
 *
 * @param {string} d IP address
 *
 * @returns {Uint8Array} Parsed IPv6 Address
 *
 * @throws {Exception} When the given ip address was not an IPv6 addr
 *
 */
export function parseHostname(d) {
  if (d.length <= 0) {
    throw new Exception("Invalid address");
  }

  if (!isHostname(d)) {
    throw new Exception("Invalid address");
  }

  if (!hostnameVerifier.test(d)) {
    throw new Exception("Invalid address");
  }

  return strToUint8Array(d);
}

/**
 * Attempt to parse a bare address string (no port) as IPv4, then IPv6, then
 * hostname. Returns the first successful result.
 *
 * @private
 * @param {string} d Address string without port.
 * @returns {{ type: string, data: Uint8Array }} Parsed address with a
 *   discriminant `type` of `"IPv4"`, `"IPv6"`, or `"Hostname"` and the
 *   corresponding byte representation.
 * @throws {Exception} When the string cannot be parsed as any address type.
 */
function parseAddr(d) {
  try {
    return {
      type: "IPv4",
      data: parseIPv4(d),
    };
  } catch (e) {
    // Do nothing
  }

  try {
    return {
      type: "IPv6",
      data: new Uint8Array(parseIPv6(d).buffer),
    };
  } catch (e) {
    // Do nothing
  }

  return {
    type: "Hostname",
    data: parseHostname(d),
  };
}

/**
 * Split a `host:port` string into its components and parse the host.
 *
 * Handles IPv6 bracket notation (`[::1]:22`), bare IPv4/hostname with an
 * implicit default port, and explicit port suffixes. The returned object
 * contains the address type tag, the parsed address bytes, and the port.
 *
 * @param {string} d Host:port string supplied by the user.
 * @param {number} defPort Default port to use when no port is present.
 * @returns {{ type: string, addr: Uint8Array, port: number }} Parsed result.
 * @throws {Exception} When the string is malformed.
 */
export function splitHostPort(d, defPort) {
  let hps = d.lastIndexOf(":"),
    fhps = d.indexOf(":"),
    ipv6hps = d.indexOf("[");

  if ((hps < 0 || hps != fhps) && ipv6hps < 0) {
    let a = parseAddr(d);

    return {
      type: a.type,
      addr: a.data,
      port: defPort,
    };
  }

  if (ipv6hps > 0) {
    throw new Exception("Invalid address");
  } else if (ipv6hps === 0) {
    let ipv6hpse = d.lastIndexOf("]");

    if (ipv6hpse <= ipv6hps || ipv6hpse + 1 != hps) {
      throw new Exception("Invalid address");
    }
  }

  let addr = d.slice(0, hps),
    port = d.slice(hps + 1, d.length);

  if (!isNumber(port)) {
    throw new Exception("Invalid address");
  }

  let portNum = parseInt(port, 10),
    a = parseAddr(addr);

  return {
    type: a.type,
    addr: a.data,
    port: portNum,
  };
}
