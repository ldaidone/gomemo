package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ldaidone/gomemo/memo"
	"github.com/ldaidone/gomemo/pkg/backends"
	_ "github.com/ldaidone/gomemo/pkg/backends/memory"
)

func RunFibonacci() {
	backend, err := backends.NewBackend("memory")
	if err != nil {
		panic(err)
	}
	m := memo.New(memo.WithBackend(backend), memo.WithTTL(10*time.Minute), memo.WithMetrics(true))
	ctx := context.Background()

	// same memoized recursive fib as before
	var fib func(int) int
	fib = func(n int) int {
		key := fmt.Sprintf("fib-%d", n)
		v, err := m.Get(ctx, key, func() (any, error) {
			if n <= 1 {
				return n, nil
			}
			return fib(n-1) + fib(n-2), nil
		})
		if err != nil {
			panic(err)
		}
		return v.(int)
	}

	// compute several fibs concurrently where many overlap
	var wg sync.WaitGroup
	inputs := []int{35, 34, 33, 32, 31}
	wg.Add(len(inputs))
	start := time.Now()
	for _, n := range inputs {
		n := n
		go func() {
			defer wg.Done()
			fmt.Printf("fib(%d) = %d\n", n, fib(n))
		}()
	}
	wg.Wait()
	fmt.Printf("Concurrent run finished in %s\n", time.Since(start))

	// metrics
	stats := m.Metrics().Snapshot()
	fmt.Printf("Hits=%d, Misses=%d, Requests=%d, HitRatio=%.2f%%\n",
		stats.Hits, stats.Misses, stats.Requests, stats.HitRatio()*100)
}
