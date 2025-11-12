---
layout: default
title: Middleware Inline Parameters
parent: API Reference
nav_order: 10
---

# Middleware Inline Parameters

Starting from **Lokstra v1.1**, you can pass parameters directly to middleware without pre-registering them in configuration.

## Overview

Instead of:
```yaml
middleware-definitions:
  rate-limit-strict:
    type: rate-limit
    config:
      max: 100
      window: "1m"
```

You can now write:
```go
r.GET("/api/users", handler, `rate-limit max="100", window="1m"`)
```

## Syntax

```
"middleware-name param1=value1, param2=value2"
```

**Quotes are optional for simple values:**
```go
// Without quotes (recommended for simple values)
"rate-limit max=100, window=1m"

// With quotes (required for values with spaces or special chars)
"cors origins=\"https://example.com\", methods=\"GET,POST\""

// Mixed (use quotes only when needed)
"auth max=100, secret=\"my-s3cr3t!\", issuer=local"
```

**Rules:**
- Middleware name comes first
- Parameters are separated by space from name
- Format: `key=value` or `key="value"`
- Multiple parameters separated by commas
- Use quotes for values with spaces or special characters
- Quotes are optional for simple alphanumeric values

## Examples

### Basic Usage

```go
import "github.com/primadi/lokstra"

r := lokstra.NewRouter("api")

// No parameters
r.GET("/public", handler, "cors")

// Single parameter (no quotes needed)
r.GET("/limited", handler, "rate-limit max=100")

// Multiple parameters (no quotes needed)
r.GET("/api/users", handler, "rate-limit max=1000, window=1h")

// Values with spaces or special chars (use quotes)
r.GET("/api/search", handler, `cors origins="https://app.example.com, https://admin.example.com"`)

// Complex example (mixed)
r.GET("/auth", handler, `jwt secret="my-super-secret-key", issuer="https://auth.example.com", debug=true`)
```

### With Route-Level Middleware

**Without quotes (recommended):**
```go
// @Route "GET /users/{id}", middlewares=["auth", "rate-limit max=1000, window=1h"]
func (s *UserService) GetByID(p *GetUserParams) (*User, error) {
    return s.UserRepo.MustGet().GetByID(p.ID)
}
```

**With quotes (when value has spaces):**
```go
// @Route "GET /search", middlewares=["auth", "logger prefix=\"Search API\""]
func (s *SearchService) Search(p *SearchParams) ([]*Result, error) {
    return s.search(p)
}
```

**Note:** In annotations, you still need `\"` for values with spaces. For simple values, omit quotes entirely.

### Combining with Registered Middleware

```yaml
# config.yaml
middleware-definitions:
  api-logger:
    type: logger
    config:
      prefix: "API"
      level: "info"
```

```go
// Use registered middleware
r.GET("/info", handler, "api-logger")

// Override specific parameters (no quotes needed)
r.GET("/debug", handler, "api-logger level=debug")

// Add new parameters to registered middleware
r.GET("/custom", handler, `api-logger output="/var/log/custom.log"`)
```

**Behavior:**
- Inline parameters **override** registered config
- Base config is preserved for non-overridden keys
- Inline-only params are added to config

## Built-in Middleware Examples

### CORS

```go
// Allow all origins
r.Use("cors")

// Custom origins (quotes needed for URL with special chars)
r.Use(`cors origins="https://app.example.com,https://admin.example.com"`)

// Simple origin
r.Use("cors origins=http://localhost:3000")
```

### Rate Limit

```go
// 100 requests per minute (no quotes needed)
r.GET("/api/search", handler, "rate-limit max=100, window=1m")

// 1000 requests per hour
r.GET("/api/users", handler, "rate-limit max=1000, window=1h")
```

### Logger

```go
// Default logger
r.Use("request-logger")

// Custom prefix (quotes needed for spaces)
r.Use(`request-logger prefix="Admin API"`)

// Skip specific paths
r.Use("request-logger skip_paths=/health,/metrics")
```

### Body Limit

```go
// 10MB limit (no quotes needed)
r.POST("/upload", handler, "body-limit max_size=10485760")

// Skip for specific paths
r.Use("body-limit max_size=1048576, skip_on_path=/upload/**")
```

## Generated Code

When using `@Route` annotations with inline params (no quotes needed for simple values):

```go
// @Route "GET /users/{id}", middlewares=["auth", "rate-limit max=100"]
func (s *UserService) GetByID(p *GetUserParams) (*User, error) { ... }
```

Generates:

