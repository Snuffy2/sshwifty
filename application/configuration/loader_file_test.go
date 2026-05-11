// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFileRejectsPresetSecretKey(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "shellport.conf.json")
	content := []byte(`{
  "HostName": "localhost",
  "PresetSecretKey": "not-allowed",
  "Servers": [
    {"ListenInterface": "127.0.0.1", "ListenPort": 8182}
  ]
}`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	_, _, err := loadFile(configPath)
	if err == nil {
		t.Fatal("loadFile returned nil error, want preset secret key error")
	}
	if !strings.Contains(err.Error(), PresetSecretKeyEnv) {
		t.Fatalf("loadFile error = %q, want %s", err, PresetSecretKeyEnv)
	}
}

func TestLoadFileRejectsPresetSecretKeyEnvName(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "shellport.conf.json")
	content := []byte(`{
  "HostName": "localhost",
  "SHELLPORT_PRESET_SECRET_KEY": "not-allowed",
  "Servers": [
    {"ListenInterface": "127.0.0.1", "ListenPort": 8182}
  ]
}`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	_, _, err := loadFile(configPath)
	if err == nil {
		t.Fatal("loadFile returned nil error, want preset secret key error")
	}
	if !strings.Contains(err.Error(), PresetSecretKeyEnv) {
		t.Fatalf("loadFile error = %q, want %s", err, PresetSecretKeyEnv)
	}
}
