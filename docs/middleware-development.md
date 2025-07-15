# Middleware Development Guide

This guide explains how to create custom middleware in Lokstra.

## Middleware Function Type

Middleware in Lokstra follows the standard pattern:

```go
type HandlerFunc = func(ctx *Context) error
type MiddlewareFunc = func(next HandlerFunc) HandlerFunc
```

## Middleware Module Interface

To make your middleware configurable and registerable, implement the `MiddlewareModule` interface:

```go
type MiddlewareModule interface {
    Name() string
    Factory(config any) MiddlewareFunc
    Meta() *MiddlewareMeta
}
```

## Creating Custom Middleware

### 1. Basic Middleware Structure

```go
package mymiddleware

import "lokstra"

const NAME = "myapp.mymiddleware"

type MyMiddleware struct{}

func (m *MyMiddleware) Name() string {
    return NAME
}

func (m *MyMiddleware) Meta() *lokstra.MiddlewareMeta {
    return &lokstra.MiddlewareMeta{
        Priority:    50, // 1-100, lower = higher priority
        Description: "My custom middleware for doing something",
        Tags:        []string{"custom", "security"},
    }
}

func (m *MyMiddleware) Factory(config any) lokstra.MiddlewareFunc {
    // Parse configuration
    configMap := make(map[string]any)
    if cfg, ok := config.(map[string]any); ok {
        configMap = cfg
    }
    
    // Extract configuration values
    enabled := true
    if e, ok := configMap["enabled"].(bool); ok {
        enabled = e
    }
    
    // Return middleware function
    return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
        return func(ctx *lokstra.Context) error {
            if !enabled {
                return next(ctx)
            }
            
            // Pre-processing
            // ... your logic here
            
            // Call next middleware/handler
            err := next(ctx)
            
            // Post-processing
            // ... your logic here
            
            return err
        }
    }
}

var _ lokstra.MiddlewareModule = (*MyMiddleware)(nil)

func GetModule() lokstra.MiddlewareModule {
    return &MyMiddleware{}
}
```

### 2. Authentication Middleware Example

```go
package auth

import (
    "lokstra"
    "strings"
)

type AuthMiddleware struct{}

func (a *AuthMiddleware) Name() string {
    return "myapp.auth"
}

func (a *AuthMiddleware) Meta() *lokstra.MiddlewareMeta {
    return &lokstra.MiddlewareMeta{
        Priority:    20, // High priority for security
        Description: "Authentication middleware",
        Tags:        []string{"auth", "security"},
    }
}

func (a *AuthMiddleware) Factory(config any) lokstra.MiddlewareFunc {
    configMap := make(map[string]any)
    if cfg, ok := config.(map[string]any); ok {
        configMap = cfg
    }
    
    secretKey := "default-secret"
    if key, ok := configMap["secret_key"].(string); ok {
        secretKey = key
    }
    
    return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
        return func(ctx *lokstra.Context) error {
            // Check for API key
            apiKey := ctx.Headers.Get("X-API-Key")
            if apiKey == "" {
                return ctx.ErrorUnauthorized("Missing API key")
            }
            
            // Validate API key
            if apiKey != secretKey {
                return ctx.ErrorUnauthorized("Invalid API key")
            }
            
            // Set user context
            ctx.Set("authenticated", true)
            ctx.Set("api_key", apiKey)
            
            return next(ctx)
        }
    }
}
```

### 3. Logging Middleware Example

```go
package logging

import (
    "lokstra"
    "time"
)

type LoggingMiddleware struct{}

func (l *LoggingMiddleware) Name() string {
    return "myapp.logging"
}

func (l *LoggingMiddleware) Meta() *lokstra.MiddlewareMeta {
    return &lokstra.MiddlewareMeta{
        Priority:    90, // Low priority, runs last
        Description: "Request logging middleware",
        Tags:        []string{"logging", "monitoring"},
    }
}

func (l *LoggingMiddleware) Factory(config any) lokstra.MiddlewareFunc {
    return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
        return func(ctx *lokstra.Context) error {
            start := time.Now()
            
            // Log request
            lokstra.Logger.WithField("method", ctx.Method()).
                WithField("path", ctx.Path()).
                WithField("ip", ctx.ClientIP()).
                Infof("Request started")
            
            // Process request
            err := next(ctx)
            
            // Log response
            duration := time.Since(start)
            status := ctx.StatusCode()
            
            lokstra.Logger.WithField("method", ctx.Method()).
                WithField("path", ctx.Path()).
                WithField("status", status).
                WithField("duration", duration).
                Infof("Request completed")
            
            return err
        }
    }
}
```

