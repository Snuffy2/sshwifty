// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import "testing"

func TestParseMoshConnectLine(t *testing.T) {
	t.Run("parses valid mosh connect output", func(t *testing.T) {
		info, err := parseMoshConnectLine(
			"warning line\nMOSH CONNECT 60001 abc123\nMOSH CONNECT 60002 ignored",
		)
		if err != nil {
			t.Fatalf("expected parse to succeed, got %v", err)
		}

		if info.Port != 60001 {
			t.Fatalf("expected port 60001, got %d", info.Port)
		}

		if info.Key != "abc123" {
			t.Fatalf("expected key abc123, got %q", info.Key)
		}
	})

	t.Run("rejects malformed mosh connect output", func(t *testing.T) {
		testCases := []string{
			"hello world",
			"MOSH CONNECT 60001",
			"MOSH CONNECT 0 abc123",
			"MOSH CONNECT nope abc123",
			"MOSH CONNECT 65536 abc123",
		}

		for _, output := range testCases {
			if _, err := parseMoshConnectLine(output); err == nil {
				t.Fatalf("expected parse failure for %q", output)
			}
		}
	})
}

func TestRenderMoshServerCommand(t *testing.T) {
	t.Run("defaults to mosh-server", func(t *testing.T) {
		got := renderMoshServerCommand(nil)
		want := "'mosh-server' 'new' '-s' '-c' '256' '-l' 'LANG=en_US.UTF-8'"
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("uses configured mosh server", func(t *testing.T) {
		got := renderMoshServerCommand(map[string]string{
			"Mosh Server": "/usr/local/bin/mosh-server",
		})
		want := "'/usr/local/bin/mosh-server' 'new' '-s' '-c' '256' '-l' 'LANG=en_US.UTF-8'"
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}
