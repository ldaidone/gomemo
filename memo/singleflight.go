// Package memo provides generic memoization functionality with pluggable backends.
package memo

import (
	"context"
	"sync"
)

// SingleFlight ensures that only one execution is in-flight for a given key at a time.
// It prevents duplicate work by having concurrent requests for the same key
// wait for the result of the first request rather than executing multiple times.
type SingleFlight struct {
	mu sync.Mutex       // protects m
	m  map[string]*call // lazily initialized
}

// call represents a single call to the function with a specific key.
type call struct {
	wg  sync.WaitGroup // Used to wait for a singleflight call to complete
	val any            // The result value
	err error          // The error result
}

// NewSingleFlight creates a new SingleFlight instance.
// This is used internally by Memoizer to prevent duplicate executions.
func NewSingleFlight() *SingleFlight {
	return &SingleFlight{m: make(map[string]*call)}
}

// Do executes the function fn once for the given key and returns the result.
// If another call with the same key is already in progress, Do waits for it to complete
// and returns the same result, preventing duplicate work.
//
// The function takes a context for cancellation and timeout handling.
// The bool return value indicates whether the function was executed (true) or
// whether this was a duplicate request that waited for the original (false).
func (g *SingleFlight) Do(ctx context.Context, key string, fn func(context.Context) (any, error)) (any, error, bool) {
	g.mu.Lock()
	if c, ok := g.m[key]; ok {
		// There's already a call in progress for this key
		g.mu.Unlock()
		done := make(chan struct{})
		go func() {
			c.wg.Wait()
			close(done)
		}()

		select {
		case <-ctx.Done():
			return nil, ctx.Err(), false
		case <-done:
			return c.val, c.err, false
		}
	}

	// Start a new call for this key
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	// Execute the function and store the result
	c.val, c.err = fn(ctx)
	c.wg.Done()

	// Clean up the call from the map
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err, true
}