## Using Your Middleware

### In Code

```go
func main() {
    ctx := lokstra.NewGlobalContext()
    
    // Register middleware module
    ctx.RegisterMiddlewareModule(mymiddleware.GetModule())
    
    app := lokstra.NewApp(ctx, "my-app", ":8080")
    
    // Use middleware globally
    app.Use("myapp.mymiddleware")
    
    // Use middleware on specific routes
    app.GET("/protected", "handler.protected", "myapp.auth")
    
    // Use middleware on route groups
    protectedGroup := app.Group("/api", "myapp.auth")
    protectedGroup.GET("/users", "user.list")
    
    app.Start()
}
```

### In YAML Configuration

```yaml
apps:
  - name: api-app
    address: :8080
    middleware:
      - name: myapp.mymiddleware
        enabled: true
        config:
          secret_key: "my-secret"
    groups:
      - prefix: /api
        middleware:
          - name: myapp.auth
            enabled: true
        routes:
          - method: GET
            path: /users
            handler: user.list
```

## Middleware Patterns

### 1. Early Return Pattern

```go
return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
    return func(ctx *lokstra.Context) error {
        // Check condition
        if !shouldProcess(ctx) {
            return ctx.ErrorForbidden("Access denied")
        }
        
        return next(ctx)
    }
}
```

### 2. Request/Response Modification

```go
return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
    return func(ctx *lokstra.Context) error {
        // Modify request
        ctx.Headers.Set("X-Custom-Header", "value")
        
        // Process
        err := next(ctx)
        
        // Modify response
        ctx.Headers.Set("X-Response-Time", time.Now().String())
        
        return err
    }
}
```

### 3. Error Handling

```go
return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
    return func(ctx *lokstra.Context) error {
        err := next(ctx)
        if err != nil {
            // Log error
            lokstra.Logger.WithError(err).Errorf("Handler error")
            
            // Transform error
            return ctx.ErrorInternal("Something went wrong")
        }
        return nil
    }
}
```

### 4. Context Data

```go
return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
    return func(ctx *lokstra.Context) error {
        // Set context data
        ctx.Set("start_time", time.Now())
        ctx.Set("request_id", generateID())
        
        err := next(ctx)
        
        // Use context data
        startTime := ctx.Get("start_time").(time.Time)
        duration := time.Since(startTime)
        
        return err
    }
}
```

## Best Practices

### 1. Configuration

- Always handle missing configuration gracefully
- Provide sensible defaults
- Validate configuration at startup

### 2. Performance

- Minimize allocations in hot paths
- Cache expensive computations
- Use sync.Pool for reusable objects

### 3. Error Handling

- Don't swallow errors unless intentional
- Provide meaningful error messages
- Log errors appropriately

### 4. Context Usage

- Use `ctx.Set()` and `ctx.Get()` for passing data
- Don't modify the HTTP request/response directly when possible
- Respect context cancellation

### 5. Priority

- Authentication: 10-30 (high priority)
- Rate limiting: 30-40
- General processing: 40-60
- Logging/monitoring: 80-100 (low priority)

## Testing Middleware

```go
func TestMyMiddleware(t *testing.T) {
    middleware := &MyMiddleware{}
    
    // Create test handler
    called := false
    handler := func(ctx *lokstra.Context) error {
        called = true
        return nil
    }
    
    // Create middleware function
    mw := middleware.Factory(map[string]any{
        "enabled": true,
    })
    
    // Wrap handler
    wrappedHandler := mw(handler)
    
    // Create test context
    ctx := createTestContext()
    
    // Execute
    err := wrappedHandler(ctx)
    
    // Assert
    assert.NoError(t, err)
    assert.True(t, called)
}
```

## Examples

See the following examples for reference:
- [Recovery Middleware](../middleware/recovery/) - Panic recovery
- [CORS Middleware](../middleware/cors/) - Cross-origin requests
- [Rate Limit Middleware](../middleware/ratelimit/) - Request rate limiting
- [JWT Auth Middleware](../modules/jwt_auth_basic/) - JWT authentication
