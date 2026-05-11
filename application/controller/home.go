// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"net/http"

	"github.com/Snuffy2/shellport/application/log"
)

// home is the controller for the root path ("/"). It embeds baseController so
// that all HTTP verbs other than GET return ErrControllerNotImplemented
// without additional boilerplate.
type home struct {
	baseController
}

// Get handles HTTP GET requests for the application root. It serves the
// embedded "index.html" page with a 200 OK status code.
func (h home) Get(w *ResponseWriter, r *http.Request, l log.Logger) error {
	return serveStaticPage("index.html", http.StatusOK, w, r, l)
}
