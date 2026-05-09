// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"testing"
	"time"
)

func TestMoshGoSessionImplementsInterface(t *testing.T) {
	var session moshSession = &moshGoSession{}
	if session == nil {
		t.Fatal("expected moshGoSession to implement moshSession")
	}
}

func TestMoshSessionReceiveTimeoutIsNonFatal(t *testing.T) {
	session := moshGoSession{
		client: moshGoClientFunc{
			recv: func(time.Duration) []byte {
				return nil
			},
		},
	}

	output, err := session.Recv(250 * time.Millisecond)
	if err != nil {
		t.Fatalf("expected timeout to be non-fatal, got %v", err)
	}

	if output != nil {
		t.Fatalf("expected nil output on timeout, got %q", output)
	}
}

type moshGoClientFunc struct {
	send   func([]byte)
	recv   func(time.Duration) []byte
	resize func(uint16, uint16)
	close  func()
}

func (m moshGoClientFunc) Send(payload []byte) {
	if m.send != nil {
		m.send(payload)
	}
}

func (m moshGoClientFunc) Recv(timeout time.Duration) []byte {
	if m.recv == nil {
		return nil
	}

	return m.recv(timeout)
}

func (m moshGoClientFunc) Resize(cols uint16, rows uint16) {
	if m.resize != nil {
		m.resize(cols, rows)
	}
}

func (m moshGoClientFunc) Close() {
	if m.close != nil {
		m.close()
	}
}
