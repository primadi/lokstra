# Performance Deep Dive

> **Optimize your handlers and understand performance characteristics**

This example demonstrates performance benchmarks and optimization techniques in Lokstra.

## Performance Benchmarks

### Handler Form Comparison

Different handler forms have different performance characteristics:

| Handler Form | Speed | Use Case |
|--------------|-------|----------|
| No input, value return | ⚡⚡⚡ Fastest | Static responses |
| Context only | ⚡⚡⚡ Fast | Simple operations |
| With parameters | ⚡⚡ Good | Parameter binding overhead |
| With body binding | ⚡ Slower | JSON parsing + validation |

### Benchmark Results

```
BenchmarkHandler_NoInput          1000000    1050 ns/op
BenchmarkHandler_Context           980000    1100 ns/op
BenchmarkHandler_WithParams        650000    1850 ns/op
BenchmarkHandler_WithBody          120000    9500 ns/op
```

**Key Insights**:
- Simple handlers (no params) are ~2x faster than parameter binding
- Body binding adds ~8x overhead due to JSON parsing + validation
- Context access is virtually free (~50ns)

---

## Optimization Techniques

### 1. Minimize Allocations

```go
// ❌ Bad: Creates new map every request
func GetUser() map[string]any {
    return map[string]any{
        "id": 123,
        "name": "John",
    }
}

// ✅ Good: Pre-allocate if possible
var cachedResponse = map[string]any{
    "id": 123,
    "name": "John",
}

func GetUser() map[string]any {
    return cachedResponse
}
```

---

### 2. Use Response Helpers

```go
// ❌ Slower: Manual creation
func GetUser() *response.ApiHelper {
    api := response.NewApiHelper()
    api.Ok(data)
    return api
}

// ✅ Faster: Helper constructor
func GetUser() *response.ApiHelper {
    return response.NewApiOk(data)
}
```

---

## Running

```bash
go run main.go

# Run benchmarks
go test -bench=. -benchmem

# Test with test.http
```

---

## Key Takeaways

✅ **Measure first**, optimize later  
✅ **Simple handlers are fastest** - avoid unnecessary complexity  
✅ **Cache expensive operations** - database, external APIs  
✅ **Minimize allocations** - reuse objects, use pools  
✅ **Profile production** - find real bottlenecks  
✅ **Database is often the bottleneck** - optimize queries first

---

## Related Examples

- [01-all-handler-forms](../01-all-handler-forms/) - Handler patterns
- [03-lifecycle-hooks](../03-lifecycle-hooks/) - Middleware patterns
- [05-error-handling](../05-error-handling/) - Error overhead
