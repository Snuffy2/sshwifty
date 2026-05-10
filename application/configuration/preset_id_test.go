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

func TestEnsurePresetIDsWithEmptySliceReturnsUnchanged(t *testing.T) {
	presets, changed, err := EnsurePresetIDs([]Preset{})
	if err != nil {
		t.Fatalf("EnsurePresetIDs returned error: %v", err)
	}
	if changed {
		t.Fatal("EnsurePresetIDs changed = true, want false for empty slice")
	}
	if len(presets) != 0 {
		t.Fatalf("EnsurePresetIDs returned %d presets, want 0", len(presets))
	}
}

func TestEnsurePresetIDsGeneratedIDHasPresetPrefix(t *testing.T) {
	presets, _, err := EnsurePresetIDs([]Preset{
		{Title: "Atlantis", Type: "SSH", Host: "atlantis.home:22"},
	})
	if err != nil {
		t.Fatalf("EnsurePresetIDs returned error: %v", err)
	}
	id := presets[0].ID
	if len(id) < len("preset-") {
		t.Fatalf("generated ID %q is too short to have preset- prefix", id)
	}
	if id[:len("preset-")] != "preset-" {
		t.Fatalf("generated ID %q does not start with preset-", id)
	}
}

func TestEnsurePresetIDsMixedExistingAndMissingIDs(t *testing.T) {
	presets, changed, err := EnsurePresetIDs([]Preset{
		{ID: "preset-atlantis", Title: "Atlantis", Type: "SSH", Host: "atlantis.home:22"},
		{Title: "Columbia", Type: "SSH", Host: "columbia.home:22"},
	})
	if err != nil {
		t.Fatalf("EnsurePresetIDs returned error: %v", err)
	}
	if !changed {
		t.Fatal("EnsurePresetIDs changed = false, want true when some IDs are missing")
	}
	if presets[0].ID != "preset-atlantis" {
		t.Fatalf("first preset ID = %q, want preset-atlantis", presets[0].ID)
	}
	if presets[1].ID == "" {
		t.Fatal("second preset ID is empty, want generated ID")
	}
}
