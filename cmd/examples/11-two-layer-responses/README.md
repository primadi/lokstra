# Two-Layer Response Pattern

This example demonstrates Lokstra's **simplified 2-layer response pattern** with **registry-based formatters**.

## 🎯 Architecture Overview

### Layer 1: Base Response (Unopinionated)
- **Purpose**: Maximum control and flexibility
- **Methods**: `c.Resp.Json()`, `c.Resp.Text()`, `c.Resp.Html()`, `c.Resp.Raw()`
- **Use Case**: Custom formats, non-JSON responses, full control needed

### Layer 2: API Response (Configurable)
- **Purpose**: Consistent API structure with configurable formats
- **Methods**: `c.Api.Ok()`, `c.Api.Created()`, `c.Api.ValidationError()`
- **Use Case**: REST APIs, consistent response format, error handling

## 🔧 Registry Pattern

The Layer 2 API response uses **formatter registry** similar to router engines:

```go
// Built-in formatters
response.RegisterFormatter("api", NewApiResponseFormatter)      // Default structured
response.RegisterFormatter("simple", NewSimpleResponseFormatter) // Minimal JSON
response.RegisterFormatter("legacy", NewLegacyResponseFormatter) // Legacy systems

// Switch formatters
response.SetApiResponseFormatterByName("legacy")
return c.Api.Ok(data) // Uses legacy format
```

## 🚀 Quick Start

```bash
go run main.go
```

Test different endpoints:

```bash
# Layer 1: Base Response
curl http://localhost:8080/base/users     # Direct JSON
curl http://localhost:8080/base/health    # Plain text

# Layer 2: API Response (default 'api' formatter)
curl http://localhost:8080/api/users      # Structured format
curl http://localhost:8080/api/users/404  # Structured error

# Dynamic formatter switching
curl http://localhost:8080/api/simple     # Switches to simple format
curl http://localhost:8080/api/legacy     # Switches to legacy format
curl http://localhost:8080/api/structured # Back to structured format
```

## 📄 Response Format Comparison

### Layer 1: Base Response
```json
[
  {"id": 1, "name": "John Doe", "email": "john@example.com"}
]
```

### Layer 2: API Response (api formatter)
```json
{
  "status": "success",
  "data": [
    {"id": 1, "name": "John Doe", "email": "john@example.com"}
  ]
}
```

### Layer 2: API Response (simple formatter)
```json
[
  {"id": 1, "name": "John Doe", "email": "john@example.com"}
]
```

### Layer 2: API Response (legacy formatter)
```json
{
  "success": true,
  "result": [
    {"id": 1, "name": "John Doe", "email": "john@example.com"}
  ]
}
```

## 🎚️ When to Use Each Layer

**Layer 1 (Base Response)**:
- Custom response formats
- Non-JSON responses (HTML, XML, binary)
- Full control over structure
- Prototyping and debugging

**Layer 2 (API Response)**:
- REST API development
- Consistent error handling
- Team standardization
- Legacy system integration (via custom formatters)

## ✨ Benefits

1. **Simplified**: Only 2 layers instead of confusing 3-layer pattern
2. **Configurable**: Registry pattern allows format switching
3. **Extensible**: Register custom formatters for specific needs
4. **Compatible**: Legacy systems supported via custom formatters
5. **Clear**: Each layer has distinct, non-overlapping purposes