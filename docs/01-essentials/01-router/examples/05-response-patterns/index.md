# Example 05: Response Patterns

> **Master all 3 response methods and 2 response paths**  
> **Time**: 15 minutes • **Concepts**: Response types, paths, when to use each

---

## 🎯 What You'll Learn

Lokstra provides **flexibility** in how you send responses:

### 3 Response Types:
1. **Manual** - Full control using `http.ResponseWriter`
2. **Generic** - Unopinionated using `response.Response` (JSON, HTML, text, etc)
3. **Opinionated** - Structured API using `response.ApiHelper` (JSON only)

### 2 Response Paths:
1. **Via Context** - Use `request.Context` to write response
2. **Via Return** - Return data or response objects

---

## 🚀 Run It

```bash
cd docs/01-essentials/01-router/examples/05-response-patterns
go run main.go
```

**Server starts on**: `http://localhost:3000`

---

## 📝 Understanding Response Types

### Type 1: Manual Response (http.ResponseWriter)

**Full manual control** - You write everything yourself.

```go
r.GET("/manual/json", func(ctx *request.Context) error {
    ctx.W.Header().Set("Content-Type", "application/json")
    ctx.W.Header().Set("X-Custom-Header", "value")
    ctx.W.WriteHeader(200)
    ctx.W.Write([]byte(`{"message":"Manual response"}`))
    return nil
})
```

**When to use**:
- ✅ Streaming responses
- ✅ Custom binary formats
- ✅ Need absolute control
- ✅ Performance-critical paths

**Pros**:
- ✅ Maximum control
- ✅ No framework overhead
- ✅ Can do anything

**Cons**:
- ❌ Most verbose
- ❌ Easy to make mistakes
- ❌ No structure

---

### Type 2: Generic Response (response.Response)

**Unopinionated** - Can send JSON, HTML, text, or any format.

```go
// JSON
r.GET("/response/json", func() *response.Response {
    resp := response.NewResponse()
    resp.RespHeaders = map[string][]string{
        "X-Custom": {"value"},
    }
    resp.WithStatus(200).Json(data)
    return resp
})

// HTML
r.GET("/response/html", func() *response.Response {
    resp := response.NewResponse()
    resp.Html("<html>...</html>")
    return resp
})

// Plain Text
r.GET("/response/text", func() *response.Response {
    resp := response.NewResponse()
    resp.Text("Plain text")
    return resp
})
```

**When to use**:
- ✅ Mixed content types (JSON, HTML, text)
- ✅ Need custom headers/status
- ✅ Don't want opinionated structure
- ✅ File downloads, streams

**Pros**:
- ✅ Flexible (multiple formats)
- ✅ Clean API
- ✅ Custom headers/status easy

**Cons**:
- ⚠️ No standard JSON structure
- ⚠️ You define error format

---

### Type 3: Opinionated API (response.ApiHelper)

**Structured JSON API** - Standard response format enforced.

```go
// Success
r.GET("/api/success", func() *response.ApiHelper {
    api := response.NewApiHelper()
    api.Ok(data)  // Standard success format
    return api
})

// Success with message
r.GET("/api/message", func() *response.ApiHelper {
    api := response.NewApiHelper()
    api.OkWithMessage(data, "Operation successful")
    return api
})

// Created (201)
r.POST("/api/created", func() *response.ApiHelper {
    api := response.NewApiHelper()
    api.Created(newResource, "Resource created")
    return api
})

// Error
r.GET("/api/error", func() *response.ApiHelper {
    api := response.NewApiHelper()
    api.NotFound("Resource not found")
    return api
})
```

**Standard Response Format**:

**Success**:
```json
{
  "status": "success",
  "data": { ... }
}
```

**Success with message**:
```json
{
  "status": "success",
  "message": "Operation successful",
  "data": { ... }
}
```

**Error**:
```json
{
  "status": "error",
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found"
  }
}
```

**When to use**:
- ✅ **REST APIs** (highly recommended!)
- ✅ Consistent JSON structure
- ✅ Multiple clients need same format
- ✅ API documentation

**Pros**:
- ✅ Consistent structure
- ✅ Easy for clients to parse
- ✅ Clear success/error distinction
- ✅ Standard HTTP status codes

**Cons**:
- ❌ JSON only (no HTML/text)
- ❌ Opinionated structure
- ❌ Can't customize format easily

---

## 🔀 Understanding Response Paths

