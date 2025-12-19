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

// shardCount is the number of shards in a ShardMap.
// 256 provides a good balance between memory overhead and lock contention reduction.
// With 256 shards, lock contention is reduced by ~99.6% compared to a single lock.
const shardCount = 256

// shard represents a single partition of a ShardMap.
// Each shard has its own mutex for independent locking.
type shard[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

// ShardMap is a concurrent map implementation using 256-way sharding.
// It reduces lock contention by distributing entries across multiple shards,
// each with its own RWMutex. This is particularly effective for high-throughput
// scenarios like Discord bots with 10,000+ guilds.
//
// Performance: ~99.6% reduction in lock contention vs single mutex.
type ShardMap[K comparable, V any] struct {
	shards [shardCount]shard[K, V]
	hasher func(K) uint8
}

// NewShardMap creates a new ShardMap with the given hash function.
// The hash function should distribute keys evenly across 0-255.
func NewShardMap[K comparable, V any](hasher func(K) uint8) *ShardMap[K, V] {
	m := &ShardMap[K, V]{hasher: hasher}
	for i := range m.shards {
		m.shards[i].data = make(map[K]V)
	}
	return m
}

// NewSnowflakeShardMap creates a ShardMap optimized for Snowflake keys.
// Uses the low 8 bits of the snowflake for sharding, which provides
// excellent distribution since Discord snowflakes are sequential with
// embedded sequence numbers.
func NewSnowflakeShardMap[V any]() *ShardMap[Snowflake, V] {
	return NewShardMap[Snowflake, V](func(k Snowflake) uint8 {
		return uint8(k & 0xFF)
	})
}

// NewSnowflakePairShardMap creates a ShardMap for SnowflakePairKey keys.
// Uses XOR of both snowflake low bits for distribution.
func NewSnowflakePairShardMap[V any]() *ShardMap[SnowflakePairKey, V] {
	return NewShardMap[SnowflakePairKey, V](func(k SnowflakePairKey) uint8 {
		return uint8((k.A ^ k.B) & 0xFF)
	})
}

// getShard returns the shard for a given key.
//
//go:nosplit
func (m *ShardMap[K, V]) getShard(key K) *shard[K, V] {
	return &m.shards[m.hasher(key)]
}

// Get retrieves a value from the map.
// Returns the value and true if found, zero value and false otherwise.
func (m *ShardMap[K, V]) Get(key K) (V, bool) {
	s := m.getShard(key)
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()
	return v, ok
}

// Set stores a value in the map.
func (m *ShardMap[K, V]) Set(key K, value V) {
	s := m.getShard(key)
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
}

// Delete removes a value from the map.
// Returns true if the key existed, false otherwise.
func (m *ShardMap[K, V]) Delete(key K) bool {
	s := m.getShard(key)
	s.mu.Lock()
	_, existed := s.data[key]
	if existed {
		delete(s.data, key)
	}
	s.mu.Unlock()
	return existed
}

// Has checks if a key exists in the map.
func (m *ShardMap[K, V]) Has(key K) bool {
	s := m.getShard(key)
	s.mu.RLock()
	_, ok := s.data[key]
	s.mu.RUnlock()
	return ok
}

// Len returns the total number of entries across all shards.
// Note: This acquires read locks on all shards sequentially,
// so the result may be slightly stale in concurrent scenarios.
func (m *ShardMap[K, V]) Len() int {
	total := 0
	for i := range m.shards {
		m.shards[i].mu.RLock()
		total += len(m.shards[i].data)
		m.shards[i].mu.RUnlock()
	}
	return total
}

// Range calls the given function for each key-value pair in the map.
// If the function returns false, iteration stops.
// The function is called with the shard lock held, so it should be fast.
// Do not call ShardMap methods from within the function to avoid deadlock.
func (m *ShardMap[K, V]) Range(fn func(K, V) bool) {
	for i := range m.shards {
		m.shards[i].mu.RLock()
		for k, v := range m.shards[i].data {
			if !fn(k, v) {
				m.shards[i].mu.RUnlock()
				return
			}
		}
		m.shards[i].mu.RUnlock()
	}
}

// GetOrSet retrieves a value or sets it if not present.
// Returns the existing value and true, or the new value and false.
func (m *ShardMap[K, V]) GetOrSet(key K, value V) (V, bool) {
	s := m.getShard(key)
	s.mu.Lock()
	if v, ok := s.data[key]; ok {
		s.mu.Unlock()
		return v, true
	}
	s.data[key] = value
	s.mu.Unlock()
	return value, false
}

// Update atomically updates a value using the provided function.
// The function receives the current value (or zero value if not present)
// and a boolean indicating if the key existed.
// Returns the new value.
func (m *ShardMap[K, V]) Update(key K, fn func(V, bool) V) V {
	s := m.getShard(key)
	s.mu.Lock()
	current, existed := s.data[key]
	newValue := fn(current, existed)
	s.data[key] = newValue
	s.mu.Unlock()
	return newValue
}

// Clear removes all entries from the map.
func (m *ShardMap[K, V]) Clear() {
	for i := range m.shards {
		m.shards[i].mu.Lock()
		m.shards[i].data = make(map[K]V)
		m.shards[i].mu.Unlock()
	}
}

// Keys returns all keys in the map.
// Note: This is a snapshot and may be stale immediately after return.
func (m *ShardMap[K, V]) Keys() []K {
	keys := make([]K, 0, m.Len())
	for i := range m.shards {
		m.shards[i].mu.RLock()
		for k := range m.shards[i].data {
			keys = append(keys, k)
		}
		m.shards[i].mu.RUnlock()
	}
	return keys
}

// Values returns all values in the map.
// Note: This is a snapshot and may be stale immediately after return.
func (m *ShardMap[K, V]) Values() []V {
	values := make([]V, 0, m.Len())
	for i := range m.shards {
		m.shards[i].mu.RLock()
		for _, v := range m.shards[i].data {
			values = append(values, v)
		}
		m.shards[i].mu.RUnlock()
	}
	return values
}

// shardedIndex is a helper for managing guild-to-entity index maps.
// It maps a primary Snowflake (e.g., guildID) to a set of secondary Snowflakes
// (e.g., member user IDs).
type shardedIndex struct {
	shards [shardCount]struct {
		mu   sync.RWMutex
		data map[Snowflake]map[Snowflake]struct{}
	}
}

func newShardedIndex() *shardedIndex {
	idx := &shardedIndex{}
	for i := range idx.shards {
		idx.shards[i].data = make(map[Snowflake]map[Snowflake]struct{})
	}
	return idx
}

func (idx *shardedIndex) getShard(key Snowflake) *struct {
	mu   sync.RWMutex
	data map[Snowflake]map[Snowflake]struct{}
} {
	return &idx.shards[uint8(key&0xFF)]
}

func (idx *shardedIndex) Add(primary, secondary Snowflake) {
	s := idx.getShard(primary)
	s.mu.Lock()
	if _, exists := s.data[primary]; !exists {
		s.data[primary] = make(map[Snowflake]struct{})
	}
	s.data[primary][secondary] = struct{}{}
	s.mu.Unlock()
}

func (idx *shardedIndex) Remove(primary, secondary Snowflake) bool {
	s := idx.getShard(primary)
	s.mu.Lock()
	defer s.mu.Unlock()

	if m, exists := s.data[primary]; exists {
		if _, has := m[secondary]; has {
			delete(m, secondary)
			if len(m) == 0 {
				delete(s.data, primary)
			}
			return true
		}
	}
	return false
}

func (idx *shardedIndex) Get(primary Snowflake) (map[Snowflake]struct{}, bool) {
	s := idx.getShard(primary)
	s.mu.RLock()
	m, ok := s.data[primary]
	if !ok {
		s.mu.RUnlock()
		return nil, false
	}
	// Return a copy to avoid holding the lock
	result := make(map[Snowflake]struct{}, len(m))
	for k := range m {
		result[k] = struct{}{}
	}
	s.mu.RUnlock()
	return result, true
}

func (idx *shardedIndex) Has(primary Snowflake) bool {
	s := idx.getShard(primary)
	s.mu.RLock()
	_, ok := s.data[primary]
	s.mu.RUnlock()
	return ok
}

func (idx *shardedIndex) Count(primary Snowflake) int {
	s := idx.getShard(primary)
	s.mu.RLock()
	m, ok := s.data[primary]
	if !ok {
		s.mu.RUnlock()
		return 0
	}
	count := len(m)
	s.mu.RUnlock()
	return count
}

func (idx *shardedIndex) Delete(primary Snowflake) (map[Snowflake]struct{}, bool) {
	s := idx.getShard(primary)
	s.mu.Lock()
	m, ok := s.data[primary]
	if ok {
		delete(s.data, primary)
	}
	s.mu.Unlock()
	return m, ok
}
