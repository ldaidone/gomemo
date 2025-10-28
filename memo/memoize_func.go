// Package memo provides generic memoization functionality with pluggable backends.
package memo

import (
	"context"
	"fmt"
)

// MemoizeFunc wraps a function with memoization capabilities.
// It returns a new function that caches results based on the arguments provided.
// This is a higher-level API that automatically handles key generation based on arguments.
//
// The returned function will cache results using the memoizer's backend and
// apply singleflight deduplication for concurrent calls with the same arguments.
//
// Example:
//
//	m := memo.New()
//	expensiveFunc := func(ctx context.Context, x int) (int, error) {
//	    // Expensive computation
//	    return x * 2, nil
//	}
//	memoized := m.MemoizeFunc(expensiveFunc)
//	result, err := memoized(ctx, 42) // First call computes and caches
//	result, err := memoized(ctx, 42) // Second call returns cached value
func (m *Memoizer) MemoizeFunc(fn func(ctx context.Context, args ...any) (any, error)) func(context.Context, ...any) (any, error) {
	return func(ctx context.Context, args ...any) (any, error) {
		// Generate a key based on function and arguments
		key := "memoized_func_" // This could be more sophisticated to include args

		// If we have a key function defined in options, use it
		if m.opts.KeyFunc != nil {
			key = m.opts.KeyFunc(args...)
		} else {
			// Default key generation - convert args to string representation
			// For now, using a simple approach - in production we'd hash the args
			key += fmt.Sprintf("%v", args)
		}

		// Use the existing Get method which handles singleflight and caching
		result, err := m.Get(ctx, key, func() (any, error) {
			return fn(ctx, args...)
		})

		return result, err
	}
}
