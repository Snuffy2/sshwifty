// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

// Package rw provides low-level I/O primitives used by the command layer:
// a buffered fetch-style reader (FetchReader), a size-limited reader
// (LimitedReader), and helper functions for reading fixed-size byte sequences.
package rw

import "errors"

// ErrFetchReaderNotEnoughBuffer is returned by FetchReader.Export when the
// internal buffer does not contain the requested number of bytes.
var (
	ErrFetchReaderNotEnoughBuffer = errors.New(
		"not enough buffer")
)

// FetchReaderFetcher is a callback that produces the next chunk of data from
// an underlying source (e.g. a WebSocket message). It is called by FetchReader
// when its internal buffer is exhausted.
type FetchReaderFetcher func() ([]byte, error)

// FetchReader wraps a FetchReaderFetcher and presents a byte-by-byte or
// chunk-by-chunk read interface. It buffers the most recently fetched chunk
// and tracks how much of it has been consumed.
type FetchReader struct {
	source   FetchReaderFetcher // Source data fetcher
	data     []byte             // Fetched source data
	dataUsed int                // Used source data
}

// Fetch is the function signature for requesting up to n bytes from a
// FetchReader. It is passed around as a first-class value to allow callers to
// read without holding a pointer to the reader.
type Fetch func(n int) ([]byte, error)

// FetchOneByte blocks until f returns at least one byte or an error. It retries
// if f returns a zero-length slice with no error.
func FetchOneByte(f Fetch) ([]byte, error) {
	for {
		d, dErr := f(1)

		if dErr != nil {
			return nil, dErr
		}

		if len(d) <= 0 {
			continue
		}

		return d, nil
	}
}

// NewFetchReader creates a new FetchReader
func NewFetchReader(g FetchReaderFetcher) FetchReader {
	return FetchReader{
		source:   g,
		data:     nil,
		dataUsed: 0,
	}
}

// dataRemain returns the number of unconsumed bytes in the internal buffer.
func (r FetchReader) dataRemain() int {
	return len(r.data) - r.dataUsed
}

// Remain returns the number of bytes currently available in the internal
// buffer without fetching from the source.
func (r *FetchReader) Remain() int {
	return r.dataRemain()
}

// Export directly exports from buffer, never read from source
//
// Params:
//   - n: Exact amount of bytes to fetch (0 to n, n included). If number n is
//     unreachable, an error will be returned, and no internal status will
//     be changed
//
// Returns:
//   - Fetched data
//   - Read error
func (r *FetchReader) Export(n int) ([]byte, error) {
	remain := r.dataRemain()

	if n > remain {
		return nil, ErrFetchReaderNotEnoughBuffer
	}

	newUsed := r.dataUsed + n
	data := r.data[r.dataUsed:newUsed]

	r.dataUsed = newUsed

	return data, nil
}

// Fetch fetchs data from the source
//
// Params:
//   - n: Max bytes to fetch (0 to n, n included)
//
// Returns:
//   - Fetched data
//   - Read error
func (r *FetchReader) Fetch(n int) ([]byte, error) {
	remain := r.dataRemain()

	if remain <= 0 {
		data, dataFetchErr := r.source()

		if dataFetchErr != nil {
			return nil, dataFetchErr
		}

		r.data = data
		r.dataUsed = 0

		remain = r.dataRemain()
	}

	if n > remain {
		n = remain
	}

	newUsed := r.dataUsed + n
	data := r.data[r.dataUsed:newUsed]

	r.dataUsed = newUsed

	return data, nil
}

// Read implements io.Reader by calling Fetch with len(b) and copying the
// returned bytes into b.
func (r *FetchReader) Read(b []byte) (int, error) {
	d, dErr := r.Fetch(len(b))

	if dErr != nil {
		return 0, dErr
	}

	return copy(b, d), nil
}
