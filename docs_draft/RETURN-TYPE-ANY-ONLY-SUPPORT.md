# Return Type `any` Only Support

## Overview

Framework Lokstra sekarang mendukung handler yang return `any` saja (tanpa error), selain pattern standar `(any, error)`. Ini memberikan fleksibilitas lebih untuk kasus-kasus sederhana di mana error handling tidak diperlukan.

## Supported Return Types

### 1. Standard Pattern (dengan error)
```go
// Data return
func(c *Context) (any, error)
func(c *Context) (map[string]any, error)

// Response return
func(c *Context) (*Response, error)
func(c *Context) (Response, error)

// ApiHelper return
func(c *Context) (*ApiHelper, error)
func(c *Context) (ApiHelper, error)

// Error only
func(c *Context) error
```

### 2. Simple Pattern (tanpa error) ✨ NEW
```go
// Data return only
func(c *Context) any
func(c *Context) map[string]any

// Response return only
func(c *Context) *Response
func(c *Context) Response

// ApiHelper return only
func(c *Context) *ApiHelper
func(c *Context) ApiHelper
```

## Use Cases

### ✅ Kapan Menggunakan Return `any` Saja

1. **Static/Mock Data**
   ```go
   func (s *MockService) GetStatus(c *statusParams) map[string]any {
       return map[string]any{
           "status":  "ok",
           "version": "1.0.0",
       }
   }
   ```

2. **Simple Getters**
   ```go
   func (s *Service) GetConfig(c *configParams) *Response {
       resp := response.NewResponse()
       resp.Json(s.config)
       return resp
   }
   ```

3. **Guaranteed Success Operations**
   ```go
   func (s *Service) GetMetrics(c *metricsParams) *ApiHelper {
       api := response.NewApiHelper()
       api.Ok(s.collectMetrics())
       return api
   }
   ```

4. **Testing/Demo Handlers**
   ```go
   func (s *DemoService) GetProducts(c *getProductsParams) any {
       return []map[string]any{
           {"id": 1, "name": "Product A"},
           {"id": 2, "name": "Product B"},
       }
   }
   ```

### ⚠️ Kapan Menggunakan Return `(any, error)`

1. **Production Code** - Explicit error handling lebih baik
2. **Database Operations** - Bisa fail
3. **External API Calls** - Network errors
4. **Validation Logic** - User input errors
5. **Business Logic** - Complex operations

## Examples

### Example 1: Simple Data Return
```go
type UserService struct{}

type getUsersParams struct{}

func (s *UserService) GetUsers(c *getUsersParams) map[string]any {
    return map[string]any{
        "users": []map[string]any{
            {"id": 1, "name": "Alice"},
            {"id": 2, "name": "Bob"},
        },
    }
}

// Route registration
r.GET("/users", userService.GetUsers)
```

### Example 2: Response with Custom Status
```go
func (s *Service) GetStatus(c *statusParams) *response.Response {
    resp := response.NewResponse()
    resp.WithStatus(http.StatusTeapot).Json(map[string]string{
        "message": "I'm a teapot",
    })
    return resp
}
```

### Example 3: ApiHelper with Custom Message
```go
func (s *Service) CreateResource(c *createParams) *response.ApiHelper {
    api := response.NewApiHelper()
    api.Created(c.Data, "Resource created successfully")
    return api
}
```

### Example 4: Nil Return (Default Success)
```go
func (s *Service) Ping(c *pingParams) *response.Response {
    // Returning nil triggers default success response
    return nil
}
// Equivalent to: return ctx.Api.Ok(nil)
```

## Priority Rules

Return value **SELALU** override `c.Resp` atau `c.Api`:

```go
func Handler(c *Context) *Response {
    // This is IGNORED
    c.Resp.WithStatus(200).Json(map[string]any{"ignored": true})
    
    // This is USED
    resp := response.NewResponse()
    resp.WithStatus(201).Json(map[string]any{"used": true})
    return resp
}
```

## Implementation Details

### Detection Logic
Framework mendeteksi return type dengan reflection:
1. Check jumlah return values (1 atau 2)
2. Jika 2 return values → must be `(any, error)`
3. Jika 1 return value:
   - Check apakah implements `error` interface
   - Jika **tidak** → treat as data/Response/ApiHelper return
   - Jika **ya** → treat as error-only return

### Handling Logic
```go
// Single return value (not error)
if numOut == 1 && !isError {
    // Check type
    if *Response or Response → use directly
    if *ApiHelper or ApiHelper → extract and use
    if any other → wrap with Api.Ok()
}
```

## Testing

Comprehensive test coverage added:
- ✅ `helper_response_test.go`: 8 new test cases
- ✅ `helper_priority_test.go`: 5 new test cases

### Test Coverage
1. Data return only
2. *Response return only
3. Response value return only
4. *ApiHelper return only
5. ApiHelper value return only
6. Struct param with data return only
7. No context with Response return only
8. Nil *Response return
9. Priority: return overrides c.Resp
10. Priority: return overrides c.Api
11. Priority: data return overrides c.Api
12. Priority: nil return sends default success
13. Priority: return overrides WriterFunc

## Performance

**Zero overhead** - Pattern detection happens once during route registration, not per request.

## Backward Compatibility

✅ **100% backward compatible** - Existing handlers with `(any, error)` pattern tetap bekerja normal.

## Best Practices

### ✅ DO
```go
// Simple handlers tanpa side effects
func GetVersion() string {
    return "1.0.0"
}

// Handlers yang guaranteed success
func GetStatus() *Response {
    resp := response.NewResponse()
    resp.Json(map[string]any{"status": "ok"})
    return resp
}
```

### ❌ DON'T
```go
// Handler dengan database operation
func GetUser(id int) User {
    user := db.Find(id) // Bisa error!
    return user
}
// ❌ BAD: Error tidak ter-handle, akan panic

// Seharusnya:
func GetUser(id int) (User, error) {
    user, err := db.Find(id)
    if err != nil {
        return User{}, err
    }
    return user, nil
}
// ✅ GOOD: Error ter-handle dengan baik
```

## Migration Guide

### From Error Return to Data-Only Return

**Before:**
```go
func GetConfig(c *Context) (map[string]any, error) {
    return map[string]any{"config": "value"}, nil
}
```

**After:**
```go
func GetConfig(c *Context) map[string]any {
    return map[string]any{"config": "value"}
}
```

### From Error Return to Response-Only Return

**Before:**
```go
func GetStatus(c *Context) (*Response, error) {
    resp := response.NewResponse()
    resp.Json(map[string]any{"status": "ok"})
    return resp, nil
}
```

**After:**
```go
func GetStatus(c *Context) *Response {
    resp := response.NewResponse()
    resp.Json(map[string]any{"status": "ok"})
    return resp
}
```

## Summary

| Pattern | Use Case | Error Handling |
|---------|----------|----------------|
| `func() (any, error)` | Production code | ✅ Explicit |
| `func() any` | Simple/Mock/Demo | ❌ Via panic |
| `func() (*Response, error)` | Production with full control | ✅ Explicit |
| `func() *Response` | Simple with full control | ❌ Via panic |
| `func() (*ApiHelper, error)` | Production with API format | ✅ Explicit |
| `func() *ApiHelper` | Simple with API format | ❌ Via panic |

**Rekomendasi**: Gunakan `(any, error)` untuk production code, gunakan `any` saja untuk mock/testing/demo handlers.
