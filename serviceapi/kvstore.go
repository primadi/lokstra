package serviceapi

import (
	"context"
	"time"
)

type KvStore interface {
	// Set sets a value with a key and optional TTL.
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Get retrieves a value by key.
	Get(ctx context.Context, key string, dest any) error

	// Delete removes a single key.
	Delete(ctx context.Context, key string) error

	// Deletes multiple keys.
	DeleteKeys(ctx context.Context, keys ...string) error

	// Gets all keys matching the pattern.
	Keys(ctx context.Context, pattern string) ([]string, error)

	// Sets Prefix for all keys
	SetPrefix(prefix string)

	// Gets Prefix for all keys
	GetPrefix() string
}
