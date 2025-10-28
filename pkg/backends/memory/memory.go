// Package memory provides an in-memory cache backend implementation.
package memory

import (
	"github.com/ldaidone/gomemo/pkg/backends"
	"sync"
	"time"
)

// Memory is an in-memory cache backend implementation.
// It stores values in a map and automatically removes expired entries.
type Memory struct {
	entries map[string]backends.CacheEntry
	mu      sync.RWMutex
}

// New creates a new in-memory cache backend.
// It starts a cleanup goroutine that periodically removes expired entries.
func New() *Memory {
	m := &Memory{
		entries: make(map[string]backends.CacheEntry),
	}

	// Start cleanup goroutine to remove expired entries periodically
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // Cleanup every minute
		defer ticker.Stop()

		for range ticker.C {
			for key, entry := range m.entries {
				if entry.IsExpired() {
					delete(m.entries, key)
				}
			}
		}
	}()

	return m
}

// init registers the memory backend with the factory
func init() {
	backends.RegisterBackend("memory", func() backends.Backend {
		return New()
	})
}

// Get retrieves a value from the cache by key.
// Returns the value and true if found and not expired, nil and false otherwise.
func (m *Memory) Get(key string) (value any, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.entries[key]
	if !exists {
		return nil, false
	}

	if entry.IsExpired() {
		delete(m.entries, key) // Clean up expired entry
		return nil, false
	}

	return entry.Value, true
}

// Set stores a value in the cache with the given TTL (time-to-live).
// If TTL is 0 or negative, the value will not expire.
func (m *Memory) Set(key string, value any, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var entry backends.CacheEntry

	entry = backends.NewEntry(value, ttl, entry.BumpVersion())
	m.entries[key] = entry
}

// Delete removes a value from the cache.
func (m *Memory) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.entries, key)
}

// Clear removes all values from the cache.
func (m *Memory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	clear(m.entries)
}
