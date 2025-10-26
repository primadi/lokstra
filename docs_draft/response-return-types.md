# Response Return Types - Handler Flexibility

## Overview

Framework Lokstra sekarang mendukung berbagai cara untuk mengembalikan response dari handler, memberikan fleksibilitas maksimal kepada developer.

## Supported Return Types

### 1. **Regular Data Return** (Existing)
Handler mengembalikan data langsung, yang akan di-wrap dalam API response format.

```go
// Return (data, error)
func GetUser(c *request.Context) (User, error) {
    user := User{ID: 1, Name: "John"}
    return user, nil
}

// Return (any, error)
func GetUsers(c *request.Context) ([]User, error) {
    users := []User{{ID: 1, Name: "John"}}
    return users, nil
}
```

**Behavior**: Data otomatis di-wrap menggunakan `ctx.Api.Ok(data)`

---

### 2. **Response Pointer Return** (New)
Handler mengembalikan `*response.Response` untuk kontrol penuh atas response.

```go
func CustomResponse(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(201).Json(map[string]string{
        "message": "Created successfully",
    })
    return resp, nil
}

func PlainTextResponse(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(200).Text("Hello World")
    return resp, nil
}

func StreamResponse(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.Stream("text/event-stream", func(w http.ResponseWriter) error {
        fmt.Fprintf(w, "data: %s\n\n", "streaming data")
        return nil
    })
    return resp, nil
}
```

**Behavior**: 
- Response langsung digunakan (tidak di-wrap)
- Full control: status code, headers, content-type, body
- Equivalent dengan akses `c.Resp`

---

### 3. **Response Value Return** (New)
Handler mengembalikan `response.Response` (value, bukan pointer).

```go
func ResponseValue(c *request.Context) (response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(202).Html("<h1>Accepted</h1>")
    return *resp, nil // Return value
}
```

**Behavior**: Sama seperti return pointer, tapi menggunakan value.

---

### 4. **ApiHelper Pointer Return** (New)
Handler mengembalikan `*response.ApiHelper` untuk menggunakan API formatting helpers.

```go
func CreateResource(c *request.Context) (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    
    resource := Resource{ID: "123", Name: "New Resource"}
    api.Created(resource, "Resource created successfully")
    
    return api, nil
}

func ListWithPagination(c *request.Context) (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    
    users := []User{{ID: 1, Name: "John"}}
    meta := &api_formatter.ListMeta{
        Page:       1,
        PerPage:    10,
        TotalItems: 100,
        TotalPages: 10,
    }
    
    api.OkList(users, meta)
    return api, nil
}

func CustomHeaderResponse(c *request.Context) (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    
    // Access underlying Response for custom headers
    resp := api.Resp()
    resp.RespHeaders = map[string][]string{
        "X-Custom-Header": {"custom-value"},
    }
    
    api.Ok(map[string]string{"data": "with custom header"})
    return api, nil
}
```

**Behavior**: 
- Menggunakan API response formatting (success, error, list, dll)
- Akses ke underlying Response melalui `api.Resp()`
- Equivalent dengan akses `c.Api`

---

### 5. **ApiHelper Value Return** (New)
Handler mengembalikan `response.ApiHelper` (value, bukan pointer).

```go
func ApiValue(c *request.Context) (response.ApiHelper, error) {
    api := response.NewApiHelper()
    api.Ok(map[string]int{"count": 42})
    return *api, nil // Return value
}
```

**Behavior**: Sama seperti return pointer.

---

## Error Handling Priority

**PENTING**: Jika error tidak nil, error **selalu diprioritaskan** meskipun Response/ApiHelper memiliki status code.

```go
func ErrorTakesPrecedence(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(200).Json(map[string]string{
        "message": "This will be IGNORED",
    })
    
    // Error takes precedence!
    return resp, errors.New("something went wrong")
}
```

**Behavior**: Framework akan memprioritaskan error dan mengabaikan Response yang sudah dibuat.

---

## Nil Pointer Handling

Jika handler mengembalikan nil pointer untuk Response/ApiHelper:

```go
func NilResponse(c *request.Context) (*response.Response, error) {
    return nil, nil // No error, no response
}

func NilApiHelper(c *request.Context) (*response.ApiHelper, error) {
    return nil, nil // No error, no helper
}
```

**Behavior**: Framework akan mengirim default success response menggunakan `ctx.Api.Ok(nil)`

---

## All Supported Signatures

