// Package redis provides a Redis cache backend implementation.
package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ldaidone/gomemo/pkg/backends"
	goredis "github.com/redis/go-redis/v9"
)

// redisBackend implements the backends.Backend interface for Redis.
// It stores values in Redis with serialization using gob encoding
// and manages expiration times using Redis TTL.
type redisBackend struct {
	client *goredis.Client // Redis client connection
	prefix string          // Key prefix to namespace gomemo keys
	ctx    context.Context // Context for Redis operations
}

var _ backends.Backend = (*redisBackend)(nil)

// New creates a new Redis backend with the specified address, prefix, and database.
// If prefix is empty, it defaults to "gomemo:".
// The backend automatically registers itself with the backend factory system.
func New(addr, prefix string, db int) backends.Backend {
	if prefix == "" {
		prefix = "gomemo:"
	}
	client := goredis.NewClient(&goredis.Options{
		Addr: addr,
		DB:   db,
	})

	return &redisBackend{
		client: client,
		prefix: prefix,
		ctx:    context.Background(),
	}
}

func init() {
	backends.RegisterBackend("redis", func() backends.Backend {
		return New("127.0.0.1:6379", "gomemo:", 0)
	})
}

// -----------------------------------------------------------------------------
// Backend interface
// -----------------------------------------------------------------------------

func (r *redisBackend) Get(key string) (any, bool) {
	var err error
	var data []byte

	data, err = r.client.Get(r.ctx, r.prefixed(key)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			log.Printf("[gomemo][redis] data error: %v\n", err)
			return nil, false
		}
		return nil, false
	}

	var entry backends.CacheEntry
	if err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&entry); err != nil {
		log.Printf("[gomemo][redis] decode error: %v\n", err)
		return nil, false
	}

	// Check if expired (using entry.IsExpired())
	if entry.IsExpired() {
		// proactive cleanup
		if err = r.client.Del(r.ctx, r.prefixed(key)).Err(); err != nil {
			log.Printf("[gomemo][redis] expiry error: %v\n", err)
		}
		log.Printf("[gomemo][redis] expired entry: %s\n", key)
		return nil, false
	}

	return entry.Value, true
}

func (r *redisBackend) Set(key string, value any, ttl time.Duration) {
	var entry backends.CacheEntry
	var buf bytes.Buffer
	var err error

	entry = backends.NewEntry(value, ttl, 0)

	if err = gob.NewEncoder(&buf).Encode(entry); err != nil {
		log.Printf("[gomemo][redis] encode error: %v\n", err)
		return
	}

	err = r.client.Set(r.ctx, r.prefixed(key), buf.Bytes(), ttl).Err()
	if err != nil {
		log.Printf("[gomemo][redis] set error: %v\n", err)
	}
}

func (r *redisBackend) Delete(key string) {
	var err error
	if err = r.client.Del(r.ctx, r.prefixed(key)).Err(); err != nil {
		log.Printf("[gomemo][redis] delete error: %v\n", err)
	}
}

func (r *redisBackend) Clear() {
	var err error
	var cursor, next uint64
	var keys []string

	for {
		keys, next, err = r.client.Scan(r.ctx, cursor, r.prefix+"*", 100).Result()
		if err != nil {
			fmt.Printf("[gomemo][redis] scan error: %v\n", err)
			return
		}
		if len(keys) > 0 {
			if err = r.client.Del(r.ctx, keys...).Err(); err != nil {
				log.Printf("[gomemo][redis] clear error: %v\n", err)
			}

		}
		if next == 0 {
			break
		}
		cursor = next
	}
}

func (r *redisBackend) prefixed(key string) string {
	return r.prefix + key
}
