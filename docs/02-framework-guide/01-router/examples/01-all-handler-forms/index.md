# All 69 Handler Forms

> **Complete reference of all handler signatures supported by Lokstra**

This example demonstrates **all 69 handler forms** supported by Lokstra, systematically organized by input and output variations.

## Formula

**6 Input Variations × 11 Output Variations = 66 combinations**  
**+ 3 special forms = 69 total handler forms**

---

## Input Variations (6)

1. **`()`** - No parameters
2. **`(*request.Context)`** - Context only (access to request details)
3. **`(*request.Context, *Param)`** - Context + pointer to param struct
4. **`(*request.Context, Param)`** - Context + value param struct
5. **`(*Param)`** - Pointer param only (no context)
6. **`(Param)`** - Value param only (no context)

## Output Variations (11)

1. **`error`** - Error only (nil = 200 OK)
2. **`any`** - Any type (map, struct, slice, primitive) - auto JSON
3. **`(any, error)`** - Data + error
4. **`*response.Response`** - Unopinionated helper (full control)
5. **`response.Response`** - Value form of Response
6. **`*response.ApiHelper`** - Opinionated JSON API (structured format)
7. **`response.ApiHelper`** - Value form of ApiHelper
8. **`(*response.Response, error)`** - Response + error
9. **`(response.Response, error)`** - Response value + error
10. **`(*response.ApiHelper, error)`** - ApiHelper + error
11. **`(response.ApiHelper, error)`** - ApiHelper value + error

---

## Complete Handler List

### GROUP 1: No Input - `func()` (01-11)
```go
func() error                              // 01
func() any                                // 02
func() (any, error)                       // 03
func() *response.Response                 // 04
func() response.Response                  // 05
func() *response.ApiHelper                // 06
func() response.ApiHelper                 // 07
func() (*response.Response, error)        // 08
func() (response.Response, error)         // 09
func() (*response.ApiHelper, error)       // 10
func() (response.ApiHelper, error)        // 11
```

### GROUP 2: Context Only - `func(*request.Context)` (12-22)
```go
func(*request.Context) error                              // 12
func(*request.Context) any                                // 13
func(*request.Context) (any, error)                       // 14
func(*request.Context) *response.Response                 // 15
func(*request.Context) response.Response                  // 16
func(*request.Context) *response.ApiHelper                // 17
func(*request.Context) response.ApiHelper                 // 18
func(*request.Context) (*response.Response, error)        // 19
func(*request.Context) (response.Response, error)         // 20
func(*request.Context) (*response.ApiHelper, error)       // 21
func(*request.Context) (response.ApiHelper, error)        // 22
```

### GROUP 3: Context + *Param - `func(*request.Context, *Param)` (23-33)
```go
func(*request.Context, *Param) error                              // 23
func(*request.Context, *Param) any                                // 24
func(*request.Context, *Param) (any, error)                       // 25
func(*request.Context, *Param) *response.Response                 // 26
func(*request.Context, *Param) response.Response                  // 27
func(*request.Context, *Param) *response.ApiHelper                // 28
func(*request.Context, *Param) response.ApiHelper                 // 29
func(*request.Context, *Param) (*response.Response, error)        // 30
func(*request.Context, *Param) (response.Response, error)         // 31
func(*request.Context, *Param) (*response.ApiHelper, error)       // 32
func(*request.Context, *Param) (response.ApiHelper, error)        // 33
```

### GROUP 4: Context + Param - `func(*request.Context, Param)` (34-44)
```go
func(*request.Context, Param) error                              // 34
func(*request.Context, Param) any                                // 35
func(*request.Context, Param) (any, error)                       // 36 ⭐ MOST COMMON
func(*request.Context, Param) *response.Response                 // 37
func(*request.Context, Param) response.Response                  // 38
func(*request.Context, Param) *response.ApiHelper                // 39
func(*request.Context, Param) response.ApiHelper                 // 40
func(*request.Context, Param) (*response.Response, error)        // 41
func(*request.Context, Param) (response.Response, error)         // 42
func(*request.Context, Param) (*response.ApiHelper, error)       // 43
func(*request.Context, Param) (response.ApiHelper, error)        // 44
```

### GROUP 5: *Param Only - `func(*Param)` (45-55)
```go
func(*Param) error                              // 45
func(*Param) any                                // 46
func(*Param) (any, error)                       // 47
func(*Param) *response.Response                 // 48
func(*Param) response.Response                  // 49
func(*Param) *response.ApiHelper                // 50
func(*Param) response.ApiHelper                 // 51
func(*Param) (*response.Response, error)        // 52
func(*Param) (response.Response, error)         // 53
func(*Param) (*response.ApiHelper, error)       // 54
func(*Param) (response.ApiHelper, error)        // 55
```

