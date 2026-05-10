// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"bytes"
	"encoding/base64"
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

func newAuthenticatedTestPresetConfig(t *testing.T, configPath string) presetConfig {
	t.Helper()

	controller := newTestPresetConfig(t, configPath)
	controller.commonCfg.SharedKey = "test-shared-key"
	return controller
}

func newAdminTestPresetConfig(t *testing.T, configPath string) presetConfig {
	t.Helper()

	controller := newAuthenticatedTestPresetConfig(t, configPath)
	controller.commonCfg.PresetAdminKey = "test-preset-admin-key"
	return controller
}

func authorizePresetConfigRequest(controller presetConfig, request *http.Request) {
	verifier := newSocketVerification(
		socket{commonCfg: controller.commonCfg},
		configuration.Server{},
		controller.commonCfg,
	)
	request.Header.Set(
		"X-Key",
		base64.StdEncoding.EncodeToString(verifier.authKey(request)),
	)
}

func authorizePresetAdminRequest(controller presetConfig, request *http.Request) {
	request.Header.Set(
		presetAdminKeyHeader,
		base64.StdEncoding.EncodeToString(
			presetAdminAuthKey(controller.commonCfg.PresetAdminKey),
		),
	)
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
	controller := newAdminTestPresetConfig(t, configPath)
	body := []byte(`{"presets":[{"title":"Columbia","type":"SSH","host":"columbia.home","meta":{"User":"pi"}}]}`)
	request := httptest.NewRequest(http.MethodPut, "/sshwifty/config/presets", bytes.NewReader(body))
	authorizePresetConfigRequest(controller, request)
	authorizePresetAdminRequest(controller, request)
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

func TestPresetConfigPutRemovesSupportedPresetsAndPreservesRawUnsupported(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, []map[string]any{
		{"ID": "preset-atlantis", "Title": "Atlantis", "Type": "SSH", "Host": "atlantis.home"},
		{"ID": "preset-future", "Title": "Future", "Type": "Future", "Host": "future.home"},
	})
	controller := newAdminTestPresetConfig(t, configPath)
	body := []byte(`{"presets":[{"id":"preset-columbia","title":"Columbia","type":"SSH","host":"columbia.home","meta":{"User":"pi"}}]}`)
	request := httptest.NewRequest(http.MethodPut, "/sshwifty/config/presets", bytes.NewReader(body))
	authorizePresetConfigRequest(controller, request)
	authorizePresetAdminRequest(controller, request)
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	if err := controller.Put(&writer, request, log.Ditch{}); err != nil {
		t.Fatalf("Put returned error: %v", err)
	}

	var raw struct {
		Presets []struct {
			ID string
		}
	}
	f, err := os.Open(configPath)
	if err != nil {
		t.Fatalf("os.Open returned error: %v", err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		t.Fatalf("json Decode returned error: %v", err)
	}
	if len(raw.Presets) != 2 {
		t.Fatalf("raw preset count = %d, want 2", len(raw.Presets))
	}
	if raw.Presets[0].ID != "preset-columbia" {
		t.Fatalf("first raw preset ID = %q, want preset-columbia", raw.Presets[0].ID)
	}
	if raw.Presets[1].ID != "preset-future" {
		t.Fatalf("second raw preset ID = %q, want preset-future", raw.Presets[1].ID)
	}
}

func TestPresetConfigPutRejectsDuplicateIDs(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, nil)
	controller := newAdminTestPresetConfig(t, configPath)
	body := []byte(`{"presets":[{"id":"dup","title":"A","type":"SSH","host":"a.home"},{"id":"dup","title":"B","type":"SSH","host":"b.home"}]}`)
	request := httptest.NewRequest(http.MethodPut, "/sshwifty/config/presets", bytes.NewReader(body))
	authorizePresetConfigRequest(controller, request)
	authorizePresetAdminRequest(controller, request)
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	err := controller.Put(&writer, request, log.Ditch{})
	if err == nil {
		t.Fatal("Put returned nil error, want duplicate ID error")
	}
}

