# 📋 Response Pattern FAQ & Complete Guide

This document answers common questions about Lokstra's dual response pattern and provides comprehensive guidance.

## ❓ **Frequently Asked Questions**

### 1. **"Ada c.Api.Ok, ada c.Api.Ok, dll. Apakah sudah dijelaskan bedanya dalam dokumentasi?"**

**Jawaban**: **TIDAK ada lagi `c.Api.Ok`!** Lokstra sekarang menggunakan **2-layer pattern** yang lebih sederhana:

#### Layer 1 - `c.Resp` Methods (Base Response - Unopinionated):
- `c.Resp.Json(data)` - Direct JSON response
- `c.Resp.Text(text)` - Plain text response  
- `c.Resp.Html(html)` - HTML response
- `c.Resp.Raw(contentType, bytes)` - Binary response
- `c.Resp.Stream(contentType, fn)` - Streaming response
- `c.Resp.WithStatus(code)` - Set HTTP status

#### Layer 2 - `c.Api` Methods (Configurable API Response via Registry):
- `c.Api.Ok(data)` - Success response using configured formatter
- `c.Api.OkWithMessage(data, msg)` - Success with message
- `c.Api.Created(data, msg)` - 201 Created response
- `c.Api.NotFound(msg)` - 404 Not Found error
- `c.Api.ValidationError(msg, fields)` - Structured validation errors

**Perbedaan Utama**:
```go
// Layer 1 - Direct output (full control)
return c.Resp.Json(users)
// Output: [{"id":1,"name":"John"}]

// Layer 2 - Configurable format (via registry)
return c.Api.Ok(users)
// Output depends on active formatter:

// 'api' formatter: {"status":"success","data":[{"id":1,"name":"John"}]}
// 'simple' formatter: [{"id":1,"name":"John"}]  
// 'legacy' formatter: {"success":true,"result":[{"id":1,"name":"John"}]}
```

**Registry Pattern**:
```go
// Switch formatter at startup or runtime
response.SetApiResponseFormatterByName("legacy")
return c.Api.Ok(users) // Uses legacy format
```

### 2. **"Apakah sudah ada list lengkap dari isi Api dan Resp di dokumentasi?"**

**Jawaban**: **YA**, sudah ada dokumentasi lengkap di:

- 📄 **`docs/method-reference.md`** - Complete method comparison
- 📄 **`docs/response-architecture.md`** - Architecture overview
- 📄 **`cmd/examples/10-response-patterns/`** - Working examples

#### Complete Method Reference:

| Category | `c.Resp` Methods | `c.Api` Methods |
|----------|------------------|-----------------|
| **Success** | `Json(data)` | `Ok(data)`, `OkWithMessage()`, `Created()` |
| **Lists** | `Json(array)` | `OkList(data, meta)`, `OkListWithMeta()` |
| **Errors** | `WithStatus(code).Json(error)` | `BadRequest()`, `NotFound()`, `ValidationError()` |
| **Other** | `Text()`, `Html()`, `Raw()`, `Stream()` | N/A (use c.Resp) |

### 3. **"Apakah Api object bisa diganti? Misalnya Lokstra digunakan di legacy system yang punya opini berbeda dengan Api structure?"**

**Jawaban**: **YA**, Api object bisa diganti sepenuhnya! Implementasi menggunakan **Builder Pattern**:

#### Cara Mengganti API Response Format:

```go
// 1. Buat custom builder
type LegacySystemBuilder struct{}

func (l *LegacySystemBuilder) BuildSuccess(data any) any {
    return map[string]any{
        "success": true,
        "result":  data,
        "timestamp": time.Now(),
    }
}

func (l *LegacySystemBuilder) BuildError(code, message string) any {
    return map[string]any{
        "success":     false,
        "error_code":  code,
        "error_msg":   message,
        "timestamp":   time.Now(),
    }
}

// 2. Set builder at startup
func main() {
    response.SetGlobalApiBuilder(&LegacySystemBuilder{})
    
    // Semua c.Api calls sekarang menggunakan format legacy
    app := lokstra.NewApp("legacy-api", ":8080")
    // ...
}

// 3. Handler code TIDAK berubah
func GetUsers(c *request.Context) error {
    // Same code, different output format
    return c.Api.Ok(users)
}
```

