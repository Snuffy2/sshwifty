// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"context"
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

func TestMoshSessionAwaitReadyReturnsInitialOutput(t *testing.T) {
	activityAt := time.Now().Add(25 * time.Millisecond)
	session := moshGoSession{
		client: moshGoClientFunc{
			recv: func(timeout time.Duration) []byte {
				if timeout != 0 {
					t.Fatalf("expected immediate initial output read, got timeout %s", timeout)
				}
				return []byte("ready")
			},
		},
		readyRecvBaseline: activityAt.Add(-10 * time.Millisecond),
		lastRecv: func() time.Time {
			if time.Now().After(activityAt) {
				return activityAt
			}
			return activityAt.Add(-50 * time.Millisecond)
		},
	}

	output, err := session.AwaitReady(context.Background(), 250*time.Millisecond)
	if err != nil {
		t.Fatalf("expected readiness to succeed, got %v", err)
	}

	if string(output) != "ready" {
		t.Fatalf("expected readiness output %q, got %q", "ready", output)
	}
}

func TestMoshSessionAwaitReadySucceedsWithoutInitialOutput(t *testing.T) {
	now := time.Now()
	session := moshGoSession{
		client: moshGoClientFunc{
			recv: func(timeout time.Duration) []byte {
				if timeout != 0 {
					t.Fatalf("expected immediate initial output read, got timeout %s", timeout)
				}
				return nil
			},
		},
		readyRecvBaseline: now,
		lastRecv: func() time.Time {
			return now.Add(10 * time.Millisecond)
		},
	}

	output, err := session.AwaitReady(context.Background(), 250*time.Millisecond)
	if err != nil {
		t.Fatalf("expected quiet readiness to succeed, got %v", err)
	}

	if output != nil {
		t.Fatalf("expected nil output for quiet ready session, got %q", output)
	}
}

func TestMoshSessionAwaitReadyTimesOutWithoutActivity(t *testing.T) {
	now := time.Now()
	session := moshGoSession{
		client:            moshGoClientFunc{},
		readyRecvBaseline: now,
		lastRecv: func() time.Time {
			return now
		},
	}

	output, err := session.AwaitReady(context.Background(), 250*time.Millisecond)
	if err == nil {
		t.Fatal("expected readiness timeout to fail")
	}

	if output != nil {
		t.Fatalf("expected nil output on readiness timeout, got %q", output)
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
