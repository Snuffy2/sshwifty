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
		"SSH",
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
		"SSH",
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

func TestPresetPasswordCredentialMatchesCommandType(t *testing.T) {
	credential, ok := presetPasswordCredential(
		command.Configuration{
			Presets: []configuration.Preset{
				{
					Type: "Mosh",
					Host: "atlantis.home:22",
					Meta: map[string]string{
						"Authentication": "Password",
						"User":           "pi",
						"Password":       "moshpassword",
					},
				},
				{
					Type: "SSH",
					Host: "atlantis.home:22",
					Meta: map[string]string{
						"Authentication": "Password",
						"User":           "pi",
						"Password":       "sshpassword",
					},
				},
			},
		},
		"SSH",
		"pi",
		"atlantis.home:22",
	)

	if !ok {
		t.Fatal("presetPasswordCredential ok = false, want true")
	}
	if credential != "sshpassword" {
		t.Fatalf("credential = %q, want sshpassword", credential)
	}
}
