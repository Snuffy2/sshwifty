// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"context"
	"errors"
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

func TestMoshSessionReceiveReturnsErrorAfterClose(t *testing.T) {
	session := moshGoSession{
		client: moshGoClientFunc{},
		closed: make(chan struct{}),
	}

	if err := session.Close(); err != nil {
		t.Fatalf("expected close to succeed, got %v", err)
	}

	output, err := session.Recv(250 * time.Millisecond)
	if !errors.Is(err, ErrMoshSessionClosed) {
		t.Fatalf("expected session closed error, got %v", err)
	}

	if output != nil {
		t.Fatalf("expected nil output after close, got %q", output)
	}
}

func TestMoshSessionCloseInterruptsBlockedReceive(t *testing.T) {
	recvEntered := make(chan struct{})
	releaseRecv := make(chan struct{})
	closeEntered := make(chan struct{})
	session := moshGoSession{
		client: moshGoClientFunc{
			recv: func(time.Duration) []byte {
				close(recvEntered)
				<-releaseRecv
				return nil
			},
			close: func() {
				close(closeEntered)
			},
		},
		closed: make(chan struct{}),
	}

	recvDone := make(chan error, 1)
	go func() {
		_, err := session.Recv(5 * time.Second)
		recvDone <- err
	}()

	select {
	case <-recvEntered:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected receive to enter client")
	}

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- session.Close()
	}()

	var closeErr error
	select {
	case <-closeEntered:
	case <-time.After(250 * time.Millisecond):
		close(releaseRecv)
		t.Fatal("expected close to interrupt blocked receive promptly")
	}

	select {
	case closeErr = <-closeDone:
	case <-time.After(250 * time.Millisecond):
		close(releaseRecv)
		t.Fatal("expected close to finish promptly")
	}
	if closeErr != nil {
		close(releaseRecv)
		t.Fatalf("expected close to succeed, got %v", closeErr)
	}

	close(releaseRecv)
	select {
	case err := <-recvDone:
		if !errors.Is(err, ErrMoshSessionClosed) {
			t.Fatalf("expected blocked receive to observe close, got %v", err)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected receive to finish after close")
	}
}

func TestMoshSessionCloseWaitsForInFlightSendAndBlocksNewSends(t *testing.T) {
	sendEntered := make(chan struct{})
	releaseSend := make(chan struct{})
	closeEntered := make(chan struct{})
	var sentCount int
	session := moshGoSession{
		client: moshGoClientFunc{
			send: func([]byte) {
				sentCount++
				close(sendEntered)
				<-releaseSend
			},
			close: func() {
				close(closeEntered)
			},
		},
		closed: make(chan struct{}),
	}

	sendDone := make(chan error, 1)
	go func() {
		sendDone <- session.Send([]byte("first"))
	}()

	select {
	case <-sendEntered:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected first send to enter client")
	}

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- session.Close()
	}()

	select {
	case err := <-closeDone:
		t.Fatalf("expected close to wait for in-flight send, returned %v", err)
	case <-time.After(25 * time.Millisecond):
	}

	secondSendDone := make(chan error, 1)
	go func() {
		secondSendDone <- session.Send([]byte("second"))
	}()

	select {
	case err := <-secondSendDone:
		t.Fatalf("expected second send to wait for close, returned %v", err)
	case <-time.After(25 * time.Millisecond):
	}

	close(releaseSend)

	select {
	case err := <-sendDone:
		if err != nil {
			t.Fatalf("expected first send to complete successfully, got %v", err)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected first send to finish")
	}

	select {
	case <-closeEntered:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected client close to run")
	}

	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("expected close to succeed, got %v", err)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected close to finish")
	}

	select {
	case err := <-secondSendDone:
		if !errors.Is(err, ErrMoshSessionClosed) {
			t.Fatalf("expected second send to fail closed, got %v", err)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected second send to finish")
	}

	if sentCount != 1 {
		t.Fatalf("expected only in-flight send to reach client, got %d sends", sentCount)
	}
}

func TestMoshSessionCloseInterruptsAwaitReady(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	now := time.Now()
	recvEntered := make(chan struct{})
	closeEntered := make(chan struct{})
	session := moshGoSession{
		client: moshGoClientFunc{
			recv: func(timeout time.Duration) []byte {
				if timeout != 0 {
					t.Fatalf("expected immediate readiness receive, got timeout %s", timeout)
				}
				select {
				case <-recvEntered:
				default:
					close(recvEntered)
				}
				return nil
			},
			close: func() {
				close(closeEntered)
			},
		},
		readyRecvBaseline: now,
		lastRecv: func() time.Time {
			return now
		},
		closed: make(chan struct{}),
	}

	readyDone := make(chan error, 1)
	go func() {
		_, err := session.AwaitReady(ctx, 5*time.Second)
		readyDone <- err
	}()

	select {
	case <-recvEntered:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected readiness wait to poll client")
	}

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- session.Close()
	}()

	select {
	case <-closeEntered:
	case <-time.After(250 * time.Millisecond):
		cancel()
		t.Fatal("expected close to interrupt readiness wait promptly")
	}

	select {
	case err := <-closeDone:
		if err != nil {
			cancel()
			t.Fatalf("expected close to succeed, got %v", err)
		}
	case <-time.After(250 * time.Millisecond):
		cancel()
		t.Fatal("expected close to finish promptly")
	}

	select {
	case err := <-readyDone:
		if !errors.Is(err, ErrMoshSessionClosed) {
			t.Fatalf("expected readiness wait to observe close, got %v", err)
		}
	case <-time.After(250 * time.Millisecond):
		cancel()
		t.Fatal("expected readiness wait to finish after close")
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

func TestMoshSessionAwaitReadyDrainsBufferedOutputBeforeActivityCheck(t *testing.T) {
	now := time.Now()
	session := moshGoSession{
		client: moshGoClientFunc{
			recv: func(timeout time.Duration) []byte {
				if timeout != 0 {
					t.Fatalf("expected immediate buffered output read, got timeout %s", timeout)
				}
				return []byte("early")
			},
		},
		readyRecvBaseline: now,
		lastRecv: func() time.Time {
			return now
		},
	}

	output, err := session.AwaitReady(context.Background(), 250*time.Millisecond)
	if err != nil {
		t.Fatalf("expected buffered readiness to succeed, got %v", err)
	}

	if string(output) != "early" {
		t.Fatalf("expected buffered readiness output %q, got %q", "early", output)
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