func TestPresetConfigPutRequiresSharedKey(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, nil)
	controller := newTestPresetConfig(t, configPath)
	body := []byte(`{"presets":[{"id":"preset-atlantis","title":"Atlantis","type":"SSH","host":"atlantis.home"}]}`)
	request := httptest.NewRequest(http.MethodPut, "/sshwifty/config/presets", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	err := controller.Put(&writer, request, log.Ditch{})
	if err == nil {
		t.Fatal("Put returned nil error, want authentication requirement")
	}
}

func TestPresetConfigPutRequiresPresetAdminKeyForFullReplacement(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, []map[string]any{
		{"ID": "preset-atlantis", "Title": "Atlantis", "Type": "SSH", "Host": "atlantis.home"},
	})
	controller := newAuthenticatedTestPresetConfig(t, configPath)
	body := []byte(`{"presets":[{"id":"preset-columbia","title":"Columbia","type":"SSH","host":"columbia.home"}]}`)
	request := httptest.NewRequest(http.MethodPut, "/sshwifty/config/presets", bytes.NewReader(body))
	authorizePresetConfigRequest(controller, request)
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	err := controller.Put(&writer, request, log.Ditch{})
	if err == nil {
		t.Fatal("Put returned nil error, want preset admin authentication error")
	}
}

func TestPresetConfigPutRejectsIDsDuplicatedAfterTrimming(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, nil)
	controller := newAdminTestPresetConfig(t, configPath)
	body := []byte(`{"presets":[{"id":" dup ","title":"A","type":"SSH","host":"a.home"},{"id":"dup","title":"B","type":"SSH","host":"b.home"}]}`)
	request := httptest.NewRequest(
		http.MethodPut,
		"/sshwifty/config/presets",
		bytes.NewReader(body),
	)
	authorizePresetConfigRequest(controller, request)
	authorizePresetAdminRequest(controller, request)
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	err := controller.Put(&writer, request, log.Ditch{})
	if err == nil {
		t.Fatal("Put returned nil error, want duplicate ID error")
	}
}

func TestPresetConfigPutEncryptsPlaintextPasswordsWhenKeyIsSet(t *testing.T) {
	t.Setenv(
		configuration.PresetSecretKeyEnv,
		base64.StdEncoding.EncodeToString(
			[]byte("0123456789abcdef0123456789abcdef"),
		),
	)
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, nil)
	controller := newAdminTestPresetConfig(t, configPath)
	body := []byte(`{"presets":[{"id":"preset-atlantis","title":"Atlantis","type":"SSH","host":"atlantis.home","meta":{"User":"pi","Authentication":"Password","Password":"mypassword"}}]}`)
	request := httptest.NewRequest(
		http.MethodPut,
		"/sshwifty/config/presets",
		bytes.NewReader(body),
	)
	authorizePresetConfigRequest(controller, request)
	authorizePresetAdminRequest(controller, request)
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	if err := controller.Put(&writer, request, log.Ditch{}); err != nil {
		t.Fatalf("Put returned error: %v", err)
	}

	var response presetConfigResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("json Decode returned error: %v", err)
	}
	if _, ok := response.Presets[0].Meta[configuration.PresetMetaPassword]; ok {
		t.Fatal("response still contains plaintext Password")
	}
	if _, ok := response.Presets[0].Meta[configuration.PresetMetaEncryptedPassword]; ok {
		t.Fatal("response exposed Encrypted Password")
	}

	_, reloaded, err := configuration.CustomFile(configPath)(log.Ditch{})
	if err != nil {
		t.Fatalf("CustomFile returned error: %v", err)
	}
	if _, ok := reloaded.Presets[0].Meta[configuration.PresetMetaPassword]; ok {
		t.Fatal("persisted config still contains plaintext Password")
	}
	if reloaded.Presets[0].Meta[configuration.PresetMetaEncryptedPassword] == "" {
		t.Fatal("persisted config missing Encrypted Password")
	}
}

