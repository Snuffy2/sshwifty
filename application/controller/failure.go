// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"net/http"

	"github.com/Snuffy2/shellport/application/log"
)

// serveFailure writes an HTTP error response by rendering the embedded
// "error.html" static page with the status code carried by err. It returns any
// write error produced by the underlying static-page handler.
func serveFailure(
	err Error,
	w http.ResponseWriter,
	r *http.Request,
	l log.Logger,
) error {
	return serveStaticPage("error.html", err.Code(), w, r, l)
}
