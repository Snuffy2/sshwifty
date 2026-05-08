// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

// Preset describes a pre-configured remote endpoint displayed in the Sshwifty
// UI. Each Preset is associated with a command type (e.g. "SSH" or "Telnet")
// and may carry command-specific metadata in the Meta map.
type Preset struct {
	// Title is the human-readable label shown in the UI tab.
	Title string
	// Type identifies the command that handles this preset (e.g. "SSH").
	Type string
	// Host is the address (and optional port) of the remote endpoint.
	Host string
	// TabColor is an optional CSS colour string used to tint the UI tab.
	TabColor string
	// Meta holds command-specific key/value options (e.g. SSH username).
	Meta map[string]string
}
