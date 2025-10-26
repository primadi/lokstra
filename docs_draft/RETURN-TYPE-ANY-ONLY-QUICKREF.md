# Quick Reference - Return Type `any` Only

## TL;DR

✅ Handler sekarang bisa return `any` **TANPA** `error`
✅ Support: `*Response`, `*ApiHelper`, atau data biasa
✅ 100% backward compatible dengan pattern `(any, error)`

## Comparison

### Before (dengan error)
```go
func GetUsers(c *Context) (map[string]any, error) {
    return map[string]any{"users": []...}, nil
}
```

### After (tanpa error) ✨
```go
func GetUsers(c *Context) map[string]any {
    return map[string]any{"users": []...}
}
```

## Quick Patterns

### 1️⃣ Data Return Only
```go
func Handler(c *Context) map[string]any {
    return map[string]any{"status": "ok"}
}
```

### 2️⃣ *Response Return Only
```go
func Handler(c *Context) *response.Response {
    resp := response.NewResponse()
    resp.WithStatus(201).Json(data)
    return resp
}
```

### 3️⃣ *ApiHelper Return Only
```go
func Handler(c *Context) *response.ApiHelper {
    api := response.NewApiHelper()
    api.Created(data, "Success")
    return api
}
```

### 4️⃣ With Struct Param
```go
type params struct {
    ID int `path:"id"`
}

func Handler(p *params) map[string]any {
    return map[string]any{"id": p.ID}
}
```

### 5️⃣ No Context
```go
func Handler() *response.Response {
    resp := response.NewResponse()
    resp.Json(map[string]any{"static": "data"})
    return resp
}
```

## When to Use

### ✅ Use Return `any` Only
- Static/mock data
- Guaranteed success operations
- Simple getters
- Testing/demo handlers

### ⚠️ Use Return `(any, error)`
- Production code
- Database operations
- External API calls
- Validation logic
- Any operation that can fail

## Priority Rules

Return value **ALWAYS** overrides `c.Resp` or `c.Api`:

```go
func Handler(c *Context) *Response {
    c.Resp.Json(data1)  // ❌ IGNORED
    
    resp := response.NewResponse()
    resp.Json(data2)     // ✅ USED
    return resp
}
```

## All Supported Signatures

### With Context
```go
func(*Context) any
func(*Context) (any, error)
func(*Context) *Response
func(*Context) (*Response, error)
func(*Context) *ApiHelper
func(*Context) (*ApiHelper, error)
func(*Context) error
```

### With Struct Param
```go
func(*Struct) any
func(*Struct) (any, error)
func(*Struct) *Response
func(*Struct) (*Response, error)
func(*Struct) *ApiHelper
func(*Struct) (*ApiHelper, error)
func(*Struct) error
```

### With Both
```go
func(*Context, *Struct) any
func(*Context, *Struct) (any, error)
func(*Context, *Struct) *Response
func(*Context, *Struct) (*Response, error)
func(*Context, *Struct) *ApiHelper
func(*Context, *Struct) (*ApiHelper, error)
func(*Context, *Struct) error
```

### No Params
```go
func() any
func() (any, error)
func() *Response
func() (*Response, error)
func() *ApiHelper
func() (*ApiHelper, error)
func() error
```

## Examples from Services.go

### UserService (Updated) ✨
```go
type getUsersParams struct{}

func (s *UserService) GetUsers(c *getUsersParams) (map[string]any, error) {
    return map[string]any{
        "users": []map[string]any{...},
        "source": "App1",
    }, nil
}
```

### ProductService (Updated) ✨
```go
type getProductsParams struct{}

// ✨ NEW: Return data only (no error)
func (s *ProductService) GetProducts(c *getProductsParams) map[string]any {
    return map[string]any{
        "products": []map[string]any{...},
        "source": "App2",
    }
}
```

## Error Handling

### With Error Return
```go
func Handler(c *Context) (any, error) {
    data, err := db.Query()
    if err != nil {
        return nil, err  // ✅ Explicit error handling
    }
    return data, nil
}
```

### Without Error Return (Panic on Error)
```go
func Handler(c *Context) any {
    data := getStaticData()  // Must not error
    return data
    // If error occurs, will panic ⚠️
}
```

## Testing

All tests PASSED ✅
- 29 response handler tests
- 13 priority tests
- 8 integration tests
- 9 edge case tests

See: `docs_draft/TEST-RESULTS-RETURN-TYPE-ANY-ONLY.md`

## Documentation

Full docs: `docs_draft/RETURN-TYPE-ANY-ONLY-SUPPORT.md`

## Files Changed

### Core
- ✅ `core/router/helper.go` - adaptSmart updated

### Tests
- ✅ `core/router/helper_response_test.go` - 8 new tests
- ✅ `core/router/helper_priority_test.go` - 5 new tests

### Examples
- ✅ `cmd_draft/examples/return-type-any-only/main.go` - demo server

### Docs
- ✅ `docs_draft/RETURN-TYPE-ANY-ONLY-SUPPORT.md`
- ✅ `docs_draft/TEST-RESULTS-RETURN-TYPE-ANY-ONLY.md`
- ✅ `docs_draft/RETURN-TYPE-ANY-ONLY-QUICKREF.md` (this file)
