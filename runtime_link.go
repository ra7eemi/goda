/************************************************************************************
 *
 * goda (Golang Optimized Discord API), A Lightweight Go library for Discord API
 *
 * SPDX-License-Identifier: BSD-3-Clause
 *
 * Copyright 2025 Marouane Souiri
 *
 * Licensed under the BSD 3-Clause License.
 * See the LICENSE file for details.
 *
 ************************************************************************************/

package goda

import (
	_ "unsafe" // Required for go:linkname
)

// nanotime returns the current time in nanoseconds from a monotonic clock.
// This is linked to the runtime's internal nanotime function for maximum precision.
//
// Unlike time.Now(), this function:
// - Has sub-microsecond precision
// - Is monotonic (never goes backwards)
// - Is faster (no syscall, no timezone conversion)
//
// This is used for precise heartbeat timing in the gateway connection.
//
// WARNING: This uses go:linkname which may break in future Go versions.
// The function is intentionally unexported to limit its use to internal code.
//
//go:linkname nanotime runtime.nanotime
func nanotime() int64

// MonotonicNow returns the current monotonic time in nanoseconds.
// This is a safe wrapper around the linked nanotime function.
func MonotonicNow() int64 {
	return nanotime()
}

// MonotonicSince returns the time elapsed since the given start time in nanoseconds.
// Both start and the return value are in nanoseconds.
func MonotonicSince(start int64) int64 {
	return nanotime() - start
}

// MonotonicSinceMs returns the time elapsed since the given start time in milliseconds.
func MonotonicSinceMs(start int64) int64 {
	return (nanotime() - start) / 1_000_000
}
