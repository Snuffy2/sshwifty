// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Preset describes a pre-configured remote endpoint displayed in the Sshwifty
// UI. Each Preset is associated with a command type (e.g. "SSH" or "Telnet")
// and may carry command-specific metadata in the Meta map.
type Preset struct {
	// ID is the stable identifier used when editing this preset through the API.
	ID string
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

// EnsurePresetIDs returns a copy of presets with every missing ID filled.
//
// Existing IDs are preserved. Duplicate non-empty IDs return an error because
// the preset update API uses IDs as stable identifiers.
func EnsurePresetIDs(presets []Preset) ([]Preset, bool, error) {
	normalized := make([]Preset, len(presets))
	copy(normalized, presets)

	seen := make(map[string]struct{}, len(normalized))
	changed := false
	for i := range normalized {
		if normalized[i].ID == "" {
			id, err := newPresetID()
			if err != nil {
				return nil, false, err
			}
			normalized[i].ID = id
			changed = true
		}
		if _, ok := seen[normalized[i].ID]; ok {
			return nil, false, fmt.Errorf("duplicate preset ID %q", normalized[i].ID)
		}
		seen[normalized[i].ID] = struct{}{}
	}
	return normalized, changed, nil
}

// newPresetID returns a random URL-safe preset identifier.
func newPresetID() (string, error) {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("generate preset ID: %w", err)
	}
	return "preset-" + hex.EncodeToString(data[:]), nil
}
