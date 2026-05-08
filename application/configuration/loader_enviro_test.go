// Sshwifty - A Web SSH client
//
// Copyright (C) 2019-2025 Ni Rui <ranqus@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
