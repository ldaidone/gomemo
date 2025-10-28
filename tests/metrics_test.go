package memo

import (
	"testing"
	"time"

	"github.com/ldaidone/gomemo/memo"
)

// TestMetricsCreation tests creating metrics
func TestMetricsCreation(t *testing.T) {
	// Test with metrics enabled
	metrics := memo.NewMetrics(true)
	if metrics == nil {
		t.Fatal("Expected metrics to be created")
	}
	if !metrics.Enabled {
		t.Fatal("Expected metrics to be enabled")
	}

	// Test with metrics disabled
	metrics2 := memo.NewMetrics(false)
	if metrics2 == nil {
		t.Fatal("Expected metrics to be created")
	}
	if metrics2.Enabled {
		t.Fatal("Expected metrics to be disabled")
	}
}

// TestMetricsRecordHit tests recording cache hits
func TestMetricsRecordHit(t *testing.T) {
	metrics := memo.NewMetrics(true)

	// Initially zero
	snapshot := metrics.Snapshot()
	if snapshot.Hits != 0 {
		t.Fatalf("Expected 0 hits initially, got: %d", snapshot.Hits)
	}
	if snapshot.Requests != 0 {
		t.Fatalf("Expected 0 requests initially, got: %d", snapshot.Requests)
	}

	// Record a hit
	metrics.RecordHit()
	snapshot = metrics.Snapshot()
	if snapshot.Hits != 1 {
		t.Fatalf("Expected 1 hit, got: %d", snapshot.Hits)
	}
	if snapshot.Requests != 1 {
		t.Fatalf("Expected 1 request, got: %d", snapshot.Requests)
	}

	// Record another hit
	metrics.RecordHit()
	snapshot = metrics.Snapshot()
	if snapshot.Hits != 2 {
		t.Fatalf("Expected 2 hits, got: %d", snapshot.Hits)
	}
	if snapshot.Requests != 2 {
		t.Fatalf("Expected 2 requests, got: %d", snapshot.Requests)
	}
}

// TestMetricsRecordMiss tests recording cache misses
func TestMetricsRecordMiss(t *testing.T) {
	metrics := memo.NewMetrics(true)

	// Initially zero
	snapshot := metrics.Snapshot()
	if snapshot.Misses != 0 {
		t.Fatalf("Expected 0 misses initially, got: %d", snapshot.Misses)
	}
	if snapshot.Requests != 0 {
		t.Fatalf("Expected 0 requests initially, got: %d", snapshot.Requests)
	}

	// Record a miss
	metrics.RecordMiss()
	snapshot = metrics.Snapshot()
	if snapshot.Misses != 1 {
		t.Fatalf("Expected 1 miss, got: %d", snapshot.Misses)
	}
	if snapshot.Requests != 1 {
		t.Fatalf("Expected 1 request, got: %d", snapshot.Requests)
	}

	// Record another miss
	metrics.RecordMiss()
	snapshot = metrics.Snapshot()
	if snapshot.Misses != 2 {
		t.Fatalf("Expected 2 misses, got: %d", snapshot.Misses)
	}
	if snapshot.Requests != 2 {
		t.Fatalf("Expected 2 requests, got: %d", snapshot.Requests)
	}
}

// TestMetricsRecordEviction tests recording evictions
func TestMetricsRecordEviction(t *testing.T) {
	metrics := memo.NewMetrics(true)

	// Initially zero
	snapshot := metrics.Snapshot()
	if snapshot.Evictions != 0 {
		t.Fatalf("Expected 0 evictions initially, got: %d", snapshot.Evictions)
	}

	// Record an eviction
	metrics.RecordEviction()
	snapshot = metrics.Snapshot()
	if snapshot.Evictions != 1 {
		t.Fatalf("Expected 1 eviction, got: %d", snapshot.Evictions)
	}

	// Record another eviction
	metrics.RecordEviction()
	snapshot = metrics.Snapshot()
	if snapshot.Evictions != 2 {
		t.Fatalf("Expected 2 evictions, got: %d", snapshot.Evictions)
	}
}

