// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"bytes"
	"strings"
	"testing"
)

func testParseAddress(
	t *testing.T,
	input []byte,
	buf []byte,
	expectedType AddressType,
	expectedData []byte,
	expectedPort uint16,
	expectedHostPortString string,
) {
	source := bytes.NewBuffer(input)
	addr, addrErr := ParseAddress(source.Read, buf)

	if addrErr != nil {
		t.Error("Failed to parse due to error:", addrErr)

		return
	}

	if addr.Type() != expectedType {
		t.Errorf("Expecting the Type to be %d, got %d instead",
			expectedType, addr.Type())

		return
	}

	if !bytes.Equal(addr.Data(), expectedData) {
		t.Errorf("Expecting the Data to be %d, got %d instead",
			expectedData, addr.Data())

		return
	}

	if addr.Port() != expectedPort {
		t.Errorf("Expecting the Port to be %d, got %d instead",
			expectedPort, addr.Port())

		return
	}

	if addr.String() != expectedHostPortString {
		t.Errorf("Expecting the Host Port string to be \"%s\", "+
			"got \"%s\" instead",
			expectedHostPortString, addr.String())

		return
	}

	output := make([]byte, len(input))
	mLen, mErr := addr.Marshal(output)

	if mErr != nil {
		t.Error("Failed to marshal due to error:", mErr)

		return
	}

	if !bytes.Equal(output[:mLen], input) {
		t.Errorf("Expecting marshaled result to be %d, got %d instead",
			input, output[:mLen])

		return
	}
}

func TestParseAddress(t *testing.T) {
	testParseAddress(
		t, []byte{0x04, 0x1e, 0x00}, make([]byte, 3), LoopbackAddr, nil, 1054,
		"localhost:1054")

	testParseAddress(
		t,
		[]byte{
			0x04, 0x1e, byte(IPv4Addr),
			0x7f, 0x00, 0x00, 0x01,
		},
		make([]byte, 4), IPv4Addr, []byte{0x7f, 0x00, 0x00, 0x01}, 1054,
		"127.0.0.1:1054")

	testParseAddress(
		t,
		[]byte{
			0x04, 0x1e, byte(IPv6Addr),
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x7f, 0x00, 0x00, 0x01,
		},
		make([]byte, 16), IPv6Addr, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x7f, 0x00, 0x00, 0x01}, 1054,
		"[::7f00:1]:1054")

	testParseAddress(
		t,
		[]byte{
			0x04, 0x1e, byte(HostNameAddr), 73,
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'1', '2', '3',
		},
		make([]byte, 73), HostNameAddr, []byte{
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'1', '2', '3',
		}, 1054,
		strings.Repeat("ABCDEFGHIJ", 7)+"123:1054")
}
