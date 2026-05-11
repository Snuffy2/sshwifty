// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"testing"

	"github.com/Snuffy2/shellport/application/command"
	"github.com/Snuffy2/shellport/application/configuration"
	"github.com/Snuffy2/shellport/application/log"
)

// TestSSHCommandKeepsBufferPoolScopedToSession verifies that SSH clients retain
// the per-session buffer pool supplied by the command handler.
func TestSSHCommandKeepsBufferPoolScopedToSession(t *testing.T) {
	bufferPool := command.NewBufferPool(4096)
	poolPtr := &bufferPool

	got := newSSH(
		log.NewDitch(),
		command.NewHooks(configuration.HookSettings{}),
		command.StreamResponder{},
		command.Configuration{},
		poolPtr,
	)
	client, ok := got.(*sshClient)
	if !ok {
		t.Fatalf("expected *sshClient, got %T", got)
	}

	if client.bufferPool != poolPtr {
		t.Fatalf(
			"expected ssh client buffer pool %p, got %p",
			poolPtr,
			client.bufferPool,
		)
	}
}
