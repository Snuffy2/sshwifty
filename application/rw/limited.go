// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package rw

import (
	"errors"
	"io"
)

// ErrReadUntilCompletedBufferFull is returned by ReadUntilCompleted when the
// provided buffer b is too small to hold all remaining bytes in the reader.
var (
	ErrReadUntilCompletedBufferFull = errors.New(
		"cannot read more, not enough data buffer")
)

// LimitedReader wraps a FetchReader and allows at most n bytes to be read.
// Once n bytes have been consumed, Completed returns true and further reads
// return io.EOF.
type LimitedReader struct {
	// r is the underlying FetchReader.
	r *FetchReader
	// n is the number of bytes remaining before the reader is exhausted.
	n int
}

// ReadUntilCompleted reads from r into b until r.Completed() returns true or
// an error occurs. It returns ErrReadUntilCompletedBufferFull if b is not
// large enough to hold all remaining bytes.
func ReadUntilCompleted(r *LimitedReader, b []byte) (int, error) {
	bCur := 0
	bLen := len(b)

	for !r.Completed() {
		if bCur >= bLen {
			return bCur, ErrReadUntilCompletedBufferFull
		}

		rLen, rErr := r.Read(b[bCur:])

		if rErr != nil {
			return bCur + rLen, rErr
		}

		bCur += rLen
	}

	return bCur, nil
}

// NewLimitedReader creates a new LimitedReader
func NewLimitedReader(r *FetchReader, n int) LimitedReader {
	return LimitedReader{
		r: r,
		n: n,
	}
}

// Buffered returns all remaining bytes by calling Fetch(Remains()). It returns
// io.EOF when the reader is completed.
func (l *LimitedReader) Buffered() ([]byte, error) {
	return l.Fetch(l.Remains())
}

// Fetch returns up to n bytes from the underlying FetchReader, capped at
// Remains(). It returns io.EOF when the reader is already completed.
func (l *LimitedReader) Fetch(n int) ([]byte, error) {
	if l.Completed() {
		return nil, io.EOF
	}

	if n > l.Remains() {
		n = l.Remains()
	}

	exported, eErr := l.r.Fetch(n)

	l.n -= len(exported)

	return exported, eErr
}

// Read implements io.Reader, reading up to min(len(b), Remains()) bytes into b.
// It returns io.EOF when the reader is completed.
func (l *LimitedReader) Read(b []byte) (int, error) {
	if l.Completed() {
		return 0, io.EOF
	}

	toRead := len(b)

	if toRead > l.Remains() {
		toRead = l.Remains()
	}

	rLen, rErr := l.r.Read(b[:toRead])

	l.n -= rLen

	return rLen, rErr
}

// Ditch discards all remaining bytes in the reader by repeatedly reading into
// buf until Completed returns true. It returns the first read error encountered.
func (l *LimitedReader) Ditch(buf []byte) error {
	for !l.Completed() {
		_, rErr := l.Read(buf)

		if rErr != nil {
			return rErr
		}
	}

	return nil
}

// Remains returns the number of bytes that can still be read before the reader
// is completed.
func (l LimitedReader) Remains() int {
	return l.n
}

// Completed returns true when all n bytes have been consumed and no further
// reads are permitted.
func (l LimitedReader) Completed() bool {
	return l.n <= 0
}
