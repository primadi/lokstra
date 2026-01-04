package kvstore_inmemory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
)

const SERVICE_TYPE = "kvstore_inmemory"

var ErrKeyNotFound = errors.New("key not found")

var (
	mu   sync.RWMutex
	data = make(map[string]kvEntry)

	MaxCounter     int = 30
	cleanupCounter int
)

type kvEntry struct {
	value     any
	expiresAt *time.Time
}

type kvStoreInMemory struct {
	prefix string
}

func checkCleanUp() {
	go func() {
		mu.Lock()
		defer mu.Unlock()

		if cleanupCounter < MaxCounter {
			cleanupCounter++
			return
		}

		// Perform cleanup
		now := time.Now()
		for key, entry := range data {
			if entry.expiresAt != nil && now.After(*entry.expiresAt) {
				delete(data, key)
			}
		}
		cleanupCounter = 0
	}()
}

func (k *kvStoreInMemory) prefixKey(key string) string {
	if k.prefix != "" {
		return k.prefix + ":" + key
	}
	return key
}

// Delete implements [serviceapi.KvStore].
func (k *kvStoreInMemory) Delete(ctx context.Context, key string) error {
	mu.Lock()
	delete(data, k.prefixKey(key))
	mu.Unlock()

	checkCleanUp()
	return nil
}

// DeleteKeys implements [serviceapi.KvStore].
func (k *kvStoreInMemory) DeleteKeys(ctx context.Context, keys ...string) error {
	mu.Lock()
	for _, key := range keys {
		delete(data, k.prefixKey(key))
	}
	mu.Unlock()

	checkCleanUp()
	return nil
}

// Get implements [serviceapi.KvStore].
func (k *kvStoreInMemory) Get(ctx context.Context, key string, dest any) error {
	mu.RLock()
	defer mu.RUnlock()

	entry, exists := data[k.prefixKey(key)]
	if !exists {
		return ErrKeyNotFound
	}
	if entry.expiresAt != nil && time.Now().After(*entry.expiresAt) {
		return ErrKeyNotFound
	}
	dest = entry.value
	return nil
}

// GetPrefix implements [serviceapi.KvStore].
func (k *kvStoreInMemory) GetPrefix() string {
	return k.prefix
}

// Keys implements [serviceapi.KvStore].
func (k *kvStoreInMemory) Keys(ctx context.Context, pattern string) ([]string, error) {
	pattern = k.prefixKey(pattern)
	lenPattern := len(pattern)
	wildcardPattern := lenPattern > 0 && pattern[lenPattern-1] == '*'
	startItem := len(k.prefix)
	if startItem > 0 {
		startItem++ // to account for the colon
	}

	var keys []string

	mu.RLock()
	defer mu.RUnlock()
	for key, value := range data {
		// Simple pattern matching: only supports '*' at the end
		if wildcardPattern {
			prefix := pattern[:lenPattern-1]
			if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
				if value.expiresAt == nil || time.Now().Before(*value.expiresAt) {
					keys = append(keys, key[startItem:])
				}
			}
		} else if key == pattern {
			if value.expiresAt == nil || time.Now().Before(*value.expiresAt) {
				keys = append(keys, key[startItem:])
			}
		}
	}
	return keys, nil
}

// Set implements [serviceapi.KvStore].
func (k *kvStoreInMemory) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	mu.Lock()
	var expiresAt *time.Time
	if ttl > 0 {
		t := time.Now().Add(ttl)
		expiresAt = &t
	}
	data[k.prefixKey(key)] = kvEntry{
		value:     value,
		expiresAt: expiresAt,
	}
	mu.Unlock()

	checkCleanUp()
	return nil
}

// SetPrefix implements [serviceapi.KvStore].
func (k *kvStoreInMemory) SetPrefix(prefix string) {
	k.prefix = prefix
}

var _ serviceapi.KvStore = (*kvStoreInMemory)(nil)

// creates a new instance of kvStoreInMemory service.
func Service(prefix string) *kvStoreInMemory {
	return &kvStoreInMemory{
		prefix: prefix,
	}
}

// the factory function for kvStoreInMemory service.
func ServiceFactory(config map[string]any) any {
	prefix := utils.GetValueFromMap(config, "prefix", "")
	return Service(prefix)
}

// registers the kvStoreInMemory service type.
func Register() {
	lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory)
}
