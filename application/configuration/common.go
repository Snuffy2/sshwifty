// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

// Package configuration defines the data types and loader infrastructure used
// to supply runtime settings to the Sshwifty application. It supports multiple
// configuration sources (environment variables, JSON files, direct injection)
// through the Loader function type and a Redundant combinator.
package configuration

import (
	"time"

	"github.com/Snuffy2/sshwifty/application/network"
)

// Common holds the configuration settings that are shared across all server
// instances within a single Configuration. It is derived from a Configuration
// via Configuration.Common() and passed to each server at startup.
type Common struct {
	// HostName is the public hostname used in generated links and TLS validation.
	HostName string
	// SharedKey is the pre-shared secret required for client authentication;
	// an empty value disables authentication.
	SharedKey string
	// Dialer is the function used to open outbound network connections,
	// optionally via SOCKS5 or with access-control restrictions.
	Dialer network.Dial
	// DialTimeout is the maximum duration permitted for a single outbound dial.
	DialTimeout time.Duration
	// Presets is the list of pre-configured remote endpoints shown in the UI.
	Presets []Preset
	// Hooks contains the hook settings that govern lifecycle callbacks.
	Hooks HookSettings
	// OnlyAllowPresetRemotes restricts outbound connections to hosts listed in
	// Presets when true.
	OnlyAllowPresetRemotes bool
}
