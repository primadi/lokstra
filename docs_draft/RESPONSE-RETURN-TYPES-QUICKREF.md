# Response Return Types - Quick Reference

## Overview
Handler sekarang bisa return `*response.Response` atau `*response.ApiHelper` untuk kontrol penuh atas response.

---

## Basic Usage

### 1. Regular Data Return (Existing)
```go
func GetUser(c *request.Context) (User, error) {
    return user, nil  // Auto-wrapped with Api.Ok()
}
```

### 2. Response Return (New)
```go
func CustomResponse(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(201).Json(data)
    return resp, nil  // Full control!
}
```

### 3. ApiHelper Return (New)
```go
func ApiResponse(c *request.Context) (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    api.Created(data, "Success")
    return api, nil  // API formatted!
}
```

---

## When to Use What

| Return Type | Use Case | Example |
|------------|----------|---------|
| `(data, error)` | Standard API responses | CRUD operations |
| `(*Response, error)` | Custom headers/content-type | File download, streaming |
| `(*ApiHelper, error)` | API format + custom headers | Paginated lists, errors |

---

## Key Features

### ✅ Error Priority
Error **always** takes precedence over Response:
```go
func Handler() (*response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(200).Json(data)  // IGNORED!
    return resp, errors.New("failed") // ERROR returned
}
```

### ✅ Nil Handling
Nil pointer returns default success:
```go
func Handler() (*response.Response, error) {
    return nil, nil  // Sends Api.Ok(nil)
}
```

### ✅ Full Control
```go
func Handler() (*response.Response, error) {
    resp := response.NewResponse()
    resp.RespHeaders = map[string][]string{
        "X-Custom": {"value"},
    }
    resp.WithStatus(201).Json(data)
    return resp, nil
}
```

---

## All Supported Signatures

```go
// With Context
func(*Context) error
func(*Context) (data, error)
func(*Context) (*Response, error)          // NEW
func(*Context) (*ApiHelper, error)         // NEW

// With Context + Struct
func(*Context, *Struct) error
func(*Context, *Struct) (data, error)
func(*Context, *Struct) (*Response, error) // NEW
func(*Context, *Struct) (*ApiHelper, error)// NEW

// Without Context
func(*Struct) error
func(*Struct) (data, error)
func(*Struct) (*Response, error)           // NEW
func(*Struct) (*ApiHelper, error)          // NEW

func() error
func() (data, error)
func() (*Response, error)                  // NEW
func() (*ApiHelper, error)                 // NEW
```

---

## Quick Examples

### Custom Status Code
```go
func(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(http.StatusCreated).Json(data)
    return resp, nil
}
```

### Plain Text
```go
func(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(200).Text("Hello World")
    return resp, nil
}
```

### HTML
```go
func(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(200).Html("<h1>Hello</h1>")
    return resp, nil
}
```

### Streaming
```go
func(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.Stream("text/event-stream", func(w http.ResponseWriter) error {
        fmt.Fprintf(w, "data: %s\n\n", "streaming")
        return nil
    })
    return resp, nil
}
```

### API Formatted with Pagination
```go
func(c *request.Context) (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    
    meta := &api_formatter.ListMeta{
        Page:     1,
        PageSize: 10,
        Total:    100,
    }
    
    api.OkList(items, meta)
    return api, nil
}
```

### API with Custom Headers
```go
func(c *request.Context) (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    
    api.Resp().RespHeaders = map[string][]string{
        "X-Resource-ID": {"123"},
    }
    
    api.Ok(data)
    return api, nil
}
```

---

## Comparison: c.Resp vs Return

### Old Way (Still Works!)
```go
func Handler(c *request.Context) error {
    c.Resp.WithStatus(201).Json(data)
    return nil
}
```

### New Way (More Flexible!)
```go
func Handler(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(201).Json(data)
    return resp, nil
}
```

**Both ways work!** Choose based on your needs.

---

## Testing

```bash
# Run tests
go test ./core/router -run TestAdaptSmart_Returns -v

# Run example
go run cmd_draft/examples/response-return-types/main.go
```

---

## Documentation

- Full docs: `docs_draft/response-return-types.md`
- Summary: `docs_draft/RESPONSE-RETURN-TYPES-SUMMARY.md`
- Tests: `core/router/helper_response_test.go`
- Example: `cmd_draft/examples/response-return-types/main.go`

---

## Key Points

1. ⚡ **Performance**: Type detection once at registration, not per-request
2. ✅ **Backward Compatible**: Existing handlers work without changes
3. 🎯 **Flexible**: Choose control level based on needs
4. 🔒 **Type Safe**: Compile-time type checking
5. ⚠️ **Error Priority**: Error always wins over Response

---

**Happy Coding!** 🚀
