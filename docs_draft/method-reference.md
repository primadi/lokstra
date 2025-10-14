# Complete Method Reference: Two-Layer Pattern

Lokstra provides **two layers** of response methods with **registry-based formatters** for maximum flexibility and clarity:

1. **Base Response (`c.Resp`)** - Unopinionated core response methods
2. **API Response (`c.Api`)** - Configurable structured responses via registry pattern

> **Note**: The old confusing 3-layer pattern with JSON helpers has been **removed** for simplicity.

## 📋 **Complete Method List**

### Layer 1: Base Response (`c.Resp`) - Unopinionated Methods

| Method | Description | Parameters | Output |
|--------|-------------|------------|---------|
| `Json(data)` | Direct JSON response | `data any` | Raw JSON |
| `Text(text)` | Plain text response | `text string` | Plain text |
| `Html(html)` | HTML response | `html string` | HTML content |
| `Raw(contentType, bytes)` | Binary response | `contentType string, bytes []byte` | Binary data |
| `Stream(contentType, fn)` | Streaming response | `contentType string, fn func(w http.ResponseWriter) error` | Streamed data |
| `WithStatus(code)` | Set HTTP status | `code int` | Chainable Response |

**Characteristics**:
- ✅ Maximum control over output format
- ✅ Any content type supported
- ✅ No assumptions about structure
- ✅ Direct, predictable output

### Layer 2: API Response (`c.Api`) - Configurable Methods

| Method | Description | Parameters | Status | Output Format |
|--------|-------------|------------|--------|---------------|
| `Ok(data)` | Successful response | `data any` | 200 | Formatter-dependent |
| `OkWithMessage(data, msg)` | Success with message | `data any, message string` | 200 | Formatter-dependent |
| `Created(data, msg)` | Resource created | `data any, message string` | 201 | Formatter-dependent |
| `OkList(data, meta)` | Paginated list | `data any, meta *ListMeta` | 200 | Formatter-dependent |
| `ValidationError(msg, fields)` | Validation error | `message string, fields []FieldError` | 400 | Formatter-dependent |
| `BadRequest(code, msg)` | Bad request error | `code string, message string` | 400 | Formatter-dependent |
| `Unauthorized(msg)` | Unauthorized error | `message string` | 401 | Formatter-dependent |
| `Forbidden(msg)` | Forbidden error | `message string` | 403 | Formatter-dependent |
| `NotFound(msg)` | Not found error | `message string` | 404 | Formatter-dependent |
| `InternalError(msg)` | Server error | `message string` | 500 | Formatter-dependent |

**Characteristics**:
- ✅ Consistent API structure
- ✅ Configurable output format via registry
- ✅ Built-in error handling patterns
- ✅ Rich metadata support

## 🔧 **Built-in Response Formatters**

### 1. `api` (Default) - Structured API Format
```go
response.SetApiResponseFormatterByName("api")
return c.Api.Ok(users)
```

**Output**:
```json
{
  "status": "success", 
  "data": [{"id": 1, "name": "John"}]
}
```

### 2. `simple` - Minimal JSON Format  
```go
response.SetApiResponseFormatterByName("simple")
return c.Api.Ok(users)
```

**Output**:
```json
[{"id": 1, "name": "John"}]
```

### 3. `legacy` - Legacy System Format
```go
response.SetApiResponseFormatterByName("legacy") 
return c.Api.Ok(users)
```

**Output**:
```json
{
  "success": true,
  "result": [{"id": 1, "name": "John"}]
}
```

## 📄 **Response Format Comparison**

### Success Response Comparison

| Layer | Method | Output |
|-------|--------|--------|
| **Base** | `c.Resp.Json(users)` | `[{"id":1,"name":"John"}]` |
| **API (api)** | `c.Api.Ok(users)` | `{"status":"success","data":[{"id":1,"name":"John"}]}` |
| **API (simple)** | `c.Api.Ok(users)` | `[{"id":1,"name":"John"}]` |
| **API (legacy)** | `c.Api.Ok(users)` | `{"success":true,"result":[{"id":1,"name":"John"}]}` |

### Error Response Comparison

