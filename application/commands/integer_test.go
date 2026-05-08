// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"bytes"
	"testing"
)

func TestInteger(t *testing.T) {
	ii := Integer(0x3fff)
	result := Integer(0)
	buf := make([]byte, 2)

	mLen, mErr := ii.Marshal(buf)

	if mErr != nil {
		t.Error("Failed to marshal:", mErr)

		return
	}

	mData := bytes.NewBuffer(buf[:mLen])
	mErr = result.Unmarshal(mData.Read)

	if mErr != nil {
		t.Error("Failed to unmarshal:", mErr)

		return
	}

	if result != ii {
		t.Errorf("Expecting result to be %d, got %d instead", ii, result)

		return
	}
}

func TestIntegerSingleByte1(t *testing.T) {
	ii := Integer(102)
	result := Integer(0)
	buf := make([]byte, 2)

	mLen, mErr := ii.Marshal(buf)

	if mErr != nil {
		t.Error("Failed to marshal:", mErr)

		return
	}

	if mLen != 1 {
		t.Errorf("Expecting the Integer to be marshalled into %d bytes, got "+
			"%d instead", 1, mLen)

		return
	}

	mData := bytes.NewBuffer(buf[:mLen])
	mErr = result.Unmarshal(mData.Read)

	if mErr != nil {
		t.Error("Failed to unmarshal:", mErr)

		return
	}

	if result != ii {
		t.Errorf("Expecting result to be %d, got %d instead", ii, result)

		return
	}
}

func TestIntegerSingleByte2(t *testing.T) {
	ii := Integer(127)
	result := Integer(0)
	buf := make([]byte, 2)

	mLen, mErr := ii.Marshal(buf)

	if mErr != nil {
		t.Error("Failed to marshal:", mErr)

		return
	}

	if mLen != 1 {
		t.Errorf("Expecting the Integer to be marshalled into %d bytes, got "+
			"%d instead", 1, mLen)

		return
	}

	mData := bytes.NewBuffer(buf[:mLen])
	mErr = result.Unmarshal(mData.Read)

	if mErr != nil {
		t.Error("Failed to unmarshal:", mErr)

		return
	}

	if result != ii {
		t.Errorf("Expecting result to be %d, got %d instead", ii, result)

		return
	}
}
