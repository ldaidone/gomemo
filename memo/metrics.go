// Package memo provides generic memoization functionality with pluggable backends.
package memo

import (
	"sync/atomic"
	"time"
)

// Metrics stores atomic counters for cache performance tracking.
// All fields are thread-safe and updated atomically.
// Use the methods on this struct to safely access and manipulate the metrics.
type Metrics struct {
	// Enabled indicates whether metrics collection is active.
	Enabled bool

	// Hits counts the number of successful cache hits.
	Hits uint64

	// Misses counts the number of cache misses.
	Misses uint64

	// Evictions counts the number of entries removed due to expiration or cleanup.
	Evictions uint64

	// Requests counts the total number of cache requests (hits + misses).
	Requests uint64

	// totalLatency is the sum of all recorded latencies (in microseconds).
	totalLatency uint64
	// countLatency is the number of latency samples recorded.
	countLatency uint64
	// minLatency is the minimum latency observed (in microseconds).
	minLatency int64
	// maxLatency is the maximum latency observed (in microseconds).
	maxLatency int64
	// lastLatency is the duration of the last recorded computation (in microseconds).
	lastLatency int64
}

// NewMetrics creates a new metrics collector.
func NewMetrics(enabled bool) *Metrics {
	m := &Metrics{Enabled: enabled}
	if enabled {
		atomic.StoreInt64(&m.minLatency, int64(^uint64(0)>>1)) // set to max int64
	}
	return m
}

// RecordHit increments hit counters.
func (m *Metrics) RecordHit() {
	if !m.Enabled {
		return
	}
	atomic.AddUint64(&m.Hits, 1)
	atomic.AddUint64(&m.Requests, 1)
}

// RecordMiss increments miss counters.
func (m *Metrics) RecordMiss() {
	if !m.Enabled {
		return
	}
	atomic.AddUint64(&m.Misses, 1)
	atomic.AddUint64(&m.Requests, 1)
}

// RecordEviction increments eviction counter.
func (m *Metrics) RecordEviction() {
	if !m.Enabled {
		return
	}
	atomic.AddUint64(&m.Evictions, 1)
}

// RecordLatency tracks compute duration in microseconds.
func (m *Metrics) RecordLatency(duration time.Duration) {
	if !m.Enabled {
		return
	}

	microseconds := duration.Microseconds()
	atomic.StoreInt64(&m.lastLatency, microseconds)
	atomic.AddUint64(&m.totalLatency, uint64(microseconds))
	atomic.AddUint64(&m.countLatency, 1)

	for {
		oldMin := atomic.LoadInt64(&m.minLatency)
		if microseconds >= oldMin {
			break
		}
		if atomic.CompareAndSwapInt64(&m.minLatency, oldMin, microseconds) {
			break
		}
	}

	for {
		oldMax := atomic.LoadInt64(&m.maxLatency)
		if microseconds <= oldMax {
			break
		}
		if atomic.CompareAndSwapInt64(&m.maxLatency, oldMax, microseconds) {
			break
		}
	}
}

// Snapshot returns a copy of current metrics safely.
func (m *Metrics) Snapshot() Metrics {
	dupe := Metrics{
		Enabled:      m.Enabled,
		Hits:         atomic.LoadUint64(&m.Hits),
		Misses:       atomic.LoadUint64(&m.Misses),
		Evictions:    atomic.LoadUint64(&m.Evictions),
		Requests:     atomic.LoadUint64(&m.Requests),
		totalLatency: atomic.LoadUint64(&m.totalLatency),
		countLatency: atomic.LoadUint64(&m.countLatency),
		minLatency:   atomic.LoadInt64(&m.minLatency),
		maxLatency:   atomic.LoadInt64(&m.maxLatency),
		lastLatency:  atomic.LoadInt64(&m.lastLatency),
	}
	return dupe
}

// HitRatio returns cache efficiency (hits / total).
func (m *Metrics) HitRatio() float64 {
	if !m.Enabled {
		return 0.0
	}
	total := atomic.LoadUint64(&m.Requests)
	if total == 0 {
		return 0.0
	}
	hits := atomic.LoadUint64(&m.Hits)
	return float64(hits) / float64(total)
}

// AvgLatency returns the average latency (microseconds).
func (m *Metrics) AvgLatency() float64 {
	count := atomic.LoadUint64(&m.countLatency)
	if count == 0 {
		return 0.0
	}
	total := atomic.LoadUint64(&m.totalLatency)
	return float64(total) / float64(count)
}

// MinLatency returns minimum observed latency.
func (m *Metrics) MinLatency() time.Duration {
	microseconds := atomic.LoadInt64(&m.minLatency)
	if microseconds < 0 {
		return 0
	}
	return time.Duration(microseconds) * time.Microsecond
}

// MaxLatency returns maximum observed latency.
func (m *Metrics) MaxLatency() time.Duration {
	return time.Duration(atomic.LoadInt64(&m.maxLatency)) * time.Microsecond
}

// LastLatency returns the duration of the last recorded computation.
func (m *Metrics) LastLatency() time.Duration {
	return time.Duration(atomic.LoadInt64(&m.lastLatency)) * time.Microsecond
}
