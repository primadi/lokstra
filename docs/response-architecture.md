# Response Architecture: Three-Layer Pattern

Lokstra provides a **three-layer response architecture** that gives developers progressive levels of structure and convenience:

1. **Layer 1: Base Response (`c.Resp`)** - Core methods with maximum control
2. **Layer 2: JSON Helpers (`c.Resp`)** - Convenient methods for common patterns
3. **Layer 3: API Response (`c.Api`)** - Opinionated, structured API responses

## Architecture Overview

```
┌─────────────────────────────────────────┐
│              Context                     │
├─────────────────────────────────────────┤
│  Req: *RequestHelper                    │
│  Resp: *Response          ◄─────────────┼─── Layer 1: Core Methods (Json/Text/Raw)
│          └─ Helpers       ◄─────────────┼─── Layer 2: JSON Convenience (Ok/ErrorNotFound)
│  Api: *ApiHelper          ◄─────────────┼─── Layer 3: Structured API (Ok/ValidationError)
│  W: *ResponseWriter                     │
│  R: *http.Request                       │
└─────────────────────────────────────────┘
```

## Core Components

### Layer 1: Base Response (`*Response` Core Methods)

**Purpose**: Maximum control over HTTP response format and content type.

**Methods**:
- `c.Resp.Json(data)` - Direct JSON response
- `c.Resp.Text(text)` - Plain text response
- `c.Resp.Html(html)` - HTML response
- `c.Resp.Raw(contentType, bytes)` - Binary response  
- `c.Resp.Stream(contentType, writerFunc)` - Streaming response
- `c.Resp.WithStatus(code)` - Set HTTP status code

**Use Cases**:
- Custom response formats
- File downloads and streaming
- Non-JSON APIs (GraphQL, RPC, XML)
- Maximum performance requirements
- Microservice protocols

### Layer 2: JSON Helpers (`*Response` Convenience Methods)

**Purpose**: Convenient JSON responses with common HTTP status codes and minimal structure.

**Methods**:
- `c.Api.Ok(data)` - 200 JSON response
- `c.Api.OkCreated(data)` - 201 JSON response
- `c.Api.OkNoContent()` - 204 No content
- `c.Resp.ErrorBadRequest(err)` - 400 error response
- `c.Resp.ErrorNotFound(err)` - 404 error response
- `c.Resp.ErrorInternal(err)` - 500 error response

**Use Cases**:
- Simple JSON APIs
- Quick development with standard HTTP codes
- Existing systems needing minimal structure
- Lightweight services

### Layer 3: API Response (`*ApiHelper` Structured Methods)

**Purpose**: Consistent, opinionated API responses with rich structure and metadata.

**Methods**:
- `c.Api.Ok(data)` - Success response with data
- `c.Api.OkWithMessage(data, msg)` - Success with message
- `c.Api.Created(data, msg)` - 201 Created response
- `c.Api.OkList(data, meta)` - Paginated list response
- `c.Api.BadRequest(code, msg)` - 400 Bad Request
- `c.Api.NotFound(msg)` - 404 Not Found
- `c.Api.ValidationError(msg, fields)` - Structured validation errors

**Use Cases**:
- REST API development
- Frontend-friendly responses with consistent structure
- Complex validation with field-level errors
- Paginated data with metadata
- Team collaboration with standards

## Response Structure Comparison

### Base Response Output

```go
// Direct array
return c.Resp.Json(users)
```
```json
[
  {"id": 1, "name": "John"},
  {"id": 2, "name": "Jane"}
]
```

### API Response Output

```go
// Wrapped in ApiResponse
return c.Api.Ok(users)
```
```json
{
  "status": "success",
  "data": [
    {"id": 1, "name": "John"},
    {"id": 2, "name": "Jane"}
  ]
}
```

## Implementation Details

### Context Structure
```go
type Context struct {
    Req  *RequestHelper    // Request operations
    Resp *Response         // Direct response control
    Api  *ApiHelper        // Opinionated API responses
    // ... other fields
}
```

### ApiHelper Initialization
```go
func NewContext(w http.ResponseWriter, r *http.Request, handlers []HandlerFunc) *Context {
    resp := &response.Response{}
    
    ctx := &Context{
        Resp: resp,
        Api:  response.NewApiHelper(resp), // Shares same Response instance
        // ...
    }
    return ctx
}
```

## Usage Examples

### Error Handling Patterns

