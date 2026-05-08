// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

//go:build windows

package command

import (
	"os/exec"
	"syscall"
)

// configureExecCommand configures given `e` for Windows
func configureExecCommand(e *exec.Cmd) {
	e.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
