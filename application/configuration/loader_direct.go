// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"github.com/Snuffy2/shellport/application/log"
)

const (
	directTypeName = "Direct"
)

// Direct creates a loader that return raw configuration data directly.
// Good for integration.
func Direct(cfg Configuration) Loader {
	return func(log log.Logger) (string, Configuration, error) {
		return directTypeName, cfg, nil
	}
}
