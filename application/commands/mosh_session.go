// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"fmt"
	"time"

	mosh "github.com/unixshells/mosh-go"
)

type moshSession interface {
	Send([]byte) error
	Recv(time.Duration) ([]byte, error)
	AwaitReady(time.Duration) ([]byte, error)
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
	client moshGoClient
}

func newMoshGoSession(host string, port int, key string) (moshSession, error) {
	client, err := mosh.Dial(host, port, key)
	if err != nil {
		return nil, err
	}

	return &moshGoSession{client: client}, nil
}

func (m *moshGoSession) Send(payload []byte) error {
	m.client.Send(payload)
	return nil
}

func (m *moshGoSession) Recv(timeout time.Duration) ([]byte, error) {
	return m.client.Recv(timeout), nil
}

func (m *moshGoSession) AwaitReady(timeout time.Duration) ([]byte, error) {
	output, err := m.Recv(timeout)
	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, fmt.Errorf("timed out waiting for initial mosh server response within %s", timeout)
	}

	return output, nil
}

func (m *moshGoSession) Resize(cols uint16, rows uint16) error {
	m.client.Resize(cols, rows)
	return nil
}

func (m *moshGoSession) Close() error {
	m.client.Close()
	return nil
}
