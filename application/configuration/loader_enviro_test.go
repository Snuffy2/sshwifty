// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"testing"
	"time"

	"github.com/Snuffy2/sshwifty/application/log"
)

func TestEnvironUsesWriteDelayEnvironmentVariable(t *testing.T) {
	t.Setenv("SSHWIFTY_WRITEDELAY", "25")
	t.Setenv("SSHWIFTY_WRITEELAY", "99")

	name, cfg, err := Environ()(log.NewDitch())
	if err != nil {
		t.Fatalf("Environ returned an error: %s", err)
	}
	if name != environTypeName {
		t.Fatalf("Expected loader name %q, got %q", environTypeName, name)
	}
	if len(cfg.Servers) != 1 {
		t.Fatalf("Expected one server, got %d", len(cfg.Servers))
	}
	if cfg.Servers[0].WriteDelay != 25*time.Millisecond {
		t.Fatalf(
			"Expected WriteDelay to use SSHWIFTY_WRITEDELAY, got %s",
			cfg.Servers[0].WriteDelay,
		)
	}
}
