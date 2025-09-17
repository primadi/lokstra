# Middleware

Lokstra's middleware system provides a powerful and flexible way to process HTTP requests before they reach your handlers. Middleware can handle authentication, logging, CORS, rate limiting, and much more.

## Understanding Middleware

Middleware functions wrap your handlers to provide cross-cutting functionality. They follow a simple pattern:

```go
type Func = func(next request.HandlerFunc) request.HandlerFunc
```

Each middleware function:
1. Receives the next handler in the chain
2. Returns a new handler that may execute code before and/or after calling next
3. Can short-circuit the request by not calling next

## Basic Middleware Structure

```go
func myMiddleware(next request.HandlerFunc) request.HandlerFunc {
    return func(ctx *request.Context) error {
        // Code before handler execution
        fmt.Println("Before handler")
        
        // Call the next handler
        err := next(ctx)
        
        // Code after handler execution  
        fmt.Println("After handler")
        
        return err
    }
}
```

## Middleware Registration

### Global Registration

Register middleware factories globally for reuse across your application:

```go
// Basic registration with default priority (50)
regCtx.RegisterMiddlewareFactory("cors", func(cfg any) midware.Func {
    return func(next request.HandlerFunc) request.HandlerFunc {
        return func(ctx *request.Context) error {
            // Set CORS headers
            ctx.SetHeader("Access-Control-Allow-Origin", "*")
            ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
            return next(ctx)
        }
    }
})

// Registration with custom priority (lower number = higher priority)
regCtx.RegisterMiddlewareFactoryWithPriority("auth", authFactory, 10)

// Register a simple function (no configuration)
regCtx.RegisterMiddlewareFunc("request-logger", func(next request.HandlerFunc) request.HandlerFunc {
    return func(ctx *request.Context) error {
        start := time.Now()
        method := ctx.GetMethod()
        path := ctx.GetPath()
        
        err := next(ctx)
        
        duration := time.Since(start)
        fmt.Printf("%s %s - %v\n", method, path, duration)
        
        return err
    }
})
```

### Configurable Middleware

Create middleware that accepts configuration:

```go
type CorsConfig struct {
    AllowOrigins []string
    AllowMethods []string
    AllowHeaders []string
}

func corsFactory(cfg any) midware.Func {
    config := cfg.(CorsConfig) // Type assertion
    
    return func(next request.HandlerFunc) request.HandlerFunc {
        return func(ctx *request.Context) error {
            // Apply configuration
            ctx.SetHeader("Access-Control-Allow-Origin", strings.Join(config.AllowOrigins, ", "))
            ctx.SetHeader("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
            ctx.SetHeader("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
            
            // Handle preflight requests
            if ctx.GetMethod() == "OPTIONS" {
                return ctx.OkNoContent()
            }
            
            return next(ctx)
        }
    }
}

// Register with priority
regCtx.RegisterMiddlewareFactoryWithPriority("cors", corsFactory, 30)
```

## Using Middleware

### Router-Level Middleware

Apply middleware to all routes in the router:

```go
// Named middleware
r.Use("cors")
r.Use("request-logger")

// Inline middleware function
r.Use(func(next request.HandlerFunc) request.HandlerFunc {
    return func(ctx *request.Context) error {
        ctx.SetHeader("X-Api-Version", "v1")
        return next(ctx)
    }
})

// All routes will have these middleware applied
r.GET("/users", getUsersHandler)
r.POST("/users", createUserHandler)
```

### Group-Level Middleware

Apply middleware to route groups:

```go
// API group with authentication
api := r.Group("/api", "auth")
api.GET("/profile", getProfileHandler)
api.POST("/settings", updateSettingsHandler)

// Admin group with additional authorization
admin := api.Group("/admin", "admin-role")
admin.GET("/stats", getStatsHandler)
admin.DELETE("/users/{id}", deleteUserHandler)
```

### Route-Level Middleware

Apply middleware to specific routes:

```go
// Single middleware
r.GET("/profile", getProfileHandler, "auth")

// Multiple middleware
r.POST("/admin/sensitive", sensitiveHandler, "auth", "admin-role", "audit")

// Mix named and inline middleware
r.PUT("/settings", updateSettingsHandler, "auth", func(next request.HandlerFunc) request.HandlerFunc {
    return func(ctx *request.Context) error {
        // Custom validation for this route
        if !isValidUpdateRequest(ctx) {
            return ctx.ErrorBadRequest("Invalid update request")
        }
        return next(ctx)
    }
})
```

### Middleware with Configuration

Use configured middleware instances:

```go
import "github.com/primadi/lokstra/core/midware"

// Use middleware with specific configuration
corsConfig := CorsConfig{
    AllowOrigins: []string{"https://example.com", "https://app.example.com"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders: []string{"Content-Type", "Authorization"},
}

r.Use(midware.Named("cors", corsConfig))
```

## Middleware Priority and Execution Order

Middleware execution is controlled by priority and insertion order:

