# Custom Response Formatters

This example demonstrates how to **register and use custom response formatters** with Lokstra's registry pattern.

## 🎯 Overview

Instead of the confusing 3-layer approach, Lokstra now uses a **simple 2-layer pattern** with **configurable formatters**:

1. **Layer 1**: Base Response (`c.Resp`) - Unopinionated, full control
2. **Layer 2**: API Response (`c.Api`) - Structured, configurable via registry

## 🔧 Custom Formatter Implementation

### Step 1: Implement ResponseFormatter Interface

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

### Step 2: Register Your Formatter

```go
func main() {
    // Register at application startup
    response.RegisterFormatter("corporate", NewCustomFormatter)
    
    // Set as global formatter
    response.SetApiResponseFormatterByName("corporate")
    
    // Now all c.Api.Ok() calls use corporate format
}
```

### Step 3: Use in Your Handlers

```go
func GetUsers(c *request.Context) error {
    // Uses currently set formatter (corporate in this example)
    return c.Api.Ok(users)
}
```

## 🏢 Built-in Formatters

| Name | Description | Use Case |
|------|-------------|-----------|
| `api` | Structured API format (default) | Modern REST APIs |
| `simple` | Minimal JSON format | Simple APIs, prototypes |
| `legacy` | Legacy system format | Backward compatibility |

## 🚀 Quick Start

```bash
go run main.go
```

Test different formatters:

```bash
# Default formatter
curl http://localhost:8080/default/users

# Corporate formatter  
curl http://localhost:8080/corporate/users

# Mobile formatter
curl http://localhost:8080/mobile/users

# Reset to default
curl http://localhost:8080/reset
```

## 📄 Response Format Examples

### Default Formatter (api)
```json
{
  "status": "success",
  "data": [
    {"id": 1, "name": "John Doe"}
  ]
}
```

### Corporate Formatter
```json
{
  "responseCode": "00",
  "responseStatus": "SUCCESS", 
  "payload": [
    {"id": 1, "name": "John Doe"}
  ],
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Mobile Formatter  
```json
{
  "ok": true,
  "data": [
    {"id": 1, "name": "John Doe"}
  ]
}
```

## 🎚️ Formatter Switching Strategies

### 1. Global at Startup
```go
func main() {
    response.SetApiResponseFormatterByName("corporate")
    // All API responses use corporate format
}
```

### 2. Per Route/Handler
```go
r.GET("/mobile/users", func(c *request.Context) error {
    response.SetApiResponseFormatterByName("mobile")
    return c.Api.Ok(users)
})
```

### 3. Middleware-Based
```go
func CorporateFormatMiddleware(c *request.Context) error {
    response.SetApiResponseFormatterByName("corporate")
    return c.Next()
}

r.Use("/api/v1/*", CorporateFormatMiddleware)
```

### 4. Client-Based
```go
func GetUsers(c *request.Context) error {
    clientType := c.Req.Header.Get("X-Client-Type")
    
    switch clientType {
    case "mobile":
        response.SetApiResponseFormatterByName("mobile")
    case "corporate":
        response.SetApiResponseFormatterByName("corporate")
    default:
        response.SetApiResponseFormatterByName("api")
    }
    
    return c.Api.Ok(users)
}
```

## ✨ Benefits

1. **Simplified Architecture**: Only 2 layers, no confusion
2. **Registry Pattern**: Same pattern as router engines
3. **Runtime Flexibility**: Switch formatters as needed
4. **Legacy Support**: Maintain existing API contracts
5. **Team Standards**: Enforce consistent response formats
6. **Environment-Specific**: Different formats for different clients

## 🔄 Migration from Old 3-Layer

**Before (Confusing 3-layer)**:
```go
c.Api.Ok(data)           // Layer 2 - JSON Helper (removed!)
c.Api.Ok(data)           // Layer 3 - Structured API
```

**After (Clean 2-layer)**:
```go
c.Resp.Json(data)        // Layer 1 - Base Response  
c.Api.Ok(data)           // Layer 2 - Configurable API (via registry)
```

The old `c.Api.Ok()` JSON helpers have been **removed** to eliminate confusion!