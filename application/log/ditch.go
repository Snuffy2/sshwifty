// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

// Package log defines the Logger interface used throughout the application and
// provides two concrete implementations: Writer (full logging) and
// NonDebugWriter (debug messages suppressed), plus Ditch (all messages
// silently discarded).
package log

// Ditch is a no-op Logger implementation that discards every log message and
// write. It is useful in tests or when a logger must be provided but output is
// not desired.
type Ditch struct{}

// NewDitch creates and returns a Ditch logger.
func NewDitch() Ditch {
	return Ditch{}
}

// Context returns the same Ditch logger; no sub-context is created.
func (w Ditch) Context(name string) Logger {
	return w
}

// TitledContext returns the same Ditch logger; no sub-context is created.
func (w Ditch) TitledContext(name string, params ...any) Logger {
	return w
}

// Write discards b and reports success to satisfy io.Writer.
func (w Ditch) Write(b []byte) (int, error) {
	return len(b), nil
}

// Info write an info message
func (w Ditch) Info(msg string, params ...any) {}

// Debug write an debug message
func (w Ditch) Debug(msg string, params ...any) {}

// Warning write an warning message
func (w Ditch) Warning(msg string, params ...any) {}

// Error write an error message
func (w Ditch) Error(msg string, params ...any) {}
