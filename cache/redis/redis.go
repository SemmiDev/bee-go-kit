// Package redis provides a production-ready Redis implementation of the
// cache.Cache interface using github.com/redis/go-redis/v9.
//
// Usage:
//
//	cfg := redis.DefaultConfig()
//	cfg.Addr = "localhost:6379"
//
//	rdb, err := redis.New(cfg)
//	if err != nil { ... }
//	defer rdb.Close()
//
//	// rdb satisfies cache.Cache
//	err = rdb.Set(ctx, "user:1", userData, 10*time.Minute)
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/semmidev/bee-go-kit/cache"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// Config holds connection and pool parameters for the Redis client.
type Config struct {
	Addr         string        // Redis server address (host:port).
	Password     string        // Redis password (empty = no auth).
	DB           int           // Redis database number.
	PoolSize     int           // Maximum number of socket connections.
	MinIdleConns int           // Minimum idle connections in the pool.
	DialTimeout  time.Duration // Timeout for establishing new connections.
	ReadTimeout  time.Duration // Timeout for read operations.
	WriteTimeout time.Duration // Timeout for write operations.
	MaxRetries   int           // Maximum number of retries before giving up.
}

// DefaultConfig returns a Config with production-safe defaults.
func DefaultConfig() Config {
	return Config{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		MaxRetries:   3,
	}
}

// ---------------------------------------------------------------------------
// RedisCache
// ---------------------------------------------------------------------------

// RedisCache implements cache.Cache backed by Redis.
type RedisCache struct {
	client *goredis.Client
}

// New creates a new RedisCache and pings the server to verify connectivity.
// Returns an error if the ping fails.
func New(cfg Config) (*RedisCache, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		MaxRetries:   cfg.MaxRetries,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis: ping failed: %w", err)
	}

	return &RedisCache{client: rdb}, nil
}

// ---------------------------------------------------------------------------
// Internal serialisation helpers
// ---------------------------------------------------------------------------

// marshal converts a value to a JSON string for storage.
// Plain strings are stored as-is (no extra quoting).
func marshal(value any) (string, error) {
	if s, ok := value.(string); ok {
		return s, nil
	}
	b, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("redis: marshal error: %w", err)
	}
	return string(b), nil
}

// unmarshal parses a raw Redis string into dest.
// If dest is *string, the raw value is assigned directly.
func unmarshal(raw string, dest any) error {
	if dest == nil {
		return cache.ErrNilValue
	}
	if s, ok := dest.(*string); ok {
		*s = raw
		return nil
	}
	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		return fmt.Errorf("redis: unmarshal error: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// cache.Cache implementation – String operations
// ---------------------------------------------------------------------------

// Set stores a value with the given TTL.
func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	v, err := marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, v, ttl).Err()
}

// Get retrieves and deserialises the value for key into dest.
// Returns cache.ErrNotFound if the key does not exist.
func (r *RedisCache) Get(ctx context.Context, key string, dest any) error {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return cache.ErrNotFound
	}
	if err != nil {
		return err
	}
	return unmarshal(val, dest)
}

// Delete removes one or more keys.
func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists returns the number of specified keys that exist.
func (r *RedisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire updates the TTL on an existing key.
func (r *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

// TTL returns the remaining TTL for a key.
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// ---------------------------------------------------------------------------
// cache.Cache implementation – Counter operations
// ---------------------------------------------------------------------------

// Incr atomically increments the value at key by 1.
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// IncrBy atomically increments the value at key by val.
func (r *RedisCache) IncrBy(ctx context.Context, key string, val int64) (int64, error) {
	return r.client.IncrBy(ctx, key, val).Result()
}

// Decr atomically decrements the value at key by 1.
func (r *RedisCache) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// ---------------------------------------------------------------------------
// cache.Cache implementation – Lifecycle
// ---------------------------------------------------------------------------

// Ping verifies the Redis connection.
func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close releases the Redis client resources.
func (r *RedisCache) Close() error {
	return r.client.Close()
}
