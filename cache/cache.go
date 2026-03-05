// Package cache defines the Cache interface for key-value storage operations.
//
// The interface is intentionally small and maps directly to common Redis/
// Memcached operations, but it is backend-agnostic. Use the redis sub-package
// for a production-ready Redis implementation.
//
// Sentinel errors ErrNotFound and ErrNilValue are provided for consistent
// error handling across implementations.
package cache

import (
	"context"
	"errors"
	"time"
)

// ---------------------------------------------------------------------------
// Sentinel errors
// ---------------------------------------------------------------------------

var (
	// ErrNotFound is returned when a key does not exist in the cache.
	ErrNotFound = errors.New("cache: key not found")

	// ErrNilValue is returned when a nil destination is passed to Get.
	ErrNilValue = errors.New("cache: nil value")
)

// ---------------------------------------------------------------------------
// Cache interface
// ---------------------------------------------------------------------------

// Cache is the primary key-value abstraction. Implementations must be safe
// for concurrent use by multiple goroutines.
type Cache interface {
	// ---- String operations ----

	// Set stores value under key with the given TTL. A zero TTL means the key
	// does not expire.
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Get retrieves the value for key and deserialises it into dest.
	// Returns ErrNotFound if the key does not exist.
	Get(ctx context.Context, key string, dest any) error

	// Delete removes one or more keys. Non-existent keys are silently ignored.
	Delete(ctx context.Context, keys ...string) error

	// Exists returns the number of keys that exist.
	Exists(ctx context.Context, keys ...string) (int64, error)

	// Expire sets a new TTL on an existing key.
	Expire(ctx context.Context, key string, ttl time.Duration) error

	// TTL returns the remaining time-to-live for a key.
	TTL(ctx context.Context, key string) (time.Duration, error)

	// ---- Atomic counter operations ----

	// Incr atomically increments the integer value at key by 1.
	Incr(ctx context.Context, key string) (int64, error)

	// IncrBy atomically increments the integer value at key by val.
	IncrBy(ctx context.Context, key string, val int64) (int64, error)

	// Decr atomically decrements the integer value at key by 1.
	Decr(ctx context.Context, key string) (int64, error)

	// ---- Lifecycle ----

	// Ping verifies the connection to the cache backend is alive.
	Ping(ctx context.Context) error

	// Close releases any resources held by the cache (connections, pools, etc.).
	Close() error
}
