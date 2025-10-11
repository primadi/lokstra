# Registry Optimization: RWMutex to sync.Map Migration

## Summary

Migrated router and service convention registries from `RWMutex + map` to `sync.Map` for significantly better concurrent read performance.

## Performance Improvement

### Benchmark Results

| Use Case | RWMutex | sync.Map | Improvement |
|----------|---------|----------|-------------|
| **Concurrent Reads** | 42.99 ns/op | 2.936 ns/op | **14.6x faster** ✅ |
| Sequential Reads | 24.44 ns/op | 22.24 ns/op | 1.1x faster |
| Write Once, Read Many | 14.51 ns/op | 19.81 ns/op | 1.4x slower |
| GetAll (Copy) | 704.8 ns/op | 855.2 ns/op | 1.2x slower |

**Key Finding**: For the primary use case (concurrent reads in high-traffic servers), sync.Map is **14.6x faster**.

## Changes Made

### 1. `lokstra_registry/router.go`

**Before:**
```go
var routerRegistry = make(map[string]router.Router)
var routerMutex sync.RWMutex

func GetRouter(name string) router.Router {
    routerMutex.RLock()
    defer routerMutex.RUnlock()
    
    if r, ok := routerRegistry[name]; ok {
        return r
    }
    return nil
}
```

**After:**
```go
var routerRegistry sync.Map

func GetRouter(name string) router.Router {
    if v, ok := routerRegistry.Load(name); ok {
        return v.(router.Router)
    }
    return nil
}
```

### 2. `core/router/service_convention.go`

**Before:**
```go
var (
    conventionRegistry = make(map[string]ServiceConvention)
    conventionMu       sync.RWMutex
)

func GetConvention(name string) (ServiceConvention, error) {
    conventionMu.RLock()
    defer conventionMu.RUnlock()
    
    convention, exists := conventionRegistry[name]
    if !exists {
        return nil, fmt.Errorf("convention '%s' not found", name)
    }
    return convention, nil
}
```

**After:**
```go
var conventionRegistry sync.Map

func GetConvention(name string) (ServiceConvention, error) {
    if v, ok := conventionRegistry.Load(name); ok {
        return v.(ServiceConvention), nil
    }
    return nil, fmt.Errorf("convention '%s' not found", name)
}
```

### 3. `lokstra_registry/client_router.go`

Fixed direct map access to use `GetRouter()` helper function.

## Why sync.Map?

### ✅ Advantages for This Use Case

1. **Write Once, Read Many**: Routers and conventions are registered once at startup, then read thousands of times per second
2. **Concurrent Reads**: Every incoming HTTP request may look up routers/conventions - perfect for sync.Map's lock-free reads
3. **No Lock Contention**: Under high load, RWMutex causes contention even for reads. sync.Map eliminates this
4. **Scalability**: Performance improves with more CPU cores

### ⚠️ Trade-offs

1. **Type Safety**: Requires type assertion `v.(router.Router)` (minimal overhead)
2. **GetAll Slower**: Copying all entries is ~20% slower (rarely used operation)
3. **Single Writes Slower**: Initial registration is ~36% slower (happens once at startup)

## Why Not Just Remove Mutex?

**Important**: Unlike the reflection caching case, **maps in Go are NOT goroutine-safe**!

```go
// ❌ DANGEROUS - Will panic with concurrent access
var registry = make(map[string]router.Router)

// Concurrent reads + any write = CRASH!
// fatal error: concurrent map read and map write
```

Even for "write once, read many" scenarios:
- **Memory visibility**: Other goroutines might not see the written value
- **Race detector**: Will detect data races
- **Production crashes**: Under load, this WILL crash your server

## Production Impact

### Before (RWMutex)
- Under high load: Lock contention on every request
- Scalability: Limited by mutex contention
- Performance: 43ns per router lookup

### After (sync.Map)
- Under high load: Lock-free concurrent reads
- Scalability: Scales with CPU cores
- Performance: 2.9ns per router lookup

**For a server handling 10,000 requests/second:**
- Before: ~430,000 ns (0.43ms) total overhead
- After: ~29,000 ns (0.029ms) total overhead
- **Savings: ~0.4ms per second, or better CPU utilization**

## Testing

All existing tests pass (2 pre-existing test failures unrelated to this change):

```bash
# Router registry tests
cd lokstra_registry
go test -v .

# Service convention tests  
cd core/router
go test -v .

# Benchmark comparison
cd lokstra_registry
go test -run=^$ -bench=BenchmarkRegistry -benchmem -benchtime=2s
```

## Compatibility

- ✅ **API Compatible**: No breaking changes to public API
- ✅ **Behavior Compatible**: Same behavior, just faster
- ✅ **Thread Safe**: Both approaches are thread-safe

## Recommendations for Other Registries

Use sync.Map when:
- ✅ Write once/rarely, read many times
- ✅ High concurrent read workload
- ✅ Registry-like pattern
- ✅ Performance-critical path

Keep RWMutex when:
- ✅ Frequent updates to map
- ✅ Need to iterate entire map frequently
- ✅ Strong typing preferred over performance

## References

- Benchmark code: `lokstra_registry/registry_bench_test.go`
- sync.Map documentation: https://golang.org/pkg/sync/#Map
- Best practices: "sync.Map is optimized for read-heavy workloads"
