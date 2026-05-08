// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import "fmt"

// Meta contains data of a Key -> Value map which can be use to store
// dynamically structured configuration options
type Meta map[string]String

// Concretize returns an concretized Meta as a `map[string]string`
func (m Meta) Concretize() (mm map[string]string, err error) {
	mm = make(map[string]string, len(m))
	for k, v := range m {
		var result string
		if result, err = v.Parse(); err != nil {
			err = fmt.Errorf("unable to parse Meta \"%s\": %s", k, err)
			return
		}
		mm[k] = result
	}
	return
}
