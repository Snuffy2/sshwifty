// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import "math"

// SignedIntegers is a type constraint that matches all built-in signed integer
// kinds. It is used by the generic numeric helper functions in this file.
type SignedIntegers interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// UnsignedIntegers is a type constraint that matches all built-in unsigned
// integer kinds.
type UnsignedIntegers interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Integers is a type constraint that matches both signed and unsigned integer
// kinds, enabling generic min/max/clamp helpers to work across all integral
// types.
type Integers interface {
	SignedIntegers | UnsignedIntegers
}

// atMost returns input clamped to upperLimit, i.e. the minimum of the two values.
func atMost[I Integers](input I, upperLimit I) I {
	return min(input, upperLimit)
}

// atLeast returns input clamped to lowerLimit, i.e. the maximum of the two values.
func atLeast[I Integers](input I, lowerLimit I) I {
	return max(input, lowerLimit)
}

// clampRange returns input clamped to the inclusive range [lowerLimit, upperLimit].
func clampRange[I Integers](input I, upperLimit I, lowerLimit I) I {
	return atMost(atLeast(input, lowerLimit), upperLimit)
}

// setZeroUintToDefault returns defaultVal when input is zero or negative,
// otherwise it returns input unchanged. It is used to apply configuration
// defaults for unspecified numeric fields.
func setZeroUintToDefault[I Integers](input I, defaultVal I) I {
	if input <= 0 {
		return defaultVal
	}
	return input
}

// castUintToIntMax safely converts the unsigned integer input to signed type S.
// If the conversion would overflow or change sign, it returns max instead of
// producing an incorrect value.
func castUintToIntMax[U UnsignedIntegers, S SignedIntegers](input U, max S) S {
	c := S(input)
	if U(c) == input && (input >= 0) == (c >= 0) {
		return c
	}
	return max
}

// castUintToInt safely converts input to int, capping at math.MaxInt if the
// value would overflow.
func castUintToInt[U UnsignedIntegers](input U) int {
	return castUintToIntMax(input, math.MaxInt)
}
