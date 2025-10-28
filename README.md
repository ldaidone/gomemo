# gomemo

Elegant, thread-safe memoization for Go — context-aware, pluggable, and generic.

[![Go Report Card](https://goreportcard.com/badge/github.com/ldaidone/gomemo)](https://goreportcard.com/report/github.com/ldaidone/gomemo)
[![GoDoc](https://godoc.org/github.com/ldaidone/gomemo?status.svg)](https://godoc.org/github.com/ldaidone/gomemo)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![GitHub stars](https://img.shields.io/github/stars/ldaidone/gomemo.svg)](https://github.com/ldaidone/gomemo/stargazers)
[![Tests](https://github.com/ldaidone/gomemo/actions/workflows/test.yml/badge.svg)](https://github.com/ldaidone/gomemo/actions/workflows/test.yml)
[![Coverage](https://img.shields.io/badge/Coverage-92.1%25-brightgreen.svg)](https://github.com/ldaidone/gomemo)

[//]: # ([!["Buy Me A Coffee"]&#40;https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png&#41;]&#40;https://buymeacoffee.com/leodaido&#41;)

## Overview

`gomemo` is a high-performance, thread-safe memoization library for Go that provides automatic deduplication of concurrent calls for the same input. It features pluggable backends, context awareness, automatic TTL management, performance metrics, and comprehensive test coverage (92.1%). The library uses atomic operations for thread safety and provides a clean, idiomatic Go API.

## Features

- **Thread-safe**: Built with concurrent access in mind using singleflight pattern with atomic operations
- **Pluggable backends**: Support for memory, Redis, and other custom backends via registration system
- **Context-aware**: Full support for Go contexts for cancellation and timeouts
- **TTL management**: Automatic expiration of cached values with configurable cleanup
- **Performance metrics**: Comprehensive metrics collection with hit/miss ratios, latency tracking, and real-time statistics
- **Generic**: Works with any function type that returns `(any, error)`
- **Deduplication**: Prevents duplicate execution of same function with same arguments
- **High test coverage**: 92.1% test coverage with comprehensive test suite including race detection
- **Cache entry management**: Thread-safe cache entries with versioning and expiration tracking

## Installation

```bash
go get github.com/ldaidone/gomemo
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/ldaidone/gomemo/memo"
    "github.com/ldaidone/gomemo/pkg/backends"
)

func main() {
    // Create a memoizer with memory backend
    memBackend, _ := backends.NewBackend("memory")
    m := memo.New(
        memo.WithBackend(memBackend),
        memo.WithTTL(30*time.Second),
        memo.WithMetrics(true),
    )

    expensiveOp := func() (any, error) {
        time.Sleep(1 * time.Second) // Simulate expensive operation
        return "computed result", nil
    }

    ctx := context.Background()
    
    // First call - executes the function and caches the result
    result1, err := m.Get(ctx, "my-key", expensiveOp)
    if err != nil {
        panic(err)
    }
    fmt.Printf("First call: %v\n", result1)

    // Second call - returns cached result immediately
    result2, err := m.Get(ctx, "my-key", expensiveOp)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Second call: %v\n", result2)
}
```

### Using MemoizeFunc for Automatic Key Generation

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/ldaidone/gomemo/memo"
)

func main() {
    m := memo.New(memo.WithTTL(30 * time.Second))

    // Create a memoized function
    expensiveFunc := func(ctx context.Context, x int) (int, error) {
        time.Sleep(100 * time.Millisecond) // Simulate work
        return x * 2, nil
    }

    memoized := m.MemoizeFunc(func(ctx context.Context, args ...any) (any, error) {
        x := args[0].(int)
        return expensiveFunc(ctx, x)
    })

    ctx := context.Background()

    // First call with argument 5 - computes and caches
    result1, _ := memoized(ctx, 5)
    fmt.Printf("Result 1: %v\n", result1)

    // Second call with argument 5 - returns cached value
    result2, _ := memoized(ctx, 5)
    fmt.Printf("Result 2: %v\n", result2)

    // Call with different argument - computes and caches
    result3, _ := memoized(ctx, 10)
    fmt.Printf("Result 3: %v\n", result3)
}
```

## Backends

### Memory Backend (Default)

```go
// Memory backend with default configuration
memBackend, err := backends.NewBackend("memory")
if err != nil {
    panic(err)
}
```

### Redis Backend

```go
// Redis backend with connection parameters
redisBackend := redis.New("localhost:6379", "gomemo:", 0)
```

The Redis backend provides distributed caching capabilities with automatic serialization of cache entries using gob encoding. It handles TTL through Redis's native expiration mechanism.

You can easily add custom backends by implementing the `backends.Backend` interface and registering them using `backends.RegisterBackend()`:

```go
func init() {
    backends.RegisterBackend("mybackend", func() backends.Backend {
        return NewMyBackend()
    })
}
```

## Configuration Options

### Available Options

- `WithTTL(duration)`: Set time-to-live for cached values
- `WithBackend(backend)`: Specify a cache backend
- `WithKeyFunc(fn)`: Custom function for generating cache keys
- `WithCleanupInterval(duration)`: Set cleanup interval for expired entries
- `WithCacheOnCancel(bool)`: Cache results even when context is cancelled
- `WithMetrics(bool)`: Enable/disable performance metrics

### Example Configuration

```go
m := memo.New(
    memo.WithTTL(5 * time.Minute),
    memo.WithBackend(myCustomBackend),
    memo.WithMetrics(true),
    memo.WithKeyFunc(func(args ...any) string {
        // Custom key generation
        return fmt.Sprintf("%v", args)
    }),
)
```

## Performance Metrics

The library includes built-in performance metrics:

```go
// Enable metrics
m := memo.New(memo.WithMetrics(true))

// Use the memoizer...
// ...

// Access metrics
metrics := m.Metrics()
snapshot := metrics.Snapshot()

fmt.Printf("Requests: %d\n", snapshot.Requests)
fmt.Printf("Hits: %d\n", snapshot.Hits)
fmt.Printf("Misses: %d\n", snapshot.Misses)
fmt.Printf("Hit Ratio: %.2f%%\n", snapshot.HitRatio()*100)
fmt.Printf("Avg Latency: %v\n", time.Duration(metrics.AvgLatency())*time.Microsecond)
```

## Architecture

### Core Components

- **Memoizer**: Main struct that coordinates caching logic and deduplication
- **Backend Interface**: Pluggable storage systems (memory, Redis, etc.) with registration system
- **SingleFlight**: Deduplication mechanism to prevent duplicate work with context support
- **Metrics**: Comprehensive performance tracking with hit ratios, latency calculations, and atomic counters
- **CacheEntry**: Thread-safe cache entry with atomic expiration, versioning, and TTL tracking
- **HashUtil**: Deterministic key generation with fallback mechanisms for complex types

### Thread Safety

All operations are thread-safe using atomic operations and mutex synchronization where needed. The library is designed for safe concurrent access across multiple goroutines.

## Development

### Running Tests

```bash
# Run all tests with verbose output
go test -v ./...

# Run tests with race detection (recommended)
go test -race -v ./...

# Run with coverage report
go test -coverpkg=./memo/... -v ./tests/

# Run benchmarks
go test -bench=. -benchmem ./tests/
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality (targeting 92.1%+ coverage)
5. Run `go test -race ./...` to ensure no race conditions
6. Submit a pull request

### Adding New Backends

To add a new backend, implement the `backends.Backend` interface and register it:

```go
func init() {
    backends.RegisterBackend("mybackend", func() backends.Backend {
        return NewMyBackend()
    })
}
```

## License

MIT License - see the [LICENSE](LICENSE) file for details.

## Support

If you find this library useful, consider [buying me a coffee](https://www.buymeacoffee.com/your_coffee_link)! ☕️

## Acknowledgments

- Inspired by the need for efficient memoization in concurrent Go applications  
- Built with performance, safety, and ease of use in mind
- Comprehensive test suite with 92.1% coverage ensures reliability
- Pluggable backend architecture enables flexible deployment scenarios
- Atomic operations and proper synchronization provide thread safety