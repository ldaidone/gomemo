// Package backends provides interfaces and implementations for cache storage backends.
package backends

import (
	"fmt"
	"sync"
	"time"
)

// Backend defines a pluggable cache storage interface.
// Different implementations can provide different storage characteristics
// such as in-memory storage, Redis, or other persistent storage systems.
type Backend interface {
	// Get retrieves a value from the cache by key.
	// Returns the value and true if found, nil and false otherwise.
	Get(key string) (value any, ok bool)

	// Set stores a value in the cache with an optional TTL (time-to-live).
	// If TTL is 0 or negative, the value will not expire.
	Set(key string, value any, ttl time.Duration)

	// Delete removes a value from the cache.
	Delete(key string)

	// Clear removes all values from the cache.
	Clear()
}

// BackendFactory is a function that creates a new backend instance.
// It is used by the registration system to dynamically create backends.
type BackendFactory func() Backend

// registry holds the available backend factories, mapped by name.
var (
	registry = make(map[string]BackendFactory)
	mutex    = sync.RWMutex{}
)

// RegisterBackend registers a new backend factory with the given name.
// This function should typically be called from an init() function in
// backend implementation packages to make them available for use.
//
// Panics if the factory is nil or if a backend with the same name is already registered.
func RegisterBackend(name string, factory BackendFactory) {
	mutex.Lock()
	defer mutex.Unlock()

	if factory == nil {
		panic("backend factory cannot be nil")
	}

	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("backend factory already registered: %s", name))
	}

	registry[name] = factory
}

// NewBackend creates a new backend instance by the given type name.
// Returns an error if no backend with the given name is registered.
func NewBackend(backendType string) (Backend, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	factory, exists := registry[backendType]
	if !exists {
		return nil, fmt.Errorf("unknown backend type: %s", backendType)
	}

	return factory(), nil
}

// ListBackends returns a list of all registered backend type names.
// This is useful for discovering available backends at runtime.
func ListBackends() []string {
	mutex.RLock()
	defer mutex.RUnlock()

	backends := make([]string, 0, len(registry))
	for name := range registry {
		backends = append(backends, name)
	}

	return backends
}

// BackendExists checks if a backend type with the given name is registered.
// Returns true if the backend type exists, false otherwise.
func BackendExists(name string) bool {
	mutex.RLock()
	defer mutex.RUnlock()

	_, exists := registry[name]
	return exists
}
