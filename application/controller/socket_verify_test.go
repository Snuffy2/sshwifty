// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"testing"

	"github.com/Snuffy2/sshwifty/application/configuration"
)

func TestSanitizeSocketPresetMetaRemovesPasswordKeys(t *testing.T) {
	meta := map[string]string{
		"User":                                   "pi",
		configuration.PresetMetaPassword:          "secret",
		configuration.PresetMetaEncryptedPassword: "v1:aes-256-gcm:...",
		"Authentication":                          "Password",
	}

	sanitized := sanitizeSocketPresetMeta(meta)

	if _, ok := sanitized[configuration.PresetMetaPassword]; ok {
		t.Fatal("sanitized meta still contains plaintext Password key")
	}
	if _, ok := sanitized[configuration.PresetMetaEncryptedPassword]; ok {
		t.Fatal("sanitized meta still contains Encrypted Password key")
	}
	if sanitized["User"] != "pi" {
		t.Fatalf("User = %q, want pi", sanitized["User"])
	}
	if sanitized["Authentication"] != "Password" {
		t.Fatalf("Authentication = %q, want Password", sanitized["Authentication"])
	}
}

func TestSanitizeSocketPresetMetaWithNoPasswordKeys(t *testing.T) {
	meta := map[string]string{
		"User":           "pi",
		"Authentication": "Private Key",
		"Fingerprint":    "SHA256:abc",
	}

	sanitized := sanitizeSocketPresetMeta(meta)

	if len(sanitized) != 3 {
		t.Fatalf("sanitized meta len = %d, want 3", len(sanitized))
	}
	if sanitized["Fingerprint"] != "SHA256:abc" {
		t.Fatalf("Fingerprint = %q, want SHA256:abc", sanitized["Fingerprint"])
	}
}

func TestSanitizeSocketPresetMetaWithNilMetaReturnsEmpty(t *testing.T) {
	sanitized := sanitizeSocketPresetMeta(nil)

	if sanitized == nil {
		t.Fatal("sanitizeSocketPresetMeta(nil) returned nil, want empty map")
	}
	if len(sanitized) != 0 {
		t.Fatalf("sanitized meta len = %d, want 0", len(sanitized))
	}
}

func TestSanitizeSocketPresetMetaWithEmptyMetaReturnsEmpty(t *testing.T) {
	sanitized := sanitizeSocketPresetMeta(map[string]string{})

	if len(sanitized) != 0 {
		t.Fatalf("sanitized meta len = %d, want 0", len(sanitized))
	}
}

func TestNewSocketAccessConfigurationIncludesPresetID(t *testing.T) {
	accessCfg := newSocketAccessConfiguration(
		[]configuration.Preset{
			{
				ID:    "preset-atlantis",
				Title: "Atlantis",
				Type:  "SSH",
				Host:  "atlantis.home:22",
				Meta:  map[string]string{"User": "pi"},
			},
		},
		"",
		false,
	)

	if len(accessCfg.Presets) != 1 {
		t.Fatalf("Presets len = %d, want 1", len(accessCfg.Presets))
	}
	if accessCfg.Presets[0].ID != "preset-atlantis" {
		t.Fatalf("Presets[0].ID = %q, want preset-atlantis", accessCfg.Presets[0].ID)
	}
}

func TestNewSocketAccessConfigurationPresetConfigWritableFalse(t *testing.T) {
	accessCfg := newSocketAccessConfiguration(nil, "", false)

	if accessCfg.PresetConfigWritable {
		t.Fatal("PresetConfigWritable = true, want false")
	}
}

func TestNewSocketAccessConfigurationEscapesServerMessage(t *testing.T) {
	accessCfg := newSocketAccessConfiguration(
		nil,
		"<script>alert('xss')</script>",
		false,
	)

	if accessCfg.ServerMessage == "<script>alert('xss')</script>" {
		t.Fatal("server message was not HTML escaped")
	}
}

func TestNewSocketAccessConfigurationWithEmptyPresets(t *testing.T) {
	accessCfg := newSocketAccessConfiguration([]configuration.Preset{}, "", false)

	if len(accessCfg.Presets) != 0 {
		t.Fatalf("Presets len = %d, want 0", len(accessCfg.Presets))
	}
}

func TestNewSocketAccessConfigurationDoesNotExposeSecretMeta(t *testing.T) {
	accessCfg := newSocketAccessConfiguration(
		[]configuration.Preset{
			{
				ID:   "preset-a",
				Host: "a.home:22",
				Meta: map[string]string{
					"User": "pi",
				},
				SecretMeta: map[string]string{
					configuration.PresetMetaPassword: "secret",
				},
			},
		},
		"",
		false,
	)

	// SecretMeta is server-only and should not appear in the socket meta.
	// The socket response only encodes preset.Meta, not preset.SecretMeta.
	if _, ok := accessCfg.Presets[0].Meta[configuration.PresetMetaPassword]; ok {
		t.Fatal("socket preset meta exposed server-only SecretMeta password")
	}
}