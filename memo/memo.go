package memo

import (
	"context"
	"errors"
	"github.com/ldaidone/gomemo/pkg/backends"
	"time"
)

// Memoizer coordinates caching logic using a backend and singleflight for deduplication.
// It provides thread-safe memoization with automatic deduplication of concurrent calls
// for the same key, preventing redundant computations.
type Memoizer struct {
	backend backends.Backend // cache storage backend
	opts    Options          // configuration options
	group   *SingleFlight    // singleflight group for deduplication
	metrics *Metrics         // metrics collector
}

// Validate checks if the Options are properly configured.
// It returns an error if required fields are missing or invalid.
func (o *Options) Validate() error {
	if o.Backend == nil {
		return errors.New("backend cannot be nil")
	}
	if o.TTL <= 0 {
		return errors.New("TTL must be positive")
	}
	return nil
}

// New creates a new Memoizer instance with the provided options.
// It configures the memoizer with a backend and optional settings.
// If no backend is provided via options, it defaults to an in-memory backend.
func New(opts ...Option) *Memoizer {
	cfg := DefaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}

	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	return &Memoizer{
		backend: cfg.Backend,
		opts:    *cfg,
		group:   NewSingleFlight(),
		metrics: NewMetrics(cfg.MetricsEnabled),
	}
}

// Get retrieves a cached value or computes and stores it if missing.
// This method implements thread-safe memoization with singleflight deduplication.
//
// The fn parameter should be a function that returns (any, error).
// If multiple goroutines call Get with the same key simultaneously,
// only one will execute fn while others wait for the result.
//
// Example:
//
//	m := memo.New(memo.WithTTL(30 * time.Second))
//	result, err := m.Get(ctx, "my-key", func() (any, error) {
//	    // Expensive computation here
//	    return expensiveOperation()
//	})
func (m *Memoizer) Get(ctx context.Context, key string, fn func() (any, error)) (any, error) {
	// 1. Attempt to get from cache
	if val, ok := m.backend.Get(key); ok {
		m.metrics.RecordHit()
		return val, nil
	}

	m.metrics.RecordMiss()
	start := time.Now()

	// 2. Prevent duplicate calls via singleflight
	v, err, _ := m.group.Do(ctx, key, func(ctx2 context.Context) (any, error) {
		// Check cache again after acquiring lock (race condition guard)
		if val, ok := m.backend.Get(key); ok {
			m.metrics.RecordHit()
			return val, nil
		}

		result, err := fn()
		if err != nil {
			return nil, err
		}

		// Store computed value
		m.backend.Set(key, result, m.opts.TTL)
		return result, nil
	})

	elapsed := time.Since(start)
	m.metrics.RecordLatency(elapsed)

	return v, err
}

// Delete removes an entry from cache.
// It removes the value associated with the given key from the backend.
func (m *Memoizer) Delete(key string) {
	m.backend.Delete(key)
}

// Clear purges all entries from the backend.
// It removes all cached values, effectively resetting the cache to empty state.
func (m *Memoizer) Clear() {
	m.backend.Clear()
}

// Metrics returns the metrics collector for this memoizer.
// The returned metrics contain statistics about cache hit/miss ratios,
// request counts, and performance metrics if metrics collection is enabled.
func (m *Memoizer) Metrics() *Metrics {
	return m.metrics
}
