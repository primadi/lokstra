# 02. Request Context - Complete Reference

**Foundation:** Understanding `request.Context` is essential before writing handlers and middleware.

## What is Context?

`request.Context` is the central object in Lokstra. Every handler receives it:

```go
func MyHandler(c *request.Context) error {
    // c contains everything you need
    return nil
}
```

## The Structure

```go
type Context struct {
    context.Context       // Embedded Go context

    // HELPERS (Recommended) - use these 95% of the time
    Req  *RequestHelper   // Request data access
    Resp *Response        // Response building
    Api  *ApiHelper       // API responses

    // PRIMITIVES (Advanced) - direct HTTP access
    R *http.Request       // Raw HTTP request
    W *writerWrapper      // Raw response writer

    // INTERNAL - framework internals
    index    int
    handlers []HandlerFunc
}
```

## Three Helper Layers

### 1. Req Helper - Unopinionated Request Access

**Philosophy:** Complete access to request data without imposing structure.

```go
// Single values with defaults
id := c.Req.PathParam("id", "")
page := c.Req.QueryParam("page", "1")
auth := c.Req.HeaderParam("Authorization", "")

// Multiple values
tags := c.Req.QueryParams("tag")  // []string

// Bulk access
allParams := c.Req.AllQueryParams()  // map[string][]string

// Body
body, err := c.Req.RawRequestBody()  // []byte
```

**Binding (Recommended):** Automatic validation + type conversion

```go
type CreateUserInput struct {
    ID   string   `path:"id"`
    Page int      `query:"page"`
    Auth string   `header:"Authorization"`
    Name string   `json:"name" validate:"required,min=3"`
    Age  int      `json:"age" validate:"required,gte=0"`
}

var input CreateUserInput
if err := c.Req.BindAll(&input); err != nil {
    // Returns ValidationError with field-level details
    return c.Api.BadRequest("BIND_ERROR", err.Error())
}

// input.Name, input.Age ready to use with validation passed
```

**Available Bind Methods:**
- `BindPath(&struct)` - Path parameters only
- `BindQuery(&struct)` - Query parameters only
- `BindHeader(&struct)` - Headers only
- `BindBody(&struct)` - Body only (requires Content-Type)
- `BindAll(&struct)` - Everything combined
- `BindBodyAutoContentType(&struct)` - Auto-detect JSON/XML/form
- `BindAllAutoContentType(&struct)` - Everything with auto-detect

### 2. Resp Helper - Unopinionated Response Building

**Philosophy:** Maximum flexibility. You control the entire response structure.

```go
// JSON with custom structure
return c.Resp.Json(map[string]any{
    "message": "Success",
    "data": userData,
    "timestamp": time.Now(),
})

// Custom status codes
return c.Resp.WithStatus(201).Json(map[string]any{
    "id": newID,
})

// HTML
return c.Resp.Html("<h1>Admin Panel</h1>")

// Plain text
return c.Resp.Text("Health check OK")

// Raw bytes
return c.Resp.Raw("application/pdf", pdfBytes)

// Streaming for large files
return c.Resp.Stream("application/octet-stream", func(w http.ResponseWriter) error {
    // Write data in chunks
    return writeFileInChunks(w, filePath)
})
```

**Use Cases:**
- Admin panels / internal UIs
- Webhooks (external services expect specific formats)
- File downloads / streaming
- Custom protocols
- Any non-REST-API endpoint

### 3. Api Helper - Opinionated API Responses

**Philosophy:** Consistent structure for REST APIs with pluggable formatting.

```go
// Success responses
return c.Api.Ok(user)
return c.Api.OkWithMessage(user, "User retrieved successfully")
return c.Api.Created(newUser, "User created")

// Lists with pagination
return c.Api.OkList(users, &api_formatter.ListMeta{
    Page:  page,
    Limit: limit,
    Total: totalCount,
})

// Error responses
return c.Api.Error(400, "INVALID_INPUT", "Name is required")
return c.Api.ValidationError("Validation failed", []FieldError{
    {Field: "email", Message: "Invalid email format"},
    {Field: "age", Message: "Must be >= 18"},
})

// Error shortcuts
return c.Api.BadRequest("INVALID_ID", "ID must be numeric")
return c.Api.Unauthorized("Invalid token")
return c.Api.Forbidden("Access denied")
return c.Api.NotFound("User not found")
return c.Api.InternalError("Database connection failed")
```

**Default Output Format:**

```json
{
  "success": true,
  "data": { ... },
  "message": "Success"
}
```

**Pluggable Formatters:**

```go
// Custom formatter
response.SetApiResponseFormatter(func(data any, meta any, statusCode int) any {
    return map[string]any{
        "status": statusCode,
        "payload": data,
    }
})

// Or use built-in formatters
response.SetApiResponseFormatterByName("json-api")
```

