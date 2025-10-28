package memo

import (
	"context"
	"github.com/ldaidone/gomemo/memo"
	"sync/atomic"
	"testing"
	"time"
)

// TestTTLExpiration tests that cached values expire after their TTL duration.
// It verifies that after the TTL period, the function is executed again
// instead of returning the cached value.
func TestTTLExpiration(t *testing.T) {
	m := memo.New(memo.WithTTL(100 * time.Millisecond))

	var calls int32
	slow := func(ctx context.Context, v int) (int, error) {
		atomic.AddInt32(&calls, 1)
		return v + 1, nil
	}

	memoized := m.MemoizeFunc(func(ctx context.Context, args ...any) (any, error) {
		return slow(ctx, args[0].(int))
	})

	ctx := context.Background()

	_, err := memoized(ctx, 1)
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	// immediate second call should be a cache hit
	_, err = memoized(ctx, 1)
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 call so far, got %d", calls)
	}

	// wait until TTL expires
	time.Sleep(150 * time.Millisecond)

	_, err = memoized(ctx, 1)
	if err != nil {
		t.Fatalf("third call error: %v", err)
	}

	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected function to be called again after TTL expiration, calls=%d", calls)
	}
}
