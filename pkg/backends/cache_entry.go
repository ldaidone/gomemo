package backends

import (
	"sync/atomic"
	"time"
)

// CacheEntry holds the stored value together with metadata used by backends.
// It's intentionally small and safe for being copied as value (except for the pointer fields).
type CacheEntry struct {
	// value stored; backends may store raw bytes or concrete types.
	Value any

	// expiry in unix nanoseconds; 0 means no expiration.
	expiry int64

	// version is a monotonic counter incremented on writes (useful for CAS/diffs).
	version uint64
}

// NewEntry creates a CacheEntry with optional ttl.
func NewEntry(v any, ttl time.Duration, ver uint64) CacheEntry {
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}
	return CacheEntry{
		Value:   v,
		expiry:  exp,
		version: ver,
	}
}

// IsExpired returns true if the entry's TTL has elapsed.
func (e *CacheEntry) IsExpired() bool {
	exp := atomic.LoadInt64(&e.expiry)
	if exp == 0 {
		return false
	}
	return time.Now().UnixNano() > exp
}

// TTLRemaining returns the remaining duration until expiration, or zero if expired or no TTL.
func (e *CacheEntry) TTLRemaining() time.Duration {
	exp := atomic.LoadInt64(&e.expiry)
	if exp == 0 {
		return 0
	}
	now := time.Now().UnixNano()
	rem := exp - now
	if rem <= 0 {
		return 0
	}
	return time.Duration(rem)
}

// SetExpiry replaces the expiry atomically (useful for resets/refresh).
func (e *CacheEntry) SetExpiry(ttl time.Duration) {
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}
	atomic.StoreInt64(&e.expiry, exp)
}

// Version returns the entry's version.
func (e *CacheEntry) Version() uint64 {
	return atomic.LoadUint64(&e.version)
}

// BumpVersion increments version and returns new value.
func (e *CacheEntry) BumpVersion() uint64 {
	return atomic.AddUint64(&e.version, 1)
}
