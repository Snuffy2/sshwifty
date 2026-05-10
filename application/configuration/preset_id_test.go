// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import "testing"

func TestEnsurePresetIDsAddsMissingIDs(t *testing.T) {
	presets, changed, err := EnsurePresetIDs([]Preset{
		{Title: "Atlantis", Type: "SSH", Host: "atlantis.home:22"},
	})
	if err != nil {
		t.Fatalf("EnsurePresetIDs returned error: %v", err)
	}
	if !changed {
		t.Fatal("EnsurePresetIDs changed = false, want true")
	}
	if presets[0].ID == "" {
		t.Fatal("EnsurePresetIDs left missing preset ID empty")
	}
}

func TestEnsurePresetIDsPreservesExistingUniqueIDs(t *testing.T) {
	presets, changed, err := EnsurePresetIDs([]Preset{
		{ID: "preset-atlantis", Title: "Atlantis", Type: "SSH", Host: "atlantis.home:22"},
	})
	if err != nil {
		t.Fatalf("EnsurePresetIDs returned error: %v", err)
	}
	if changed {
		t.Fatal("EnsurePresetIDs changed = true, want false")
	}
	if presets[0].ID != "preset-atlantis" {
		t.Fatalf("ID = %q, want preset-atlantis", presets[0].ID)
	}
}

func TestEnsurePresetIDsRejectsDuplicateIDs(t *testing.T) {
	_, _, err := EnsurePresetIDs([]Preset{
		{ID: "duplicate", Title: "Atlantis", Type: "SSH", Host: "atlantis.home:22"},
		{ID: "duplicate", Title: "Columbia", Type: "SSH", Host: "columbia.home:22"},
	})
	if err == nil {
		t.Fatal("EnsurePresetIDs returned nil error, want duplicate ID error")
	}
}
