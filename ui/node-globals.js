// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

/**
 * @file Browser-facing Node global shims required by iconv-lite dependencies.
 */

import { Buffer as __Buffer } from "buffer";
import __process from "process";

globalThis.Buffer ??= __Buffer;
globalThis.process ??= __process;