#### Format Output Comparison:

**Default Lokstra Format**:
```json
{
  "status": "success",
  "data": [{"id": 1, "name": "John"}]
}
```

**Custom Legacy Format** (same `c.Api.Ok()` call):
```json
{
  "success": true,
  "result": [{"id": 1, "name": "John"}],
  "timestamp": "2024-10-01T10:00:00Z"
}
```

## 📚 **Complete Documentation Index**

### Core Documentation
1. **`docs/response-architecture.md`** - Overall architecture explanation
2. **`docs/method-reference.md`** - Complete method comparison with examples
3. **`docs/api-standard.md`** - API standards and PagingRequest usage

### Examples
1. **`cmd/examples/09-api-standard/`** - API standards with PagingRequest
2. **`cmd/examples/10-response-patterns/`** - c.Resp vs c.Api comparison
3. **`cmd/examples/11-custom-api-builder/`** - Custom API response formats

## 🎯 **Decision Matrix: When to Use What**

| Scenario | Use `c.Resp` | Use `c.Api` | Custom Builder |
|----------|--------------|-------------|----------------|
| **New REST API** | ❌ | ✅ | Maybe |
| **Legacy Integration** | ✅ | ❌ | ✅ |
| **File Downloads** | ✅ | ❌ | ❌ |
| **Streaming Data** | ✅ | ❌ | ❌ |
| **Consistent Team API** | ❌ | ✅ | Maybe |
| **Microservice Protocol** | ✅ | ❌ | ❌ |
| **GraphQL/RPC** | ✅ | ❌ | ❌ |
| **Frontend-friendly API** | ❌ | ✅ | ✅ |

## 🔧 **Advanced Use Cases**

### 1. Multi-Version API Support
```go
// Different formats for different API versions
func SetupVersionedAPIs() {
    // v1 - Legacy format
    response.SetGlobalApiBuilder(&LegacyV1Builder{})
    v1Router.GET("/v1/users", GetUsersV1)
    
    // v2 - Modern format  
    response.SetGlobalApiBuilder(&ModernV2Builder{})
    v2Router.GET("/v2/users", GetUsersV2)
}
```

### 2. Conditional Response Format
```go
// Choose format based on request
func SmartApiMiddleware(c *request.Context) error {
    clientType := c.Req.HeaderParam("Client-Type", "web")
    
    switch clientType {
    case "mobile":
        response.SetGlobalApiBuilder(&MobileOptimizedBuilder{})
    case "legacy":
        response.SetGlobalApiBuilder(&LegacyCompatBuilder{})
    default:
        response.SetGlobalApiBuilder(&DefaultBuilder{})
    }
    
    return c.Next()
}
```

### 3. A/B Testing Different Response Formats
```go
// Test different response structures
func ABTestMiddleware(c *request.Context) error {
    if isTestGroup(c.Req.HeaderParam("User-ID", "")) {
        response.SetGlobalApiBuilder(&ExperimentalBuilder{})
    }
    
    return c.Next()
}
```

## ✅ **Key Benefits Summary**

### Developer Experience
- **No method confusion** - clear distinction between `c.Resp` and `c.Api`
- **Complete flexibility** - use either approach as needed
- **Replaceable API format** - adapt to any existing system
- **Comprehensive documentation** - all methods and patterns covered

### Migration & Integration  
- **Zero breaking changes** - existing code continues to work
- **Legacy system support** - custom builders for any API format
- **Gradual adoption** - mix both patterns during transition
- **Team autonomy** - each team can choose their approach

### Consistency & Standards
- **Unified method calls** - same `c.Api` methods regardless of output format
- **Configurable standards** - define API format once, use everywhere
- **Type-safe builders** - interface ensures all methods implemented
- **Runtime flexibility** - change formats based on request context

## 📖 **Next Steps**

1. **Read**: `docs/method-reference.md` for complete method list
2. **Try**: `cmd/examples/10-response-patterns/` for hands-on comparison  
3. **Implement**: Custom builder using `cmd/examples/11-custom-api-builder/` as template
4. **Decide**: Choose pattern based on your use case and existing systems

This comprehensive approach ensures Lokstra can adapt to any existing system while providing modern, consistent API development patterns for new projects.