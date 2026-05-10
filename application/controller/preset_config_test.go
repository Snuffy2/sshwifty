// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Snuffy2/sshwifty/application/commands"
	"github.com/Snuffy2/sshwifty/application/configuration"
	"github.com/Snuffy2/sshwifty/application/log"
)

func writePresetAPIConfig(t *testing.T, path string, presets []map[string]any) {
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

func newTestPresetConfig(t *testing.T, configPath string) presetConfig {
	t.Helper()

	_, cfg, err := configuration.CustomFile(configPath)(log.Ditch{})
	if err != nil {
		t.Fatalf("CustomFile returned error: %v", err)
	}
	cfg, err = normalizeStartupPresetIDsForTest(cfg)
	if err != nil {
		t.Fatalf("normalizeStartupPresetIDsForTest returned error: %v", err)
	}
	cfg.Presets, err = commands.New().Reconfigure(cfg.Presets)
	if err != nil {
		t.Fatalf("Reconfigure returned error: %v", err)
	}
	return newPresetConfig(cfg.Common(), commands.New())
}

func normalizeStartupPresetIDsForTest(
	cfg configuration.Configuration,
) (configuration.Configuration, error) {
	presets, _, err := configuration.EnsurePresetIDs(cfg.Presets)
	if err != nil {
		return configuration.Configuration{}, err
	}
	cfg.Presets = presets
	return cfg, nil
}

func TestPresetConfigGetReturnsPresetIDs(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, []map[string]any{
		{"ID": "preset-atlantis", "Title": "Atlantis", "Type": "SSH", "Host": "atlantis.home"},
	})
	controller := newTestPresetConfig(t, configPath)
	request := httptest.NewRequest(http.MethodGet, "/sshwifty/config/presets", nil)
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	if err := controller.Get(&writer, request, log.Ditch{}); err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	var response presetConfigResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("json Decode returned error: %v", err)
	}
	if response.Presets[0].ID != "preset-atlantis" {
		t.Fatalf("preset ID = %q, want preset-atlantis", response.Presets[0].ID)
	}
}

func TestPresetConfigPutAddsMissingIDsAndPersists(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, []map[string]any{
		{"ID": "preset-atlantis", "Title": "Atlantis", "Type": "SSH", "Host": "atlantis.home"},
	})
	controller := newTestPresetConfig(t, configPath)
	body := []byte(`{"presets":[{"title":"Columbia","type":"SSH","host":"columbia.home","meta":{"User":"pi"}}]}`)
	request := httptest.NewRequest(http.MethodPut, "/sshwifty/config/presets", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	if err := controller.Put(&writer, request, log.Ditch{}); err != nil {
		t.Fatalf("Put returned error: %v", err)
	}

	var response presetConfigResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("json Decode returned error: %v", err)
	}
	if response.Presets[0].ID == "" {
		t.Fatal("response preset ID is empty")
	}
	if response.Presets[0].Host != "columbia.home:22" {
		t.Fatalf("response host = %q, want columbia.home:22", response.Presets[0].Host)
	}

	_, reloaded, err := configuration.CustomFile(configPath)(log.Ditch{})
	if err != nil {
		t.Fatalf("CustomFile returned error: %v", err)
	}
	if reloaded.Presets[0].ID != response.Presets[0].ID {
		t.Fatalf("persisted ID = %q, want %q", reloaded.Presets[0].ID, response.Presets[0].ID)
	}
}

func TestPresetConfigPutRejectsDuplicateIDs(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, nil)
	controller := newTestPresetConfig(t, configPath)
	body := []byte(`{"presets":[{"id":"dup","title":"A","type":"SSH","host":"a.home"},{"id":"dup","title":"B","type":"SSH","host":"b.home"}]}`)
	request := httptest.NewRequest(http.MethodPut, "/sshwifty/config/presets", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	err := controller.Put(&writer, request, log.Ditch{})
	if err == nil {
		t.Fatal("Put returned nil error, want duplicate ID error")
	}
}

func TestSocketAccessConfigurationIncludesPresetConfigWritable(t *testing.T) {
	accessConfig := newSocketAccessConfiguration(nil, "", true)

	if !accessConfig.PresetConfigWritable {
		t.Fatal("PresetConfigWritable = false, want true")
	}
}

func TestSocketAccessConfigurationPresetConfigWritableFalseByDefault(t *testing.T) {
	accessConfig := newSocketAccessConfiguration(nil, "", false)

	if accessConfig.PresetConfigWritable {
		t.Fatal("PresetConfigWritable = true, want false")
	}
}

