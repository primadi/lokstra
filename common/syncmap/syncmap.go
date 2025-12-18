package syncmap

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
)

// SyncMap is a distributed map that synchronizes across multiple nodes/servers.
// It provides map-like operations with automatic synchronization via backend storage.
// All operations are type-safe through Go generics.
//
// Key features:
// - Automatic synchronization across nodes
// - Real-time change notifications via Subscribe
// - Type-safe operations
// - Prefix-based key isolation
//
// Implementation uses SyncConfig backend (backend-agnostic).
// The actual storage implementation (PostgreSQL, Redis, etc.) is determined by the SyncConfig instance.
//
// Usage:
//
//	syncConfig := sync_config_pg.NewSyncConfigPG(cfg)
//	sm := syncmap.NewSyncMap[string](syncConfig, "myprefix")
//	sm.Set(ctx, "key", "value")
//	val, err := sm.Get(ctx, "key")

// SyncMap implements a distributed map using SyncConfig backend
// The implementation is backend-agnostic - it works with any SyncConfig implementation
type SyncMap[V any] struct {
	config serviceapi.SyncConfig
	prefix string
	mu     sync.RWMutex
	subs   map[string]func(key string, value V) // subscription ID -> callback
}

// creates a new distributed SyncMap with the given SyncConfig backend and key prefix
// The prefix is used to isolate this map's keys from other maps using the same backend
// The backend implementation (PostgreSQL, Redis, etc.) is determined by the SyncConfig instance
func NewSyncMapWithConfig[V any](config serviceapi.SyncConfig, prefix string) *SyncMap[V] {
	sm := &SyncMap[V]{
		config: config,
		prefix: prefix,
		subs:   make(map[string]func(key string, value V)),
	}

	// Load initial data from SyncConfig (data that already exists in database)
	// This ensures existing database entries are available in the SyncMap
	if allData, err := config.GetAll(context.Background()); err == nil {
		for fullKey, value := range allData {
			if _, ok := sm.stripPrefix(fullKey); ok {
				// This key belongs to our prefix - process it
				sm.handleNotification(fullKey, value)
			}
		}
	}

	// Subscribe to config changes and filter by prefix (for future changes)
	config.Subscribe(func(fullKey string, value any) {
		sm.handleNotification(fullKey, value)
	})

	return sm
}

// creates a new distributed SyncMap using the default SyncConfig service from the registry
// The prefix is used to isolate this map's keys from other maps using the same backend
// The backend implementation (PostgreSQL, Redis, etc.) is determined by the SyncConfig instance
func NewSyncMap[V any](prefix string) *SyncMap[V] {
	syncConfigService := lokstra_registry.GetService[serviceapi.SyncConfig]("sync-config")
	if syncConfigService == nil {
		panic("SyncConfig service not found in registry")
	}
	return NewSyncMapWithConfig[V](syncConfigService, prefix)
}

func (sm *SyncMap[V]) makeFullKey(key string) string {
	return sm.prefix + ":" + key
}

func (sm *SyncMap[V]) stripPrefix(fullKey string) (string, bool) {
	prefixWithColon := sm.prefix + ":"
	if strings.HasPrefix(fullKey, prefixWithColon) {
		return strings.TrimPrefix(fullKey, prefixWithColon), true
	}
	return "", false
}

func (sm *SyncMap[V]) convertValue(val any) (V, error) {
	var zero V

	if val == nil {
		return zero, nil
	}

	// Try direct type assertion first
	if typed, ok := val.(V); ok {
		return typed, nil
	}

	// Fall back to JSON conversion
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return zero, fmt.Errorf("failed to marshal value: %w", err)
	}

	var result V
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return zero, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return result, nil
}

func (sm *SyncMap[V]) handleNotification(fullKey string, value any) {
	// Check if this key belongs to our prefix
	key, ok := sm.stripPrefix(fullKey)
	if !ok {
		return // Not our prefix, ignore
	}

	// Convert value to our type
	typedVal, err := sm.convertValue(value)
	if err != nil {
		return // Conversion failed, ignore
	}

	// Notify all subscribers
	sm.mu.RLock()
	callbacks := make([]func(string, V), 0, len(sm.subs))
	for _, cb := range sm.subs {
		callbacks = append(callbacks, cb)
	}
	sm.mu.RUnlock()

	for _, cb := range callbacks {
		go cb(key, typedVal)
	}
}

// sets a value in the SyncMap
func (sm *SyncMap[V]) Set(ctx context.Context, key string, value V) error {
	fullKey := sm.makeFullKey(key)
	return sm.config.Set(ctx, fullKey, value)
}