### GROUP 6: Param Only - `func(Param)` (56-66)
```go
func(Param) error                              // 56
func(Param) any                                // 57
func(Param) (any, error)                       // 58
func(Param) *response.Response                 // 59
func(Param) response.Response                  // 60
func(Param) *response.ApiHelper                // 61
func(Param) response.ApiHelper                 // 62
func(Param) (*response.Response, error)        // 63
func(Param) (response.Response, error)         // 64
func(Param) (*response.ApiHelper, error)       // 65
func(Param) (response.ApiHelper, error)        // 66
```

### SPECIAL FORMS (+3)
```go
http.Handler                                   // 67
http.HandlerFunc                               // 68
request.HandlerFunc                            // 69 (alias for func(*request.Context) error)
```

---

## Response Helper Guide

### `any` - Auto JSON Serialization
```go
// Return any type - Lokstra converts to JSON automatically
func Handler() map[string]string {
    return map[string]string{"message": "success"}
}
// → {"message": "success"}
```

### `*response.Response` - Unopinionated (Full Control)
```go
func Handler() *response.Response {
    r := response.NewResponse()
    r.RespStatusCode = 200
    r.RespData = map[string]string{"message": "success"}
    return r
}
// → {"message": "success"}
```

### `*response.ApiHelper` - Opinionated (Structured Format)
```go
func Handler() *response.ApiHelper {
    api := response.NewApiHelper()
    api.Ok(map[string]string{"name": "John"})
    return api
}
// → {"success": true, "data": {"name": "John"}}
```

---

## Selection Guide

| Use Case | Recommended Form | Handler # |
|----------|------------------|-----------|
| Health check | `func() any` | 02 |
| Simple with error | `func() (any, error)` | 03 |
| Need request details | `func(ctx) (any, error)` | 14 |
| **REST API standard** | `func(ctx, Param) (any, error)` | **36** ⭐ |
| Full error control | `func(ctx, Param) error` | 34 |
| Structured API response | `func(ctx, Param) *response.ApiHelper` | 39 |
| Legacy compatibility | `http.HandlerFunc` | 68 |

---

## Parameter Binding Example

```go
type Param struct {
    ID   int    `path:"id"`         // From URL path
    Name string `query:"name"`      // From query string
    Key  string `header:"X-API-Key"` // From header
    Data string `json:"data"`       // From JSON body
}

// All 4 parameter sources bound automatically
func Handler36(ctx *request.Context, p Param) (any, error) {
    return map[string]any{
        "id":   p.ID,
        "name": p.Name,
        "key":  p.Key,
        "data": p.Data,
    }, nil
}
```

Route: `GET /h36/:id?name=test` with header `X-API-Key: abc123`

---

## Pointer vs Value Parameters

### When to use `*Param` (pointer):
- Large structs (> 100 bytes)
- Need to distinguish nil vs zero value
- Want to modify param inside handler

### When to use `Param` (value):
- Small structs (< 100 bytes)
- Read-only access
- Immutable patterns preferred

**Technical note**: Both are valid and work identically in Lokstra. Difference is purely in Go semantics.

---

## Running the Example

```bash
# Start server
cd docs/02-deep-dive/01-router/examples/01-all-handler-forms
go run main.go

# Server starts on :8080
# Test using test.http file (VS Code REST Client extension)
```

### Test Routes
- **Group 1** (no params): `/h01` - `/h11`
- **Group 2** (context): `/h12` - `/h22`
- **Group 3-6** (with params): `/h23/:id` - `/h66/:id`
- **Special**: `/h67`, `/h68`, `/h69`

**Example test:**
```http
GET http://localhost:8080/h36/123?name=Alice
X-API-Key: test-key
```

---

## Performance Considerations

| Pattern | Overhead | When to Use |
|---------|----------|-------------|
| `func()` | None | Health checks, static responses |
| `func(ctx)` | ~50ns | Need request metadata |
| `func(Param)` | ~100ns | Need parameter binding |
| `func(ctx, Param)` | ~150ns | Full REST API (worth it) |

**Recommendation**: Use the form that matches your needs. The overhead is negligible for real API calls (ms scale).

---

## Key Takeaways

✅ **69 total handler forms**: 66 systematic + 3 special  
✅ **Formula**: 6 inputs × 11 outputs = complete coverage  
✅ **Most common**: Handler #36 `func(ctx, Param) (any, error)` for REST APIs  
✅ **Type-safe**: All forms compile-time checked  
✅ **Flexible**: Choose simplest form that meets requirements  
✅ **Compatible**: Supports standard `http.Handler` and `http.HandlerFunc`

---

## See Also

- [Parameter Binding Example](../02-parameter-binding/) - Deep dive into parameter types
- [Lifecycle Hooks](../03-lifecycle-hooks/) - Middleware before/after patterns
- [Response Helpers Guide](../../../00-introduction/03-response-types/) - Response.Response vs ApiHelper
