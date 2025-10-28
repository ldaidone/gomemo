package memo

import (
	"testing"
	"time"

	"github.com/ldaidone/gomemo/pkg/backends"
)

// TestCacheEntryCreation tests creating cache entries
func TestCacheEntryCreation(t *testing.T) {
	// Test with TTL
	entry1 := backends.NewEntry("test-value", 5*time.Second, 1)
	if entry1.Value != "test-value" {
		t.Fatalf("Expected 'test-value', got: %v", entry1.Value)
	}

	// Test with no TTL (0 duration)
	entry2 := backends.NewEntry("no-expiry-value", 0, 2)
	if entry2.Value != "no-expiry-value" {
		t.Fatalf("Expected 'no-expiry-value', got: %v", entry2.Value)
	}
	if entry2.IsExpired() {
		t.Fatal("Entry with zero TTL should not be expired initially")
	}

	// Test with negative TTL
	entry3 := backends.NewEntry("negative-ttl-value", -1*time.Second, 3)
	if entry3.Value != "negative-ttl-value" {
		t.Fatalf("Expected 'negative-ttl-value', got: %v", entry3.Value)
	}
}

// TestCacheEntryExpiration tests TTL expiration functionality
func TestCacheEntryExpiration(t *testing.T) {
	// Create entry with 10ms TTL
	entry := backends.NewEntry("expiring-value", 10*time.Millisecond, 1)

	// Should not be expired initially
	if entry.IsExpired() {
		t.Fatal("Entry should not be expired immediately after creation")
	}

	// Wait for TTL to expire
	time.Sleep(15 * time.Millisecond)

	// Should be expired now
	if !entry.IsExpired() {
		t.Fatal("Entry should be expired after TTL")
	}

	// Test TTL remaining functionality
	entry2 := backends.NewEntry("remaining-test", 100*time.Millisecond, 2)
	remaining := entry2.TTLRemaining()
	if remaining <= 0 {
		t.Fatal("TTL remaining should be positive initially")
	}
	if remaining > 100*time.Millisecond {
		t.Fatalf("TTL remaining should not exceed 100ms, got: %v", remaining)
	}

	// Wait and check again
	time.Sleep(20 * time.Millisecond)
	remaining2 := entry2.TTLRemaining()
	if remaining2 > remaining {
		t.Fatalf("TTL remaining should be decreasing, was: %v, now: %v", remaining, remaining2)
	}

	// Create expired entry and check TTL remaining
	time.Sleep(20 * time.Millisecond) // Wait more to ensure previous entry is expired
	if entry.TTLRemaining() != 0 {
		t.Fatalf("Expired entry should have 0 TTL remaining, got: %v", entry.TTLRemaining())
	}
}

// TestCacheEntryVersion tests version management
func TestCacheEntryVersion(t *testing.T) {
	// Create entry with initial version
	entry := backends.NewEntry("versioned-value", 10*time.Second, 5)
	if entry.Version() != 5 {
		t.Fatalf("Expected version 5, got: %d", entry.Version())
	}

	// Bump version
	newVersion := entry.BumpVersion()
	if newVersion != 6 {
		t.Fatalf("Expected new version 6, got: %d", newVersion)
	}
	if entry.Version() != 6 {
		t.Fatalf("Expected version 6 after bump, got: %d", entry.Version())
	}

	// Bump again
	newVersion2 := entry.BumpVersion()
	if newVersion2 != 7 {
		t.Fatalf("Expected new version 7, got: %d", newVersion2)
	}
	if entry.Version() != 7 {
		t.Fatalf("Expected version 7 after second bump, got: %d", entry.Version())
	}
}

// TestCacheEntrySetExpiry tests setting expiry after creation
func TestCacheEntrySetExpiry(t *testing.T) {
	entry := backends.NewEntry("expiry-test", 1*time.Millisecond, 1)

	// Wait for original TTL to expire
	time.Sleep(5 * time.Millisecond)
	if !entry.IsExpired() {
		t.Fatal("Entry should be expired after original TTL")
	}

	// Set new expiry in the future
	entry.SetExpiry(100 * time.Millisecond)
	if entry.IsExpired() {
		t.Fatal("Entry should not be expired after setting new future expiry")
	}

	// Set expiry to 0 (no expiration)
	entry.SetExpiry(0)
	if entry.IsExpired() {
		t.Fatal("Entry with 0 TTL should not be considered expired")
	}

	// Set expiry to past (negative effective TTL)
	time.Sleep(10 * time.Millisecond) // Ensure some time has passed
	entry2 := backends.NewEntry("past-expiry", 10*time.Millisecond, 1)
	time.Sleep(20 * time.Millisecond) // Ensure it would be expired
	if !entry2.IsExpired() {
		t.Fatal("Entry should be expired after TTL duration")
	}
}
