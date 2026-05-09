// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package main

import "testing"

func TestShouldPrintVersionAcceptsVersionFlags(t *testing.T) {
	for _, args := range [][]string{
		{"sshwifty", "-V"},
		{"sshwifty", "--version"},
	} {
		if !shouldPrintVersion(args) {
			t.Fatalf("shouldPrintVersion(%v) = false, want true", args)
		}
	}
}

func TestShouldPrintVersionRejectsOtherArgs(t *testing.T) {
	for _, args := range [][]string{
		{"sshwifty"},
		{"sshwifty", "--help"},
		{"sshwifty", "--version", "--help"},
	} {
		if shouldPrintVersion(args) {
			t.Fatalf("shouldPrintVersion(%v) = true, want false", args)
		}
	}
}
