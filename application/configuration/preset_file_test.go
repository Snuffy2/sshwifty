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

func TestPersistPresetIDsPreservesUnknownTopLevelFields(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	content := []byte(`{
  "Servers": [
    {"ListenInterface": "127.0.0.1", "ListenPort": 8182}
  ],
  "Presets": [
    {"Title": "Atlantis", "Type": "SSH", "Host": "atlantis.home"}
  ],
  "FutureTopLevel": {"enabled": true}
}`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	_, cfg, err := loadFile(configPath)
	if err != nil {
		t.Fatalf("loadFile returned error: %v", err)
	}
	presets, _, err := EnsurePresetIDs(cfg.Presets)
	if err != nil {
		t.Fatalf("EnsurePresetIDs returned error: %v", err)
	}
	if err := PersistPresetIDs(cfg.SourceFile, presets); err != nil {
		t.Fatalf("PersistPresetIDs returned error: %v", err)
	}

	var raw map[string]json.RawMessage
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("os.ReadFile returned error: %v", err)
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if _, ok := raw["FutureTopLevel"]; !ok {
		t.Fatal("unknown top-level field was not preserved")
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

func TestReplaceFilePresetsPreservesUnknownTopLevelFields(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	content := []byte(`{
  "Servers": [
    {"ListenInterface": "127.0.0.1", "ListenPort": 8182}
  ],
  "Presets": [
    {"ID": "preset-atlantis", "Title": "Atlantis", "Type": "SSH", "Host": "atlantis.home:22"}
  ],
  "FutureTopLevel": {"enabled": true}
}`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

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

	var raw map[string]json.RawMessage
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("os.ReadFile returned error: %v", err)
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if _, ok := raw["FutureTopLevel"]; !ok {
		t.Fatal("unknown top-level field was not preserved")
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
