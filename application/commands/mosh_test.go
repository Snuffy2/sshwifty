// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"bytes"
	"context"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/Snuffy2/sshwifty/application/command"
	"github.com/Snuffy2/sshwifty/application/configuration"
	"github.com/Snuffy2/sshwifty/application/rw"
)

func TestCommandsIncludesMosh(t *testing.T) {
	commands := New()
	expectedNames := map[byte]string{
		0x00: "Telnet",
		0x01: "SSH",
		0x02: "Mosh",
	}

	for id, expectedName := range expectedNames {
		name := reflect.ValueOf(commands[id]).FieldByName("name").String()
		if name != expectedName {
			t.Fatalf("expected command %d to be %q, got %q", id, expectedName, name)
		}
	}
}

func TestParseMoshConfigNormalizesHost(t *testing.T) {
	cfg, err := parseMoshConfig(configuration.Preset{Host: "example.com"})
	if err != nil {
		t.Fatalf("expected config parsing to succeed, got %v", err)
	}

	if cfg.Host != "example.com:22" {
		t.Fatalf("expected normalized host to be %q, got %q", "example.com:22", cfg.Host)
	}
}

func TestMoshBuildRemoteSessionPassesIPv4LiteralUnchanged(t *testing.T) {
	client := &moshClient{
		baseCtx: context.Background(),
		hostResolver: func(context.Context, string) ([]net.IP, error) {
			t.Fatal("expected IPv4 literal to bypass DNS resolution")
			return nil, nil
		},
	}

	var called bool
	client.sessionBuilder = func(host string, port int, key string) (moshSession, error) {
		called = true
		if host != "192.0.2.10" {
			t.Fatalf("expected session builder host to remain IPv4 literal, got %q", host)
		}
		if port != 60001 {
			t.Fatalf("expected port 60001, got %d", port)
		}
		if key != "secret" {
			t.Fatalf("expected key secret, got %q", key)
		}

		return &fakeMoshSession{}, nil
	}

	if _, err := client.buildRemoteSession("192.0.2.10:22", 60001, "secret"); err != nil {
		t.Fatalf("expected remote session build to succeed, got %v", err)
	}

	if !called {
		t.Fatal("expected session builder to be called")
	}
}

func TestMoshBuildRemoteSessionResolvesHostnameToIPv4(t *testing.T) {
	client := &moshClient{
		baseCtx: context.Background(),
		cfg: command.Configuration{
			DialTimeout: 2 * time.Second,
		},
		hostResolver: func(ctx context.Context, host string) ([]net.IP, error) {
			deadline, hasDeadline := ctx.Deadline()
			if !hasDeadline {
				t.Fatal("expected hostname resolution context to carry a deadline")
			}
			if time.Until(deadline) <= 0 {
				t.Fatal("expected hostname resolution deadline to be in the future")
			}
			if host != "example.com" {
				t.Fatalf("expected resolver host example.com, got %q", host)
			}

			return []net.IP{
				net.ParseIP("2001:db8::1"),
				net.ParseIP("198.51.100.23"),
			}, nil
		},
	}

	var resolvedHost string
	client.sessionBuilder = func(host string, port int, key string) (moshSession, error) {
		resolvedHost = host
		return &fakeMoshSession{}, nil
	}

	if _, err := client.buildRemoteSession("example.com:22", 60001, "secret"); err != nil {
		t.Fatalf("expected hostname-backed session build to succeed, got %v", err)
	}

	if resolvedHost != "198.51.100.23" {
		t.Fatalf("expected session builder to receive resolved IPv4 literal, got %q", resolvedHost)
	}
}

func TestMoshBuildRemoteSessionRejectsIPv6OnlyTargets(t *testing.T) {
	client := &moshClient{
		baseCtx: context.Background(),
		hostResolver: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("2001:db8::1")}, nil
		},
		sessionBuilder: func(string, int, string) (moshSession, error) {
			t.Fatal("expected IPv6-only target to fail before session dial")
			return nil, nil
		},
	}

	_, err := client.buildRemoteSession("example.com:22", 60001, "secret")
	if err == nil {
		t.Fatal("expected IPv6-only target to be rejected")
	}
}

func TestMoshBootupRejectsSocks5(t *testing.T) {
	bufferPool := command.NewBufferPool(4096)
	client := newMosh(
		nil,
		command.NewHooks(configuration.HookSettings{}),
		command.StreamResponder{},
		command.Configuration{Socks5Configured: true},
		&bufferPool,
	)

	state, err := client.Bootup(nil, nil)
	if state != nil {
		t.Fatalf("expected bootup state to stay nil on SOCKS5 rejection, got %v", state)
	}

	if err.Error() != ErrMoshSocks5Unsupported.Error() {
		t.Fatalf("expected SOCKS5 bootup error, got %v", err)
	}

	if err.Code() != MoshRequestErrorUnsupportedProxy {
		t.Fatalf("expected unsupported proxy code, got %d", err.Code())
	}
}

