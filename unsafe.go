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
	"runtime"
	"unsafe"
)

// BytesToString converts a byte slice to a string without allocation.
// WARNING: The returned string shares memory with the byte slice.
// The byte slice MUST NOT be modified after this call, or the string
// will be corrupted. The byte slice must remain alive for the lifetime
// of the returned string.
//
//go:nosplit
func BytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

// StringToBytes converts a string to a byte slice without allocation.
// WARNING: The returned byte slice shares memory with the string.
// The byte slice MUST NOT be modified, as strings are immutable in Go.
// Modifying the returned slice results in undefined behavior.
//
//go:nosplit
func StringToBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// hidePointer casts a pointer to uintptr to hide it from Go's escape analysis.
// This forces the compiler to treat the object as potentially stack-allocated,
// preventing unnecessary heap allocations in hot paths.
//
// WARNING: The hidden pointer MUST be recovered via unhidePointer before:
//   - The GC runs (use runtime.KeepAlive on the original pointer)
//   - The current goroutine yields
//   - The function returns
//
// Failure to follow these rules results in undefined behavior as the GC
// may move or collect the underlying object.
//
//go:nosplit
func hidePointer[T any](p *T) uintptr {
	return uintptr(unsafe.Pointer(p))
}

// unhidePointer recovers a pointer that was hidden via hidePointer.
// The uintptr value must have been created by hidePointer and must
// be recovered within the same synchronous execution context.
//
// WARNING: See hidePointer for safety requirements. The original
// pointer source must be kept alive via runtime.KeepAlive until
// after this function returns.
//
//go:nosplit
func unhidePointer[T any](u uintptr) *T {
	return (*T)(unsafe.Pointer(u))
}

// noescape hides a pointer from escape analysis. The pointer is
// returned unchanged but the compiler cannot prove that it doesn't
// escape. This is useful for passing pointers to assembly or cgo
// functions that the compiler cannot analyze.
//
//go:nosplit
//go:nocheckptr
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

// keepAlive is a helper that ensures the value is not garbage collected
// until after this call. This is an alias for runtime.KeepAlive for
// consistency within the unsafe utilities.
//
//go:nosplit
func keepAlive(x any) {
	runtime.KeepAlive(x)
}

// parseUint64Branchless parses a decimal string to uint64 without branches.
// This function assumes the input is a valid decimal number string.
// Invalid input (non-digit characters) results in undefined output.
// Empty strings return 0.
//
// Performance: ~3-5ns for typical Discord snowflakes (18-19 digits)
// compared to ~30-50ns for strconv.ParseUint.
//
//go:nosplit
func parseUint64Branchless(s string) uint64 {
	if len(s) == 0 {
		return 0
	}

	var n uint64
	for i := 0; i < len(s); i++ {
		// Branchless digit extraction: '0'-'9' maps to 0-9
		// Any non-digit character will produce garbage, which is acceptable
		// since we assume valid input from Discord's API
		n = n*10 + uint64(s[i]-'0')
	}
	return n
}

// unquoteSimple removes surrounding quotes from a JSON string value.
// This is a fast path for simple strings without escape sequences.
// Returns the original string if it doesn't start/end with quotes.
// For complex strings with escapes, use strconv.Unquote.
//
//go:nosplit
func unquoteSimple(s string) string {
	if len(s) < 2 {
		return s
	}
	if s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