### Path 1: Via Context (func(ctx *request.Context) error)

Write response **using context helpers**. Works for all response types.

```go
// Manual - Direct control with ctx.W
r.GET("/manual/json", func(ctx *request.Context) error {
    ctx.W.Header().Set("Content-Type", "application/json")
    ctx.W.Write([]byte(`{"message":"hello"}`))
    return nil
})

// Generic Response - Using ctx.Resp (can return directly!)
r.GET("/ctx-resp/json", func(ctx *request.Context) error {
    return ctx.Resp.WithStatus(200).Json(data)
})

r.GET("/ctx-resp/html", func(ctx *request.Context) error {
    return ctx.Resp.Html("<html>...</html>")
})

// Opinionated API - Using ctx.Api (can return directly!)
r.GET("/ctx-api/success", func(ctx *request.Context) error {
    return ctx.Api.Ok(data)
})

r.GET("/ctx-api/error", func(ctx *request.Context) error {
    return ctx.Api.NotFound("Resource not found")
})
```

**Characteristics**:
- Must have `*request.Context` parameter
- Use `ctx.W` for manual control (return nil)
- Use `ctx.Resp` for generic responses - **can return directly!**
- Use `ctx.Api` for opinionated API - **can return directly!**
- Methods return `error`, so you can `return ctx.Api.Ok(data)` directly

---

### Path 2: Via Return (func() T or func() (T, error))

**Return** response object or data. Works for all response types except manual.

```go
// Return plain data (auto JSON)
r.GET("/return/data", func() any {
    return map[string]string{"message": "hello"}
})

// Return data with error
r.GET("/return/data-error", func() (any, error) {
    return data, nil
})

// Return Response object
r.GET("/return/response", func() *response.Response {
    resp := response.NewResponse()
    resp.Json(data)
    return resp
})

// Return Response with error handling
r.GET("/return/response-error", func() (*response.Response, error) {
    resp := response.NewResponse()
    resp.Json(data)
    return resp, nil
})

// Return ApiHelper object
r.GET("/return/api", func() *response.ApiHelper {
    api := response.NewApiHelper()
    api.Ok(data)
    return api
})

// Return ApiHelper with error handling
r.GET("/return/api-error", func() (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    users, err := db.GetUsers()
    if err != nil {
        api.InternalError("Database error")
        return api, nil
    }
    api.Ok(users)
    return api, nil
})
```

**Characteristics**:
- No `*request.Context` needed (optional)
- Return response or data
- Lokstra handles response writing
- Cleaner for most cases
- Cannot be used for manual `ctx.W` responses

---

## 📊 Complete Response Matrix

| Response Type | Via Context | Via Return | Use Case |
|---------------|-------------|------------|----------|
| **Manual** | ✅ `ctx.W.Write()` | ❌ Not supported | Streaming, binary |
| **Generic Response** | ✅ `ctx.Resp.Json()` | ✅ `return resp` or `return resp, nil` | Mixed formats (JSON/HTML/text) |
| **Opinionated API** | ✅ `ctx.Api.Ok()` | ✅ `return api` or `return api, nil` | REST APIs |
| **Plain Data** | ❌ Not supported | ✅ `return data` | Simple JSON |

**Key Points**: 
- **Manual (`ctx.W`)**: Only via context, cannot be returned
- **Generic (`response.Response`)**: Can use `ctx.Resp` OR return
- **Opinionated (`response.ApiHelper`)**: Can use `ctx.Api` OR return
- **Plain Data**: Only via return

---

## 🧪 Test Examples

### Manual Response
```bash
curl http://localhost:3000/manual/json
curl http://localhost:3000/manual/text
```

**Response** (manual/json):
```json
{"message":"Manual JSON response","method":"http.ResponseWriter"}
```

---

### Generic Response
```bash
curl http://localhost:3000/response/json
curl http://localhost:3000/response/html
curl http://localhost:3000/response/text
curl -i http://localhost:3000/response/custom-status
```

**Response** (response/json):
```json
{
  "message": "Generic JSON using response.Response",
  "data": [...]
}
```

---

### Opinionated API
```bash
curl http://localhost:3000/api/success
curl http://localhost:3000/api/success-message
curl -X POST http://localhost:3000/api/created
curl http://localhost:3000/api/error-notfound
```

