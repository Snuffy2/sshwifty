// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

//go:generate go fmt ./...

package controller

import (
	"bytes"
	"compress/gzip"
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Snuffy2/sshwifty/application/log"
)

// static_assets is populated by npm run generate before production builds and
// full tests. index.html and error.html are named explicitly so direct Go
// builds fail when the generated frontend assets are absent.
//
//go:embed static_assets/index.html static_assets/error.html static_assets/*
var embeddedStaticAssets embed.FS

// staticData holds an embedded static file together with its optional
// gzip-compressed variant, content type, creation timestamp, and ETag hash.
// Files that do not benefit from compression (images, fonts) will have an
// empty compressed slice.
type staticData struct {
	// data is the raw (uncompressed) content of the file.
	data []byte
	// compressed is the gzip-compressed form of data. It is empty when
	// compression was skipped for this file type.
	compressed []byte
	// created records when the static file was compiled into the binary.
	created time.Time
	// contentType is the MIME type used in the Content-Type response header.
	contentType string
}

// hasCompressed reports whether a pre-compressed variant of this file is
// available.
func (s staticData) hasCompressed() bool {
	return len(s.compressed) > 0
}

// staticFileExt returns the lowercased file extension of fileName, including
// the leading dot (e.g. ".js"). It returns an empty string if fileName
// contains no dot.
func staticFileExt(fileName string) string {
	extIdx := strings.LastIndex(fileName, ".")
	if extIdx < 0 {
		return ""
	}
	return strings.ToLower(fileName[extIdx:])
}

// staticContentType returns the HTTP content type for a static file name.
// It preserves project-specific overrides for extensions that the standard
// library may not recognize or may map incorrectly.
func staticContentType(fileName string) string {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".ico":
		return "image/x-icon"
	case ".md":
		return "text/markdown"
	case ".map":
		return "text/plain"
	case ".txt":
		return "text/plain"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".webmanifest":
		return "application/manifest+json"
	default:
		contentType := mime.TypeByExtension(filepath.Ext(fileName))
		if contentType == "" {
			return "application/binary"
		}
		return contentType
	}
}

// shouldCompressStatic reports whether a content type should get a prebuilt
// gzip variant for runtime negotiation.
func shouldCompressStatic(contentType string) bool {
	return !strings.HasPrefix(contentType, "image/") &&
		!strings.HasPrefix(contentType, "font/woff") &&
		contentType != "text/plain"
}

// gzipStaticContent returns the best-compression gzip encoding for content.
// It panics on writer creation or write/close errors to match build-time
// generator failure semantics during process startup.
func gzipStaticContent(content []byte) []byte {
	var compressed bytes.Buffer
	writer, err := gzip.NewWriterLevel(&compressed, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	if _, err = writer.Write(content); err != nil {
		panic(err)
	}
	if err = writer.Close(); err != nil {
		panic(err)
	}
	return compressed.Bytes()
}

// loadStaticPages builds the runtime static page map from the embedded asset
// directory.
func loadStaticPages() map[string]staticData {
	return loadStaticPagesFromFS(embeddedStaticAssets, "static_assets")
}

// loadStaticPagesFromFS builds the runtime static page map from root in fileSystem.
func loadStaticPagesFromFS(fileSystem fs.FS, root string) map[string]staticData {
	pages := make(map[string]staticData)
	created := time.Now()

	walkErr := fs.WalkDir(fileSystem, root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		name := strings.TrimPrefix(path, root+"/")
		content, readErr := fs.ReadFile(fileSystem, path)
		if readErr != nil {
			return readErr
		}

		contentType := staticContentType(name)
		data := staticData{
			data:        content,
			created:     created,
			contentType: contentType,
		}
		if shouldCompressStatic(contentType) {
			data.compressed = gzipStaticContent(content)
		}
		pages[name] = data
		return nil
	})
	if walkErr != nil {
		panic(walkErr)
	}
	requireStaticPage(pages, "index.html")
	requireStaticPage(pages, "error.html")

	return pages
}

// requireStaticPage panics when a generated shell page is missing from the
// embedded asset set.
func requireStaticPage(pages map[string]staticData, name string) {
	if _, found := pages[name]; !found {
		panic("generated static asset " + name + " is missing; run npm run generate before building")
	}
}

var staticPages = loadStaticPages()

// serveStaticCacheData serves an embedded static asset identified by dataName
// with long-lived caching headers, but rejects HTML/HTM files by returning
// ErrNotFound so they cannot be served directly as assets (they are served
// only by their dedicated routes). For all other file types it delegates to
// serveStaticCachePage.
func serveStaticCacheData(
	dataName string,
	fileExt string,
	w http.ResponseWriter,
	r *http.Request,
	l log.Logger,
) error {
	if fileExt == ".html" || fileExt == ".htm" {
		return ErrNotFound
	}
	return serveStaticCachePage(dataName, w, r, l)
}

// serveStaticCachePage writes the embedded static file identified by dataName
// to w with a long-lived public cache header ("max-age=5184000"). If the
// client advertises gzip support and a compressed variant is available, it
// sends the compressed form along with the appropriate Vary and
// Content-Encoding headers. It returns ErrNotFound when dataName is not
// present in the static page map.
func serveStaticCachePage(
	dataName string,
	w http.ResponseWriter,
	r *http.Request,
	l log.Logger,
) error {
	d, dFound := staticPages[dataName]
	if !dFound {
		return ErrNotFound
	}
	selectedData := d.data
	if clientSupportGZIP(r) && d.hasCompressed() {
		selectedData = d.compressed
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Add("Content-Encoding", "gzip")
	}
	w.Header().Add("Cache-Control", "public, max-age=5184000")
	w.Header().Add("Content-Type", d.contentType)
	_, wErr := w.Write(selectedData)
	return wErr
}

// serveStaticPage writes the embedded static page identified by dataName to w
// using the provided HTTP status code. It negotiates gzip encoding the same
// way as serveStaticCachePage but does not set caching headers, making it
// appropriate for HTML pages that should not be cached at the proxy layer.
// It returns ErrNotFound when dataName is absent from the static page map.
func serveStaticPage(
	dataName string,
	code int,
	w http.ResponseWriter,
	r *http.Request,
	l log.Logger,
) error {
	d, dFound := staticPages[dataName]
	if !dFound {
		return ErrNotFound
	}
	selectedData := d.data
	if clientSupportGZIP(r) && d.hasCompressed() {
		selectedData = d.compressed
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Add("Content-Encoding", "gzip")
	}
	w.Header().Add("Content-Type", d.contentType)
	w.WriteHeader(code)
	_, wErr := w.Write(selectedData)
	return wErr
}
