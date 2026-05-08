// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"net/http"
	"regexp"
	"strings"
)

// clientSupportGZIP reports whether the client advertises gzip support in its
// Accept-Encoding header. It uses a simple substring check rather than full
// header parsing, which is sufficient for the expected browser traffic.
func clientSupportGZIP(r *http.Request) bool {
	// Should be good enough
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}

// serverMessageFormatLink is a compiled regular expression that matches
// Markdown-style inline links of the form [title](url) within a server
// message string.
var (
	serverMessageFormatLink = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
)

// parseServerMessage transforms [title](url) patterns into HTML anchor tags.
// It does not perform HTML escaping; input must already be trusted/sanitized,
// or callers must escape/encode title and URL values before embedding output.
// All other text is returned verbatim.
func parseServerMessage(input string) (result string) {
	// Yep, this is a new low, throwing regexp at a flat text format now...will
	// rewrite the entire thing in a new version with a proper parser, maybe
	// Con: Barely work when we only need to support exactly one text format
	// Pro: Expecting a debugging battle, wrote the thing in one go instead
	found := serverMessageFormatLink.FindAllStringSubmatchIndex(input, -1)
	if len(found) <= 0 {
		return input
	}
	currentStart := 0
	for _, f := range found {
		if len(f) != 6 { // Expecting 6 parameters from the given expression
			return input
		}
		segStart, segEnd, titleStart, titleEnd, linkStart, linkEnd :=
			f[0], f[1], f[2], f[3], f[4], f[5]
		result += input[currentStart:segStart]
		result += "<a href=\"" +
			input[linkStart:linkEnd] +
			"\" target=\"_blank\">" +
			input[titleStart:titleEnd] +
			"</a>"
		currentStart = segEnd
	}
	result += input[currentStart:]
	return
}
