// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"encoding/base64"
	"strings"
	"testing"
)

func presetSecretTestKey(t *testing.T) string {
	t.Helper()

	return base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
}

func TestApplyPresetSecretsLeavesPlaintextWithoutKey(t *testing.T) {
	t.Setenv(PresetSecretKeyEnv, "")

	presets, changed, err := ApplyPresetSecrets([]Preset{
		{
			Title: "Atlantis",
			Type:  "SSH",
			Host:  "atlantis.home:22",
			Meta: map[string]string{
				PresetMetaPassword: "mypassword",
			},
		},
	})
	if err != nil {
		t.Fatalf("ApplyPresetSecrets returned error: %v", err)
	}
	if changed {
		t.Fatal("ApplyPresetSecrets changed = true, want false")
	}
	if presets[0].Meta[PresetMetaPassword] != "mypassword" {
		t.Fatal("plaintext password was not preserved")
	}
}

func TestApplyPresetSecretsEncryptsPlaintextWithKey(t *testing.T) {
	t.Setenv(PresetSecretKeyEnv, presetSecretTestKey(t))

	presets, changed, err := ApplyPresetSecrets([]Preset{
		{
			Title: "Atlantis",
			Type:  "SSH",
			Host:  "atlantis.home:22",
			Meta: map[string]string{
				PresetMetaPassword: "mypassword",
			},
		},
	})
	if err != nil {
		t.Fatalf("ApplyPresetSecrets returned error: %v", err)
	}
	if !changed {
		t.Fatal("ApplyPresetSecrets changed = false, want true")
	}
	if _, ok := presets[0].Meta[PresetMetaPassword]; ok {
		t.Fatal("plaintext password was still present in preset meta")
	}
	if !strings.HasPrefix(
		presets[0].Meta[PresetMetaEncryptedPassword],
		"v1:aes-256-gcm:",
	) {
		t.Fatalf(
			"encrypted password = %q, want v1 aes-gcm format",
			presets[0].Meta[PresetMetaEncryptedPassword],
		)
	}
	if presets[0].SecretMeta[PresetMetaPassword] != "mypassword" {
		t.Fatal("decrypted password was not stored in SecretMeta")
	}
}

func TestApplyPresetSecretsDecryptsEncryptedPasswordWithKey(t *testing.T) {
	t.Setenv(PresetSecretKeyEnv, presetSecretTestKey(t))
	encrypted, changed, err := ApplyPresetSecrets([]Preset{
		{
			Meta: map[string]string{
				PresetMetaPassword: "mypassword",
			},
		},
	})
	if err != nil {
		t.Fatalf("initial ApplyPresetSecrets returned error: %v", err)
	}
	if !changed {
		t.Fatal("initial ApplyPresetSecrets changed = false, want true")
	}

	presets, changed, err := ApplyPresetSecrets([]Preset{
		{
			Meta: map[string]string{
				PresetMetaEncryptedPassword: encrypted[0].Meta[PresetMetaEncryptedPassword],
			},
		},
	})
	if err != nil {
		t.Fatalf("ApplyPresetSecrets returned error: %v", err)
	}
	if changed {
		t.Fatal("ApplyPresetSecrets changed = true, want false")
	}
	if presets[0].SecretMeta[PresetMetaPassword] != "mypassword" {
		t.Fatal("encrypted password was not decrypted into SecretMeta")
	}
}

func TestApplyPresetSecretsRejectsEncryptedPasswordWithoutKey(t *testing.T) {
	t.Setenv(PresetSecretKeyEnv, "")

	_, _, err := ApplyPresetSecrets([]Preset{
		{
			Meta: map[string]string{
				PresetMetaEncryptedPassword: "v1:aes-256-gcm:nonce:ciphertext",
			},
		},
	})
	if err == nil {
		t.Fatal("ApplyPresetSecrets returned nil error, want missing key error")
	}
}
