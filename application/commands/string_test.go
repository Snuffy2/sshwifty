// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"bytes"
	"testing"
)

func testString(t *testing.T, str []byte) {
	ss := NewString(str)
	mm := make([]byte, len(str)+2)

	mLen, mErr := ss.Marshal(mm)

	if mErr != nil {
		t.Error("Failed to marshal:", mErr)

		return
	}

	buf := make([]byte, mLen)
	source := bytes.NewBuffer(mm[:mLen])
	result, rErr := ParseString(source.Read, buf)

	if rErr != nil {
		t.Error("Failed to parse:", rErr)

		return
	}

	if !bytes.Equal(result.Data(), ss.Data()) {
		t.Errorf("Expecting the data to be %d, got %d instead",
			ss.Data(), result.Data())

		return
	}
}

func TestString(t *testing.T) {
	testString(t, []byte{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
	})

	testString(t, []byte{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'i',
	})
}
