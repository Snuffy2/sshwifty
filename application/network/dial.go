// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package network

import (
	"context"
	"net"
)

// Dial is the function signature used throughout the application to open
// outbound network connections. It mirrors net.Dialer.DialContext, accepting
// a context for cancellation, a network type (e.g. "tcp"), and a host:port
// address string.
type Dial func(
	ctx context.Context,
	network string,
	address string,
) (net.Conn, error)

// TCPDial returns a Dial that opens plain TCP connections using a default
// net.Dialer. When ctx is nil it falls back to Dial (no context), otherwise
// it uses DialContext.
func TCPDial() Dial {
	dial := net.Dialer{}
	return func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error) {
		if ctx == nil {
			return dial.Dial(network, address)
		}
		return dial.DialContext(ctx, network, address)
	}
}