**Use Cases:**
- REST APIs
- Mobile app backends
- Public APIs
- Microservice APIs
- Any API requiring consistent structure

## When to Use What?

| Scenario | Use |
|----------|-----|
| REST API endpoint | `c.Api.Ok()`, `c.Api.NotFound()` |
| Admin panel HTML | `c.Resp.Html()` |
| Webhook callback | `c.Resp.Json()` custom structure |
| File download | `c.Resp.Stream()` |
| Custom protocol | `c.R` and `c.W` primitives |

## Primitives: R and W

**Philosophy:** Direct HTTP access when helpers aren't enough.

```go
// Read request details
method := c.R.Method
url := c.R.URL.Path
clientIP := c.R.RemoteAddr
isTLS := c.R.TLS != nil
userAgent := c.R.Header.Get("User-Agent")

// Manual response (advanced)
c.W.Header().Set("X-Custom-Header", "value")
c.W.WriteHeader(200)
c.W.Write([]byte("manual response"))
return nil  // MUST return nil after manual write
```

⚠️ **Important:** After using `c.W.Write()`, `c.Resp` and `c.Api` are bypassed!

## Context Storage - Middleware Communication

Share data between middleware and handlers:

```go
// Middleware sets data
func AuthMiddleware(c *request.Context) error {
    user := authenticateToken(c)
    
    c.Set("user_id", user.ID)
    c.Set("username", user.Username)
    c.Set("role", user.Role)
    
    return c.Next()
}

// Handler reads data
func GetProfile(c *request.Context) error {
    // Type-unsafe (returns any)
    userID := c.Value("user_id")
    
    // Type-safe (recommended)
    userIDStr := request.GetContextValue(c.Context, "user_id", "")
    role := request.GetContextValue(c.Context, "role", "guest")
    
    return c.Api.Ok(map[string]any{
        "user_id": userIDStr,
        "role": role,
    })
}
```

**Common Keys:**
- `user_id`, `username`, `role` - Authentication
- `request_id`, `trace_id` - Observability
- `tenant_id` - Multi-tenancy
- `session_id` - Session management

## Middleware Chain - c.Next()

Control execution flow:

```go
func LoggingMiddleware(c *request.Context) error {
    start := time.Now()
    fmt.Printf("→ %s %s\n", c.R.Method, c.R.URL.Path)
    
    err := c.Next()  // Execute next middleware/handler
    
    duration := time.Since(start)
    fmt.Printf("← %v\n", duration)
    
    return err
}
```

**Execution Flow:**

```
Request → Middleware 1 → Middleware 2 → Handler
             ↓              ↓             ↓
          Before         Before        Execute
             ↓              ↓             ↓
         c.Next()       c.Next()       return
             ↓              ↓             ↓
          After          After         Response
```

**Early Return (Stop Chain):**

```go
func AuthMiddleware(c *request.Context) error {
    token := c.Req.HeaderParam("Authorization", "")
    
    if token == "" {
        // Don't call Next() - chain stops here
        return c.Api.Unauthorized("Token required")
    }
    
    // Continue to next handler
    return c.Next()
}
```

## Quick Reference

### Request Data
```go
c.Req.PathParam("id", "")
c.Req.QueryParam("page", "1")
c.Req.BindAll(&input)
```

### Responses
```go
// Flexible
c.Resp.Json(data)
c.Resp.WithStatus(201).Json(data)
c.Resp.Html("<h1>Hello</h1>")

// API
c.Api.Ok(data)
c.Api.Created(data, "Created")
c.Api.NotFound("Not found")
c.Api.BadRequest("CODE", "Message")
```

### Context Storage
```go
c.Set("key", value)
val := request.GetContextValue[string](c.Context, "key", "default")
```

### Middleware
```go
err := c.Next()  // Continue
return c.Api.Unauthorized("...")  // Stop
```

## Running the Example

```bash
go run .
```

This prints a comprehensive reference guide to your terminal.

## Next Steps

1. **[03-handlers](../03-handlers/)** - Use Context in real handler patterns
2. **[04-middleware](../04-middleware/)** - Build middleware with Context
3. **[05-services](../05-services/)** - Access services through Context

## Key Takeaways

✅ **Context is the foundation** - understand it before handlers/middleware  
✅ **Three helpers** - Req (request), Resp (response), Api (opinionated API)  
✅ **Req.BindAll()** - Best way to get request data (validation + types)  
✅ **Use c.Api for APIs** - Consistent structure with error handling  
✅ **Use c.Resp for custom** - Full control when you need it  
✅ **Context storage** - Share data between middleware and handlers  
✅ **c.Next()** - Control middleware execution flow  
