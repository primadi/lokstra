package old_registry

import (
	"sync"
	"testing"
)

// Simple string registry for benchmarking (simulates router registry behavior)
// Approach 1: Current implementation (RWMutex + map)
type RegistryRWMutex struct {
	registry map[string]string
	mu       sync.RWMutex
}

func (r *RegistryRWMutex) Register(name string, value string) {
	r.mu.Lock()
	r.registry[name] = value
	r.mu.Unlock()
}

func (r *RegistryRWMutex) Get(name string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.registry[name]
}

func (r *RegistryRWMutex) GetAll() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	copy := make(map[string]string, len(r.registry))
	for k, v := range r.registry {
		copy[k] = v
	}
	return copy
}

// Approach 2: sync.Map
type RegistrySyncMap struct {
	registry sync.Map
}

func (r *RegistrySyncMap) Register(name string, value string) {
	r.registry.Store(name, value)
}

func (r *RegistrySyncMap) Get(name string) string {
	if v, ok := r.registry.Load(name); ok {
		return v.(string)
	}
	return ""
}

func (r *RegistrySyncMap) GetAll() map[string]string {
	copy := make(map[string]string)
	r.registry.Range(func(key, value any) bool {
		copy[key.(string)] = value.(string)
		return true
	})
	return copy
}

// Setup functions
func setupRWMutex() *RegistryRWMutex {
	reg := &RegistryRWMutex{
		registry: make(map[string]string),
	}
	// Initialize with 10 items (simulate your use case)
	for i := 0; i < 10; i++ {
		reg.Register("router-"+string(rune('0'+i)), "value-"+string(rune('0'+i)))
	}
	return reg
}

func setupSyncMap() *RegistrySyncMap {
	reg := &RegistrySyncMap{}
	// Initialize with 10 items
	for i := 0; i < 10; i++ {
		reg.Register("router-"+string(rune('0'+i)), "value-"+string(rune('0'+i)))
	}
	return reg
}

// Benchmark: Single read (sequential)
func BenchmarkRegistry_RWMutex_SingleRead(b *testing.B) {
	reg := setupRWMutex()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = reg.Get("router-5")
	}
}

func BenchmarkRegistry_SyncMap_SingleRead(b *testing.B) {
	reg := setupSyncMap()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = reg.Get("router-5")
	}
}

// Benchmark: Concurrent reads (MOST REALISTIC for your case)
func BenchmarkRegistry_RWMutex_ConcurrentRead(b *testing.B) {
	reg := setupRWMutex()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = reg.Get("router-5")
		}
	})
}

func BenchmarkRegistry_SyncMap_ConcurrentRead(b *testing.B) {
	reg := setupSyncMap()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = reg.Get("router-5")
		}
	})
}

// Benchmark: Write once, read many (YOUR EXACT USE CASE)
func BenchmarkRegistry_RWMutex_WriteOnceReadMany(b *testing.B) {
	reg := &RegistryRWMutex{
		registry: make(map[string]string),
	}

	// Write once (startup)
	reg.Register("main-router", "main-value")

	b.ResetTimer()
	// Read many times (every request)
	for i := 0; i < b.N; i++ {
		_ = reg.Get("main-router")
	}
}

func BenchmarkRegistry_SyncMap_WriteOnceReadMany(b *testing.B) {
	reg := &RegistrySyncMap{}

	// Write once (startup)
	reg.Register("main-router", "main-value")

	b.ResetTimer()
	// Read many times (every request)
	for i := 0; i < b.N; i++ {
		_ = reg.Get("main-router")
	}
}

// Benchmark: GetAll (returns complete registry copy)
func BenchmarkRegistry_RWMutex_GetAll(b *testing.B) {
	reg := setupRWMutex()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = reg.GetAll()
	}
}

func BenchmarkRegistry_SyncMap_GetAll(b *testing.B) {
	reg := setupSyncMap()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = reg.GetAll()
	}
}