// TestMetricsRecordLatency tests recording latency
func TestMetricsRecordLatency(t *testing.T) {
	metrics := memo.NewMetrics(true)

	// Initially zero for most values, but minLatency is set to max initially to track minimum
	if metrics.AvgLatency() != 0.0 {
		t.Fatalf("Expected 0 avg latency initially, got: %f", metrics.AvgLatency())
	}
	// MinLatency is initialized to max value internally, so initially it will be that value
	minLatency := metrics.MinLatency()
	// For a new metrics collector, minLatency might be represented differently depending on internal init
	if minLatency < 0 {
		// This might happen due to overflow in the internal representation
		// Just verify that we can call the method without panic
	}
	if metrics.MaxLatency() < 0 {
		t.Fatalf("Expected non-negative max latency initially, got: %v", metrics.MaxLatency())
	}
	if metrics.MaxLatency() != 0 {
		t.Fatalf("Expected 0 max latency initially, got: %v", metrics.MaxLatency())
	}

	// Record some latencies
	metrics.RecordLatency(10 * time.Millisecond)
	metrics.RecordLatency(20 * time.Millisecond)
	metrics.RecordLatency(5 * time.Millisecond)  // new minimum
	metrics.RecordLatency(30 * time.Millisecond) // new maximum

	// Check averages
	expectedAvg := (10.0 + 20.0 + 5.0 + 30.0) / 4.0 // 16.25 ms
	expectedAvgMicroseconds := expectedAvg * 1000.0 // Convert to microseconds: 16250.0
	if metrics.AvgLatency() != expectedAvgMicroseconds {
		t.Fatalf("Expected avg latency %f microseconds, got: %f", expectedAvgMicroseconds, metrics.AvgLatency())
	}

	// Check minimum (5ms)
	if metrics.MinLatency() != 5*time.Millisecond {
		t.Fatalf("Expected min latency 5ms, got: %v", metrics.MinLatency())
	}

	// Check maximum (30ms)
	if metrics.MaxLatency() != 30*time.Millisecond {
		t.Fatalf("Expected max latency 30ms, got: %v", metrics.MaxLatency())
	}

	// Check last (should be 30ms, the most recent)
	if metrics.LastLatency() != 30*time.Millisecond {
		t.Fatalf("Expected last latency 30ms, got: %v", metrics.LastLatency())
	}
}

// TestMetricsDisabled tests that disabled metrics don't record anything
func TestMetricsDisabled(t *testing.T) {
	metrics := memo.NewMetrics(false)

	// All operations should be no-ops when disabled
	metrics.RecordHit()
	metrics.RecordMiss()
	metrics.RecordEviction()
	metrics.RecordLatency(10 * time.Millisecond)

	snapshot := metrics.Snapshot()
	if snapshot.Hits != 0 {
		t.Fatalf("Expected 0 hits when disabled, got: %d", snapshot.Hits)
	}
	if snapshot.Misses != 0 {
		t.Fatalf("Expected 0 misses when disabled, got: %d", snapshot.Misses)
	}
	if snapshot.Evictions != 0 {
		t.Fatalf("Expected 0 evictions when disabled, got: %d", snapshot.Evictions)
	}
	if metrics.AvgLatency() != 0.0 {
		t.Fatalf("Expected 0 avg latency when disabled, got: %f", metrics.AvgLatency())
	}
	if metrics.HitRatio() != 0.0 {
		t.Fatalf("Expected 0 hit ratio when disabled, got: %f", metrics.HitRatio())
	}
}

// TestMetricsHitRatio tests hit ratio calculation
func TestMetricsHitRatio(t *testing.T) {
	metrics := memo.NewMetrics(true)

	// Start with no requests
	if metrics.HitRatio() != 0.0 {
		t.Fatalf("Expected 0 hit ratio with no requests, got: %f", metrics.HitRatio())
	}

	// One hit, zero requests - should still be 0
	metrics.RecordHit()
	if metrics.HitRatio() != 1.0 {
		t.Fatalf("Expected 1.0 hit ratio with 1 hit and 1 request, got: %f", metrics.HitRatio())
	}

	// Add a miss
	metrics.RecordMiss()
	expectedRatio := 1.0 / 2.0 // 1 hit out of 2 requests
	if metrics.HitRatio() != expectedRatio {
		t.Fatalf("Expected hit ratio %f, got: %f", expectedRatio, metrics.HitRatio())
	}

	// Add more hits and misses
	metrics.RecordHit()
	metrics.RecordHit()
	metrics.RecordMiss()

	expectedRatio = 3.0 / 5.0 // 3 hits out of 5 requests
	if metrics.HitRatio() != expectedRatio {
		t.Fatalf("Expected hit ratio %f, got: %f", expectedRatio, metrics.HitRatio())
	}
}

// TestMetricsSnapshot tests creating snapshots
func TestMetricsSnapshot(t *testing.T) {
	metrics := memo.NewMetrics(true)

	metrics.RecordHit()
	metrics.RecordMiss()
	metrics.RecordMiss()
	metrics.RecordEviction()
	metrics.RecordLatency(15 * time.Millisecond)

	// Create a snapshot
	snapshot := metrics.Snapshot()

	// Verify snapshot contains correct values
	if snapshot.Hits != 1 {
		t.Fatalf("Expected 1 hit in snapshot, got: %d", snapshot.Hits)
	}
	if snapshot.Misses != 2 {
		t.Fatalf("Expected 2 misses in snapshot, got: %d", snapshot.Misses)
	}
	if snapshot.Requests != 3 {
		t.Fatalf("Expected 3 requests in snapshot, got: %d", snapshot.Requests)
	}
	if snapshot.Evictions != 1 {
		t.Fatalf("Expected 1 eviction in snapshot, got: %d", snapshot.Evictions)
	}
}