// gets a value from the SyncMap
func (sm *SyncMap[V]) Get(ctx context.Context, key string) (V, error) {
	var zero V
	fullKey := sm.makeFullKey(key)

	val, err := sm.config.Get(ctx, fullKey)
	if err != nil {
		// Key not found
		return zero, err
	}

	typedVal, err := sm.convertValue(val)
	if err != nil {
		return zero, err
	}

	return typedVal, nil
}

func (sm *SyncMap[V]) Load(key string) (V, bool) {
	var zero V
	fullKey := sm.makeFullKey(key)
	val, err := sm.config.Get(context.Background(), fullKey)
	if err != nil {
		return zero, false
	}
	typedVal, err := sm.convertValue(val)
	if err != nil {
		return zero, false
	}
	return typedVal, true
}

func (sm *SyncMap[V]) Store(key string, value V) error {
	fullKey := sm.makeFullKey(key)
	return sm.config.Set(context.Background(), fullKey, value)
}

func (sm *SyncMap[V]) LoadOrStore(key string, newFunc func() (V, error)) (V, bool, error) {
	var zero V
	fullKey := sm.makeFullKey(key)
	val, err := sm.config.Get(context.Background(), fullKey)
	if err == nil {
		typedVal, err := sm.convertValue(val)
		if err != nil {
			return zero, false, err
		}
		return typedVal, true, nil
	}

	// If not found, call newFunc to create a new value
	newVal, err := newFunc()
	if err != nil {
		return zero, false, err
	}

	// Store the new value
	if err := sm.Store(key, newVal); err != nil {
		return zero, false, err
	}

	return newVal, false, nil
}

// deletes a key from the SyncMap
func (sm *SyncMap[V]) Delete(ctx context.Context, key string) error {
	fullKey := sm.makeFullKey(key)
	return sm.config.Delete(ctx, fullKey)
}

// returns true if the key exists in the SyncMap
func (sm *SyncMap[V]) Has(ctx context.Context, key string) bool {
	_, err := sm.Get(ctx, key)
	return err == nil
}

// returns all keys in the SyncMap
func (sm *SyncMap[V]) Keys(ctx context.Context) ([]string, error) {
	all, err := sm.config.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0)

	for fullKey := range all {
		if key, ok := sm.stripPrefix(fullKey); ok {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// returns all values in the SyncMap
func (sm *SyncMap[V]) Values(ctx context.Context) ([]V, error) {
	all, err := sm.config.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	values := make([]V, 0)

	for fullKey, val := range all {
		if _, ok := sm.stripPrefix(fullKey); ok {
			typedVal, err := sm.convertValue(val)
			if err == nil {
				values = append(values, typedVal)
			}
		}
	}

	return values, nil
}

// returns all key-value pairs in the SyncMap
func (sm *SyncMap[V]) All(ctx context.Context) (map[string]V, error) {
	all, err := sm.config.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]V)

	for fullKey, val := range all {
		if key, ok := sm.stripPrefix(fullKey); ok {
			typedVal, err := sm.convertValue(val)
			if err == nil {
				result[key] = typedVal
			}
		}
	}

	return result, nil
}

// iterates over all key-value pairs in the SyncMap
func (sm *SyncMap[V]) Range(ctx context.Context, fn func(key string, value V) bool) error {
	all, err := sm.All(ctx)
	if err != nil {
		return err
	}

	for key, val := range all {
		if !fn(key, val) {
			break
		}
	}

	return nil
}

// returns the number of key-value pairs in the SyncMap
func (sm *SyncMap[V]) Len(ctx context.Context) int {
	all, err := sm.config.GetAll(ctx)
	if err != nil {
		return 0
	}

	count := 0
	for fullKey := range all {
		if _, ok := sm.stripPrefix(fullKey); ok {
			count++
		}
	}

	return count
}

// registers a callback for changes to any key in the SyncMap
func (sm *SyncMap[V]) Subscribe(callback func(key string, value V)) string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := uuid.New().String()
	sm.subs[id] = callback
	return id
}

// removes a previously registered callback by subscription ID
func (sm *SyncMap[V]) Unsubscribe(subscriptionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.subs, subscriptionID)
}

// clears all key-value pairs in the SyncMap
func (sm *SyncMap[V]) Clear(ctx context.Context) error {
	keys, err := sm.Keys(ctx)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := sm.Delete(ctx, key); err != nil {
			return err
		}
	}

	return nil
}
