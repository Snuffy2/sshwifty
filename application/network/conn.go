// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

// Package network provides network connection utilities for ShellPort, including
// timeout-enforcing net.Conn wrappers, TCP and SOCKS5 dialers, and an
// access-control dialer that restricts outbound connections to an allow-list.
package network

import (
	"time"
)

// emptyTime is the zero value of time.Time used to represent "no deadline" in
// SetDeadline/SetReadDeadline/SetWriteDeadline calls.
var (
	emptyTime = time.Time{}
)