func TestSocketAccessConfigurationPreservesPresetIDs(t *testing.T) {
	presets := []configuration.Preset{
		{ID: "preset-alpha", Title: "Alpha", Type: "SSH", Host: "alpha.home:22"},
		{ID: "preset-beta", Title: "Beta", Type: "Telnet", Host: "beta.home:23"},
	}
	accessConfig := newSocketAccessConfiguration(presets, "", false)

	if len(accessConfig.Presets) != 2 {
		t.Fatalf("Presets len = %d, want 2", len(accessConfig.Presets))
	}
	if accessConfig.Presets[0].ID != "preset-alpha" {
		t.Fatalf("Presets[0].ID = %q, want preset-alpha", accessConfig.Presets[0].ID)
	}
	if accessConfig.Presets[1].ID != "preset-beta" {
		t.Fatalf("Presets[1].ID = %q, want preset-beta", accessConfig.Presets[1].ID)
	}
}

func TestPresetConfigGetResponseHasNoCacheHeaders(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, []map[string]any{
		{"ID": "preset-atlantis", "Title": "Atlantis", "Type": "SSH", "Host": "atlantis.home"},
	})
	controller := newTestPresetConfig(t, configPath)
	request := httptest.NewRequest(http.MethodGet, "/sshwifty/config/presets", nil)
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	if err := controller.Get(&writer, request, log.Ditch{}); err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	cacheControl := recorder.Header().Get("Cache-Control")
	if cacheControl != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", cacheControl)
	}
}

func TestPresetConfigPutRejectsNonWritableConfig(t *testing.T) {
	// Build a Common config with no SourceFile so PresetConfigWritable returns false.
	commonCfg := configuration.Common{
		SourceFile:       "",
		PresetRepository: configuration.NewPresetRepository(nil),
	}
	controller := newPresetConfig(commonCfg, commands.New())
	body := []byte(`{"presets":[{"title":"X","type":"SSH","host":"x.home"}]}`)
	request := httptest.NewRequest(http.MethodPut, "/sshwifty/config/presets", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	err := controller.Put(&writer, request, log.Ditch{})
	if err == nil {
		t.Fatal("Put returned nil error for non-writable config, want error")
	}
}

func TestPresetConfigGetReturnsEmptyPresetsForEmptyConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, nil)
	controller := newTestPresetConfig(t, configPath)
	request := httptest.NewRequest(http.MethodGet, "/sshwifty/config/presets", nil)
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	if err := controller.Get(&writer, request, log.Ditch{}); err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	var response presetConfigResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("json Decode returned error: %v", err)
	}
	if len(response.Presets) != 0 {
		t.Fatalf("response preset count = %d, want 0", len(response.Presets))
	}
}

func TestPresetConfigPutInvalidJSONReturnsError(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, nil)
	controller := newTestPresetConfig(t, configPath)
	body := []byte(`{invalid json}`)
	request := httptest.NewRequest(http.MethodPut, "/sshwifty/config/presets", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	err := controller.Put(&writer, request, log.Ditch{})
	if err == nil {
		t.Fatal("Put returned nil error for invalid JSON, want error")
	}
}

func TestPresetConfigPutUpdatesLiveRepository(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, []map[string]any{
		{"ID": "preset-original", "Title": "Original", "Type": "SSH", "Host": "original.home"},
	})
	controller := newTestPresetConfig(t, configPath)

	body := []byte(`{"presets":[{"id":"preset-updated","title":"Updated","type":"SSH","host":"updated.home"}]}`)
	request := httptest.NewRequest(http.MethodPut, "/sshwifty/config/presets", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	if err := controller.Put(&writer, request, log.Ditch{}); err != nil {
		t.Fatalf("Put returned error: %v", err)
	}

	// After PUT, a subsequent GET should reflect the updated list from the repository.
	getRequest := httptest.NewRequest(http.MethodGet, "/sshwifty/config/presets", nil)
	getRecorder := httptest.NewRecorder()
	getWriter := newResponseWriter(getRecorder)

	if err := controller.Get(&getWriter, getRequest, log.Ditch{}); err != nil {
		t.Fatalf("Get after Put returned error: %v", err)
	}

	var response presetConfigResponse
	if err := json.NewDecoder(getRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("json Decode returned error: %v", err)
	}
	if len(response.Presets) != 1 {
		t.Fatalf("after Put, Get returned %d presets, want 1", len(response.Presets))
	}
	if response.Presets[0].ID != "preset-updated" {
		t.Fatalf("after Put, Get preset ID = %q, want preset-updated", response.Presets[0].ID)
	}
}
