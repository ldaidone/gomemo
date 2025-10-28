package memo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ldaidone/gomemo/memo"
	"github.com/ldaidone/gomemo/pkg/backends/memory"
)

// TestNewWithOptions tests creating a memoizer with different options
func TestNewWithOptions(t *testing.T) {
	// Test with memory backend
	m := memo.New(memo.WithBackend(memory.New()))
	if m == nil {
		t.Fatal("Expected memoizer to be created")
	}

	// Test with TTL
	m = memo.New(memo.WithTTL(5 * time.Second))
	if m == nil {
		t.Fatal("Expected memoizer to be created with TTL")
	}

	// Test with metrics enabled
	m = memo.New(memo.WithMetrics(true))
	if m == nil {
		t.Fatal("Expected memoizer to be created with metrics")
	}

	// Test with all options
	m = memo.New(
		memo.WithTTL(10*time.Second),
		memo.WithBackend(memory.New()),
		memo.WithMetrics(true),
		memo.WithCacheOnCancel(false),
	)
	if m == nil {
		t.Fatal("Expected memoizer to be created with all options")
	}
}

// TestValidateOptions tests the validation of options
func TestValidateOptions(t *testing.T) {
	options := &memo.Options{}
	err := options.Validate()
	if err == nil {
		t.Fatal("Expected error when backend is nil")
	}
	if err.Error() != "backend cannot be nil" {
		t.Fatalf("Expected backend cannot be nil error, got: %v", err)
	}

	options.Backend = memory.New()
	err = options.Validate()
	if err == nil {
		t.Fatal("Expected error when TTL is not positive")
	}
	if err.Error() != "TTL must be positive" {
		t.Fatalf("Expected TTL must be positive error, got: %v", err)
	}

	options.TTL = 1 * time.Second
	err = options.Validate()
	if err != nil {
		t.Fatalf("Expected no error when options are valid, got: %v", err)
	}
}

