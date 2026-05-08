// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"html"
	"testing"
)

func TestParseServerMessage(t *testing.T) {
	for _, test := range [][]string{
		{
			"<b>This is a [测试](http://example.com) " +
				"[for link support](http://example.com)</b>.",
			"&lt;b&gt;This is a " +
				"<a href=\"http://example.com\" target=\"_blank\">测试</a> " +
				"<a href=\"http://example.com\" target=\"_blank\">for link support</a>" +
				"&lt;/b&gt;.",
		},
		{
			"[测试](http://example.com)",
			"<a href=\"http://example.com\" target=\"_blank\">测试</a>",
		},
		{
			"[测试](http://example.com).",
			"<a href=\"http://example.com\" target=\"_blank\">测试</a>.",
		},
		{
			".[测试](http://example.com)",
			".<a href=\"http://example.com\" target=\"_blank\">测试</a>",
		},
	} {
		result := parseServerMessage(html.EscapeString(test[0]))
		if result != test[1] {
			t.Errorf("Expecting %v, got %v instead", test[1], result)
			return
		}
	}
}