func TestPresetConfigPutPreservesHiddenPasswordOnFingerprintSave(t *testing.T) {
	t.Setenv(
		configuration.PresetSecretKeyEnv,
		base64.StdEncoding.EncodeToString(
			[]byte("0123456789abcdef0123456789abcdef"),
		),
	)
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, []map[string]any{
		{
			"ID":    "preset-atlantis",
			"Title": "Atlantis",
			"Type":  "SSH",
			"Host":  "atlantis.home",
			"Meta": map[string]string{
				"User":           "pi",
				"Authentication": "Password",
				"Password":       "mypassword",
			},
		},
	})
	controller := newAuthenticatedTestPresetConfig(t, configPath)
	body := []byte(`{"presets":[{"id":"preset-atlantis","title":"Atlantis","type":"SSH","host":"atlantis.home:22","meta":{"User":"pi","Authentication":"Password","Fingerprint":"SHA256:abc"}}]}`)
	request := httptest.NewRequest(
		http.MethodPut,
		"/sshwifty/config/presets",
		bytes.NewReader(body),
	)
	authorizePresetConfigRequest(controller, request)
	request.Header.Set(preserveHiddenPresetPasswordsHeader, "yes")
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	if err := controller.Put(&writer, request, log.Ditch{}); err != nil {
		t.Fatalf("Put returned error: %v", err)
	}

	live := controller.commonCfg.CurrentPresets()
	if live[0].SecretMeta[configuration.PresetMetaPassword] != "mypassword" {
		t.Fatal("live preset lost hidden password")
	}
	if live[0].Meta["Fingerprint"] != "SHA256:abc" {
		t.Fatal("live preset missing fingerprint")
	}

	_, reloaded, err := configuration.CustomFile(configPath)(log.Ditch{})
	if err != nil {
		t.Fatalf("CustomFile returned error: %v", err)
	}
	if _, ok := reloaded.Presets[0].Meta[configuration.PresetMetaPassword]; ok {
		t.Fatal("persisted config still contains plaintext Password")
	}
	if reloaded.Presets[0].Meta[configuration.PresetMetaEncryptedPassword] == "" {
		t.Fatal("persisted config missing Encrypted Password")
	}
	if reloaded.Presets[0].Meta["Fingerprint"] != "SHA256:abc" {
		t.Fatal("persisted config missing fingerprint")
	}
}

func TestPresetConfigPutRejectsFingerprintSaveChangingHost(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, []map[string]any{
		{
			"ID":    "preset-atlantis",
			"Title": "Atlantis",
			"Type":  "SSH",
			"Host":  "atlantis.home",
			"Meta": map[string]string{
				"User": "pi",
			},
		},
	})
	controller := newAuthenticatedTestPresetConfig(t, configPath)
	body := []byte(`{"presets":[{"id":"preset-atlantis","title":"Atlantis","type":"SSH","host":"evil.home:22","meta":{"User":"pi","Fingerprint":"SHA256:abc"}}]}`)
	request := httptest.NewRequest(
		http.MethodPut,
		"/sshwifty/config/presets",
		bytes.NewReader(body),
	)
	authorizePresetConfigRequest(controller, request)
	request.Header.Set(preserveHiddenPresetPasswordsHeader, "yes")
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	err := controller.Put(&writer, request, log.Ditch{})
	if err == nil {
		t.Fatal("Put returned nil error, want fingerprint-only validation error")
	}
}