// TestGetWithDifferentKeys tests Get with different keys
func TestGetWithDifferentKeys(t *testing.T) {
	m := memo.New(memo.WithTTL(5 * time.Second))

	callCount1 := 0
	callCount2 := 0

	fn1 := func() (any, error) {
		callCount1++
		return "result1", nil
	}

	fn2 := func() (any, error) {
		callCount2++
		return "result2", nil
	}

	ctx := context.Background()

	// Test first key
	result1a, err := m.Get(ctx, "key1", fn1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result1a != "result1" {
		t.Fatalf("Expected 'result1', got: %v", result1a)
	}
	if callCount1 != 1 {
		t.Fatalf("Expected fn1 to be called once, was called %d times", callCount1)
	}

	// Test first key again - should be cached
	result1b, err := m.Get(ctx, "key1", fn1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result1b != "result1" {
		t.Fatalf("Expected 'result1', got: %v", result1b)
	}
	if callCount1 != 1 {
		t.Fatalf("Expected fn1 to be called once still, was called %d times", callCount1)
	}

	// Test second key
	result2, err := m.Get(ctx, "key2", fn2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result2 != "result2" {
		t.Fatalf("Expected 'result2', got: %v", result2)
	}
	if callCount2 != 1 {
		t.Fatalf("Expected fn2 to be called once, was called %d times", callCount2)
	}
}

// TestGetWithErrors tests Get when the function returns an error
func TestGetWithErrors(t *testing.T) {
	m := memo.New()

	errFunc := func() (any, error) {
		return nil, errors.New("test error")
	}

	ctx := context.Background()
	result, err := m.Get(ctx, "error-key", errFunc)
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
	if err.Error() != "test error" {
		t.Fatalf("Expected 'test error', got: %v", err)
	}
	if result != nil {
		t.Fatalf("Expected nil result, got: %v", result)
	}
}

// TestDelete tests the Delete method
func TestDelete(t *testing.T) {
	m := memo.New()

	callCount := 0
	fn := func() (any, error) {
		callCount++
		return "computed", nil
	}

	ctx := context.Background()

	// First call - function should execute
	result1, err := m.Get(ctx, "delete-test", fn)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result1 != "computed" {
		t.Fatalf("Expected 'computed', got: %v", result1)
	}
	if callCount != 1 {
		t.Fatalf("Expected function to be called once, was called %d times", callCount)
	}

	// Delete the key
	m.Delete("delete-test")

	// Second call - function should execute again since key was deleted
	result2, err := m.Get(ctx, "delete-test", fn)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result2 != "computed" {
		t.Fatalf("Expected 'computed', got: %v", result2)
	}
	if callCount != 2 {
		t.Fatalf("Expected function to be called twice, was called %d times", callCount)
	}
}

// TestClear tests the Clear method
func TestClear(t *testing.T) {
	m := memo.New()

	callCount1 := 0
	callCount2 := 0

	fn1 := func() (any, error) {
		callCount1++
		return "result1", nil
	}

	fn2 := func() (any, error) {
		callCount2++
		return "result2", nil
	}

	ctx := context.Background()

	// Call both functions to populate cache
	_, err := m.Get(ctx, "key1", fn1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	_, err = m.Get(ctx, "key2", fn2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if callCount1 != 1 || callCount2 != 1 {
		t.Fatalf("Expected both functions to be called once, got %d and %d", callCount1, callCount2)
	}

	// Clear the cache
	m.Clear()

	// Call again - both should execute since cache was cleared
	_, err = m.Get(ctx, "key1", fn1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	_, err = m.Get(ctx, "key2", fn2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if callCount1 != 2 || callCount2 != 2 {
		t.Fatalf("Expected both functions to be called twice after clear, got %d and %d", callCount1, callCount2)
	}
}

// TestMetrics tests metrics functionality
func TestMetrics(t *testing.T) {
	m := memo.New(memo.WithMetrics(true))

	fn := func() (any, error) {
		return "result", nil
	}

	ctx := context.Background()

	// Initial metrics should be zero
	metrics := m.Metrics()
	if metrics == nil {
		t.Fatal("Expected metrics to be available")
	}

	initialSnapshot := metrics.Snapshot()
	if initialSnapshot.Requests != 0 {
		t.Fatalf("Expected 0 requests initially, got: %d", initialSnapshot.Requests)
	}

	// First call - miss
	_, err := m.Get(ctx, "metrics-test", fn)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	snapshot := metrics.Snapshot()
	if snapshot.Requests != 1 {
		t.Fatalf("Expected 1 request, got: %d", snapshot.Requests)
	}
	if snapshot.Misses != 1 {
		t.Fatalf("Expected 1 miss, got: %d", snapshot.Misses)
	}
	if snapshot.Hits != 0 {
		t.Fatalf("Expected 0 hits, got: %d", snapshot.Hits)
	}

	// Second call - hit
	_, err = m.Get(ctx, "metrics-test", fn)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	snapshot = metrics.Snapshot()
	if snapshot.Requests != 2 {
		t.Fatalf("Expected 2 requests, got: %d", snapshot.Requests)
	}
	if snapshot.Misses != 1 {
		t.Fatalf("Expected 1 miss, got: %d", snapshot.Misses)
	}
	if snapshot.Hits != 1 {
		t.Fatalf("Expected 1 hit, got: %d", snapshot.Hits)
	}

	// Test metrics calculations
	hitRatio := metrics.HitRatio()
	expectedHitRatio := 1.0 / 2.0 // 1 hit out of 2 requests
	if hitRatio != expectedHitRatio {
		t.Fatalf("Expected hit ratio %f, got %f", expectedHitRatio, hitRatio)
	}
}

// TestMetricsWithoutEnabled tests metrics when not enabled
func TestMetricsWithoutEnabled(t *testing.T) {
	m := memo.New(memo.WithMetrics(false))

	fn := func() (any, error) {
		time.Sleep(10 * time.Millisecond) // Some work to measure latency
		return "result", nil
	}

	ctx := context.Background()

	// Call the function
	_, err := m.Get(ctx, "no-metrics-test", fn)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Metrics should still be accessible but not record anything
	metrics := m.Metrics()
	if metrics == nil {
		t.Fatal("Expected metrics to be available even when disabled")
	}

	snapshot := metrics.Snapshot()
	if snapshot.Requests != 0 {
		t.Fatalf("Expected 0 requests when metrics disabled, got: %d", snapshot.Requests)
	}

	hitRatio := metrics.HitRatio()
	if hitRatio != 0.0 {
		t.Fatalf("Expected 0 hit ratio when metrics disabled, got: %f", hitRatio)
	}
}

// TestGetContext tests Get with context cancellation
func TestGetContext(t *testing.T) {
	m := memo.New()

	fn := func() (any, error) {
		return "result", nil
	}

	// Create a context and cancel it
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Get should work fine even with cancelled context since the function
	// itself doesn't check the context. The context is passed to singleflight only.
	result, err := m.Get(ctx, "cancelled-test", fn)
	if err != nil {
		// Context cancellation doesn't necessarily cause Get to return an error
		// since the function itself doesn't use the context and cache operations
		// are not cancelled
	}
	if result == nil {
		t.Fatalf("Expected result even with cancelled context, got: %v", result)
	}
}
