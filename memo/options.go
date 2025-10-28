package memo

import (
	"github.com/ldaidone/gomemo/internals/hashutil"
	"time"

	"github.com/ldaidone/gomemo/pkg/backends"
	"github.com/ldaidone/gomemo/pkg/backends/memory"
)

// ----------------------------------------------------------------------------
// Option system
// ----------------------------------------------------------------------------

// Options holds the configuration for a Memoizer instance.
// It contains various settings that control caching behavior,
// including TTL, backend storage, and performance metrics.
type Options struct {
	// TTL specifies the time-to-live for cached values.
	// Values will be automatically removed from cache after this duration.
	TTL time.Duration

	// KeyFunc is an optional function that generates cache keys from function arguments.
	// If nil, the default key generation will be used.
	KeyFunc func(args ...any) string

	// CacheOnCancel determines whether to cache results when the context is cancelled.
	// If true, cancelled requests may still update the cache.
	CacheOnCancel bool

	// CleanupInterval specifies how frequently to clean up expired entries.
	// This is used by backends that require periodic cleanup.
	CleanupInterval time.Duration

	// Backend specifies the storage backend for the cache.
	// If nil, the default memory backend will be used.
	Backend backends.Backend

	// MetricsEnabled enables or disables performance metrics collection.
	// When enabled, cache hit/miss ratios and other statistics will be tracked.
	MetricsEnabled bool
}

// Option is a function that modifies Options.
type Option func(*Options)

// DefaultOptions returns a sane default configuration.
func DefaultOptions() *Options {
	return &Options{
		TTL:             time.Hour,
		KeyFunc:         hashutil.HashArgs,
		CacheOnCancel:   false,
		CleanupInterval: time.Hour,
		Backend:         memory.New(),
		MetricsEnabled:  false,
	}
}

// ----------------------------------------------------------------------------
// Option builders
// ----------------------------------------------------------------------------

// WithTTL sets the time-to-live for cached values.
// Values will be automatically removed from cache after this duration.
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}

// WithKeyFunc sets a custom function for generating cache keys from function arguments.
// This allows fine-grained control over key generation for memoized functions.
func WithKeyFunc(fn func(args ...any) string) Option {
	return func(o *Options) {
		o.KeyFunc = fn
	}
}

// WithBackend sets the storage backend for the cache.
// Different backends provide different storage characteristics (in-memory, Redis, etc.).
func WithBackend(b backends.Backend) Option {
	return func(o *Options) {
		o.Backend = b
	}
}

// WithCleanupInterval sets how frequently to clean up expired entries.
// This is used by backends that require periodic cleanup of expired entries.
func WithCleanupInterval(d time.Duration) Option {
	return func(o *Options) {
		o.CleanupInterval = d
	}
}

// WithCacheOnCancel determines whether to cache results when the context is cancelled.
// When enabled, cancelled requests may still update the cache with their results.
func WithCacheOnCancel(enabled bool) Option {
	return func(o *Options) {
		o.CacheOnCancel = enabled
	}
}

// WithMetrics enables or disables performance metrics collection.
// When enabled, cache hit ratios, request counts, and other statistics are tracked.
func WithMetrics(enabled bool) Option {
	return func(o *Options) {
		o.MetricsEnabled = enabled
	}
}
