// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"errors"

	"github.com/Snuffy2/sshwifty/application/rw"
)

// Errors
var (
	ErrStringParseBufferTooSmall = errors.New(
		"not enough buffer space to parse given string")

	ErrStringMarshalBufferTooSmall = errors.New(
		"not enough buffer space to marshal given string")
)

// String is a length-prefixed byte string used in the command wire protocol.
// The length field is stored as a variable-length Integer; data is the raw
// byte payload.
type String struct {
	// len is the encoded length of the data payload.
	len Integer
	// data holds the string bytes, aliased into the caller's buffer.
	data []byte
}

// ParseString decodes a length-prefixed String from reader into the provided
// scratch buffer b. It returns ErrStringParseBufferTooSmall if b cannot hold
// the declared payload length.
func ParseString(reader rw.ReaderFunc, b []byte) (String, error) {
	lenData := Integer(0)

	mErr := lenData.Unmarshal(reader)

	if mErr != nil {
		return String{}, mErr
	}

	bLen := len(b)

	if bLen < lenData.Int() {
		return String{}, ErrStringParseBufferTooSmall
	}

	_, rErr := rw.ReadFull(reader, b[:lenData])

	if rErr != nil {
		return String{}, rErr
	}

	return String{
		len:  lenData,
		data: b[:lenData],
	}, nil
}

// NewString constructs a String from raw bytes d. It panics if len(d) exceeds
// MaxInteger, which is the maximum encodable length.
func NewString(d []byte) String {
	dLen := len(d)

	if dLen > MaxInteger {
		panic("Data was too long for a String")
	}

	return String{
		len:  Integer(dLen),
		data: d,
	}
}

// Data returns the data of the string
func (s String) Data() []byte {
	return s.data
}

// Marshal encodes the String into b, writing the variable-length length prefix
// followed by the data bytes. It returns ErrStringMarshalBufferTooSmall if b
// is not large enough.
func (s String) Marshal(b []byte) (int, error) {
	bLen := len(b)

	if bLen < s.len.ByteSize()+len(s.data) {
		return 0, ErrStringMarshalBufferTooSmall
	}

	mLen, mErr := s.len.Marshal(b)

	if mErr != nil {
		return 0, mErr
	}

	return copy(b[mLen:], s.data) + mLen, nil
}
