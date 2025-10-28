package memo

import (
	"context"
	"github.com/ldaidone/gomemo/memo"
	"testing"
)

// BenchmarkMemoizeCold benchmarks the performance of memoization when the cache is cold.
// This measures the time to execute a function that has never been called before
// (no cache hits, all cache misses).
func BenchmarkMemoizeCold(b *testing.B) {
	m := memo.New()

	slow := func(ctx context.Context, v ...any) (any, error) {
		return v[0].(int) * 2, nil
	}

	memoized := m.MemoizeFunc(func(ctx context.Context, args ...any) (any, error) {
		return slow(ctx, args...)
	})

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = memoized(ctx, i)
	}
}

// BenchmarkMemoizeWarm benchmarks the performance of memoization when the cache is warm.
// This measures the time to retrieve values that are already in the cache
// (all cache hits, no computation).
func BenchmarkMemoizeWarm(b *testing.B) {
	m := memo.New()

	slow := func(ctx context.Context, v ...any) (any, error) {
		var x []int
		for _, y := range v {
			x = append(x, y.([]any)[0].(int))
		}
		return x[0] * 2, nil
	}

	memoized := m.MemoizeFunc(func(ctx context.Context, args ...any) (any, error) {
		return slow(ctx, args)
	})

	ctx := context.Background()
	// prime cache
	for i := 0; i < 1000; i++ {
		_, _ = memoized(ctx, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = memoized(ctx, i%1000)
	}
}
