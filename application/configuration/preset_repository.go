// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import "sync"

// PresetRepository stores the live preset list shared by HTTP controllers.
type PresetRepository struct {
	lock    sync.RWMutex
	presets []Preset
}

// NewPresetRepository creates a repository seeded with presets.
func NewPresetRepository(presets []Preset) *PresetRepository {
	repo := &PresetRepository{}
	repo.Replace(presets)
	return repo
}

// List returns a copy of the current presets.
func (r *PresetRepository) List() []Preset {
	r.lock.RLock()
	defer r.lock.RUnlock()

	presets := make([]Preset, len(r.presets))
	copy(presets, r.presets)
	return presets
}

// Replace swaps the live preset list.
func (r *PresetRepository) Replace(presets []Preset) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.presets = make([]Preset, len(presets))
	copy(r.presets, presets)
}

// Allowed reports whether host is in the live preset list.
func (r *PresetRepository) Allowed(host string) bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	for _, preset := range r.presets {
		if preset.Host == host {
			return true
		}
	}
	return false
}
