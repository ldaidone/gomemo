package memo

import (
	"testing"
	"time"

	"github.com/ldaidone/gomemo/pkg/backends"
	"github.com/ldaidone/gomemo/pkg/backends/memory"
)

// TestBackendRegistration tests the backend registration system
func TestBackendRegistration(t *testing.T) {
	// Clear any existing registrations by creating a new registry
	// Note: In a real scenario, we'd need to reset the global registry
	// For now, just verify that memory backend exists
	if !backends.BackendExists("memory") {
		t.Fatal("Expected memory backend to be registered")
	}

	// List available backends
	backendList := backends.ListBackends()
	foundMemory := false
	for _, name := range backendList {
		if name == "memory" {
			foundMemory = true
			break
		}
	}
	if !foundMemory {
		t.Fatal("Expected memory backend in the list")
	}
}

// TestBackendFactory tests creating backends via the factory
func TestBackendFactory(t *testing.T) {
	// Test creating memory backend
	memBackend, err := backends.NewBackend("memory")
	if err != nil {
		t.Fatalf("Expected no error creating memory backend, got: %v", err)
	}
	if memBackend == nil {
		t.Fatal("Expected memory backend to be created")
	}

	// Test creating unknown backend
	_, err = backends.NewBackend("unknown")
	if err == nil {
		t.Fatal("Expected error creating unknown backend")
	}
	if err.Error() != "unknown backend type: unknown" {
		t.Fatalf("Expected 'unknown backend type' error, got: %v", err)
	}
}

// TestMemoryBackendBasic tests basic memory backend functionality
func TestMemoryBackendBasic(t *testing.T) {
	backend := memory.New()

	// Test Set and Get
	backend.Set("test-key", "test-value", 5*time.Second)

	value, exists := backend.Get("test-key")
	if !exists {
		t.Fatal("Expected key to exist")
	}
	if value != "test-value" {
		t.Fatalf("Expected 'test-value', got: %v", value)
	}

	// Test Get non-existent key
	_, exists = backend.Get("non-existent")
	if exists {
		t.Fatal("Expected non-existent key to not exist")
	}

	// Test Delete
	backend.Delete("test-key")
	_, exists = backend.Get("test-key")
	if exists {
		t.Fatal("Expected key to be deleted")
	}
}

// TestMemoryBackendTTL tests TTL functionality in memory backend
func TestMemoryBackendTTL(t *testing.T) {
	backend := memory.New()

	// Test with TTL
	backend.Set("ttl-key", "ttl-value", 10*time.Millisecond)
	value, exists := backend.Get("ttl-key")
	if !exists {
		t.Fatal("Expected key to exist before TTL expiration")
	}
	if value != "ttl-value" {
		t.Fatalf("Expected 'ttl-value', got: %v", value)
	}

	// Wait for TTL to expire
	time.Sleep(15 * time.Millisecond)
	_, exists = backend.Get("ttl-key")
	if exists {
		t.Fatal("Expected key to be expired and removed")
	}

	// Test with no TTL (0 duration)
	backend.Set("no-ttl-key", "no-ttl-value", 0)
	value, exists = backend.Get("no-ttl-key")
	if !exists {
		t.Fatal("Expected key with no TTL to exist")
	}
	if value != "no-ttl-value" {
		t.Fatalf("Expected 'no-ttl-value', got: %v", value)
	}

	// Wait and check again - should still exist
	time.Sleep(20 * time.Millisecond)
	value, exists = backend.Get("no-ttl-key")
	if !exists {
		t.Fatal("Expected key with no TTL to still exist")
	}
	if value != "no-ttl-value" {
		t.Fatalf("Expected 'no-ttl-value', got: %v", value)
	}
}

// TestMemoryBackendClear tests Clear functionality
func TestMemoryBackendClear(t *testing.T) {
	backend := memory.New()

	// Add some values
	backend.Set("key1", "value1", 5*time.Second)
	backend.Set("key2", "value2", 5*time.Second)
	backend.Set("key3", "value3", 5*time.Second)

	// Verify they exist
	_, exists1 := backend.Get("key1")
	_, exists2 := backend.Get("key2")
	_, exists3 := backend.Get("key3")
	if !exists1 || !exists2 || !exists3 {
		t.Fatal("Expected all keys to exist before clear")
	}

	// Clear all
	backend.Clear()

	// Verify they don't exist
	_, exists1 = backend.Get("key1")
	_, exists2 = backend.Get("key2")
	_, exists3 = backend.Get("key3")
	if exists1 || exists2 || exists3 {
		t.Fatal("Expected all keys to be cleared")
	}
}

// TestMemoryBackendConcurrent tests concurrent access to memory backend
func TestMemoryBackendConcurrent(t *testing.T) {
	backend := memory.New()

	// Set a value
	backend.Set("concurrent-test", "initial", 5*time.Second)

	// Read and write concurrently
	done := make(chan bool, 2)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			backend.Set("concurrent-test", i, 5*time.Second)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			backend.Get("concurrent-test")
		}
		done <- true
	}()

	// Wait for both goroutines to finish
	<-done
	<-done
}
