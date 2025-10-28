package memo

import (
	"context"
	"github.com/ldaidone/gomemo/memo"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestSingleflightConcurrent tests that concurrent calls with the same arguments
// are deduplicated using the singleflight mechanism. This ensures that only
// one execution happens while others wait for the same result.
func TestSingleflightConcurrent(t *testing.T) {
	m := memo.New()

	var calls int32
	slow := func(ctx context.Context, v int) (int, error) {
		// artificially long
		atomic.AddInt32(&calls, 1)
		// wait on ctx or small sleep inside implementation; we keep it simple
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-timeAfter(50):
			return v + 100, nil
		}
	}

	memoized := m.MemoizeFunc(func(ctx context.Context, args ...any) (any, error) {
		return slow(ctx, args[0].(int))
	})

	ctx := context.Background()
	const concurrency = 50
	wg := sync.WaitGroup{}
	wg.Add(concurrency)

	results := make([]int, concurrency)
	errs := make([]error, concurrency)

	for i := 0; i < concurrency; i++ {
		idx := i
		go func() {
			defer wg.Done()
			r, e := memoized(ctx, 7)
			results[idx] = r.(int)
			errs[idx] = e
		}()
	}

	wg.Wait()

	// all should succeed and have same value
	for i := 0; i < concurrency; i++ {
		if errs[i] != nil {
			t.Fatalf("concurrent call %d returned error: %v", i, errs[i])
		}
		if results[i] != results[0] {
			t.Fatalf("inconsistent results: %v vs %v", results[i], results[0])
		}
	}

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("underlying function should be called only once, called=%d", calls)
	}
}

// timeAfter returns a channel that will be closed after the specified duration in milliseconds.
// This is a helper function to avoid importing time directly in the test function.
func timeAfter(ms int) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		close(ch)
	}()
	return ch
}
