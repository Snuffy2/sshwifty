// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package rw

import (
	"bytes"
	"io"
	"testing"
)

func testDummyFetchGen(data []byte) FetchReaderFetcher {
	current := 0

	return func() ([]byte, error) {
		if current >= len(data) {
			return nil, io.EOF
		}

		oldCurrent := current
		current = oldCurrent + 1

		return data[oldCurrent:current], nil
	}
}

func TestFetchReader(t *testing.T) {
	r := NewFetchReader(testDummyFetchGen([]byte("Hello World")))
	b := make([]byte, 11)

	_, rErr := io.ReadFull(&r, b)

	if rErr != nil {
		t.Error("Failed to read due to error:", rErr)

		return
	}

	if !bytes.Equal(b, []byte("Hello World")) {
		t.Errorf("Expecting data to be %s, got %s instead",
			[]byte("Hello World"), b)

		return
	}
}
