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
	"bytes"
	"compress/zlib"
	"io"
	"sync"
)

// zlibSuffix is the zlib flush suffix that Discord sends at the end of compressed messages.
// This indicates the end of a complete zlib-compressed payload.
var zlibSuffix = []byte{0x00, 0x00, 0xff, 0xff}

// zlibReaderWrapper wraps a zlib.Reader with a reusable buffer.
// This allows the reader to be reused across multiple decompressions,
// avoiding the allocation overhead of creating new readers.
type zlibReaderWrapper struct {
	reader io.ReadCloser
	buf    bytes.Buffer
}

// zlibReaderPool provides reusable zlib readers for decompressing
// gateway messages. Using a pool prevents memory spikes during
// high-traffic events by recycling decompressor instances.
var zlibReaderPool = sync.Pool{
	New: func() any {
		return &zlibReaderWrapper{
			buf: bytes.Buffer{},
		}
	},
}

// AcquireZlibReader gets a zlib reader wrapper from the pool.
func AcquireZlibReader() *zlibReaderWrapper {
	return zlibReaderPool.Get().(*zlibReaderWrapper)
}

// ReleaseZlibReader returns a zlib reader wrapper to the pool.
func ReleaseZlibReader(w *zlibReaderWrapper) {
	if w == nil {
		return
	}
	// Close the reader if open
	if w.reader != nil {
		w.reader.Close()
		w.reader = nil
	}
	// Reset the buffer
	w.buf.Reset()
	zlibReaderPool.Put(w)
}

// Decompress decompresses zlib-compressed data from the gateway.
// Returns the decompressed data as a byte slice.
//
// The wrapper's internal buffer is used to accumulate compressed data
// until a complete message (ending with zlibSuffix) is received.
func (w *zlibReaderWrapper) Decompress(data []byte) ([]byte, error) {
	// Write incoming data to buffer
	w.buf.Write(data)

	// Check if we have a complete message (ends with zlib suffix)
	if !bytes.HasSuffix(w.buf.Bytes(), zlibSuffix) {
		// Incomplete message, wait for more data
		return nil, nil
	}

	// Create or reset the zlib reader
	if w.reader == nil {
		reader, err := zlib.NewReader(&w.buf)
		if err != nil {
			return nil, err
		}
		w.reader = reader
	} else {
		// Reset for new decompression
		if resetter, ok := w.reader.(zlib.Resetter); ok {
			if err := resetter.Reset(&w.buf, nil); err != nil {
				return nil, err
			}
		}
	}

	// Read decompressed data
	decompressed, err := io.ReadAll(w.reader)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// Clear buffer for next message
	w.buf.Reset()

	return decompressed, nil
}

// DecompressOneShot decompresses a single zlib-compressed message.
// This is a convenience function for one-off decompression.
// For streaming decompression (Discord gateway), use the pooled wrapper.
func DecompressOneShot(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// IsZlibCompressed checks if data appears to be zlib-compressed.
// Zlib data starts with a specific header based on compression level.
func IsZlibCompressed(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	// Check for zlib header (CMF + FLG)
	// Common values: 0x78 0x01, 0x78 0x9C, 0x78 0xDA
	return data[0] == 0x78 && (data[1] == 0x01 || data[1] == 0x9C || data[1] == 0xDA)
}

// HasZlibSuffix checks if data ends with the Discord zlib flush suffix.
func HasZlibSuffix(data []byte) bool {
	return bytes.HasSuffix(data, zlibSuffix)
}
