// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"github.com/Snuffy2/shellport/application/log"
)

// PresetReloader reloads preset
type PresetReloader func(p Preset) (Preset, error)

// Loader Configuration loader
type Loader func(log log.Logger) (name string, cfg Configuration, err error)
