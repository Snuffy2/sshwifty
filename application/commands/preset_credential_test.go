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