### Priority Rules

- **Lower numbers = Higher priority** (execute earlier)
- Named middleware uses the priority set during registration
- Inline functions default to priority 5000 (execute last)

```go
// These will execute in this order:
regCtx.RegisterMiddlewareFactoryWithPriority("auth", authFactory, 10)        // First
regCtx.RegisterMiddlewareFactoryWithPriority("cors", corsFactory, 30)        // Second  
regCtx.RegisterMiddlewareFactoryWithPriority("logger", loggerFactory, 50)    // Third
```

### Execution Order

Within the same priority level, middleware executes in insertion order:

```go
r.Use("cors")     // Priority 30, order 0
r.Use("logger")   // Priority 50, order 1  
r.Use("auth")     // Priority 10, order 2

// Actual execution order: auth (10,2), cors (30,0), logger (50,1)
```

### Composition Pattern

Middleware wraps from inside-out:

```go
// Given middleware: [A, B, C] and handler H
// Composition: A(B(C(H)))
// Execution: A-before → B-before → C-before → H → C-after → B-after → A-after
```

## Middleware Override

Sometimes you need to bypass inherited middleware:

### Group Override

```go
// Router with global middleware
r.Use("cors")
r.Use("auth")

// Public group that bypasses auth
public := r.WithOverrideMiddleware(true).Group("/public")
public.Use("cors") // Only applies cors, not auth
public.GET("/health", healthHandler)
```

### Route Override

```go
// Route that bypasses all inherited middleware
r.HandleOverrideMiddleware(request.GET, "/metrics", metricsHandler, "basic-auth")
```

## Built-in Middleware Examples

### Authentication Middleware

```go
func authMiddleware(next request.HandlerFunc) request.HandlerFunc {
    return func(ctx *request.Context) error {
        token := ctx.GetHeader("Authorization")
        if token == "" {
            return ctx.ErrorUnauthorized("Authorization header required")
        }
        
        // Remove "Bearer " prefix
        if strings.HasPrefix(token, "Bearer ") {
            token = token[7:]
        }
        
        // Validate token
        claims, err := validateJWT(token)
        if err != nil {
            return ctx.ErrorUnauthorized("Invalid token")
        }
        
        // Store user info in context
        ctx.Set("user_id", claims.UserID)
        ctx.Set("user_role", claims.Role)
        
        return next(ctx)
    }
}
```

### Rate Limiting Middleware

```go
type RateLimitConfig struct {
    RequestsPerMinute int
    BurstSize         int
}

func rateLimitFactory(cfg any) midware.Func {
    config := cfg.(RateLimitConfig)
    limiter := rate.NewLimiter(rate.Limit(config.RequestsPerMinute), config.BurstSize)
    
    return func(next request.HandlerFunc) request.HandlerFunc {
        return func(ctx *request.Context) error {
            if !limiter.Allow() {
                return ctx.ErrorTooManyRequests("Rate limit exceeded")
            }
            
            return next(ctx)
        }
    }
}
```

### Request/Response Logging

```go
func requestLoggerMiddleware(next request.HandlerFunc) request.HandlerFunc {
    return func(ctx *request.Context) error {
        start := time.Now()
        method := ctx.GetMethod()
        path := ctx.GetPath()
        userAgent := ctx.GetHeader("User-Agent")
        
        // Log request
        fmt.Printf("[REQ] %s %s - %s\n", method, path, userAgent)
        
        err := next(ctx)
        
        duration := time.Since(start)
        statusCode := ctx.Response.StatusCode
        
        // Log response
        fmt.Printf("[RES] %s %s - %d (%v)\n", method, path, statusCode, duration)
        
        return err
    }
}
```

### Recovery Middleware

```go
func recoveryMiddleware(next request.HandlerFunc) request.HandlerFunc {
    return func(ctx *request.Context) error {
        defer func() {
            if r := recover(); r != nil {
                // Log the panic
                fmt.Printf("PANIC: %v\n%s\n", r, debug.Stack())
                
                // Return 500 error if response hasn't been written
                if ctx.Response.StatusCode == 0 {
                    ctx.ErrorInternal("Internal server error")
                }
            }
        }()
        
        return next(ctx)
    }
}
```

## Advanced Patterns

### Conditional Middleware

```go
func conditionalAuth(condition func(*request.Context) bool) midware.Func {
    return func(next request.HandlerFunc) request.HandlerFunc {
        return func(ctx *request.Context) error {
            if condition(ctx) {
                // Apply authentication logic
                if err := validateAuth(ctx); err != nil {
                    return err
                }
            }
            return next(ctx)
        }
    }
}

// Use with condition
r.Use(conditionalAuth(func(ctx *request.Context) bool {
    return strings.HasPrefix(ctx.GetPath(), "/admin")
}))
```

### Chain Building

```go
func buildAuthChain() []any {
    return []any{
        "cors",
        "rate-limit", 
        "auth",
        "audit",
    }
}

// Apply chain to group
admin := r.Group("/admin", buildAuthChain()...)
```

