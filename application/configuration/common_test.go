// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCommonCurrentPresetsReturnsStaticPresetsWhenRepositoryIsNil(t *testing.T) {
	c := Common{
		Presets: []Preset{
			{ID: "preset-a", Host: "a.home:22"},
		},
		PresetRepository: nil,
	}

	presets := c.CurrentPresets()
	if len(presets) != 1 {
		t.Fatalf("CurrentPresets() len = %d, want 1", len(presets))
	}
	if presets[0].ID != "preset-a" {
		t.Fatalf("CurrentPresets()[0].ID = %q, want preset-a", presets[0].ID)
	}
}

func TestCommonCurrentPresetsReturnsLiveRepositoryWhenSet(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-live", Host: "live.home:22"},
	})
	c := Common{
		Presets: []Preset{
			{ID: "preset-static", Host: "static.home:22"},
		},
		PresetRepository: repo,
	}

	presets := c.CurrentPresets()
	if len(presets) != 1 {
		t.Fatalf("CurrentPresets() len = %d, want 1", len(presets))
	}
	if presets[0].ID != "preset-live" {
		t.Fatalf("CurrentPresets()[0].ID = %q, want preset-live (from repo)", presets[0].ID)
	}
}

func TestCommonCurrentPresetsReflectsRepositoryUpdates(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-old", Host: "old.home:22"},
	})
	c := Common{
		PresetRepository: repo,
	}

	repo.Replace([]Preset{
		{ID: "preset-new", Host: "new.home:22"},
	})

	presets := c.CurrentPresets()
	if presets[0].ID != "preset-new" {
		t.Fatalf("CurrentPresets()[0].ID = %q after Replace, want preset-new", presets[0].ID)
	}
}

func TestCommonPresetConfigWritableReturnsFalseForEmptySourceFile(t *testing.T) {
	c := Common{SourceFile: ""}
	if c.PresetConfigWritable() {
		t.Fatal("PresetConfigWritable() = true, want false for empty SourceFile")
	}
}

func TestCommonPresetConfigWritableReturnsTrueForWritableFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "sshwifty.conf.json")
	if err := os.WriteFile(configPath, []byte(`{}`), 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	c := Common{SourceFile: configPath}
	if !c.PresetConfigWritable() {
		t.Fatalf("PresetConfigWritable() = false, want true for writable file %q", configPath)
	}
}

func TestCommonPresetConfigWritableReturnsFalseForNonExistentFile(t *testing.T) {
	c := Common{SourceFile: filepath.Join(t.TempDir(), "does-not-exist.json")}
	if c.PresetConfigWritable() {
		t.Fatal("PresetConfigWritable() = true, want false for non-existent file")
	}
}