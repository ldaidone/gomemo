package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ldaidone/gomemo/memo"
	"github.com/ldaidone/gomemo/pkg/backends"
)

func main() {
	// Create a memory backend using the factory and memoizer with metrics enabled
	memBackend, err := backends.NewBackend("memory")
	if err != nil {
		fmt.Printf("Error creating backend: %v\n", err)
		return
	}

	m := memo.New(
		memo.WithBackend(memBackend),
		memo.WithTTL(2*time.Second),
		memo.WithMetrics(true), // Enable metrics to track performance
	)

	expensiveOp := func() (any, error) {
		// Simulate an expensive operation that takes 500ms
		time.Sleep(500 * time.Millisecond)
		return "computed value", nil
	}

	ctx := context.Background()

	// Demonstrate cache hit/miss behavior
	fmt.Println("=== Cache Demo ===")
	for i := 0; i < 3; i++ {
		start := time.Now()
		val, err := m.Get(ctx, "demo-key", expensiveOp)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("[%v] Got: %v (after %v)\n", i, val, time.Since(start))
	}

	// Show metrics after operations
	metrics := m.Metrics()
	if metrics != nil && metrics.Enabled {
		snapshot := metrics.Snapshot()
		fmt.Printf("\n=== Metrics ===\n")
		fmt.Printf("Requests: %d\n", snapshot.Requests)
		fmt.Printf("Hits: %d\n", snapshot.Hits)
		fmt.Printf("Misses: %d\n", snapshot.Misses)
		fmt.Printf("Hit Ratio: %.2f%%\n", snapshot.HitRatio()*100)
	}

	// Demonstrate TTL expiration
	fmt.Println("\n=== TTL Demo ===")
	fmt.Println("Waiting for TTL to expire (2 seconds)...")
	time.Sleep(3 * time.Second)

	start := time.Now()
	val, err := m.Get(ctx, "demo-key", expensiveOp)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("After TTL: Got %v (after %v)\n", val, time.Since(start))
	}

	// Example with different key to show that only the expired key recomputes
	fmt.Println("\n=== Multiple Keys Demo ===")
	for i := 0; i < 2; i++ {
		start := time.Now()
		val, err := m.Get(ctx, "demo-key-2", expensiveOp)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("[Key2-%v] Got: %v (after %v)\n", i, val, time.Since(start))
	}

	// Show final metrics
	if metrics != nil && metrics.Enabled {
		snapshot := metrics.Snapshot()
		fmt.Printf("\n=== Final Metrics ===\n")
		fmt.Printf("Total Requests: %d\n", snapshot.Requests)
		fmt.Printf("Total Hits: %d\n", snapshot.Hits)
		fmt.Printf("Total Misses: %d\n", snapshot.Misses)
		fmt.Printf("Final Hit Ratio: %.2f%%\n", snapshot.HitRatio()*100)
	}
}
