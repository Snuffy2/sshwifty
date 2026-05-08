// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package rw

// ReaderFunc is a function type that matches the io.Reader.Read signature. It
// is used throughout the command layer as a first-class reader argument,
// allowing callers to pass method values without wrapping them in an interface.
type ReaderFunc func(b []byte) (int, error)

// ReadFull calls r repeatedly until b is fully populated or an error occurs.
// It mirrors io.ReadFull but works with ReaderFunc rather than io.Reader.
func ReadFull(r ReaderFunc, b []byte) (int, error) {
	bLen := len(b)
	readed := 0

	for {
		rLen, rErr := r(b[readed:])

		readed += rLen

		if rErr != nil {
			return readed, rErr
		}

		if readed >= bLen {
			return readed, nil
		}
	}
}
