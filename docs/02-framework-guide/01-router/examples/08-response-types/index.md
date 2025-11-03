# Response Types - Complete Guide

> **Master all response types in Lokstra: JSON, HTML, Text, XML, CSV, Binary, Streams**

This example demonstrates **all available response types** and **multiple methods** to create each type.

## Quick Reference

| Type | Helper Constructor | Manual Method | Use Case |
|------|-------------------|---------------|----------|
| JSON | `NewJsonResponse()` | `r.Json()` | APIs, data responses |
| **API Success** | `NewApiOk()` | `api.Ok()` | **REST APIs (structured)** |
| **API Error** | `NewApiBadRequest()` | `api.Error()` | **API error responses** |
| HTML | `NewHtmlResponse()` | `r.Html()` | Web pages, templates |
| Text | `NewTextResponse()` | `r.Text()` | Plain text, logs |
| Raw | `NewRawResponse()` | `r.Raw()` | CSV, XML, custom types |
| Stream | `NewStreamResponse()` | `r.Stream()` | SSE, chunked transfer |

---

## API Helper Responses (3 Methods)

### Method 1: NewApiOk() - **RECOMMENDED for REST APIs**
```go
func Handler() *response.ApiHelper {
    return response.NewApiOk(map[string]any{
        "user": "John",
        "role": "admin",
    })
}
```

**Output**:
```json
{
  "success": true,
  "data": {
    "user": "John",
    "role": "admin"
  }
}
```

**Pros**: Structured format, consistent API responses  
**When**: REST APIs, client-facing APIs

### Method 2: NewApiOkWithMessage()
```go
func Handler() *response.ApiHelper {
    return response.NewApiOkWithMessage(
        map[string]string{"id": "123"},
        "User created successfully",
    )
}
```

**Output**:
```json
{
  "success": true,
  "message": "User created successfully",
  "data": {"id": "123"}
}
```

### Method 3: Error Helpers
```go
// 400 Bad Request
return response.NewApiBadRequest("INVALID_INPUT", "Email is required")

// 401 Unauthorized
return response.NewApiUnauthorized("Please login first")

// 404 Not Found
return response.NewApiNotFound("User not found")

// 500 Internal Error
return response.NewApiInternalError("Database connection failed")
```

### Complete API Helper Constructors

| Constructor | Status | Description |
|-------------|--------|-------------|
| `NewApiOk(data)` | 200 | Success response |
| `NewApiOkWithMessage(data, msg)` | 200 | Success with message |
| `NewApiCreated(data, msg)` | 201 | Resource created |
| `NewApiOkList(data, meta)` | 200 | Paginated list |
| `NewApiBadRequest(code, msg)` | 400 | Bad request |
| `NewApiUnauthorized(msg)` | 401 | Unauthorized |
| `NewApiForbidden(msg)` | 403 | Forbidden |
| `NewApiNotFound(msg)` | 404 | Not found |
| `NewApiInternalError(msg)` | 500 | Internal error |
| `NewApiValidationError(msg, fields)` | 400 | Validation errors |

---

## JSON Responses (3 Methods)

### Method 1: NewJsonResponse() - **RECOMMENDED**
```go
func Handler() *response.Response {
    return response.NewJsonResponse(map[string]any{
        "message": "Quick and clean",
        "status": "success",
    })
}
```

**Pros**: One-liner, clean, readable  
**When**: Most JSON responses

### Method 2: Manual with Chainable Methods
```go
func Handler() *response.Response {
    r := response.NewResponse()
    r.WithStatus(200).Json(map[string]any{
        "message": "More control",
    })
    return r
}
```

**Pros**: Fine-grained control, custom status  
**When**: Need to set custom headers or status codes

### Method 3: Using ctx.Resp
```go
func Handler(ctx *request.Context) error {
    return ctx.Resp.WithStatus(200).Json(map[string]any{
        "message": "With context",
        "path": ctx.R.URL.Path,
    })
}
```

**Pros**: Access to request context  
**When**: Need request details in response

**Result**: All three produce identical JSON output

---

## HTML Responses

### Quick HTML
```go
func Handler() *response.Response {
    html := `<h1>Hello World</h1>`
    return response.NewHtmlResponse(html)
}
```

### Dynamic HTML with Context
```go
func Handler(ctx *request.Context) *response.Response {
    name := ctx.R.URL.Query().Get("name")
    html := fmt.Sprintf("<h1>Hello, %s!</h1>", name)
    return response.NewHtmlResponse(html)
}
```

**Use cases**:
- Server-rendered pages
- Error pages
- Admin dashboards
- Simple web interfaces

---

## Text Responses

### Plain Text
```go
func Handler() *response.Response {
    return response.NewTextResponse("Plain text content")
}
```

### Dynamic Text
```go
func Handler(ctx *request.Context) *response.Response {
    info := fmt.Sprintf("Method: %s\nPath: %s", 
        ctx.R.Method, ctx.R.URL.Path)
    return response.NewTextResponse(info)
}
```

**Use cases**:
- Log outputs
- Debug information
- Simple status pages
- Plain text APIs

---

## Raw Responses (Custom Content-Type)

### CSV Response
```go
func Handler() *response.Response {
    csv := "name,age,city\nJohn,30,Jakarta"
    return response.NewRawResponse("text/csv", []byte(csv))
}
```

### XML Response
```go
func Handler() *response.Response {
    xml := `<?xml version="1.0"?><data>content</data>`
    return response.NewRawResponse("application/xml", []byte(xml))
}
```

### Binary Response (PDF, Images, etc.)
```go
func Handler() *response.Response {
    data := readPDFFile() // your binary data
    r := response.NewRawResponse("application/pdf", data)
    r.RespHeaders["Content-Disposition"] = []string{
        "attachment; filename=document.pdf",
    }
    return r
}
```

