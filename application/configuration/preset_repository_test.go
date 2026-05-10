// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"sync"
	"testing"
)

func TestPresetRepositoryListReturnsCopy(t *testing.T) {
	original := []Preset{
		{ID: "preset-a", Title: "A", Host: "a.home:22"},
	}
	repo := NewPresetRepository(original)

	got := repo.List()
	if len(got) != 1 {
		t.Fatalf("List() len = %d, want 1", len(got))
	}
	if got[0].ID != "preset-a" {
		t.Fatalf("List()[0].ID = %q, want preset-a", got[0].ID)
	}

	// Mutating the returned slice must not affect the repository.
	got[0].ID = "mutated"
	again := repo.List()
	if again[0].ID != "preset-a" {
		t.Fatalf("List() returned mutable reference: ID = %q, want preset-a", again[0].ID)
	}
}

func TestPresetRepositoryReplaceUpdatesPresets(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-old", Title: "Old", Host: "old.home:22"},
	})

	repo.Replace([]Preset{
		{ID: "preset-new", Title: "New", Host: "new.home:22"},
	})

	got := repo.List()
	if len(got) != 1 {
		t.Fatalf("List() len after Replace = %d, want 1", len(got))
	}
	if got[0].ID != "preset-new" {
		t.Fatalf("List()[0].ID after Replace = %q, want preset-new", got[0].ID)
	}
}

func TestPresetRepositoryReplaceWithEmptyList(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-a", Title: "A", Host: "a.home:22"},
	})

	repo.Replace([]Preset{})

	got := repo.List()
	if len(got) != 0 {
		t.Fatalf("List() len after Replace with empty = %d, want 0", len(got))
	}
}

func TestPresetRepositoryAllowedReturnsTrueForKnownHost(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-a", Host: "a.home:22"},
		{ID: "preset-b", Host: "b.home:22"},
	})

	if !repo.Allowed("a.home:22") {
		t.Fatal("Allowed(a.home:22) = false, want true")
	}
	if !repo.Allowed("b.home:22") {
		t.Fatal("Allowed(b.home:22) = false, want true")
	}
}

func TestPresetRepositoryAllowedReturnsFalseForUnknownHost(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-a", Host: "a.home:22"},
	})

	if repo.Allowed("unknown.home:22") {
		t.Fatal("Allowed(unknown.home:22) = true, want false")
	}
}

func TestPresetRepositoryAllowedReturnsFalseWhenEmpty(t *testing.T) {
	repo := NewPresetRepository([]Preset{})

	if repo.Allowed("any.host:22") {
		t.Fatal("Allowed on empty repository = true, want false")
	}
}

func TestPresetRepositoryAllowedReflectsReplace(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-old", Host: "old.home:22"},
	})

	if !repo.Allowed("old.home:22") {
		t.Fatal("Allowed(old.home:22) before Replace = false, want true")
	}

	repo.Replace([]Preset{
		{ID: "preset-new", Host: "new.home:22"},
	})

	if repo.Allowed("old.home:22") {
		t.Fatal("Allowed(old.home:22) after Replace = true, want false")
	}
	if !repo.Allowed("new.home:22") {
		t.Fatal("Allowed(new.home:22) after Replace = false, want true")
	}
}

func TestPresetRepositoryNewSeedsFromPresets(t *testing.T) {
	presets := []Preset{
		{ID: "preset-seed", Title: "Seed", Host: "seed.home:22"},
	}
	repo := NewPresetRepository(presets)

	got := repo.List()
	if len(got) != 1 {
		t.Fatalf("NewPresetRepository seeded len = %d, want 1", len(got))
	}
	if got[0].ID != "preset-seed" {
		t.Fatalf("NewPresetRepository seeded ID = %q, want preset-seed", got[0].ID)
	}
}

func TestPresetRepositoryNewWithNilPresets(t *testing.T) {
	repo := NewPresetRepository(nil)

	got := repo.List()
	if got == nil {
		// A nil return is acceptable; check length.
		t.Fatal("List() returned nil, want empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("NewPresetRepository(nil) seeded len = %d, want 0", len(got))
	}
}

func TestPresetRepositoryConcurrentReadWrite(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-init", Host: "init.home:22"},
	})

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			if n%2 == 0 {
				repo.Replace([]Preset{{ID: "preset-concurrent", Host: "concurrent.home:22"}})
			} else {
				_ = repo.List()
				_ = repo.Allowed("concurrent.home:22")
			}
		}(i)
	}

	wg.Wait()
	// No race condition panics; verify the repo is still functional.
	got := repo.List()
	_ = got
}