# Handler Forms Example

> **Demonstrates handler forms and 3 response methods in Lokstra**

Related to: [Key Features - Handler Forms](../../key-features.md#feature-1-29-handler-forms)

---

## 📖 What This Example Shows

### Handler Forms:
- ✅ Simple handlers (return values - auto wrapped in ApiHelper)
- ✅ Handlers with error handling
- ✅ Request binding (JSON, path params, headers)
- ✅ Context access (headers, params)

### Response Methods (3 Ways):
1. **ApiHelper** (opiniated, default for simple functions)
   - `ctx.Api.Ok()`, `ctx.Api.Created()`, `ctx.Api.Error()`, etc.
   - Wraps data in standard JSON structure
   - Can be customized if API response format differs

2. **response.Response** (generic helper, not opiniated)
   - `resp.Json()`, `resp.Html()`, `resp.Text()`, `resp.Stream()`
   - Full control over response format
   - Custom status codes and headers

3. **Manual** (http.ResponseWriter)
   - Direct `ctx.W.Write()` for maximum control
   - For special cases or custom protocols

---

## 🚀 Run the Example

```bash
# From this directory
go run main.go
```

Server will start on `http://localhost:3001`

---

## 🧪 Test the Endpoints

Use the `test.http` file in VS Code with REST Client extension.

### Simple Forms (Auto ApiHelper)

### Simple Forms (Auto ApiHelper)

```bash
# Return string (auto wrapped)
curl http://localhost:3001/ping

# Return map (auto wrapped)
curl http://localhost:3001/time

# Return slice (auto wrapped)
curl http://localhost:3001/users
```

### Request Binding

```bash
# Create user (JSON binding + validation)
curl -X POST http://localhost:3001/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}'

# Get user by ID (path param binding)
curl http://localhost:3001/users/123

# Update user (context + binding)
curl -X PUT http://localhost:3001/users/456 \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane","email":"jane@example.com"}'

# Header binding
curl http://localhost:3001/headers \
  -H "Authorization: Bearer my-token" \
  -H "X-Custom-Header: custom-value"
```

### ApiHelper Responses (Opiniated)

```bash
# Standard success
curl http://localhost:3001/api-ok

# Success with message
curl http://localhost:3001/api-ok-message

# Created (201)
curl -X POST http://localhost:3001/api-created \
  -H "Content-Type: application/json" \
  -d '{"name":"test"}'

# Error responses
curl http://localhost:3001/api-not-found
curl http://localhost:3001/api-bad-request
```

### response.Response (Generic, Not Opiniated)

```bash
# JSON response
curl http://localhost:3001/resp-json

# HTML response
curl http://localhost:3001/resp-html

# Text response
curl http://localhost:3001/resp-text

# Custom status (202)
curl -X POST http://localhost:3001/resp-custom-status

# File download (stream)
curl http://localhost:3001/resp-download
```

### Manual Responses (http.ResponseWriter)

```bash
# Manual JSON
curl http://localhost:3001/manual-json

# Manual with custom headers
curl -i http://localhost:3001/manual-custom

# Manual text
curl http://localhost:3001/manual-text
```

### Error Handling

```bash
# Auto 500 error
curl http://localhost:3001/error-500

# Validation error (missing fields)
curl -X POST http://localhost:3001/validate \
  -H "Content-Type: application/json" \
  -d '{}'

# Validation success
curl -X POST http://localhost:3001/validate \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com"}'
```

---

## � Response Examples

### ApiHelper (Default, Opiniated)

**Request**: `GET /ping`

**Response**:
```json
{
  "status": "success",
  "data": "pong"
}
```

**Request**: `GET /api-ok-message`

**Response**:
```json
{
  "status": "success",
  "message": "Operation completed successfully",
  "data": {
    "status": "completed"
  }
}
```

### response.Response (Generic)

**Request**: `GET /resp-json`

**Response**:
```json
{
  "message": "Generic JSON response",
  "type": "custom"
}
```

Note: No wrapper - you control exact JSON structure

### Manual Response

**Request**: `GET /manual-custom`

**Response Headers**:
```
HTTP/1.1 200 OK
Content-Type: application/json
X-Custom-Header: custom-value
X-Request-ID: req-123
```

**Response Body**:
```json
{
  "message": "Manual response with custom headers"
}
```

---

## 💡 Key Concepts

### 1. Three Response Methods

#### Method 1: ApiHelper (Recommended for APIs)
```go
r.GET("/users", func(ctx *request.Context) error {
    return ctx.Api.Ok(users)  // Auto-wrapped in standard format
})
```

**When to use**:
- Building REST APIs
- Want consistent response format
- Need standard success/error structure

**Output**:
```json
{
  "status": "success",
  "data": {...}
}
```

#### Method 2: response.Response (For Custom Formats)
```go
r.GET("/custom", func() (*response.Response, error) {
    resp := response.NewResponse()
    resp.Json(customData)  // Your exact JSON structure
    return resp, nil
})
```

**When to use**:
- Need exact control over JSON structure
- Non-JSON responses (HTML, text, files)
- Custom content types

**Output**: Whatever you define

#### Method 3: Manual (For Special Cases)
```go
r.GET("/special", func(ctx *request.Context) error {
    ctx.W.Header().Set("Content-Type", "application/xml")
    ctx.W.Write([]byte("<root>...</root>"))
    return nil
})
```

**When to use**:
- Streaming responses
- Binary protocols
- WebSocket upgrades
- Custom HTTP behaviors

### 2. Handler Forms by Complexity

#### Level 1: Simple (Auto ApiHelper)
```go
// Return value - auto wrapped
r.GET("/ping", func() string {
    return "pong"  // → {"status":"success","data":"pong"}
})
```

#### Level 2: With Error Handling
```go
// Return value + error
r.GET("/users", func() ([]User, error) {
    users, err := db.GetAll()
    if err != nil {
        return nil, err  // → 500 error
    }
    return users, nil  // → {"status":"success","data":[...]}
})
```

#### Level 3: Request Binding

**IMPORTANT**: Request binding **requires explicit struct with tags**. You cannot use `map[string]any` directly.

```go
// ❌ WRONG - Cannot bind directly to map
r.POST("/users", func(ctx *request.Context, req map[string]any) error {
    // This will NOT work!
})

// ✅ CORRECT - Must use struct with tags
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

r.POST("/users", func(req *CreateUserRequest) (*User, error) {
    return createUser(req)  // Validation auto-handled
})

// ✅ CORRECT - Bind entire JSON body to map
type ApiCreatedParams struct {
    Data map[string]any `json:"*"` // `json:"*"` binds entire body
}

r.POST("/api-created", func(ctx *request.Context, req *ApiCreatedParams) error {
    return ctx.Api.Created(map[string]any{
        "id":   123,
        "data": req.Data, // Access via struct field
    }, "Resource created successfully")
})

// ✅ CORRECT - Bind from HTTP headers
type HeaderParams struct {
    Authorization string `header:"Authorization"`
    UserAgent     string `header:"User-Agent"`
    CustomHeader  string `header:"X-Custom-Header"`
}

r.GET("/headers", func(req *HeaderParams) (map[string]any, error) {
    return map[string]any{
        "authorization": req.Authorization,
        "user_agent":    req.UserAgent,
        "custom_header": req.CustomHeader,
    }, nil
})
```

**Supported Binding Tags**:
- `json:"field_name"` - Bind from JSON body field
- `json:"*"` - Bind entire JSON body to a field
- `path:"param_name"` - Bind from URL path parameter
- `query:"param_name"` - Bind from query string
- `header:"Header-Name"` - Bind from HTTP header
```

**Key Rules**:
- Request parameters must be **struct with tags** (`json`, `path`, `query`, `header`)
- Use `json:"*"` to bind entire JSON body to a field
- Validation happens automatically with `validate` tags

#### Level 4: Full Control with Context
```go
// Access everything + choose response method
r.GET("/custom", func(ctx *request.Context) error {
    // Option A: Use ApiHelper
    return ctx.Api.Ok(data)
    
    // Option B: Use response.Response
    resp := response.NewResponse()
    resp.Html("<h1>Hello</h1>")
    // ... manually write response
    
    // Option C: Manual
    ctx.W.Write([]byte("..."))
    return nil
})
```

### 3. Error Handling

#### Auto 500 (Return Error)
```go
r.GET("/error", func() (string, error) {
    return "", fmt.Errorf("oops")  // → 500 Internal Server Error
})
```

#### Custom Error (ApiHelper)
```go
r.GET("/not-found", func(ctx *request.Context) error {
    return ctx.Api.Error(404, "NOT_FOUND", "Resource not found")
    // → 404 with standard error format
})
```

#### Validation Errors (Auto)
```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

r.POST("/users", func(req *CreateUserRequest) (*User, error) {
    // If validation fails → Auto 400 with field errors
    return createUser(req)
})
```

---

## 🎯 Choosing Response Method

| Scenario | Method | Why |
|----------|--------|-----|
| REST API | ApiHelper | Consistent format, easy to use |
| Custom JSON | response.Response | Full control over structure |
| HTML pages | response.Response | `.Html()` method |
| File download | response.Response | `.Stream()` method |
| Binary data | Manual | Direct write |
| WebSocket | Manual | Protocol upgrade |

---

## 🔍 What's Next?

Try modifying:
- Customize ApiHelper response format
- Add more validation rules
- Implement file upload
- Add custom middleware

See more examples:
- [CRUD API](../03-crud-api/) - Complete REST API with services
- [Multi-Deployment](../04-multi-deployment/) - Monolith vs Microservices

---

**Questions?** Check the [Key Features Guide](../../key-features.md)
