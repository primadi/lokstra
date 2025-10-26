# Fast Path Optimization - Performance Benchmark Results

## Test Environment
- **CPU:** AMD Ryzen 9 5900HX with Radeon Graphics
- **OS:** Windows
- **Go Version:** 1.23+
- **Date:** October 14, 2025

---

## 📊 Benchmark Results

### Handler Performance (with response generation)

| Pattern | ns/op | B/op | allocs/op | Tier |
|---------|-------|------|-----------|------|
| `func(*Context) error` | **1,626** | 2,063 | 27 | Tier 0 (Zero-cost) |
| `func(*Context) (any, error)` | **1,690** | 2,063 | 27 | Tier 1 (Fast path) ✨ |
| `func(*Context) (*Response, error)` | **1,451** | 2,039 | 26 | Tier 1 (Fast path) ✨ |
| `func(*Context) (*ApiHelper, error)` | **1,694** | 2,151 | 29 | Tier 1 (Fast path) ✨ |
| `func(*Context) any` | **1,645** | 2,063 | 27 | Tier 1 (Fast path) ✨ |
| `func() (any, error)` | **1,651** | 2,063 | 27 | Tier 1 (Fast path) ✨ |
| `func(*Struct) (any, error)` | **2,600** | 2,732 | 34 | Tier 2 (Reflection) |
| `http.HandlerFunc` | **434** | 552 | 11 | Tier 1 (HTTP compat) |

**Key Findings:**
- ✅ Fast path patterns: **1,450-1,700 ns/op**
- ⚠️ Reflection pattern: **2,600 ns/op** (~1.5x slower)
- 🚀 Standard HTTP handler: **434 ns/op** (fastest, no API wrapper)

---

### Pure Overhead Comparison (minimal handlers)

| Tier | Pattern | ns/op | B/op | allocs/op | Overhead |
|------|---------|-------|------|-----------|----------|
| **Tier 0** | `func(*Context) error` | **349** | 472 | 9 | **0ns** (baseline) |
| **Tier 1** | `func(*Context) (any, error)` | **1,028** | 1,401 | 18 | **+679ns** |
| **Tier 2** | `func(*Struct) (any, error)` | **1,661** | 2,041 | 23 | **+1,312ns** |

**Key Findings:**
- ✅ Tier 0 (direct): **349 ns** - Pure router overhead
- ✅ Tier 1 (fast path): **+679 ns** - Wrapper + response formatting
- ⚠️ Tier 2 (reflection): **+1,312 ns** - Reflection + binding + formatting

**Performance Improvement:**
- **Tier 1 vs Tier 2:** ~1.6x faster (679ns vs 1,312ns overhead)
- **Saved per request:** ~633ns (~38% reduction)

---

## 🎯 Coverage Analysis

### Fast Path Coverage (Tier 0 & 1)

**Total: 21 patterns** (out of 29 total variants)

#### With *Context (11 patterns):
1. ✅ `func(*Context) error` - Tier 0
2. ✅ `func(*Context) (any, error)` - Tier 1
3. ✅ `func(*Context) (*Response, error)` - Tier 1
4. ✅ `func(*Context) (*ApiHelper, error)` - Tier 1
5. ✅ `func(*Context) any` - Tier 1
6. ✅ `func(*Context) *Response` - Tier 1
7. ✅ `func(*Context) *ApiHelper` - Tier 1
8. ✅ `func(*Context) (Response, error)` - Tier 1
9. ✅ `func(*Context) (ApiHelper, error)` - Tier 1
10. ✅ `func(*Context) Response` - Tier 1
11. ✅ `func(*Context) ApiHelper` - Tier 1

#### Without Context (7 patterns):
12. ✅ `func() (any, error)` - Tier 1
13. ✅ `func() (*Response, error)` - Tier 1
14. ✅ `func() (*ApiHelper, error)` - Tier 1
15. ✅ `func() any` - Tier 1
16. ✅ `func() *Response` - Tier 1
17. ✅ `func() *ApiHelper` - Tier 1
18. ✅ `func() error` - Tier 1

