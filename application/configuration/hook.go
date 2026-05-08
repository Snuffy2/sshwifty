// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"fmt"
	"time"
)

// HookType is the string identifier for a lifecycle hook event. It is used as
// the map key in the Hooks configuration to associate commands with events.
type HookType string

// HOOK_BEFORE_CONNECTING is the hook type fired immediately before an outbound
// connection attempt is made, allowing operators to run pre-flight scripts.
const (
	HOOK_BEFORE_CONNECTING HookType = "before_connecting"
)

// verify returns nil if h is a recognised HookType, or a descriptive error
// listing the supported types if it is not.
func (h HookType) verify() error {
	switch h {
	case "before_connecting":
		return nil
	default:
		return fmt.Errorf(
			"unsupported Hook type: %q. Supported types are: %q",
			h,
			[]HookType{
				HOOK_BEFORE_CONNECTING,
			},
		)
	}
}

// HookCommand is a single executable command and its arguments, represented as
// a string slice where element 0 is the executable path.
type HookCommand []string

// Hooks maps each HookType to the ordered list of commands to run when that
// hook fires. Multiple commands may be registered for the same type.
type Hooks map[HookType][]HookCommand

// verify validates all HookType keys and their command lists. Unsupported keys
// fail via HookType.verify. Empty command lists are allowed and ignored; only
// individual empty HookCommand entries (len(v[i]) <= 0) produce an error.
func (h Hooks) verify() error {
	for k, v := range h {
		if err := k.verify(); err != nil {
			return err
		}
		if len(v) <= 0 {
			continue
		}
		for i := range v {
			if len(v[i]) <= 0 {
				return fmt.Errorf(
					"the command %d for Hook type %q must not be empty",
					i,
					k,
				)
			}
		}
	}
	return nil
}

// HookSettings bundles the hook command registry with the shared execution
// timeout. It is derived from a Configuration and passed into the command layer
// via Common.
type HookSettings struct {
	// Timeout is the maximum duration any single hook invocation may run.
	Timeout time.Duration
	// Hooks is the map of hook types to their ordered command lists.
	Hooks Hooks
}
