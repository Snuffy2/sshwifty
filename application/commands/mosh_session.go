// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	mosh "github.com/unixshells/mosh-go"
)

type moshSession interface {
	Send([]byte) error
	Recv(time.Duration) ([]byte, error)
	AwaitReady(context.Context, time.Duration) ([]byte, error)
	Resize(cols uint16, rows uint16) error
	Close() error
}

type moshGoClient interface {
	Send([]byte)
	Recv(time.Duration) []byte
	Resize(cols uint16, rows uint16)
	Close()
}

type moshGoSession struct {
	client            moshGoClient
	readyRecvBaseline time.Time
	lastRecv          func() time.Time
	mu                sync.RWMutex
	closed            chan struct{}
	closeOnce         sync.Once
}

var ErrMoshSessionClosed = errors.New("mosh session closed")

func newMoshGoSession(host string, port int, key string) (moshSession, error) {
	client, err := mosh.Dial(host, port, key)
	if err != nil {
		return nil, err
	}

	return &moshGoSession{
		client:            client,
		readyRecvBaseline: client.Transport().LastRecv(),
		lastRecv:          client.Transport().LastRecv,
		closed:            make(chan struct{}),
	}, nil
}

func (m *moshGoSession) Send(payload []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	select {
	case <-m.closed:
		return ErrMoshSessionClosed
	default:
	}

	m.client.Send(payload)
	return nil
}

func (m *moshGoSession) Recv(timeout time.Duration) ([]byte, error) {
	select {
	case <-m.closed:
		return nil, ErrMoshSessionClosed
	default:
	}

	output := m.client.Recv(timeout)
	if len(output) > 0 {
		return output, nil
	}

	select {
	case <-m.closed:
		return nil, ErrMoshSessionClosed
	default:
		return output, nil
	}
}

func (m *moshGoSession) AwaitReady(ctx context.Context, timeout time.Duration) ([]byte, error) {
	select {
	case <-m.closed:
		return nil, ErrMoshSessionClosed
	default:
	}

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-m.closed:
			return nil, ErrMoshSessionClosed
		default:
		}

		if output := m.client.Recv(0); len(output) > 0 {
			return output, nil
		}

		if m.lastRecv().After(m.readyRecvBaseline) {
			return nil, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-m.closed:
			return nil, ErrMoshSessionClosed
		case <-deadline.C:
			return nil, fmt.Errorf("timed out waiting for mosh session activity within %s", timeout)
		case <-ticker.C:
		}
	}
}

func (m *moshGoSession) Resize(cols uint16, rows uint16) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	select {
	case <-m.closed:
		return ErrMoshSessionClosed
	default:
	}

	m.client.Resize(cols, rows)
	return nil
}

func (m *moshGoSession) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closeOnce.Do(func() {
		if m.client != nil {
			m.client.Close()
		}
		if m.closed != nil {
			close(m.closed)
		}
	})
	return nil
}