| Layer | Method | Output |
|-------|--------|--------|
| **Base** | `c.Resp.WithStatus(404).Json({"error":"Not found"})` | `{"error":"Not found"}` |
| **API (api)** | `c.Api.NotFound("User not found")` | `{"status":"error","error":{"code":"NOT_FOUND","message":"User not found"}}` |
| **API (simple)** | `c.Api.NotFound("User not found")` | `{"error":"User not found","code":"NOT_FOUND"}` |
| **API (legacy)** | `c.Api.NotFound("User not found")` | `{"success":false,"errorCode":"NOT_FOUND","errorMsg":"User not found"}` |

## 🎚️ **Registry Pattern Usage**

### Set Global Formatter at Startup
```go
func main() {
    // Use corporate format for all API responses
    response.SetApiResponseFormatterByName("corporate")
    
    r := router.New("app")
    // All c.Api.Ok() calls now use corporate format
}
```

### Switch Formatter Per Route
```go
r.GET("/mobile/users", func(c *request.Context) error {
    response.SetApiResponseFormatterByName("mobile")
    return c.Api.Ok(users)
})

r.GET("/web/users", func(c *request.Context) error {
    response.SetApiResponseFormatterByName("api")  
    return c.Api.Ok(users)
})
```

### Client-Based Formatter Selection
```go
func GetUsers(c *request.Context) error {
    clientType := c.Req.Header.Get("X-Client-Type")
    
    switch clientType {
    case "mobile":
        response.SetApiResponseFormatterByName("mobile")
    case "legacy":
        response.SetApiResponseFormatterByName("legacy")
    default:
        response.SetApiResponseFormatterByName("api")
    }
    
    return c.Api.Ok(users)
}
```

## 🔧 **Custom Formatter Registration**

### 1. Implement ResponseFormatter Interface
```go
type CustomFormatter struct{}

func (f *CustomFormatter) Success(data any, message ...string) any {
    return map[string]any{
        "responseCode": "00",
        "payload": data,
        "timestamp": time.Now(),
    }
}

func (f *CustomFormatter) Error(code string, message string, details ...map[string]any) any {
    return map[string]any{
        "responseCode": "99", 
        "errorCode": code,
        "errorMessage": message,
    }
}

// ... implement other methods
```

### 2. Register and Use
```go
func main() {
    // Register custom formatter
    response.RegisterFormatter("corporate", NewCustomFormatter)
    
    // Set as active formatter
    response.SetApiResponseFormatterByName("corporate")
    
    // Now all c.Api.Ok() uses corporate format
}
```

## ✨ **Migration from Old 3-Layer Pattern**

| Old (Removed) | New (Recommended) | Notes |
|---------------|-------------------|-------|
| `c.Api.Ok(data)` | `c.Resp.Json(data)` or `c.Api.Ok(data)` | JSON helpers removed |
| `c.Resp.ErrorNotFound(err)` | `c.Resp.WithStatus(404).Json(...)` or `c.Api.NotFound(...)` | Use base or API layer |
| `c.Api.OkCreated(data)` | `c.Resp.WithStatus(201).Json(data)` or `c.Api.Created(data, msg)` | Use base or API layer |

## 🎯 **When to Use Each Layer**

### **Use Layer 1 (Base Response)** when:
- You need full control over response format
- Working with non-JSON responses (HTML, XML, binary)
- Building custom response structures
- Prototyping or debugging
- Response doesn't fit standard patterns

### **Use Layer 2 (API Response)** when:
- Building REST APIs
- Need consistent error handling
- Want team standardization
- Supporting multiple client types (via formatters)
- Integrating with legacy systems (via custom formatters)

## ✅ **Benefits of Two-Layer Pattern**

1. **Simplified Architecture**: No confusing middle layer
2. **Registry-Based**: Same pattern as router engines  
3. **Configurable**: Switch formatters as needed
4. **Legacy Support**: Custom formatters for existing systems
5. **Clear Separation**: Each layer has distinct, non-overlapping purposes
6. **Team Consistency**: Enforce standards via formatter selection
7. **Runtime Flexibility**: Change formats per request/client