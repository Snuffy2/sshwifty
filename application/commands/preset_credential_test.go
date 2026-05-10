// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"testing"

	"github.com/Snuffy2/sshwifty/application/command"
	"github.com/Snuffy2/sshwifty/application/configuration"
)

func TestPresetPasswordCredentialMatchesHostUserAndPasswordAuth(t *testing.T) {
	credential, ok := presetPasswordCredential(
		command.Configuration{
			Presets: []configuration.Preset{
				{
					Type: "SSH",
					Host: "atlantis.home:22",
					Meta: map[string]string{
						"Authentication": "Password",
						"User":           "pi",
					},
					SecretMeta: map[string]string{
						"Password": "mypassword",
					},
				},
			},
		},
		"pi",
		"atlantis.home:22",
	)

	if !ok {
		t.Fatal("presetPasswordCredential ok = false, want true")
	}
	if credential != "mypassword" {
		t.Fatalf("credential = %q, want mypassword", credential)
	}
}

func TestPresetPasswordCredentialReturnsFalseWhenHostDoesNotMatch(t *testing.T) {
	_, ok := presetPasswordCredential(
		command.Configuration{
			Presets: []configuration.Preset{
				{
					Type: "SSH",
					Host: "atlantis.home:22",
					Meta: map[string]string{
						"Authentication": "Password",
						"User":           "pi",
						"Password":       "mypassword",
					},
				},
			},
		},
		"pi",
		"other.home:22",
	)

	if ok {
		t.Fatal("presetPasswordCredential ok = true, want false for non-matching host")
	}
}

func TestPresetPasswordCredentialReturnsFalseWhenUserDoesNotMatch(t *testing.T) {
	_, ok := presetPasswordCredential(
		command.Configuration{
			Presets: []configuration.Preset{
				{
					Type: "SSH",
					Host: "atlantis.home:22",
					Meta: map[string]string{
						"Authentication": "Password",
						"User":           "pi",
						"Password":       "mypassword",
					},
				},
			},
		},
		"otheruser",
		"atlantis.home:22",
	)

	if ok {
		t.Fatal("presetPasswordCredential ok = true, want false for non-matching user")
	}
}

func TestPresetPasswordCredentialReturnsFalseWhenAuthMethodIsNotPassword(t *testing.T) {
	_, ok := presetPasswordCredential(
		command.Configuration{
			Presets: []configuration.Preset{
				{
					Type: "SSH",
					Host: "atlantis.home:22",
					Meta: map[string]string{
						"Authentication": "Private Key",
						"User":           "pi",
					},
					SecretMeta: map[string]string{
						"Password": "mypassword",
					},
				},
			},
		},
		"pi",
		"atlantis.home:22",
	)

	if ok {
		t.Fatal("presetPasswordCredential ok = true, want false when Authentication is not Password")
	}
}

func TestPresetPasswordCredentialReturnsFalseWhenNoPresetsAvailable(t *testing.T) {
	_, ok := presetPasswordCredential(
		command.Configuration{
			Presets: []configuration.Preset{},
		},
		"pi",
		"atlantis.home:22",
	)

	if ok {
		t.Fatal("presetPasswordCredential ok = true, want false for empty preset list")
	}
}

func TestPresetPasswordCredentialFallsBackToPlaintextMetaWhenSecretMetaEmpty(t *testing.T) {
	credential, ok := presetPasswordCredential(
		command.Configuration{
			Presets: []configuration.Preset{
				{
					Type: "SSH",
					Host: "atlantis.home:22",
					Meta: map[string]string{
						"Authentication": "Password",
						"User":           "pi",
						"Password":       "plainpassword",
					},
					// SecretMeta is nil - should fall back to Meta password
				},
			},
		},
		"pi",
		"atlantis.home:22",
	)

	if !ok {
		t.Fatal("presetPasswordCredential ok = false, want true for plaintext Meta password")
	}
	if credential != "plainpassword" {
		t.Fatalf("credential = %q, want plainpassword", credential)
	}
}

func TestPresetPasswordCredentialUsesLivePresetRepository(t *testing.T) {
	repo := configuration.NewPresetRepository([]configuration.Preset{
		{
			Type: "SSH",
			Host: "atlantis.home:22",
			Meta: map[string]string{
				"Authentication": "Password",
				"User":           "pi",
				"Password":       "oldpassword",
			},
		},
	})
	repo.Replace([]configuration.Preset{
		{
			Type: "SSH",
			Host: "atlantis.home:22",
			Meta: map[string]string{
				"Authentication": "Password",
				"User":           "pi",
				"Password":       "newpassword",
			},
		},
	})

	credential, ok := presetPasswordCredential(
		command.Configuration{
			PresetRepository: repo,
			Presets: []configuration.Preset{
				{
					Type: "SSH",
					Host: "atlantis.home:22",
					Meta: map[string]string{
						"Authentication": "Password",
						"User":           "pi",
						"Password":       "oldpassword",
					},
				},
			},
		},
		"pi",
		"atlantis.home:22",
	)

	if !ok {
		t.Fatal("presetPasswordCredential ok = false, want true")
	}
	if credential != "newpassword" {
		t.Fatalf("credential = %q, want newpassword", credential)
	}
}
