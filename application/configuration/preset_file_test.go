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
