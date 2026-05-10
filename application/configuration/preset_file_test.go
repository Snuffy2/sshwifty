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

func TestReplaceFilePresetsPreservesRawMetaValues(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	keyPath := filepath.Join(t.TempDir(), "id_ed25519")
	if err := os.WriteFile(keyPath, []byte("PRIVATE KEY DATA"), 0o600); err != nil {
		t.Fatalf("os.WriteFile key returned error: %v", err)
	}
	writePresetConfig(t, configPath, []map[string]any{
		{
			"ID":    "preset-atlantis",
			"Title": "Atlantis",
			"Type":  "SSH",
			"Host":  "atlantis.home:22",
			"Meta": map[string]any{
				"User":           "pi",
				"Authentication": "Private Key",
				"Private Key":    "file://" + keyPath,
			},
		},
	})

	_, cfg, err := loadFile(configPath)
	if err != nil {
		t.Fatalf("loadFile returned error: %v", err)
	}
	preset := cfg.Presets[0]
	preset.Meta["Fingerprint"] = "SHA256:abc"
	if err := ReplaceFilePresets(configPath, []Preset{preset}); err != nil {
		t.Fatalf("ReplaceFilePresets returned error: %v", err)
	}

	raw, _, err := readCommonInputFile(configPath)
	if err != nil {
		t.Fatalf("readCommonInputFile returned error: %v", err)
	}
	if raw.Presets[0].Meta["Private Key"] != String("file://"+keyPath) {
		t.Fatalf(
			"raw private key = %q, want file URI",
			raw.Presets[0].Meta["Private Key"],
		)
	}
	if raw.Presets[0].Meta["Fingerprint"] != "SHA256:abc" {
		t.Fatal("raw fingerprint was not updated")
	}
}

func TestReplaceFilePresetsPreservesUnsupportedRawPresets(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetConfig(t, configPath, []map[string]any{
		{"ID": "preset-atlantis", "Title": "Atlantis", "Type": "SSH", "Host": "atlantis.home:22"},
		{"ID": "preset-future", "Title": "Future", "Type": "Future", "Host": "future.home"},
	})

	if err := ReplaceFilePresets(configPath, []Preset{
		{
			ID:    "preset-atlantis",
			Title: "Atlantis",
			Type:  "SSH",
			Host:  "atlantis.home:22",
			Meta: map[string]string{
				"Fingerprint": "SHA256:abc",
			},
		},
	}); err != nil {
		t.Fatalf("ReplaceFilePresets returned error: %v", err)
	}

	raw, _, err := readCommonInputFile(configPath)
	if err != nil {
		t.Fatalf("readCommonInputFile returned error: %v", err)
	}
	if len(raw.Presets) != 2 {
		t.Fatalf("raw preset count = %d, want 2", len(raw.Presets))
	}
	if raw.Presets[1].ID != "preset-future" {
		t.Fatalf("second raw preset ID = %q, want preset-future", raw.Presets[1].ID)
	}
}

func TestReplaceFilePresetsWithRuntimeDoesNotResolveRawMeta(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	keyPath := filepath.Join(t.TempDir(), "id_ed25519")
	writePresetConfig(t, configPath, []map[string]any{
		{
			"ID":    "preset-atlantis",
			"Title": "Atlantis",
			"Type":  "SSH",
			"Host":  "atlantis.home:22",
			"Meta": map[string]any{
				"User":           "pi",
				"Authentication": "Private Key",
				"Private Key":    "file://" + keyPath,
			},
		},
	})
	runtime := []Preset{
		{
			ID:    "preset-atlantis",
			Title: "Atlantis",
			Type:  "SSH",
			Host:  "atlantis.home:22",
			Meta: map[string]string{
				"User":           "pi",
				"Authentication": "Private Key",
				"Private Key":    "PRIVATE KEY DATA",
			},
		},
	}
	next := []Preset{runtime[0]}
	next[0].Meta["Fingerprint"] = "SHA256:abc"

	if err := ReplaceFilePresetsWithRuntime(configPath, next, runtime); err != nil {
		t.Fatalf("ReplaceFilePresetsWithRuntime returned error: %v", err)
	}

	raw, _, err := readCommonInputFile(configPath)
	if err != nil {
		t.Fatalf("readCommonInputFile returned error: %v", err)
	}
	if raw.Presets[0].Meta["Private Key"] != String("file://"+keyPath) {
		t.Fatalf("raw private key = %q, want file URI", raw.Presets[0].Meta["Private Key"])
	}
}

