// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"testing"
)

func TestCommonCurrentPresetsReturnsRepositoryListWhenSet(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-live", Title: "Live", Host: "live.home:22"},
	})
	c := Common{
		Presets:          []Preset{{ID: "preset-static", Title: "Static"}},
		PresetRepository: repo,
	}

	got := c.CurrentPresets()
	if len(got) != 1 {
		t.Fatalf("CurrentPresets len = %d, want 1", len(got))
	}
	if got[0].ID != "preset-live" {
		t.Fatalf("CurrentPresets[0].ID = %q, want preset-live", got[0].ID)
	}
}

func TestCommonCurrentPresetsReturnsStaticPresetsWhenRepositoryNil(t *testing.T) {
	c := Common{
		Presets:          []Preset{{ID: "preset-static", Title: "Static"}},
		PresetRepository: nil,
	}

	got := c.CurrentPresets()
	if len(got) != 1 {
		t.Fatalf("CurrentPresets len = %d, want 1", len(got))
	}
	if got[0].ID != "preset-static" {
		t.Fatalf("CurrentPresets[0].ID = %q, want preset-static", got[0].ID)
	}
}

func TestCommonCurrentPresetsReturnsUpdatedPresetsAfterReplace(t *testing.T) {
	repo := NewPresetRepository([]Preset{
		{ID: "preset-initial", Title: "Initial"},
	})
	c := Common{
		Presets:          []Preset{},
		PresetRepository: repo,
	}

	// Before replace
	before := c.CurrentPresets()
	if len(before) != 1 || before[0].ID != "preset-initial" {
		t.Fatalf("before Replace: CurrentPresets = %v", before)
	}

	repo.Replace([]Preset{{ID: "preset-replaced", Title: "Replaced"}})

	// After replace
	after := c.CurrentPresets()
	if len(after) != 1 || after[0].ID != "preset-replaced" {
		t.Fatalf("after Replace: CurrentPresets = %v, want preset-replaced", after)
	}
}

func TestCommonPresetConfigWritableReturnsFalseForEmptySourceFile(t *testing.T) {
	c := Common{SourceFile: ""}
	if c.PresetConfigWritable() {
		t.Fatal("PresetConfigWritable = true for empty SourceFile, want false")
	}
}

func TestConfigurationCommonPropagatesSourceFile(t *testing.T) {
	cfg := Configuration{
		SourceFile: "/etc/sshwifty.json",
		Servers: []Server{
			{ListenInterface: "127.0.0.1", ListenPort: 8182},
		},
	}
	common := cfg.Common()
	if common.SourceFile != "/etc/sshwifty.json" {
		t.Fatalf("Common.SourceFile = %q, want /etc/sshwifty.json", common.SourceFile)
	}
}

func TestConfigurationCommonCreatesPresetRepository(t *testing.T) {
	cfg := Configuration{
		Presets: []Preset{
			{ID: "preset-a", Host: "a.home:22"},
		},
	}
	common := cfg.Common()
	if common.PresetRepository == nil {
		t.Fatal("Common.PresetRepository is nil, want non-nil")
	}
	presets := common.PresetRepository.List()
	if len(presets) != 1 {
		t.Fatalf("PresetRepository.List() len = %d, want 1", len(presets))
	}
	if presets[0].ID != "preset-a" {
		t.Fatalf("PresetRepository.List()[0].ID = %q, want preset-a", presets[0].ID)
	}
}

func TestConfigurationCommonRepositoryAllowsAccessControl(t *testing.T) {
	cfg := Configuration{
		Presets: []Preset{
			{ID: "preset-allowed", Host: "allowed.home:22"},
		},
		OnlyAllowPresetRemotes: true,
	}
	common := cfg.Common()
	if common.PresetRepository == nil {
		t.Fatal("Common.PresetRepository is nil")
	}
	if !common.PresetRepository.Allowed("allowed.home:22") {
		t.Fatal("PresetRepository.Allowed(allowed.home:22) = false, want true")
	}
	if common.PresetRepository.Allowed("blocked.home:22") {
		t.Fatal("PresetRepository.Allowed(blocked.home:22) = true, want false")
	}
}

func TestConfigurationCommonPreservesPresetList(t *testing.T) {
	cfg := Configuration{
		Presets: []Preset{
			{ID: "preset-1", Title: "One"},
			{ID: "preset-2", Title: "Two"},
		},
	}
	common := cfg.Common()
	if len(common.Presets) != 2 {
		t.Fatalf("Common.Presets len = %d, want 2", len(common.Presets))
	}
}