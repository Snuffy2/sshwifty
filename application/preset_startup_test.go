// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package application

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Snuffy2/sshwifty/application/configuration"
	"github.com/Snuffy2/sshwifty/application/log"
)

func TestNormalizeStartupPresetIDsPersistsFileBackedIDs(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	configData := map[string]any{
		"Servers": []map[string]any{
			{"ListenInterface": "127.0.0.1", "ListenPort": 8182},
		},
		"Presets": []map[string]any{
			{"Title": "Atlantis", "Type": "SSH", "Host": "atlantis.home"},
		},
	}
	content, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent returned error: %v", err)
	}
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	_, cfg, err := configuration.CustomFile(configPath)(log.Ditch{})
	if err != nil {
		t.Fatalf("CustomFile returned error: %v", err)
	}
	normalized, err := normalizeStartupPresetIDs(cfg)
	if err != nil {
		t.Fatalf("normalizeStartupPresetIDs returned error: %v", err)
	}
	if normalized.Presets[0].ID == "" {
		t.Fatal("normalized preset ID is empty")
	}

	_, reloaded, err := configuration.CustomFile(configPath)(log.Ditch{})
	if err != nil {
		t.Fatalf("second CustomFile returned error: %v", err)
	}
	if reloaded.Presets[0].ID != normalized.Presets[0].ID {
		t.Fatalf("persisted ID = %q, want %q",
			reloaded.Presets[0].ID,
			normalized.Presets[0].ID,
		)
	}
}
