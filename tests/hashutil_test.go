package memo

import (
	"testing"

	"github.com/ldaidone/gomemo/internals/hashutil"
)

// TestHashArgs tests the basic hashing functionality
func TestHashArgs(t *testing.T) {
	// Test with single string
	hash1 := hashutil.HashArgs("hello")
	if hash1 == "" {
		t.Fatal("Expected non-empty hash for string")
	}

	// Test that same arguments produce same hash
	hash2 := hashutil.HashArgs("hello")
	if hash1 != hash2 {
		t.Fatal("Expected same hash for same arguments")
	}

	// Test with different arguments produce different hashes
	hash3 := hashutil.HashArgs("world")
	if hash1 == hash3 {
		t.Fatal("Expected different hash for different arguments")
	}

	// Test with multiple arguments
	hash4 := hashutil.HashArgs("hello", 42)
	hash5 := hashutil.HashArgs("hello", 42)
	if hash4 != hash5 {
		t.Fatal("Expected same hash for same multiple arguments")
	}

	hash6 := hashutil.HashArgs("hello", 43)
	if hash4 == hash6 {
		t.Fatal("Expected different hash for different multiple arguments")
	}

	// Test with different types
	hash7 := hashutil.HashArgs(42)
	hash8 := hashutil.HashArgs("42") // string vs int
	if hash7 == hash8 {
		t.Fatal("Expected different hash for different types")
	}
}

// TestHashArgsComplex tests hashing with complex data types
func TestHashArgsComplex(t *testing.T) {
	// Test with slice
	slice1 := []int{1, 2, 3}
	slice2 := []int{1, 2, 3}
	slice3 := []int{1, 2, 4}

	hash1 := hashutil.HashArgs(slice1)
	hash2 := hashutil.HashArgs(slice2)
	hash3 := hashutil.HashArgs(slice3)

	if hash1 != hash2 {
		t.Fatal("Expected same hash for equal slices")
	}
	if hash1 == hash3 {
		t.Fatal("Expected different hash for different slices")
	}

	// Test with struct-like data
	type Person struct {
		Name string
		Age  int
	}

	person1 := Person{Name: "Alice", Age: 30}
	person2 := Person{Name: "Alice", Age: 30}
	person3 := Person{Name: "Bob", Age: 30}

	hashP1 := hashutil.HashArgs(person1)
	hashP2 := hashutil.HashArgs(person2)
	hashP3 := hashutil.HashArgs(person3)

	if hashP1 != hashP2 {
		t.Fatal("Expected same hash for equal structs")
	}
	if hashP1 == hashP3 {
		t.Fatal("Expected different hash for different structs")
	}
}

// TestHashArgsEdgeCases tests edge cases in hashing
func TestHashArgsEdgeCases(t *testing.T) {
	// Test with empty arguments
	hash1 := hashutil.HashArgs()
	hash2 := hashutil.HashArgs()
	if hash1 != hash2 {
		t.Fatal("Expected same hash for empty arguments")
	}

	// Test with nil
	hash3 := hashutil.HashArgs(nil)
	hash4 := hashutil.HashArgs(nil)
	if hash3 != hash4 {
		t.Fatal("Expected same hash for nil arguments")
	}
}

// TestFallbackHash tests the fallback hashing mechanism
// This is tested implicitly through error cases in the main function
func TestFallbackHash(t *testing.T) {
	// The fallback mechanism should handle different argument types
	// This test verifies that the function doesn't panic with various inputs
	_ = hashutil.HashArgs("string")
	_ = hashutil.HashArgs(42)
	_ = hashutil.HashArgs(3.14)
	_ = hashutil.HashArgs(true)
	_ = hashutil.HashArgs([]string{"a", "b"})
	_ = hashutil.HashArgs(map[string]int{"key": 1})

	// The function should return deterministic results for the same inputs
}
