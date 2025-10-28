package memo

import (
	"context"
	"testing"
	"time"

	"github.com/ldaidone/gomemo/memo"
)

// TestSingleFlight tests the basic singleflight functionality
func TestSingleFlight(t *testing.T) {
	// This test verifies that concurrent calls to the same key get deduplicated
	sf := memo.NewSingleFlight()

	// Create a slow function to ensure overlap
	slowFn := func(ctx context.Context) (any, error) {
		time.Sleep(50 * time.Millisecond) // Give time for second call to register
		return "result", nil
	}

	ctx := context.Background()

	// Launch multiple calls concurrently to the same key
	results := make([]struct {
		result any
		err    error
		shared bool
	}, 5)

	done := make(chan int, 5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			results[i].result, results[i].err, results[i].shared = sf.Do(ctx, "key", slowFn)
			done <- i
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify all got the same result
	for i, r := range results {
		if r.err != nil {
			t.Errorf("Call %d had error: %v", i, r.err)
		}
		if r.result != "result" {
			t.Errorf("Call %d expected 'result', got: %v", i, r.result)
		}
	}

	// Exactly one call should have done the work (shared = true), others should have waited (shared = false)
	trueCount := 0
	for _, r := range results {
		if r.shared {
			trueCount++
		}
	}

	if trueCount != 1 {
		t.Errorf("Expected exactly 1 call to have shared=true, got %d", trueCount)
	}
}

// TestSingleFlightCancellation tests that context cancellation works with singleflight
func TestSingleFlightCancellation(t *testing.T) {
	sf := memo.NewSingleFlight()

	// Create a context that will be cancelled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Make two calls with cancelled context
	result1, err1, shared1 := sf.Do(ctx, "cancel-key", func(ctx2 context.Context) (any, error) {
		<-ctx2.Done() // Function would wait for context done
		return "result", ctx2.Err()
	})

	result2, err2, shared2 := sf.Do(ctx, "cancel-key", func(ctx2 context.Context) (any, error) {
		// This should potentially execute if no other call is in progress,
		// or wait if the first is still registered
		<-ctx2.Done()
		return "result", ctx2.Err()
	})

	// Both calls should have errors due to cancellation
	if err1 == nil {
		t.Fatal("Expected error due to context cancellation in first call")
	}
	if err2 == nil {
		t.Fatal("Expected error due to context cancellation in second call")
	}

	// Since we cancelled immediately, the behavior depends on timing
	// At least one of them should be the original executor
	if shared1 == false && shared2 == false {
		t.Fatal("Expected at least one call to be original executor")
	}

	// Use results to avoid "not used" errors
	_ = result1
	_ = result2
}

// TestSingleFlightRace tests race conditions in singleflight
func TestSingleFlightRace(t *testing.T) {
	sf := memo.NewSingleFlight()

	fn := func(ctx context.Context) (any, error) {
		time.Sleep(1 * time.Millisecond) // Very short work
		return "race-result", nil
	}

	ctx := context.Background()
	const concurrency = 10

	// Launch multiple goroutines calling the same key
	done := make(chan struct{})
	results := make([]struct {
		result any
		err    error
		shared bool
	}, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			results[idx].result, results[idx].err, results[idx].shared = sf.Do(ctx, "race-key", fn)
			done <- struct{}{}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// Check results
	executedCount := 0
	for i, r := range results {
		if r.err != nil {
			t.Fatalf("Goroutine %d got unexpected error: %v", i, r.err)
		}
		if r.result != "race-result" {
			t.Fatalf("Goroutine %d got unexpected result: %v", i, r.result)
		}
		if r.shared {
			executedCount++
		}
	}

	// Only one call should have executed (the shared one)
	if executedCount != 1 {
		t.Fatalf("Expected exactly 1 execution, got %d", executedCount)
	}
}
