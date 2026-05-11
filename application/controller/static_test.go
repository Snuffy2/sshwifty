// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/Snuffy2/shellport/application/log"
)

func testStaticPages() map[string]staticData {
	return loadStaticPagesFromFS(fstest.MapFS{
		"static_assets/index.html": {
			Data: []byte("<!doctype html>"),
		},
		"static_assets/error.html": {
			Data: []byte("<!doctype html>"),
		},
		"static_assets/README.md": {
			Data: []byte("readme"),
		},
	}, "static_assets")
}

func withStaticPages(t *testing.T, pages map[string]staticData) {
	t.Helper()

	original := staticPages
	staticPages = pages
	t.Cleanup(func() {
		staticPages = original
	})
}

// TestLoadStaticPagesIncludesNestedAssets verifies embedded asset loading keeps
// relative paths for files below nested asset directories.
func TestLoadStaticPagesIncludesNestedAssets(t *testing.T) {
	pages := loadStaticPagesFromFS(fstest.MapFS{
		"static_assets/index.html": {
			Data: []byte("<!doctype html>"),
		},
		"static_assets/error.html": {
			Data: []byte("<!doctype html>"),
		},
		"static_assets/assets/app.js": {
			Data: []byte("console.log('ok');"),
		},
	}, "static_assets")

	if _, ok := pages["assets/app.js"]; !ok {
		t.Fatal("expected nested asset key assets/app.js")
	}
	if _, ok := pages["index.html"]; !ok {
		t.Fatal("expected direct asset key index.html")
	}
}

// TestLoadStaticPagesRequiresShellPages verifies incomplete generated assets
// fail during startup instead of producing a binary with a missing UI.
func TestLoadStaticPagesRequiresShellPages(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("expected panic for missing error.html")
		}
	}()

	loadStaticPagesFromFS(fstest.MapFS{
		"static_assets/index.html": {
			Data: []byte("<!doctype html>"),
		},
	}, "static_assets")
}

// TestStaticContentType verifies project-specific MIME type overrides for
// embedded assets.
func TestStaticContentType(t *testing.T) {
	tests := map[string]string{
		"README.md":        "text/markdown",
		"favicon.ico":      "image/x-icon",
		"site.webmanifest": "application/manifest+json",
		"font.woff2":       "font/woff2",
	}

	for name, expected := range tests {
		if actual := staticContentType(name); actual != expected {
			t.Fatalf("staticContentType(%q) = %q, want %q",
				name, actual, expected)
		}
	}
}

// TestServeStaticCacheDataRejectsHTML verifies application shell HTML is not
// served through the cacheable asset path.
func TestServeStaticCacheDataRejectsHTML(t *testing.T) {
	withStaticPages(t, testStaticPages())

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/shellport/assets/index.html", nil)
	recorder := httptest.NewRecorder()

	err := serveStaticCacheData("index.html", ".html", recorder, req, log.Ditch{})
	if err != ErrNotFound {
		t.Fatalf("serveStaticCacheData returned %v, want ErrNotFound", err)
	}
}

// TestServeStaticPageUsesNoCacheHeader verifies shell pages remain uncached.
func TestServeStaticPageUsesNoCacheHeader(t *testing.T) {
	withStaticPages(t, testStaticPages())

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	if err := serveStaticPage("index.html", http.StatusOK, recorder, req, log.Ditch{}); err != nil {
		t.Fatalf("serveStaticPage returned error: %v", err)
	}
	if cacheControl := recorder.Header().Get("Cache-Control"); cacheControl != "" {
		t.Fatalf("Cache-Control = %q, want empty", cacheControl)
	}
}

// TestServeStaticCachePageUsesCacheAndGzip verifies cacheable assets negotiate
// gzip responses with long-lived cache headers.
func TestServeStaticCachePageUsesCacheAndGzip(t *testing.T) {
	withStaticPages(t, testStaticPages())

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/shellport/assets/README.md", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	recorder := httptest.NewRecorder()

	if err := serveStaticCachePage("README.md", recorder, req, log.Ditch{}); err != nil {
		t.Fatalf("serveStaticCachePage returned error: %v", err)
	}
	if recorder.Header().Get("Cache-Control") != "public, max-age=5184000" {
		t.Fatalf("unexpected Cache-Control header: %q",
			recorder.Header().Get("Cache-Control"))
	}
	if recorder.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip",
			recorder.Header().Get("Content-Encoding"))
	}
}
