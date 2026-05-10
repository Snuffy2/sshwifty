// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writePresetConfig(t *testing.T, path string, presets []map[string]any) {
	t.Helper()

	data := map[string]any{
		"Servers": []map[string]any{
			{"ListenInterface": "127.0.0.1", "ListenPort": 8182},
		},
		"Presets": presets,
	}
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent returned error: %v", err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}
}

func TestLoadFileRecordsSourceFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetConfig(t, configPath, []map[string]any{
		{"ID": "preset-existing", "Title": "Atlantis", "Type": "SSH", "Host": "atlantis.home"},
	})

	_, cfg, err := loadFile(configPath)
	if err != nil {
		t.Fatalf("loadFile returned error: %v", err)
	}
	if cfg.SourceFile != configPath {
		t.Fatalf("SourceFile = %q, want %q", cfg.SourceFile, configPath)
	}
}

func TestPersistPresetIDsAddsMissingIDsToConfigFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetConfig(t, configPath, []map[string]any{
		{"Title": "Atlantis", "Type": "SSH", "Host": "atlantis.home"},
	})

	_, cfg, err := loadFile(configPath)
	if err != nil {
		t.Fatalf("loadFile returned error: %v", err)
	}
	presets, changed, err := EnsurePresetIDs(cfg.Presets)
	if err != nil {
		t.Fatalf("EnsurePresetIDs returned error: %v", err)
	}
	if !changed {
		t.Fatal("EnsurePresetIDs changed = false, want true")
	}
	if err := PersistPresetIDs(cfg.SourceFile, presets); err != nil {
		t.Fatalf("PersistPresetIDs returned error: %v", err)
	}

	_, reloaded, err := loadFile(configPath)
	if err != nil {
		t.Fatalf("second loadFile returned error: %v", err)
	}
	if reloaded.Presets[0].ID == "" {
		t.Fatal("reloaded preset ID is empty")
	}
}

func TestPersistPresetIDsEmptyFilePathIsNoop(t *testing.T) {
	presets := []Preset{{ID: "preset-a", Title: "A"}}
	if err := PersistPresetIDs("", presets); err != nil {
		t.Fatalf("PersistPresetIDs with empty path returned error: %v", err)
	}
}

func TestPersistPresetIDsRejectsMismatchedCount(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetConfig(t, configPath, []map[string]any{
		{"Title": "A", "Type": "SSH", "Host": "a.home"},
	})

	// Pass two presets when the file only has one.
	err := PersistPresetIDs(configPath, []Preset{
		{ID: "preset-a"},
		{ID: "preset-b"},
	})
	if err == nil {
		t.Fatal("PersistPresetIDs returned nil error for count mismatch, want error")
	}
}

func TestPresetConfigWritableReturnsTrueForWritableFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetConfig(t, configPath, nil)

	if !PresetConfigWritable(configPath) {
		t.Fatal("PresetConfigWritable = false for a writable file, want true")
	}
}

func TestPresetConfigWritableReturnsFalseForEmptyPath(t *testing.T) {
	if PresetConfigWritable("") {
		t.Fatal("PresetConfigWritable = true for empty path, want false")
	}
}

func TestPresetConfigWritableReturnsFalseForNonexistentFile(t *testing.T) {
	nonexistent := filepath.Join(t.TempDir(), "does-not-exist.json")
	if PresetConfigWritable(nonexistent) {
		t.Fatal("PresetConfigWritable = true for nonexistent file, want false")
	}
}

func TestReplaceFilePresetsReplacesEntirePresetList(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetConfig(t, configPath, []map[string]any{
		{"ID": "preset-old", "Title": "Old", "Type": "SSH", "Host": "old.home"},
	})

	newPresets := []Preset{
		{ID: "preset-new-1", Title: "New 1", Type: "SSH", Host: "new1.home"},
		{ID: "preset-new-2", Title: "New 2", Type: "Telnet", Host: "new2.home"},
	}
	if err := ReplaceFilePresets(configPath, newPresets); err != nil {
		t.Fatalf("ReplaceFilePresets returned error: %v", err)
	}

	_, reloaded, err := loadFile(configPath)
	if err != nil {
		t.Fatalf("loadFile after ReplaceFilePresets returned error: %v", err)
	}
	if len(reloaded.Presets) != 2 {
		t.Fatalf("reloaded preset count = %d, want 2", len(reloaded.Presets))
	}
	if reloaded.Presets[0].ID != "preset-new-1" {
		t.Fatalf("reloaded Presets[0].ID = %q, want preset-new-1", reloaded.Presets[0].ID)
	}
	if reloaded.Presets[1].ID != "preset-new-2" {
		t.Fatalf("reloaded Presets[1].ID = %q, want preset-new-2", reloaded.Presets[1].ID)
	}
}

func TestReplaceFilePresetsEmptyFilePathReturnsError(t *testing.T) {
	err := ReplaceFilePresets("", []Preset{{ID: "preset-a"}})
	if err == nil {
		t.Fatal("ReplaceFilePresets with empty path returned nil error, want error")
	}
}

func TestReplaceFilePresetsPreservesMetaValues(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetConfig(t, configPath, nil)

	newPresets := []Preset{
		{
			ID:    "preset-ssh",
			Title: "SSH with meta",
			Type:  "SSH",
			Host:  "meta.home:22",
			Meta: map[string]string{
				"User": "alice",
			},
		},
	}
	if err := ReplaceFilePresets(configPath, newPresets); err != nil {
		t.Fatalf("ReplaceFilePresets returned error: %v", err)
	}

	_, reloaded, err := loadFile(configPath)
	if err != nil {
		t.Fatalf("loadFile after ReplaceFilePresets returned error: %v", err)
	}
	if len(reloaded.Presets) != 1 {
		t.Fatalf("reloaded preset count = %d, want 1", len(reloaded.Presets))
	}
	if reloaded.Presets[0].Meta["User"] != "alice" {
		t.Fatalf("Meta[User] = %q, want alice", reloaded.Presets[0].Meta["User"])
	}
}
