// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package controller

import "fmt"

// Error represents an HTTP-level error that carries both a numeric status code
// and a human-readable message. It implements the error interface so it can be
// returned from controller methods and inspected by the dispatcher to choose
// the appropriate HTTP response code.
type Error struct {
	// code is the HTTP status code associated with this error (e.g. 404, 500).
	code int
	// message is the human-readable description of the error condition.
	message string
}

// NewError creates a new Error with the given HTTP status code and message.
func NewError(code int, message string) Error {
	return Error{
		code:    code,
		message: message,
	}
}

// Code returns the HTTP status code associated with this error.
func (f Error) Code() int {
	return f.code
}

// Error returns a formatted string containing the HTTP status code and the
// error message, satisfying the error interface.
func (f Error) Error() string {
	return fmt.Sprintf("HTTP Error (%d): %s", f.code, f.message)
}