#### Base Response (Custom Format)
```go
func GetUser(c *request.Context) error {
    user, err := findUser(id)
    if err != nil {
        return c.Resp.WithStatus(404).Json(map[string]string{
            "error": "User not found",
            "code": "USER_404",
        })
    }
    return c.Resp.Json(user)
}
```

#### API Response (Standard Format)
```go  
func GetUser(c *request.Context) error {
    user, err := findUser(id)
    if err != nil {
        return c.Api.NotFound("User not found")
    }
    return c.Api.Ok(user)
}
```

### Success Response Patterns

#### Base Response
```go
// Simple success
return c.Resp.Json(data)

// Custom structure
return c.Resp.Json(map[string]any{
    "result": data,
    "timestamp": time.Now(),
    "version": "1.0",
})
```

#### API Response
```go
// Simple success
return c.Api.Ok(data)

// Success with message
return c.Api.OkWithMessage(data, "Operation completed")

// Created resource
return c.Api.Created(data, "User created successfully")
```

### List Response Patterns

#### Base Response
```go
return c.Resp.Json(map[string]any{
    "items": users,
    "total": len(users),
    "page": 1,
})
```

#### API Response
```go
meta := response.CalculateListMeta(page, pageSize, total)
return c.Api.OkList(users, meta)
```

## Configuration and Extensibility

### Replaceable API Response Builder

The `ApiResponse` structure can be **completely customized** by implementing a custom `ApiResponseBuilder` and setting it globally at application startup:

```go
// Custom builder for legacy systems
type LegacyApiBuilder struct{}

func (l *LegacyApiBuilder) BuildSuccess(data any) any {
    return map[string]any{
        "success": true,
        "result":  data,
        "code":    200,
    }
}

func (l *LegacyApiBuilder) BuildError(code, message string) any {
    return map[string]any{
        "success":     false,
        "error_code":  code,
        "error_msg":   message,
        "result":      nil,
    }
}

// Set at application startup
func main() {
    response.SetGlobalApiBuilder(&LegacyApiBuilder{})
    
    // All c.Api calls now use legacy format
    app := lokstra.NewApp("legacy-api", ":8080")
    // ...
}
```

### Builder Interface

All custom builders must implement `ApiResponseBuilder`:

```go
type ApiResponseBuilder interface {
    BuildSuccess(data any) any
    BuildSuccessWithMessage(data any, message string) any
    BuildCreated(data any, message string) any
    BuildList(data any, meta *ListMeta) any
    BuildListWithMeta(data any, meta *Meta) any
    BuildError(code, message string) any
    BuildErrorWithDetails(code, message string, details map[string]any) any
    BuildValidationError(message string, fields []FieldError) any
}
```

### Middleware Integration

Both response patterns work seamlessly with middleware:

```go
func LoggingMiddleware(c *request.Context) error {
    start := time.Now()
    
    err := c.Next()
    
    duration := time.Since(start)
    log.Printf("Request took %v", duration)
    
    return err
}
```

## Migration Strategy

### Existing Code
No breaking changes - existing `c.Resp` code continues to work.

### Gradual Adoption
```go
// Old approach (still works)
return c.Resp.WithStatus(200).Json(response.NewSuccess(data))

// New approach (cleaner)
return c.Api.Ok(data)
```

### Team Guidelines
- **New projects**: Use `c.Api` for consistency
- **Legacy projects**: Mix both as needed during migration
- **Microservices**: Choose based on integration requirements

## Benefits

### Developer Experience
- **Choice**: Pick the right tool for the job
- **Consistency**: API helper ensures uniform responses
- **Flexibility**: Base response for special cases
- **Migration**: No forced breaking changes

### Performance
- **Zero overhead**: `c.Api` builds on top of `c.Resp`
- **Same execution path**: Both use identical underlying Response
- **No double serialization**: Direct JSON encoding

### Maintainability  
- **Clear intent**: Response pattern shows API design philosophy
- **Testability**: Both patterns easily unit testable
- **Documentation**: Self-documenting through method names

## Best Practices

1. **Consistency within modules**: Stick to one pattern per logical module
2. **Document choice**: Make team decisions explicit in README
3. **Error standardization**: Use `c.Api` for user-facing APIs
4. **Performance optimization**: Use `c.Resp` for high-throughput endpoints
5. **Frontend integration**: Use `c.Api` for web/mobile client APIs

This dual pattern approach provides the flexibility to build both opinionated REST APIs and custom protocol implementations within the same framework.