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
	"sync"
)

// Object pools for hot-path entities to reduce GC pressure.
// These pools recycle frequently-allocated objects during event processing.

var (
	// userPool recycles User objects during unmarshaling.
	userPool = sync.Pool{
		New: func() any { return &User{} },
	}

	// memberPool recycles Member objects during unmarshaling.
	memberPool = sync.Pool{
		New: func() any { return &Member{} },
	}

	// messagePool recycles Message objects during unmarshaling.
	messagePool = sync.Pool{
		New: func() any { return &Message{} },
	}

	// guildPool recycles Guild objects during unmarshaling.
	guildPool = sync.Pool{
		New: func() any { return &Guild{} },
	}

	// channelPool recycles various channel objects.
	// Note: Channel is an interface, but we pool the concrete types.
	textChannelPool = sync.Pool{
		New: func() any { return &TextChannel{} },
	}

	voiceChannelPool = sync.Pool{
		New: func() any { return &VoiceChannel{} },
	}
)

// AcquireUser gets a User from the pool.
// The returned User must be released back to the pool after use.
func AcquireUser() *User {
	return userPool.Get().(*User)
}

// ReleaseUser returns a User to the pool.
// The User is reset before being returned to the pool.
// Do NOT use the User after calling this function.
func ReleaseUser(u *User) {
	if u == nil {
		return
	}
	// Reset the struct to prevent data leaks
	*u = User{}
	userPool.Put(u)
}

// AcquireMember gets a Member from the pool.
// The returned Member must be released back to the pool after use.
func AcquireMember() *Member {
	return memberPool.Get().(*Member)
}

// ReleaseMember returns a Member to the pool.
// The Member is reset before being returned to the pool.
// Do NOT use the Member after calling this function.
func ReleaseMember(m *Member) {
	if m == nil {
		return
	}
	// Reset the struct to prevent data leaks
	*m = Member{}
	memberPool.Put(m)
}

// AcquireMessage gets a Message from the pool.
// The returned Message must be released back to the pool after use.
func AcquireMessage() *Message {
	return messagePool.Get().(*Message)
}

// ReleaseMessage returns a Message to the pool.
// The Message is reset before being returned to the pool.
// Do NOT use the Message after calling this function.
func ReleaseMessage(m *Message) {
	if m == nil {
		return
	}
	// Reset the struct to prevent data leaks
	// Note: This clears all fields including slices and pointers
	*m = Message{}
	messagePool.Put(m)
}

// AcquireGuild gets a Guild from the pool.
// The returned Guild must be released back to the pool after use.
func AcquireGuild() *Guild {
	return guildPool.Get().(*Guild)
}

// ReleaseGuild returns a Guild to the pool.
// The Guild is reset before being returned to the pool.
// Do NOT use the Guild after calling this function.
func ReleaseGuild(g *Guild) {
	if g == nil {
		return
	}
	// Reset the struct to prevent data leaks
	*g = Guild{}
	guildPool.Put(g)
}

// AcquireTextChannel gets a TextChannel from the pool.
func AcquireTextChannel() *TextChannel {
	return textChannelPool.Get().(*TextChannel)
}

// ReleaseTextChannel returns a TextChannel to the pool.
func ReleaseTextChannel(c *TextChannel) {
	if c == nil {
		return
	}
	*c = TextChannel{}
	textChannelPool.Put(c)
}

// AcquireVoiceChannel gets a VoiceChannel from the pool.
func AcquireVoiceChannel() *VoiceChannel {
	return voiceChannelPool.Get().(*VoiceChannel)
}

// ReleaseVoiceChannel returns a VoiceChannel to the pool.
func ReleaseVoiceChannel(c *VoiceChannel) {
	if c == nil {
		return
	}
	*c = VoiceChannel{}
	voiceChannelPool.Put(c)
}

// bytesPool provides reusable byte slices for JSON marshaling/unmarshaling.
// Using different sizes for different use cases reduces allocations.
var (
	// smallBytesPool for small JSON payloads (< 4KB)
	smallBytesPool = sync.Pool{
		New: func() any {
			b := make([]byte, 0, 4096)
			return &b
		},
	}

	// mediumBytesPool for medium JSON payloads (< 64KB)
	mediumBytesPool = sync.Pool{
		New: func() any {
			b := make([]byte, 0, 65536)
			return &b
		},
	}

	// largeBytesPool for large JSON payloads (< 1MB)
	largeBytesPool = sync.Pool{
		New: func() any {
			b := make([]byte, 0, 1048576)
			return &b
		},
	}
)

// AcquireBytes gets a byte slice from the appropriate pool based on size hint.
// The returned slice has len=0 and cap >= sizeHint.
func AcquireBytes(sizeHint int) *[]byte {
	if sizeHint <= 4096 {
		return smallBytesPool.Get().(*[]byte)
	} else if sizeHint <= 65536 {
		return mediumBytesPool.Get().(*[]byte)
	}
	return largeBytesPool.Get().(*[]byte)
}

// ReleaseBytes returns a byte slice to the appropriate pool.
// The slice is reset (len=0) but capacity is preserved.
func ReleaseBytes(b *[]byte) {
	if b == nil || *b == nil {
		return
	}

	// Reset length but keep capacity
	*b = (*b)[:0]

	cap := cap(*b)
	if cap <= 4096 {
		smallBytesPool.Put(b)
	} else if cap <= 65536 {
		mediumBytesPool.Put(b)
	} else if cap <= 1048576 {
		largeBytesPool.Put(b)
	}
	// Don't pool extremely large slices to avoid memory bloat
}