**Response** (api/success):
```json
{
  "status": "success",
  "data": [
    {"id": 1, "name": "Alice", "email": "alice@example.com"},
    {"id": 2, "name": "Bob", "email": "bob@example.com"}
  ]
}
```

**Response** (api/error-notfound):
```json
{
  "status": "error",
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found"
  }
}
```

---

### Return Values
```bash
curl http://localhost:3000/return/data
curl http://localhost:3000/return/struct
curl http://localhost:3000/return/response
curl http://localhost:3000/return/api
```

**Response** (return/data):
```json
{
  "message": "Direct data return",
  "users": [...],
  "count": 2
}
```

---

### Comparison
```bash
# Same data, 4 different methods
curl http://localhost:3000/compare/manual
curl http://localhost:3000/compare/response
curl http://localhost:3000/compare/api
curl http://localhost:3000/compare/return
```

---

## 🎯 Decision Guide

### Choose Response Type:

```
Need HTML/text/binary?
├─ YES → Use response.Response (generic)
│
└─ NO (JSON only)
    ├─ Need consistent API structure?
    │   ├─ YES → Use response.ApiHelper ⭐ (recommended for APIs)
    │   └─ NO → Use response.Response or plain return
    │
    └─ Need absolute control?
        └─ YES → Use manual (http.ResponseWriter)
```

### Choose Response Path:

```
Simple data return?
├─ YES → Use return path (func() T)
│
└─ NO
    ├─ Need to read request headers/body?
    │   └─ YES → Use context path (func(ctx))
    │
    └─ Need error handling?
        └─ YES → Use return with error (func() (T, error))
```

---

## 💡 Best Practices

### 1. **For REST APIs: Use ApiHelper**

```go
// ✅ Recommended for APIs
r.GET("/users", func() (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    users, err := db.GetUsers()
    if err != nil {
        api.InternalError("Database error")
        return api, nil
    }
    api.Ok(users)
    return api, nil
})
```

**Why?**
- Consistent structure
- Client-friendly
- Clear error format
- Standard HTTP codes

---

### 2. **For Simple Data: Use Return**

```go
// ✅ Simplest for basic data
r.GET("/stats", func() any {
    return map[string]int{
        "users": 100,
        "posts": 500,
    }
})
```

---

### 3. **For Mixed Content: Use response.Response**

```go
// ✅ When you need HTML, JSON, text, etc
r.GET("/page", func() (*response.Response, error) {
    resp := response.NewResponse()
    
    if acceptsJSON(req) {
        resp.Json(data)
    } else {
        resp.Html(htmlPage)
    }
    
    return resp, nil
})
```

---

### 4. **For Streaming: Use Manual**

```go
// ✅ For SSE, file streaming, etc
r.GET("/stream", func(ctx *request.Context) error {
    ctx.W.Header().Set("Content-Type", "text/event-stream")
    
    for event := range events {
        fmt.Fprintf(ctx.W, "data: %s\n\n", event)
        ctx.W.(http.Flusher).Flush()
    }
    
    return nil
})
```

---

## 📋 ApiHelper Methods Reference

### Success Methods
```go
api.Ok(data)                           // 200 OK
api.OkWithMessage(data, "message")     // 200 OK with message
api.Created(data, "message")           // 201 Created
api.OkList(data, meta)                 // 200 OK with pagination
```

### Error Methods
```go
api.BadRequest(code, message)          // 400 Bad Request
api.Unauthorized(message)              // 401 Unauthorized
api.Forbidden(message)                 // 403 Forbidden
api.NotFound(message)                  // 404 Not Found
api.InternalError(message)             // 500 Internal Server Error
```

---

## 🎓 What You Learned

- ✅ 3 response types: Manual, Generic, Opinionated
- ✅ 2 response paths: Via Context, Via Return
- ✅ When to use each type
- ✅ ApiHelper standard format
- ✅ response.Response flexibility
- ✅ Manual control with http.ResponseWriter
- ✅ Decision guide for choosing

---

## 🔗 Related

- **ApiHelper Guide**: [Deep Dive](../../../02-deep-dive/response/api-helper) (coming soon)
- **Response Formats**: [Deep Dive](../../../02-deep-dive/response/formats) (coming soon)
- **Error Handling**: [Guide](../../../04-guides/error-handling) (coming soon)

---

**Back**: [04 - Handler Forms](../04-handler-forms/)  
**Next**: Ready to build a [complete API](../../06-putting-it-together/)!
