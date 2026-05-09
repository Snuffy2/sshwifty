// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"reflect"
	"testing"

	"github.com/Snuffy2/sshwifty/application/command"
	"github.com/Snuffy2/sshwifty/application/configuration"
	"github.com/Snuffy2/sshwifty/application/log"
)

// TestTelnetCommandKeepsBufferPoolScopedToSession verifies that Telnet clients
// retain the per-session buffer pool supplied by the command handler.
func TestTelnetCommandKeepsBufferPoolScopedToSession(t *testing.T) {
	bufferPool := command.NewBufferPool(4096)
	poolPtr := &bufferPool

	r := newTelnet(
		log.NewDitch(),
		command.NewHooks(configuration.HookSettings{}),
		command.StreamResponder{},
		command.Configuration{},
		poolPtr,
	)
	gotType := reflect.TypeOf(r)
	client, ok := r.(*telnetClient)
	if !ok {
		t.Fatalf("expected *telnetClient, got %v", gotType)
	}

	if client.bufferPool != poolPtr {
		t.Fatalf(
			"expected telnet client buffer pool %p, got %p",
			poolPtr,
			client.bufferPool,
		)
	}
}