func TestMoshLocalStdInWritesBufferedBytes(t *testing.T) {
	session := &fakeMoshSession{}
	client := &moshClient{session: session}

	if err := client.local(
		nil,
		newLimitedReader([]byte("typed input")),
		command.StreamHeader{MoshClientStdIn << 5, 0},
		make([]byte, 4),
	); err != nil {
		t.Fatalf("expected stdin handler to succeed, got %v", err)
	}

	if len(session.sent) != 1 {
		t.Fatalf("expected one send call, got %d", len(session.sent))
	}

	if !bytes.Equal(session.sent[0], []byte("typed input")) {
		t.Fatalf("expected sent payload %q, got %q", "typed input", session.sent[0])
	}
}

func TestMoshLocalStdInWaitsForSessionThenSends(t *testing.T) {
	session := newBlockingFakeMoshSession()
	client := &moshClient{sessionReceive: make(chan moshSession, 1)}

	errCh := make(chan error, 1)
	go func() {
		errCh <- client.local(
			nil,
			newLimitedReader([]byte("typed input")),
			command.StreamHeader{MoshClientStdIn << 5, 0},
			make([]byte, 4),
		)
	}()

	select {
	case err := <-errCh:
		t.Fatalf("expected local stdin to block until session arrives, got %v", err)
	case <-time.After(25 * time.Millisecond):
	}

	client.sessionReceive <- session

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected stdin handler to succeed after session delivery, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected stdin handler to resume after session delivery")
	}

	if len(session.sent) != 1 {
		t.Fatalf("expected one send call, got %d", len(session.sent))
	}

	if !bytes.Equal(session.sent[0], []byte("typed input")) {
		t.Fatalf("expected sent payload %q, got %q", "typed input", session.sent[0])
	}
}

func TestMoshLocalResizeWaitsForSessionThenResizes(t *testing.T) {
	session := newBlockingFakeMoshSession()
	client := &moshClient{sessionReceive: make(chan moshSession, 1)}

	errCh := make(chan error, 1)
	go func() {
		errCh <- client.local(
			nil,
			newLimitedReader([]byte{0x00, 0x18, 0x00, 0x50}),
			command.StreamHeader{MoshClientResize << 5, 0x04},
			make([]byte, 4),
		)
	}()

	select {
	case err := <-errCh:
		t.Fatalf("expected local resize to block until session arrives, got %v", err)
	case <-time.After(25 * time.Millisecond):
	}

	client.sessionReceive <- session

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected resize handler to succeed after session delivery, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected resize handler to resume after session delivery")
	}

	if len(session.resizes) != 1 {
		t.Fatalf("expected one resize call, got %d", len(session.resizes))
	}

	if session.resizes[0] != (fakeMoshResize{cols: 80, rows: 24}) {
		t.Fatalf("expected resize to be cols=%d rows=%d, got cols=%d rows=%d",
			80, 24, session.resizes[0].cols, session.resizes[0].rows)
	}
}

func TestMoshCloseClosesBufferedSessionWithoutWaitingForReceiveTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	session := newBlockingFakeMoshSession()
	client := &moshClient{
		baseCtx:                              ctx,
		baseCtxCancel:                        cancel,
		credentialReceive:                    make(chan []byte, 1),
		fingerprintVerifyResultReceive:       make(chan bool, 1),
		sessionReceive:                       make(chan moshSession, 1),
		remoteCloseWait:                      sync.WaitGroup{},
		remoteReadTimeoutRetryLock:           sync.Mutex{},
		credentialReceiveClosed:              false,
		fingerprintVerifyResultReceiveClosed: false,
	}
	client.sessionReceive <- session
	client.cacheSession(session)

	client.remoteCloseWait.Add(1)
	go func() {
		defer client.remoteCloseWait.Done()
		<-session.closedCh
	}()

	done := make(chan error, 1)
	go func() {
		done <- client.Close()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected close to succeed, got %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected close to return promptly after closing buffered session")
	}

	if !session.closed {
		t.Fatal("expected buffered session to be closed")
	}
}

type fakeMoshResize struct {
	cols uint16
	rows uint16
}

type fakeMoshSession struct {
	mu       sync.Mutex
	sent     [][]byte
	resizes  []fakeMoshResize
	closed   bool
	closedCh chan struct{}
}

func (f *fakeMoshSession) Send(payload []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, append([]byte(nil), payload...))
	return nil
}

func (f *fakeMoshSession) Recv(time.Duration) ([]byte, error) {
	return nil, nil
}

func (f *fakeMoshSession) Resize(cols uint16, rows uint16) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.resizes = append(f.resizes, fakeMoshResize{cols: cols, rows: rows})
	return nil
}

func (f *fakeMoshSession) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil
	}
	f.closed = true
	if f.closedCh != nil {
		close(f.closedCh)
	}
	return nil
}

func newBlockingFakeMoshSession() *fakeMoshSession {
	return &fakeMoshSession{
		closedCh: make(chan struct{}),
	}
}

func newLimitedReader(payload []byte) *rw.LimitedReader {
	reader := rw.NewFetchReader(func() ([]byte, error) {
		return payload, nil
	})
	limited := rw.NewLimitedReader(&reader, len(payload))
	return &limited
}