```go
RouteMiddlewares: map[string][]string{
    "GetByID": { "auth", "rate-limit max=100" },
},
```

## Parameter Types

All parameters are passed as **strings** to the middleware factory. The middleware factory is responsible for type conversion:

```go
func rateLimitFactory(config map[string]any) any {
    maxStr, _ := config["max"].(string)
    max, _ := strconv.Atoi(maxStr) // Convert to int

    windowStr, _ := config["window"].(string)
    window, _ := time.ParseDuration(windowStr) // Convert to duration

    return rateLimitMiddleware(max, window)
}
```

## Caching

Middleware instances are **cached** by full name including parameters:

```go
// Creates one instance, cached as: rate-limit max="100", window="1m"
r.GET("/api/users", handler, `rate-limit max="100", window="1m"`)
r.GET("/api/posts", handler, `rate-limit max="100", window="1m"`) // Reuses cached

// Creates different instance, cached as: rate-limit max="50", window="1m"
r.GET("/api/admin", handler, `rate-limit max="50", window="1m"`)
```

## Edge Cases

### Values With Spaces (Must Use Quotes)

```go
// Correct: Use quotes
r.GET("/test", handler, `custom message="Value with spaces"`)

// Wrong: Spaces without quotes will break parsing
r.GET("/test", handler, "custom message=Value with spaces")  // ❌
```

### Escaping Quotes in Values

```go
// Value contains quotes (use backslash)
r.GET("/test", handler, `custom message="Value with \"quotes\""`)
```

### Empty Values

```go
// Empty string value (use quotes)
r.GET("/test", handler, `custom param=""`)
```

### Commas in Values (Must Use Quotes)

```go
// Comma in value (must use quotes)
r.GET("/test", handler, `custom list="item1,item2,item3"`)

// Without quotes, will be parsed as separate parameters
r.GET("/test", handler, "custom list=item1,item2")  // ❌ Breaks: "list=item1" and "item2" 
```

## Best Practices

1. **Omit quotes for simple values**
   ```go
   // Good: Clean and readable
   r.GET("/api/users", handler, "rate-limit max=100, window=1m")
   
   // Avoid: Unnecessary quotes
   r.GET("/api/users", handler, `rate-limit max="100", window="1m"`)
   ```

2. **Use quotes when needed**
   ```go
   // Good: Quotes for spaces
   r.GET("/debug", handler, `logger prefix="Debug API"`)
   
   // Good: Quotes for special characters
   r.GET("/auth", handler, `jwt secret="my-s3cr3t!"`)
   ```

3. **Register middleware for reuse**
   ```yaml
   # config.yaml
   middleware-definitions:
     strict-rate-limit:
       type: rate-limit
       config:
         max: 10
         window: "1m"
   ```
   ```go
   // Good: Reusable
   r.GET("/api/login", handler, "strict-rate-limit")
   r.POST("/api/register", handler, "strict-rate-limit")
   ```

3. **Keep inline params simple**
   ```go
   // Good: Simple and clean
   r.GET("/api/users", handler, "rate-limit max=100")

   // Avoid: Too complex, use YAML instead
   r.GET("/api/users", handler, "rate-limit max=100, window=1m, burst=50, cost=2")
   ```

4. **Validate in factory**
   ```go
   func myMiddlewareFactory(config map[string]any) any {
       maxStr, ok := config["max"].(string)
       if !ok {
           panic("rate-limit: 'max' parameter is required")
       }
       // ... validate and convert
   }
   ```

## Migration Guide

### Before (v1.0)

```yaml
middleware-definitions:
  cors-prod:
    type: cors
    config:
      origins: ["https://app.example.com"]

  rate-limit-api:
    type: rate-limit
    config:
      max: 1000
      window: "1h"
```

```go
r.GET("/api/users", handler, "cors-prod", "rate-limit-api")
```

### After (v1.1+)

```go
// No YAML needed for simple cases (no quotes!)
r.GET("/api/users", handler,
    "cors origins=https://app.example.com",
    "rate-limit max=1000, window=1h")
```

## Summary

**Inline parameters** make middleware configuration more flexible and code more readable for simple use cases, while preserving the option to use YAML for complex, reusable configurations.

**When to use inline params:**
- ✅ One-off route-specific configuration
- ✅ Quick testing/debugging
- ✅ Simple parameter overrides
- ✅ Self-documenting code

**When to use YAML config:**
- ✅ Shared middleware across many routes
- ✅ Complex configuration
- ✅ Environment-specific settings
- ✅ Centralized middleware management
