// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"sync"
	"testing"
)

func TestNewPresetRepositorySeededWithPresets(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-a", Host: "a.home:22"},
		{ID: "preset-b", Host: "b.home:22"},
	})
	presets := repo.List()
	if len(presets) != 2 {
		t.Fatalf("List() len = %d, want 2", len(presets))
	}
	if presets[0].ID != "preset-a" {
		t.Fatalf("presets[0].ID = %q, want preset-a", presets[0].ID)
	}
	if presets[1].ID != "preset-b" {
		t.Fatalf("presets[1].ID = %q, want preset-b", presets[1].ID)
	}
}

func TestNewPresetRepositoryWithEmptySlice(t *testing.T) {
	repo := NewPresetRepository([]Preset{})
	presets := repo.List()
	if len(presets) != 0 {
		t.Fatalf("List() len = %d, want 0", len(presets))
	}
}

func TestPresetRepositoryListReturnsCopy(t *testing.T) {
	original := []Preset{{ID: "preset-a", Host: "a.home:22"}}
	repo := NewPresetRepository(original)

	list1 := repo.List()
	list1[0].Host = "mutated.home:22"

	list2 := repo.List()
	if list2[0].Host == "mutated.home:22" {
		t.Fatal("List() returned reference to internal slice (not a copy)")
	}
}

func TestPresetRepositoryReplaceSwapsLiveList(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-old", Host: "old.home:22"},
	})

	repo.Replace([]Preset{
		{ID: "preset-new", Host: "new.home:22"},
	})

	presets := repo.List()
	if len(presets) != 1 {
		t.Fatalf("List() len = %d, want 1", len(presets))
	}
	if presets[0].ID != "preset-new" {
		t.Fatalf("List()[0].ID = %q, want preset-new", presets[0].ID)
	}
}

func TestPresetRepositoryReplaceWithEmptySlice(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-a", Host: "a.home:22"},
	})

	repo.Replace([]Preset{})

	presets := repo.List()
	if len(presets) != 0 {
		t.Fatalf("List() len = %d after Replace(empty), want 0", len(presets))
	}
}

func TestPresetRepositoryAllowedReturnsTrueForKnownHost(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-a", Host: "a.home:22"},
	})

	if !repo.Allowed("a.home:22") {
		t.Fatal("Allowed(\"a.home:22\") = false, want true")
	}
}

func TestPresetRepositoryAllowedReturnsFalseForUnknownHost(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-a", Host: "a.home:22"},
	})

	if repo.Allowed("b.home:22") {
		t.Fatal("Allowed(\"b.home:22\") = true, want false")
	}
}

func TestPresetRepositoryAllowedReturnsFalseWhenEmpty(t *testing.T) {
	repo := NewPresetRepository([]Preset{})

	if repo.Allowed("any.home:22") {
		t.Fatal("Allowed on empty repository = true, want false")
	}
}

func TestPresetRepositoryAllowedSeesReplacedPresets(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-old", Host: "old.home:22"},
	})

	repo.Replace([]Preset{
		{ID: "preset-new", Host: "new.home:22"},
	})

	if repo.Allowed("old.home:22") {
		t.Fatal("Allowed(\"old.home:22\") = true after Replace, want false")
	}
	if !repo.Allowed("new.home:22") {
		t.Fatal("Allowed(\"new.home:22\") = false after Replace, want true")
	}
}

func TestPresetRepositoryConcurrentListAndReplace(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-a", Host: "a.home:22"},
	})

	var wg sync.WaitGroup
	const goroutines = 20
	wg.Add(goroutines * 2)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = repo.List()
		}()
		go func(n int) {
			defer wg.Done()
			repo.Replace([]Preset{
				{ID: "preset-dynamic", Host: "dynamic.home:22"},
			})
		}(i)
	}

	wg.Wait()
}

func TestPresetRepositoryConcurrentAllowedAndReplace(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-a", Host: "a.home:22"},
	})

	var wg sync.WaitGroup
	const goroutines = 20
	wg.Add(goroutines * 2)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = repo.Allowed("a.home:22")
		}()
		go func() {
			defer wg.Done()
			repo.Replace([]Preset{
				{ID: "preset-b", Host: "b.home:22"},
			})
		}()
	}

	wg.Wait()
}