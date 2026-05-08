// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"os"

	"github.com/Snuffy2/sshwifty/application"
	"github.com/Snuffy2/sshwifty/application/commands"
	"github.com/Snuffy2/sshwifty/application/configuration"
	"github.com/Snuffy2/sshwifty/application/controller"
	"github.com/Snuffy2/sshwifty/application/log"
)

func main() {
	configLoaders := make([]configuration.Loader, 0, 2)
	if cfgFile := configuration.GetEnv("SSHWIFTY_CONFIG"); len(cfgFile) > 0 {
		configLoaders = append(configLoaders, configuration.CustomFile(cfgFile))
	} else {
		configLoaders = append(configLoaders, configuration.DefaultFile())
		configLoaders = append(configLoaders, configuration.Environ())
	}
	e := application.
		New(os.Stderr, log.NewDebugOrNonDebugWriter(
			len(configuration.GetEnv("SSHWIFTY_DEBUG")) > 0,
			application.Name,
			os.Stderr,
		)).
		Run(configuration.Redundant(configLoaders...),
			application.DefaultProccessSignallerBuilder,
			commands.New(),
			controller.Builder,
		)
	if e == nil {
		return
	}
	os.Exit(1)
}