**Use cases**:
- File downloads
- Custom formats (CSV, XML, YAML)
- Images, PDFs, documents
- Any non-JSON content

---

## Stream Responses

### Server-Sent Events (SSE)
```go
func Handler() *response.Response {
    return response.NewStreamResponse("text/event-stream", 
        func(w http.ResponseWriter) error {
            for i := 1; i <= 5; i++ {
                fmt.Fprintf(w, "data: Event %d\n\n", i)
                w.(http.Flusher).Flush()
                time.Sleep(1 * time.Second)
            }
            return nil
        })
}
```

### Chunked Transfer
```go
func Handler() *response.Response {
    return response.NewStreamResponse("text/plain", 
        func(w http.ResponseWriter) error {
            for i := 1; i <= 10; i++ {
                fmt.Fprintf(w, "Chunk %d\n", i)
                w.(http.Flusher).Flush()
                time.Sleep(500 * time.Millisecond)
            }
            return nil
        })
}
```

**Use cases**:
- Real-time updates (SSE)
- Progress tracking
- Log streaming
- Large file generation
- Live data feeds

---

## Complete Response API

### Available Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `NewResponse()` | `*Response` | Empty response |
| `NewJsonResponse(data)` | `*Response` | JSON response |
| `NewHtmlResponse(html)` | `*Response` | HTML response |
| `NewTextResponse(text)` | `*Response` | Text response |
| `NewRawResponse(type, data)` | `*Response` | Raw bytes with content-type |
| `NewStreamResponse(type, fn)` | `*Response` | Streaming response |
| **`NewApiOk(data)`** | **`*ApiHelper`** | **API success response** |
| **`NewApiOkWithMessage(data, msg)`** | **`*ApiHelper`** | **API success with message** |
| **`NewApiCreated(data, msg)`** | **`*ApiHelper`** | **201 Created** |
| **`NewApiBadRequest(code, msg)`** | **`*ApiHelper`** | **400 Bad Request** |
| **`NewApiUnauthorized(msg)`** | **`*ApiHelper`** | **401 Unauthorized** |
| **`NewApiForbidden(msg)`** | **`*ApiHelper`** | **403 Forbidden** |
| **`NewApiNotFound(msg)`** | **`*ApiHelper`** | **404 Not Found** |
| **`NewApiInternalError(msg)`** | **`*ApiHelper`** | **500 Internal Error** |

### Chainable Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `WithStatus(code)` | `*Response` | Set HTTP status code |
| `Json(data)` | `error` | Write JSON (finalize) |
| `Html(html)` | `error` | Write HTML (finalize) |
| `Text(text)` | `error` | Write text (finalize) |
| `Raw(type, bytes)` | `error` | Write raw (finalize) |
| `Stream(type, fn)` | `error` | Write stream (finalize) |

**Note**: Methods that return `error` are finalizers - use them last in the chain or with helper constructors.

---

## Best Practices

### ✅ DO

```go
// Use helper constructors for simple cases
return response.NewJsonResponse(data)
return response.NewHtmlResponse(html)

// Use API helpers for REST APIs (RECOMMENDED)
return response.NewApiOk(data)
return response.NewApiBadRequest("INVALID", "Bad input")

// Chain methods for control
r := response.NewResponse()
r.WithStatus(201).Json(data)
return r

// Use ctx.Resp when you need context
return ctx.Resp.WithStatus(200).Json(data)
```

### ❌ DON'T

```go
// Don't use manual method when helper exists
api := response.NewApiHelper()
api.Ok(data)
return api
// USE: return response.NewApiOk(data)

// Don't return error from Json() directly as *Response
func Wrong() *response.Response {
    r := response.NewResponse()
    return r.Json(data) // WRONG: Json() returns error, not *Response
}

// Don't manually set fields when helpers exist
r := response.NewResponse()
r.RespData = data // Use r.Json(data) instead
r.RespContentType = "application/json"
```

---

## Running the Example

```bash
cd docs/02-deep-dive/01-router/examples/08-response-types
go run main.go

# Open browser
open http://localhost:8080

# Or use test.http file with VS Code REST Client
```

### Test Routes

```http
GET http://localhost:8080/json/quick       # JSON response
GET http://localhost:8080/html/dynamic     # HTML response
GET http://localhost:8080/text/plain       # Text response
GET http://localhost:8080/raw/csv          # CSV download
GET http://localhost:8080/raw/xml          # XML response
GET http://localhost:8080/stream/sse       # Server-Sent Events
```

---

## Performance Comparison

| Method | Overhead | When to Use |
|--------|----------|-------------|
| `NewJsonResponse()` | Minimal | Most cases (recommended) |
| Manual `r.Json()` | +10ns | Need status/header control |
| `ctx.Resp.Json()` | +20ns | Need request context |

**Recommendation**: Use helper constructors unless you need fine-grained control.

---

## Key Takeaways

✅ **5 response types**: JSON, HTML, Text, Raw, Stream  
✅ **3 methods per type**: Helper, manual, context  
✅ **Helper constructors**: Clean, readable, recommended  
✅ **Chainable methods**: More control when needed  
✅ **Context helper**: Use when you need request details  
✅ **Streaming**: Perfect for SSE and real-time data  
✅ **Raw responses**: Support any content-type (CSV, XML, PDF, etc.)

---

## See Also

- [All Handler Forms](../01-all-handler-forms/) - 69 handler signatures
- [Parameter Binding](../02-parameter-binding/) - Request parameter handling
- [Error Handling](../05-error-handling/) - Error response patterns
