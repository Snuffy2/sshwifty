// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package application

import (
	"strings"
	"testing"
)

func TestBannerIncludesVersion(t *testing.T) {
	output := Banner()

	for _, expected := range []string{FullName, "dev", Author, URL} {
		if !strings.Contains(output, expected) {
			t.Fatalf("Banner() = %q, missing %q", output, expected)
		}
	}
}
