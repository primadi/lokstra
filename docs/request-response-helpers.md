# Request/Response Helper Usage

Setelah refactoring, Context sekarang terorganisir dengan lebih baik:

## Struktur Baru

```go
type Context struct {
    // Helper groupings for better organization
    Req  *RequestHelper  // Request-related helpers
    Resp *ResponseHelper // Response-related helpers

    // Direct access to primitives (for advanced usage)
    W *writerWrapper // Internal use
    R *http.Request  // Internal use
    
    // ... other fields
}
```

## Penggunaan Request Helper

### Parameter Access
```go
func handler(c *request.Context) error {
    // Query parameters
    name := c.Req.QueryParam("name", "anonymous")
    page := c.Req.QueryParam("page", "1")
    
    // Path parameters  
    id := c.Req.PathParam("id", "0")
    
    // Form parameters
    email := c.Req.FormParam("email", "")
    
    // Header parameters
    auth := c.Req.HeaderParam("Authorization", "")
    userAgent := c.Req.HeaderParam("User-Agent", "")
    
    return nil
}
```

### Request Body
```go
func handler(c *request.Context) error {
    // Raw body
    body, err := c.Req.GetRawRequestBody()
    if err != nil {
        return err
    }
    
    return nil
}
```

### Binding to Structs
```go
type UserRequest struct {
    ID    string `path:"id"`
    Name  string `query:"name"`
    Email string `json:"email"`
}

func handler(c *request.Context) error {
    var req UserRequest
    
    // Individual binding
    c.Req.BindPath(&req)   // Path params
    c.Req.BindQuery(&req)  // Query params 
    c.Req.BindHeader(&req) // Headers
    c.Req.BindBody(&req)   // Request body
    
    // Or bind everything at once
    if err := c.Req.BindAll(&req); err != nil {
        return err
    }
    
    // Smart binding (auto-detect content type)
    if err := c.Req.BindAllSmart(&req); err != nil {
        return err
    }
    
    return nil
}
```

## Penggunaan Response Methods

### Basic Responses
```go
func handler(c *request.Context) error {
    // JSON response
    return c.Resp.Json(map[string]string{"message": "hello"})
    
    // Text response
    return c.Resp.Text("Hello World")
    
    // HTML response
    return c.Resp.Html("<h1>Hello</h1>")
    
    // Raw response with custom content type
    return c.Resp.Raw("text/plain", []byte("hello"))
}
```

### Success Responses
```go
func handler(c *request.Context) error {
    user := getUserById(id)
    
    // 200 OK with data
    return c.Api.Ok(user)
    
    // 201 Created with data
    return c.Api.OkCreated(user)
    
    // 204 No Content
    return c.Api.OkNoContent()
}
```

### Error Responses
```go
func handler(c *request.Context) error {
    // 400 Bad Request
    return c.Resp.ErrorBadRequest("Invalid input")
    
    // 401 Unauthorized
    return c.Resp.ErrorUnauthorized("Authentication required")
    
    // 403 Forbidden
    return c.Resp.ErrorForbidden("Access denied")
    
    // 404 Not Found
    return c.Resp.ErrorNotFound("User not found")
    
    // 409 Conflict
    return c.Resp.ErrorConflict("Email already exists")
    
    // 500 Internal Server Error
    return c.Resp.ErrorInternal(err)
}
```

### Chainable Status
```go
func handler(c *request.Context) error {
    // Set custom status code
    return c.Resp.WithStatus(201).JSON(createdUser)
}
```

### Streaming Response
```go
func handler(c *request.Context) error {
    return c.Resp.Stream("text/plain", func(w http.ResponseWriter) error {
        for i := 0; i < 10; i++ {
            fmt.Fprintf(w, "chunk %d\n", i)
            w.(http.Flusher).Flush()
            time.Sleep(100 * time.Millisecond)
        }
        return nil
    })
}
```

## Direct Access (Advanced)

Untuk kasus advanced, masih bisa akses primitif:
```go
func handler(c *request.Context) error {
    // Direct access to http.Request
    method := c.R.Method
    url := c.R.URL.String()
    
    // Direct access to ResponseWriter (through wrapper)
    c.W.Header().Set("Custom-Header", "value")
    c.W.WriteHeader(200)
    c.W.Write([]byte("custom response"))
    
    return nil
}
```

## Migrasi dari Cara Lama

### Sebelum:
```go
// Old way - methods scattered on Context
name := c.QueryParam("name", "default")
c.BindPath(&req)
c.BindQuery(&req) 
c.BindAll(&req)
c.JSON(data)
c.ErrorBadRequest("error")
```

### Sesudah:
```go  
// New way - organized under helpers
name := c.Req.QueryParam("name", "default")
c.Req.BindPath(&req)
c.Req.BindQuery(&req)
c.Req.BindAll(&req)
c.Resp.Json(data)
c.Resp.ErrorBadRequest(errors.New("error"))
```

### Compatibility Notes
- Old Context methods still work but are deprecated
- All binding methods moved to `c.Req` namespace
- All response methods moved to `c.Resp` namespace  
- Direct access to `c.R` and `c.W` still available for advanced usage

## Keuntungan Struktur Baru

1. **Organized**: Helper terkelompok berdasarkan domain (Request/Response)
2. **Readable**: Auto-complete editor lebih jelas (c.Req. vs c.Resp.)
3. **Scalable**: Mudah menambah helper baru tanpa "menggemukkan" Context
4. **Backward Compatible**: Embedded Response masih bisa diakses langsung
5. **Discoverable**: Lebih mudah menemukan method yang dibutuhkan