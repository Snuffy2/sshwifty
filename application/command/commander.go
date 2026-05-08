// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package command

import (
	"io"
	"sync"
	"time"

	"github.com/Snuffy2/sshwifty/application/log"
	"github.com/Snuffy2/sshwifty/application/network"
	"github.com/Snuffy2/sshwifty/application/rw"
)

// Configuration holds the network-level settings needed to execute outbound
// connections on behalf of a command, including the dialer function and the
// maximum duration allowed for establishing a connection.
type Configuration struct {
	// Dial is the function used to open outbound network connections.
	Dial network.Dial
	// DialTimeout is the maximum duration permitted for a single dial attempt.
	DialTimeout time.Duration
}

// Commander manages the set of registered commands and produces Handler
// instances for new client sessions.
type Commander struct {
	commands Commands
}

// New creates a new Commander backed by the given set of registered commands.
func New(cs Commands) Commander {
	return Commander{
		commands: cs,
	}
}

// New creates and returns a Handler for a new client session. The Handler
// reads frames from receiver, writes responses to sender (guarded by
// senderLock), and dispatches commands to the registered handlers. receiveDelay
// and sendDelay introduce artificial latency between frames to help with
// flow control.
func (c Commander) New(
	cfg Configuration,
	receiver rw.FetchReader,
	sender io.Writer,
	senderLock *sync.Mutex,
	receiveDelay time.Duration,
	sendDelay time.Duration,
	l log.Logger,
	hooks Hooks,
	bufferPool *BufferPool,
) (Handler, error) {
	return newHandler(
		cfg,
		&c.commands,
		receiver,
		sender,
		senderLock,
		receiveDelay,
		sendDelay,
		l,
		hooks,
		bufferPool,
	), nil
}
