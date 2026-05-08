// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

//go:build !(darwin || dragonfly || freebsd || linux || netbsd || openbsd || windows)

package command

import "os/exec"

// configureExecCommand configures given `e`
func configureExecCommand(e *exec.Cmd) {
	// By default, do nothing
}
