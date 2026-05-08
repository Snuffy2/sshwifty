// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package network

import (
	"context"
	"errors"
	"net"
)

// ErrAccessControlDialTargetHostNotAllowed is returned by AccessControlDial
// when the requested address is not in the AllowedHost set.
var (
	ErrAccessControlDialTargetHostNotAllowed = errors.New(
		"unable to dial to the specified remote host due to restriction")
)

// AllowedHosts is a set of permitted host:port strings. It implements
// AllowedHost and is used to enforce the OnlyAllowPresetRemotes restriction.
type AllowedHosts map[string]struct{}

// Allowed reports whether host is in the allow-set.
func (a AllowedHosts) Allowed(host string) bool {
	_, ok := a[host]
	return ok
}

// AllowedHost is the interface checked by AccessControlDial before delegating
// each dial attempt. Implementations return true when the address is permitted.
type AllowedHost interface {
	Allowed(host string) bool
}

// AccessControlDial wraps dial with an access-control check. Before each
// connection attempt it calls allowed.Allowed(address); if the address is not
// allowed it returns ErrAccessControlDialTargetHostNotAllowed without dialing.
func AccessControlDial(allowed AllowedHost, dial Dial) Dial {
	return func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error) {
		if !allowed.Allowed(address) {
			return nil, ErrAccessControlDialTargetHostNotAllowed
		}
		return dial(ctx, network, address)
	}
}
