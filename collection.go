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

import "sync"

// Collection is a generic, thread-safe collection of Discord entities.
// It provides discord.js-style collection methods for filtering, finding,
// and iterating over entities.
type Collection[K comparable, V any] struct {
	items map[K]V
	mu    sync.RWMutex
}

// NewCollection creates a new empty collection.
func NewCollection[K comparable, V any]() *Collection[K, V] {
	return &Collection[K, V]{
		items: make(map[K]V),
	}
}

// Get retrieves an item by key.
// Returns the item and true if found, or zero value and false if not found.
func (c *Collection[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.items[key]
	return v, ok
}

// Set adds or updates an item in the collection.
func (c *Collection[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

// Has checks if a key exists in the collection.
func (c *Collection[K, V]) Has(key K) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.items[key]
	return ok
}

// Delete removes an item from the collection by key.
// Returns true if the item was found and deleted, false otherwise.
func (c *Collection[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.items[key]
	if ok {
		delete(c.items, key)
	}
	return ok
}

// Size returns the number of items in the collection.
func (c *Collection[K, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Values returns all values in the collection as a slice.
// The order is not guaranteed.
func (c *Collection[K, V]) Values() []V {
	c.mu.RLock()
	defer c.mu.RUnlock()
	values := make([]V, 0, len(c.items))
	for _, v := range c.items {
		values = append(values, v)
	}
	return values
}

// Keys returns all keys in the collection as a slice.
// The order is not guaranteed.
func (c *Collection[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]K, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	return keys
}

// Filter returns all items matching the predicate function.
func (c *Collection[K, V]) Filter(fn func(V) bool) []V {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]V, 0)
	for _, v := range c.items {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Find returns the first item matching the predicate function.
// Returns the item and true if found, or zero value and false if not found.
func (c *Collection[K, V]) Find(fn func(V) bool) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, v := range c.items {
		if fn(v) {
			return v, true
		}
	}
	var zero V
	return zero, false
}

// ForEach iterates over all items in the collection.
// The iteration order is not guaranteed.
func (c *Collection[K, V]) ForEach(fn func(K, V)) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for k, v := range c.items {
		fn(k, v)
	}
}

// Map transforms all values using the provided function.
// Returns a new slice containing the transformed values.
func (c *Collection[K, V]) Map(fn func(V) V) []V {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]V, 0, len(c.items))
	for _, v := range c.items {
		result = append(result, fn(v))
	}
	return result
}

// First returns an arbitrary item from the collection.
// Returns the item and true if the collection is not empty,
// or zero value and false if the collection is empty.
func (c *Collection[K, V]) First() (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, v := range c.items {
		return v, true
	}
	var zero V
	return zero, false
}

// Clone creates a shallow copy of the collection.
func (c *Collection[K, V]) Clone() *Collection[K, V] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	clone := NewCollection[K, V]()
	for k, v := range c.items {
		clone.items[k] = v
	}
	return clone
}

// Clear removes all items from the collection.
func (c *Collection[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]V)
}

// Some returns true if at least one item matches the predicate.
func (c *Collection[K, V]) Some(fn func(V) bool) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, v := range c.items {
		if fn(v) {
			return true
		}
	}
	return false
}

// Every returns true if all items match the predicate.
// Returns true for an empty collection.
func (c *Collection[K, V]) Every(fn func(V) bool) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, v := range c.items {
		if !fn(v) {
			return false
		}
	}
	return true
}

// Reduce reduces the collection to a single value using the accumulator function.
func (c *Collection[K, V]) Reduce(fn func(acc V, item V) V, initial V) V {
	c.mu.RLock()
	defer c.mu.RUnlock()
	acc := initial
	for _, v := range c.items {
		acc = fn(acc, v)
	}
	return acc
}

// FilterToCollection returns a new Collection containing only items matching the predicate.
func (c *Collection[K, V]) FilterToCollection(fn func(K, V) bool) *Collection[K, V] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := NewCollection[K, V]()
	for k, v := range c.items {
		if fn(k, v) {
			result.items[k] = v
		}
	}
	return result
}

// Merge adds all items from another collection to this collection.
// Existing keys will be overwritten.
func (c *Collection[K, V]) Merge(other *Collection[K, V]) {
	other.mu.RLock()
	defer other.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range other.items {
		c.items[k] = v
	}
}
