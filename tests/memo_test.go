package memo

import (
	"context"
	"github.com/ldaidone/gomemo/memo"
	"sync/atomic"
	"testing"
	"time"
)

// TestMemoizeBasic tests the basic functionality of the memoization system.
// It verifies that the same function with the same arguments is only executed once
// and subsequent calls return the cached result.
func TestMemoizeBasic(t *testing.T) {
	m := memo.New()

	var calls int32
	slow := func(ctx context.Context, v int) (int, error) {
		atomic.AddInt32(&calls, 1)
		// simulate work
		time.Sleep(50 * time.Millisecond)
		return v * 2, nil
	}

	memoized := m.MemoizeFunc(func(ctx context.Context, args ...any) (any, error) {
		// Expect single int arg
		if len(args) != 1 {
			return 0, nil
		}
		return slow(ctx, args[0].(int))
	})

	ctx := context.Background()

	res1, err := memoized(ctx, 21)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res2, err := memoized(ctx, 21)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res1 != res2 {
		t.Fatalf("expected same results, got %v and %v", res1, res2)
	}

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected underlying function to be called once, called=%d", calls)
	}
}