func TestReplaceFilePresetsReturnsErrorForEmptyFilePath(t *testing.T) {
	err := ReplaceFilePresets("", []Preset{
		{ID: "preset-atlantis", Title: "Atlantis", Type: "SSH", Host: "atlantis.home:22"},
	})
	if err == nil {
		t.Fatal("ReplaceFilePresets returned nil error, want error for empty file path")
	}
}

func TestPersistPresetIDsReturnsNilForEmptyFilePath(t *testing.T) {
	err := PersistPresetIDs("", []Preset{
		{ID: "preset-atlantis", Title: "Atlantis", Type: "SSH", Host: "atlantis.home:22"},
	})
	if err != nil {
		t.Fatalf("PersistPresetIDs returned error for empty path, want nil: %v", err)
	}
}

func TestPersistPresetIDsReturnsErrorForCountMismatch(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetConfig(t, configPath, []map[string]any{
		{"ID": "preset-atlantis", "Title": "Atlantis", "Type": "SSH", "Host": "atlantis.home:22"},
	})

	err := PersistPresetIDs(configPath, []Preset{
		{ID: "preset-atlantis", Title: "Atlantis", Type: "SSH", Host: "atlantis.home:22"},
		{ID: "preset-columbia", Title: "Columbia", Type: "SSH", Host: "columbia.home:22"},
	})
	if err == nil {
		t.Fatal("PersistPresetIDs returned nil error, want mismatch error")
	}
}

func TestPresetConfigWritableReturnsFalseForEmptyPath(t *testing.T) {
	if PresetConfigWritable("") {
		t.Fatal("PresetConfigWritable(\"\") = true, want false")
	}
}

func TestPresetConfigWritableReturnsTrueForWritableFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetConfig(t, configPath, nil)

	if !PresetConfigWritable(configPath) {
		t.Fatalf("PresetConfigWritable(%q) = false, want true", configPath)
	}
}

func TestPresetConfigWritableReturnsFalseForNonExistentFile(t *testing.T) {
	nonExistentPath := filepath.Join(t.TempDir(), "does-not-exist.json")

	if PresetConfigWritable(nonExistentPath) {
		t.Fatalf("PresetConfigWritable(%q) = true, want false", nonExistentPath)
	}
}

func TestReplaceFilePresetsDeletesOmittedMetaKeys(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetConfig(t, configPath, []map[string]any{
		{
			"ID":    "preset-atlantis",
			"Title": "Atlantis",
			"Type":  "SSH",
			"Host":  "atlantis.home:22",
			"Meta": map[string]any{
				"User":        "pi",
				"Fingerprint": "SHA256:abc",
			},
		},
	})

	if err := ReplaceFilePresets(configPath, []Preset{
		{
			ID:    "preset-atlantis",
			Title: "Atlantis",
			Type:  "SSH",
			Host:  "atlantis.home:22",
			Meta: map[string]string{
				"User": "pi",
			},
		},
	}); err != nil {
		t.Fatalf("ReplaceFilePresets returned error: %v", err)
	}

	raw, _, err := readCommonInputFile(configPath)
	if err != nil {
		t.Fatalf("readCommonInputFile returned error: %v", err)
	}
	if _, ok := raw.Presets[0].Meta["Fingerprint"]; ok {
		t.Fatal("raw fingerprint was not deleted")
	}
}
