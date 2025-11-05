# LazyLoad vs GetService - Performance Comparison

Demonstrates the performance difference between three methods of accessing services in Lokstra.

## What You'll Learn

- Understand registry lookup overhead
- See the performance benefit of LazyLoad
- Learn when to use each method
- Benchmark real-world performance

## Three Methods Compared

### Method 1: `GetService()` 
```go
userService := lokstra_registry.GetService[*UserService]("user-service")
if userService == nil {
    return error
}
```
- ❌ SLOW: ~100-200ns per call (map lookup every request)
- Returns nil if not found
- Good for: Optional services, dynamic names

### Method 2: `MustGetService()`
```go
userService := lokstra_registry.MustGetService[*UserService]("user-service")
```
- ❌ SLOW: ~100-200ns per call (map lookup every request)
- Panics if not found (clear error message)
- Good for: Development, fail-fast behavior

### Method 3: `LazyLoad()` ⭐ RECOMMENDED
```go
// Package-level (cached!)
var userService = service.LazyLoad[*UserService]("user-service")

// In handler
users := userService.MustGet().GetUsers()
```
- ✅ FAST: ~1-5ns per call (cached after first access)
- **20-100x faster than GetService!**
- Panics with clear error if not found
- Good for: Production code, high-traffic endpoints

## Running

```bash
cd docs/01-router-guide/02-service/examples/02-lazyload-vs-getservice
go run main.go
```

## Testing

Use `test.http` or curl:

**Test each method individually:**
```bash
# Method 1 (slow)
curl http://localhost:3000/method1-getservice

# Method 2 (slow)
curl http://localhost:3000/method2-mustgetservice

# Method 3 (fast) ⭐
curl http://localhost:3000/method3-lazyload
```

**Run benchmark (1000 iterations):**
```bash
curl http://localhost:3000/benchmark
```

Expected output:
```json
{
  "iterations": 1000,
  "results": {
    "GetService": {
      "avg_ns": 150,
      "note": "Map lookup every call"
    },
    "MustGetService": {
      "avg_ns": 160,
      "note": "Map lookup + panic check every call"
    },
    "LazyLoad": {
      "avg_ns": 3,
      "note": "Cached after first access"
    }
  },
  "comparison": {
    "speedup": "50.0x faster",
    "winner": "LazyLoad",
    "recommendation": "Use LazyLoad for production code!"
  }
}
```

**Check stats:**
```bash
curl http://localhost:3000/stats
```

## Key Takeaways

1. **LazyLoad is 20-100x faster** than GetService/MustGetService
2. **Use LazyLoad for production** - especially high-traffic endpoints
3. **Package-level declaration** - LazyLoad must be package-level or struct field
4. **Clear errors with MustGet()** - better than nil pointer panics
5. **GetService still useful** - for dynamic service names or optional services

## Performance Impact

On a high-traffic API (10,000 req/sec):
- **GetService**: ~1-2ms overhead per second
- **LazyLoad**: ~0.01-0.05ms overhead per second
- **Savings**: ~2ms/sec = more capacity for business logic!

## Code Pattern

```go
package handlers

import (
    "github.com/primadi/lokstra/core/service"
)

// ✅ RECOMMENDED: Package-level LazyLoad
var (
    userService    = service.LazyLoad[*UserService]("users")
    orderService   = service.LazyLoad[*OrderService]("orders")
    paymentService = service.LazyLoad[*PaymentService]("payments")
)

func GetUsersHandler() ([]User, error) {
    // Fast! Only 1-5ns overhead
    return userService.MustGet().GetAll()
}
```

```go
// ❌ NOT RECOMMENDED: Function-level (won't cache!)
func GetUsersHandler() ([]User, error) {
    // This defeats the purpose - still slow!
    userService := service.LazyLoad[*UserService]("users")
    return userService.MustGet().GetAll()
}
```

## When to Use Each Method

| Method | Use Case | Performance | Error Handling |
|--------|----------|-------------|----------------|
| `GetService()` | Optional services, dynamic names | Slow | Returns nil |
| `MustGetService()` | Development, debugging | Slow | Panics with clear message |
| `LazyLoad()` | Production, high-traffic | **Fast** ⭐ | Panics with clear message |
