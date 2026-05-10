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

func TestEnsurePresetIDsEmptyListReturnsUnchanged(t *testing.T) {
	presets, changed, err := EnsurePresetIDs([]Preset{})
	if err != nil {
		t.Fatalf("EnsurePresetIDs returned error: %v", err)
	}
	if changed {
		t.Fatal("EnsurePresetIDs changed = true for empty list, want false")
	}
	if len(presets) != 0 {
		t.Fatalf("EnsurePresetIDs returned %d presets, want 0", len(presets))
	}
}

func TestEnsurePresetIDsMixedPresetsFilledAndPreserved(t *testing.T) {
	presets, changed, err := EnsurePresetIDs([]Preset{
		{ID: "preset-existing", Title: "Existing"},
		{Title: "Missing ID"},
	})
	if err != nil {
		t.Fatalf("EnsurePresetIDs returned error: %v", err)
	}
	if !changed {
		t.Fatal("EnsurePresetIDs changed = false, want true")
	}
	if presets[0].ID != "preset-existing" {
		t.Fatalf("existing ID modified: %q, want preset-existing", presets[0].ID)
	}
	if presets[1].ID == "" {
		t.Fatal("missing ID was not filled")
	}
}

func TestEnsurePresetIDsDoesNotMutateInput(t *testing.T) {
	input := []Preset{
		{Title: "No ID"},
	}
	originalInput := input[0]
	_, _, err := EnsurePresetIDs(input)
	if err != nil {
		t.Fatalf("EnsurePresetIDs returned error: %v", err)
	}
	if input[0].ID != originalInput.ID {
		t.Fatal("EnsurePresetIDs mutated the input slice")
	}
}

func TestEnsurePresetIDsGeneratedIDHasPresetPrefix(t *testing.T) {
	presets, _, err := EnsurePresetIDs([]Preset{
		{Title: "No ID"},
	})
	if err != nil {
		t.Fatalf("EnsurePresetIDs returned error: %v", err)
	}
	id := presets[0].ID
	if len(id) < len("preset-") || id[:len("preset-")] != "preset-" {
		t.Fatalf("generated ID %q does not start with 'preset-'", id)
	}
}

func TestEnsurePresetIDsGeneratedIDsAreUnique(t *testing.T) {
	presets, changed, err := EnsurePresetIDs([]Preset{
		{Title: "A"},
		{Title: "B"},
		{Title: "C"},
	})
	if err != nil {
		t.Fatalf("EnsurePresetIDs returned error: %v", err)
	}
	if !changed {
		t.Fatal("EnsurePresetIDs changed = false, want true")
	}
	seen := make(map[string]bool)
	for _, p := range presets {
		if seen[p.ID] {
			t.Fatalf("generated duplicate ID: %q", p.ID)
		}
		seen[p.ID] = true
	}
}