#### HTTP Compatibility (3 patterns):
19. ✅ `http.HandlerFunc` - Tier 1
20. ✅ `func(http.ResponseWriter, *http.Request)` - Tier 1
21. ✅ `http.Handler` - Tier 1

### Reflection Fallback (Tier 2)

**Total: 8 patterns** (complex signatures)

1. ⚠️ `func(*Context, *Struct) error`
2. ⚠️ `func(*Context, *Struct) (any, error)`
3. ⚠️ `func(*Context, *Struct) (*Response, error)`
4. ⚠️ `func(*Context, *Struct) *Response`
5. ⚠️ `func(*Struct) error`
6. ⚠️ `func(*Struct) (any, error)`
7. ⚠️ `func(*Struct) (*Response, error)`
8. ⚠️ `func(*Struct) *Response`

---

## 📈 Real-World Performance

### Scenario: High-traffic API (10,000 req/sec)

#### Before Optimization (all reflection):
```
10,000 req/sec × 1,312ns overhead = 13.12ms/sec = 1.31% CPU
```

#### After Optimization (90% fast path):
```
Fast path: 9,000 req/sec × 679ns = 6.11ms/sec
Reflection: 1,000 req/sec × 1,312ns = 1.31ms/sec
Total: 7.42ms/sec = 0.74% CPU
```

**Savings: 0.57% CPU** (43% reduction in handler overhead)

### At scale (100,000 req/sec):
**CPU savings: 5.7%** - Significant for high-traffic services!

---

## 🔍 Pattern Recommendations

### ⭐ Tier 0: Direct (0ns overhead)
**Use for:** Maximum performance, when you need API.Ok() wrapper

```go
func Handler(c *Context) error {
    return c.Api.Ok(data)
}
```

### ⭐⭐ Tier 1: Fast Path (~679ns overhead)
**Use for:** Production code, most common patterns

```go
// Recommended for REST APIs
func Handler(c *Context) (any, error) {
    data, err := service.GetData()
    return data, err
}

// Recommended for full control
func Handler(c *Context) (*Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(201).Json(data)
    return resp, err
}
```

### ⭐ Tier 2: Reflection (~1,312ns overhead)
**Use for:** Complex parameter binding

```go
type Params struct {
    ID   int    `path:"id"`
    Name string `query:"name"`
}

func Handler(p *Params) (any, error) {
    // Automatic binding from path & query
    return service.GetByID(p.ID, p.Name), nil
}
```

---

## 💡 Performance Tips

### 1. **Choose Fast Path when possible**
```go
// ✅ GOOD: Fast path
func GetUser(c *Context) (any, error) { }

// ⚠️ OK: Reflection (if you need binding)
func GetUser(p *GetUserParams) (any, error) { }
```

### 2. **Use Tier 0 for hot paths**
```go
// ✅ BEST: Zero-cost for critical endpoints
func HealthCheck(c *Context) error {
    return c.Api.Ok(map[string]string{"status": "ok"})
}
```

### 3. **Batch parameter extraction**
```go
// ⚠️ SLOW: Multiple reflections
func Handler1(p1 *Params1) error { }
func Handler2(p2 *Params2) error { }

// ✅ FAST: Single struct with all params
type AllParams struct {
    ID   int    `path:"id"`
    Name string `query:"name"`
    Data MyData `json:"*"`
}
func Handler(p *AllParams) error { }
```

---

## 📊 Summary

| Metric | Value |
|--------|-------|
| **Fast path patterns** | 21 / 29 (72%) |
| **Performance improvement** | 1.6x faster than reflection |
| **CPU overhead reduction** | 38% for common patterns |
| **Production ready** | ✅ Yes |

---

## ✅ Conclusion

Fast path optimization provides:
1. **Significant performance improvement** for common patterns
2. **Zero breaking changes** - fully backward compatible
3. **Type-safe** handling without reflection overhead
4. **Fallback** to reflection for complex cases

**Recommendation:** Use fast path patterns (`func(*Context) (any, error)`) for production code to get best performance.