### Middleware Composition

```go
func composeMiddleware(middlewares ...midware.Func) midware.Func {
    return func(next request.HandlerFunc) request.HandlerFunc {
        // Apply middleware in reverse order
        handler := next
        for i := len(middlewares) - 1; i >= 0; i-- {
            handler = middlewares[i](handler)
        }
        return handler
    }
}
```

## Reverse Proxy Middleware

Middleware works differently with reverse proxy routes:

```go
// Proxy with middleware
r.MountReverseProxy("/api/external/", "http://localhost:8081", false, "auth", "rate-limit")

// Override mode - only use specified middleware
r.MountReverseProxy("/api/legacy/", "http://legacy.internal", true, "cors")
```

**Important**: For proxy routes:
- **Success**: Upstream writes directly to response (no JSON re-encoding)
- **Error**: Middleware errors return structured JSON responses

## Error Handling in Middleware

```go
func errorHandlingMiddleware(next request.HandlerFunc) request.HandlerFunc {
    return func(ctx *request.Context) error {
        err := next(ctx)
        
        if err != nil {
            // Log the error
            fmt.Printf("Handler error: %v\n", err)
            
            // Don't modify the error, let it propagate
            return err
        }
        
        return nil
    }
}
```

## Testing Middleware

```go
func TestAuthMiddleware(t *testing.T) {
    // Create test handler
    testHandler := func(ctx *request.Context) error {
        return ctx.Ok("success")
    }
    
    // Apply middleware
    wrapped := authMiddleware(testHandler)
    
    // Create test context
    ctx := createTestContext()
    ctx.SetHeader("Authorization", "Bearer valid-token")
    
    // Execute
    err := wrapped(ctx)
    
    // Assert results
    assert.NoError(t, err)
    assert.Equal(t, 200, ctx.Response.StatusCode)
}
```

## Best Practices

### 1. Order Matters

Design middleware order carefully:

```go
// Correct order
regCtx.RegisterMiddlewareFactoryWithPriority("recovery", recoveryFactory, 1)  // First
regCtx.RegisterMiddlewareFactoryWithPriority("cors", corsFactory, 20)
regCtx.RegisterMiddlewareFactoryWithPriority("auth", authFactory, 30)
regCtx.RegisterMiddlewareFactoryWithPriority("logger", loggerFactory, 40)    // Last
```

### 2. Error Handling

Use appropriate error responses:

```go
func authMiddleware(next request.HandlerFunc) request.HandlerFunc {
    return func(ctx *request.Context) error {
        if !isAuthenticated(ctx) {
            return ctx.ErrorUnauthorized("Authentication required")
        }
        
        if !isAuthorized(ctx) {
            return ctx.ErrorForbidden("Insufficient permissions")
        }
        
        return next(ctx)
    }
}
```

### 3. Context Storage

Store middleware data in request context:

```go
func userMiddleware(next request.HandlerFunc) request.HandlerFunc {
    return func(ctx *request.Context) error {
        userID := extractUserID(ctx)
        user, err := getUserByID(userID)
        if err != nil {
            return ctx.ErrorUnauthorized("Invalid user")
        }
        
        // Store for use in handlers
        ctx.Set("current_user", user)
        
        return next(ctx)
    }
}

// Use in handler
func getProfileHandler(ctx *request.Context) error {
    user := ctx.Get("current_user").(*User)
    return ctx.Ok(user)
}
```

### 4. Configuration

Make middleware configurable:

```go
type LoggerConfig struct {
    IncludeHeaders bool
    IncludeBody    bool
    MaxBodySize    int
}

func loggerFactory(cfg any) midware.Func {
    config := cfg.(LoggerConfig)
    
    return func(next request.HandlerFunc) request.HandlerFunc {
        return func(ctx *request.Context) error {
            if config.IncludeHeaders {
                logHeaders(ctx)
            }
            
            if config.IncludeBody {
                logBody(ctx, config.MaxBodySize)
            }
            
            return next(ctx)
        }
    }
}
```

## Troubleshooting

### Common Issues

1. **"Middleware factory 'X' not found"**
   - Ensure middleware is registered before router building
   - Check spelling and registration context

2. **Unexpected execution order**
   - Review middleware priorities
   - Check where middleware is attached (router vs group vs route)

3. **Middleware not executing**
   - Verify middleware is properly attached
   - Check for override middleware settings

4. **Context values not available**
   - Ensure middleware runs before handler
   - Check priority and execution order

## Next Steps

- [Core Concepts](./core-concepts.md) - Understand request/response handling
- [Routing](./routing.md) - Learn about route organization
- [Services](./services.md) - Integrate with dependency injection
- [Configuration](./configuration.md) - Configure middleware via YAML

---

*Middleware is the backbone of cross-cutting concerns in Lokstra applications. Master these patterns to build robust, maintainable web services.*