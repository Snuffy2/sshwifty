// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package log

// Logger is the structured logging interface used throughout the application.
// Implementations must be safe to use concurrently.
//
//   - Context creates a child logger whose output is prefixed with name.
//   - TitledContext creates a child logger with a formatted name.
//   - Write satisfies io.Writer for compatibility with the standard library logger.
//   - Info, Debug, Warning, and Error emit messages at the corresponding severity.
type Logger interface {
	// Context returns a child Logger prefixed with name.
	Context(name string) Logger
	// TitledContext returns a child Logger with a formatted name prefix.
	TitledContext(name string, params ...any) Logger
	// Write satisfies io.Writer; implementations may log at a default severity.
	Write(b []byte) (int, error)
	// Info logs an informational message.
	Info(msg string, params ...any)
	// Debug logs a diagnostic message that may be suppressed in production.
	Debug(msg string, params ...any)
	// Warning logs a warning message indicating a potentially problematic condition.
	Warning(msg string, params ...any)
	// Error logs an error message indicating a failure condition.
	Error(msg string, params ...any)
}