func TestPresetConfigPutCanDeleteHiddenPassword(t *testing.T) {
	t.Setenv(
		configuration.PresetSecretKeyEnv,
		base64.StdEncoding.EncodeToString(
			[]byte("0123456789abcdef0123456789abcdef"),
		),
	)
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, []map[string]any{
		{
			"ID":    "preset-atlantis",
			"Title": "Atlantis",
			"Type":  "SSH",
			"Host":  "atlantis.home",
			"Meta": map[string]string{
				"User":           "pi",
				"Authentication": "Password",
				"Password":       "mypassword",
			},
		},
	})
	controller := newAdminTestPresetConfig(t, configPath)
	body := []byte(`{"presets":[{"id":"preset-atlantis","title":"Atlantis","type":"SSH","host":"atlantis.home:22","meta":{"User":"pi","Authentication":"Password"}}]}`)
	request := httptest.NewRequest(
		http.MethodPut,
		"/sshwifty/config/presets",
		bytes.NewReader(body),
	)
	authorizePresetConfigRequest(controller, request)
	authorizePresetAdminRequest(controller, request)
	recorder := httptest.NewRecorder()
	writer := newResponseWriter(recorder)

	if err := controller.Put(&writer, request, log.Ditch{}); err != nil {
		t.Fatalf("Put returned error: %v", err)
	}

	live := controller.commonCfg.CurrentPresets()
	if _, ok := live[0].Meta[configuration.PresetMetaEncryptedPassword]; ok {
		t.Fatal("live preset still contains Encrypted Password")
	}
	if _, ok := live[0].SecretMeta[configuration.PresetMetaPassword]; ok {
		t.Fatal("live preset still contains hidden Password")
	}

	_, reloaded, err := configuration.CustomFile(configPath)(log.Ditch{})
	if err != nil {
		t.Fatalf("CustomFile returned error: %v", err)
	}
	if _, ok := reloaded.Presets[0].Meta[configuration.PresetMetaEncryptedPassword]; ok {
		t.Fatal("persisted config still contains Encrypted Password")
	}
	if _, ok := reloaded.Presets[0].Meta[configuration.PresetMetaPassword]; ok {
		t.Fatal("persisted config still contains Password")
	}
}

func TestSocketAccessConfigurationIncludesPresetConfigWritable(t *testing.T) {
	accessConfig := newSocketAccessConfiguration(nil, "", true)

	if !accessConfig.PresetConfigWritable {
		t.Fatal("PresetConfigWritable = false, want true")
	}
}

func TestSocketVerificationAdvertisesPresetConfigWritableOnlyWithSharedKey(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sshwifty.conf.json")
	writePresetAPIConfig(t, configPath, nil)

	withoutKey := newSocketVerification(
		socket{},
		configuration.Server{},
		configuration.Common{SourceFile: configPath},
	)
	var withoutKeyConfig socketAccessConfiguration
	if err := json.Unmarshal(withoutKey.configRspBody, &withoutKeyConfig); err != nil {
		t.Fatalf("json.Unmarshal without key returned error: %v", err)
	}
	if withoutKeyConfig.PresetConfigWritable {
		t.Fatal("PresetConfigWritable = true without SharedKey, want false")
	}

	withKey := newSocketVerification(
		socket{},
		configuration.Server{},
		configuration.Common{SourceFile: configPath, SharedKey: "secret"},
	)
	var withKeyConfig socketAccessConfiguration
	if err := json.Unmarshal(withKey.configRspBody, &withKeyConfig); err != nil {
		t.Fatalf("json.Unmarshal with key returned error: %v", err)
	}
	if !withKeyConfig.PresetConfigWritable {
		t.Fatal("PresetConfigWritable = false with SharedKey, want true")
	}
}

func TestSocketAccessConfigurationDoesNotExposePlaintextPasswords(t *testing.T) {
	accessConfig := newSocketAccessConfiguration([]configuration.Preset{
		{
			Title: "Atlantis",
			Type:  "SSH",
			Host:  "atlantis.home:22",
			Meta: map[string]string{
				configuration.PresetMetaPassword:          "mypassword",
				configuration.PresetMetaEncryptedPassword: "encrypted",
			},
		},
	}, "", true)

	if _, ok := accessConfig.Presets[0].Meta[configuration.PresetMetaPassword]; ok {
		t.Fatal("socket access configuration exposed plaintext Password")
	}
	if _, ok := accessConfig.Presets[0].Meta[configuration.PresetMetaEncryptedPassword]; ok {
		t.Fatal("socket access configuration exposed Encrypted Password")
	}
}
