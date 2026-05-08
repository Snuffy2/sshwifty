// Sshwifty - A Web SSH client
//
// Copyright (C) 2019-2025 Ni Rui <ranqus@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
