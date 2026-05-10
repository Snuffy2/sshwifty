// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"github.com/Snuffy2/sshwifty/application/command"
	"github.com/Snuffy2/sshwifty/application/configuration"
)

func presetPasswordCredential(
	cfg command.Configuration,
	user string,
	host string,
) (string, bool) {
	presets := cfg.Presets
	if cfg.PresetRepository != nil {
		presets = cfg.PresetRepository.List()
	}
	for _, preset := range presets {
		if preset.Host != host {
			continue
		}
		if preset.Meta["User"] != user {
			continue
		}
		if preset.Meta["Authentication"] != "Password" {
			continue
		}
		if preset.SecretMeta != nil {
			if credential := preset.SecretMeta[configuration.PresetMetaPassword]; credential != "" {
				return credential, true
			}
		}
		if credential := preset.Meta[configuration.PresetMetaPassword]; credential != "" {
			return credential, true
		}
	}
	return "", false
}
