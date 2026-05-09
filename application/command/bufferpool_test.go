// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package command

import "testing"

func TestBufferPoolReturnsFixedSizeBuffers(t *testing.T) {
	pool := NewBufferPool(4096)

	buffer := pool.Get()
	defer pool.Put(buffer)

	if got := len(*buffer); got != 4096 {
		t.Fatalf("buffer length = %d, want 4096", got)
	}
}

func TestBufferPoolZeroesReturnedBuffers(t *testing.T) {
	pool := NewBufferPool(4)

	buffer := pool.Get()
	copy(*buffer, []byte{1, 2, 3, 4})
	pool.Put(buffer)

	buffer = pool.Get()
	defer pool.Put(buffer)

	for idx, value := range *buffer {
		if value != 0 {
			t.Fatalf("buffer[%d] = %d, want 0", idx, value)
		}
	}
}
