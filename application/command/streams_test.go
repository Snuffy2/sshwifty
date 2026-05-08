// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package command

import (
	"testing"
)

func TestStreamInitialHeader(t *testing.T) {
	hd := streamInitialHeader{}

	hd.set(15, 128, true)

	if hd.command() != 15 {
		t.Errorf("Expecting command to be %d, got %d instead",
			15, hd.command())

		return
	}

	if hd.data() != 128 {
		t.Errorf("Expecting data to be %d, got %d instead", 128, hd.data())

		return
	}

	if hd.success() != true {
		t.Errorf("Expecting success to be %v, got %v instead",
			true, hd.success())

		return
	}

	hd.set(0, 2047, false)

	if hd.command() != 0 {
		t.Errorf("Expecting command to be %d, got %d instead",
			0, hd.command())

		return
	}

	if hd.data() != 2047 {
		t.Errorf("Expecting data to be %d, got %d instead", 2047, hd.data())

		return
	}

	if hd.success() != false {
		t.Errorf("Expecting success to be %v, got %v instead",
			false, hd.success())

		return
	}
}

func TestStreamHeader(t *testing.T) {
	s := StreamHeader{}

	s.Set(StreamHeaderMaxMarker, StreamHeaderMaxLength)

	if s.Marker() != StreamHeaderMaxMarker {
		t.Errorf("Expecting the marker to be %d, got %d instead",
			StreamHeaderMaxMarker, s.Marker())

		return
	}

	if s.Length() != StreamHeaderMaxLength {
		t.Errorf("Expecting the length to be %d, got %d instead",
			StreamHeaderMaxLength, s.Length())

		return
	}

	if s[0] != s[1] || s[0] != 0xff {
		t.Errorf("Expecting the header to be 255, 255, got %d, %d instead",
			s[0], s[1])

		return
	}
}