### With Context Parameter
```go
func(*Context) error
func(*Context) (data, error)
func(*Context) (*Response, error)
func(*Context) (Response, error)
func(*Context) (*ApiHelper, error)
func(*Context) (ApiHelper, error)

func(*Context, *Struct) error
func(*Context, *Struct) (data, error)
func(*Context, *Struct) (*Response, error)
func(*Context, *Struct) (Response, error)
func(*Context, *Struct) (*ApiHelper, error)
func(*Context, *Struct) (ApiHelper, error)
```

### Without Context Parameter
```go
func(*Struct) error
func(*Struct) (data, error)
func(*Struct) (*Response, error)
func(*Struct) (Response, error)
func(*Struct) (*ApiHelper, error)
func(*Struct) (ApiHelper, error)

func() error
func() (data, error)
func() (*Response, error)
func() (Response, error)
func() (*ApiHelper, error)
func() (ApiHelper, error)
```

---

## Use Cases & Best Practices

### When to use **Regular Data Return**
✅ Simple CRUD operations  
✅ Standard API responses  
✅ Data yang sudah ter-structure dengan baik  

```go
func GetUser(c *request.Context) (User, error) {
    return userService.GetByID(id)
}
```

---

### When to use **Response Pointer/Value**
✅ Custom status codes  
✅ Custom content-type (HTML, plain text, XML, dll)  
✅ Streaming responses  
✅ File downloads  
✅ Custom headers yang kompleks  

```go
func DownloadFile(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.RespHeaders = map[string][]string{
        "Content-Disposition": {"attachment; filename=file.pdf"},
    }
    resp.Stream("application/pdf", func(w http.ResponseWriter) error {
        return writeFileToStream(w)
    })
    return resp, nil
}
```

---

### When to use **ApiHelper Pointer/Value**
✅ Standardized API responses (success, error, list)  
✅ Pagination dan metadata  
✅ Validation errors  
✅ Perlu custom header tapi tetap menggunakan API format  

```go
func CreateUser(c *request.Context) (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    
    user, err := userService.Create(input)
    if err != nil {
        return nil, err // Error handling
    }
    
    // Set custom header
    api.Resp().RespHeaders = map[string][]string{
        "X-Resource-ID": {user.ID},
    }
    
    api.Created(user, "User created successfully")
    return api, nil
}
```

---

## Migration Guide

Existing handlers **tetap kompatibel** tanpa perubahan:

```go
// Old style - still works!
func OldHandler(c *request.Context) (User, error) {
    return user, nil
}

// New flexibility - when you need it
func NewHandler(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(201).Json(user)
    return resp, nil
}
```

Tidak ada breaking changes!

---

## Implementation Details

### Type Detection
Framework mendeteksi return type menggunakan reflection saat handler registration:

```go
// Detected types
typeOfResponse    = reflect.TypeOf((*response.Response)(nil))
typeOfApiHelper   = reflect.TypeOf((*response.ApiHelper)(nil))
typeOfResponseVal = reflect.TypeOf(response.Response{})
typeOfApiHelperVal = reflect.TypeOf(response.ApiHelper{})
```

### Handler Metadata
Metadata di-compile sekali saat registration (bukan per-request):

```go
type handlerMetadata struct {
    returnsResponse  bool // *response.Response or response.Response
    returnsApiHelper bool // *response.ApiHelper or response.ApiHelper
    isResponsePtr    bool // Pointer vs value
    isApiHelperPtr   bool // Pointer vs value
}
```

### Response Handling Flow
1. Execute handler function
2. Check error return (prioritas tertinggi)
3. Check return type (Response, ApiHelper, or regular data)
4. Process response sesuai type
5. Copy Response ke `ctx.Resp` atau wrap dengan `ctx.Api.Ok()`

---

## Performance Considerations

- **Zero overhead** untuk regular data returns (existing behavior)
- **Minimal overhead** untuk Response/ApiHelper returns (single pointer copy)
- **Type detection** dilakukan sekali saat registration, bukan per-request
- **No allocations** untuk response handling (reuses existing ctx.Resp)

---

## Testing

Lihat `core/router/helper_response_test.go` untuk contoh comprehensive testing berbagai return types.

```bash
go test ./core/router -run TestAdaptSmart_Returns
```

---

## Summary

| Return Type | Control Level | Use Case |
|------------|---------------|----------|
| `(data, error)` | Low | Standard API responses |
| `(*Response, error)` | High | Custom status, headers, content-type |
| `(*ApiHelper, error)` | Medium | Standardized API + custom headers |

Pilih sesuai kebutuhan! Framework mendukung semuanya. 🚀
