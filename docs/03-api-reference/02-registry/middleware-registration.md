# Middleware Registration

> Comprehensive guide to middleware registration patterns and factory functions

## Overview

Lokstra provides flexible middleware registration through factory functions, allowing you to create reusable middleware types with different configurations. This guide covers all middleware registration patterns.

## Import Path

```go
import (
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/core/deploy/schema"
    "github.com/primadi/lokstra/core/request"
    "github.com/primadi/lokstra/lokstra_registry"
)
```

---

## Middleware Factory Registration

### RegisterMiddlewareFactory
Registers a middleware factory function for creating middleware instances.

**Signature:**
```go
func RegisterMiddlewareFactory(
    mwType string,
    factory any,
    opts ...RegisterOption,
)
```

**Parameters:**
- `mwType` - Unique identifier for the middleware type
- `factory` - Factory function
- `opts` - Optional settings (e.g., `AllowOverride`)

**Factory Signatures:**
```go
// Modern pattern (returns any)
func(config map[string]any) any

// Old pattern (returns request.HandlerFunc directly)
func(config map[string]any) request.HandlerFunc
```

**Example:**
```go
lokstra_registry.RegisterMiddlewareFactory("logger", func(cfg map[string]any) any {
    level := cfg["level"].(string)
    colorize := cfg["colorize"].(bool)
    
    return func(ctx *request.Context) error {
        if colorize {
            log.Printf("\033[36m[%s] %s %s\033[0m", level, ctx.R.Method, ctx.R.URL.Path)
        } else {
            log.Printf("[%s] %s %s", level, ctx.R.Method, ctx.R.URL.Path)
        }
        return ctx.Next()
    }
})
```

---

### AllowOverride Option
Controls whether existing middleware types can be overridden.

**Signature:**
```go
lokstra_registry.AllowOverride(enable bool) RegisterOption
```

**Example:**
```go
// Default: Panic if already registered
lokstra_registry.RegisterMiddlewareFactory("logger", loggerFactory)

// Allow override
lokstra_registry.RegisterMiddlewareFactory("logger", newLoggerFactory,
    lokstra_registry.AllowOverride(true))
```

---

## Named Middleware Registration

### RegisterMiddlewareName
Registers a named middleware instance with specific configuration.

**Signature:**
```go
func RegisterMiddlewareName(
    mwName string,
    mwType string,
    config map[string]any,
    opts ...RegisterOption,
)
```

**Parameters:**
- `mwName` - Unique name for this middleware instance
- `mwType` - Middleware type (factory name)
- `config` - Configuration for this instance
- `opts` - Optional settings (e.g., `AllowOverride`)

**Example:**
```go
// Register factory
lokstra_registry.RegisterMiddlewareFactory("logger", loggerFactory)

// Register multiple named instances with different configs
lokstra_registry.RegisterMiddlewareName("logger-debug", "logger",
    map[string]any{
        "level":    "DEBUG",
        "colorize": true,
    })

lokstra_registry.RegisterMiddlewareName("logger-info", "logger",
    map[string]any{
        "level":    "INFO",
        "colorize": false,
    })

lokstra_registry.RegisterMiddlewareName("logger-error", "logger",
    map[string]any{
        "level":    "ERROR",
        "colorize": true,
    })
```

**Usage in Router:**
```go
router := lokstra.NewRouter()
router.Use(lokstra_registry.CreateMiddleware("logger-debug"))
router.GET("/users", handlers.GetUsers)
```

---

## Direct Middleware Registration

### RegisterMiddleware
Registers a pre-instantiated middleware directly.

**Signature:**
```go
func RegisterMiddleware(name string, handler request.HandlerFunc)
```

**Example:**
```go
logger := func(ctx *request.Context) error {
    log.Printf("%s %s", ctx.R.Method, ctx.R.URL.Path)
    return ctx.Next()
}

lokstra_registry.RegisterMiddleware("simple-logger", logger)
```

**Use Cases:**
- Quick testing
- Simple middleware without configuration
- One-off middleware instances

---

## Middleware Access

### CreateMiddleware
Creates a middleware from its definition and caches it.

**Signature:**
```go
func CreateMiddleware(name string) request.HandlerFunc
```

**Returns:**
- Middleware handler function

**Example:**
```go
logger := lokstra_registry.CreateMiddleware("logger-debug")
router.Use(logger)
```

**Behavior:**
- First call: Creates middleware using factory
- Subsequent calls: Returns cached instance
- Supports both YAML-defined and code-registered middleware

---

### GetMiddleware
Retrieves a registered middleware instance.

**Signature:**
```go
func GetMiddleware(name string) (request.HandlerFunc, bool)
```

**Returns:**
- `(handler, true)` if found
- `(nil, false)` if not found

**Example:**
```go
if logger, ok := lokstra_registry.GetMiddleware("logger-debug"); ok {
    router.Use(logger)
} else {
    log.Println("Middleware not found")
}
```

---

## YAML-Based Middleware Definition

### DefineMiddleware (YAML)
Define middleware in YAML configuration files.

**Example:**
```yaml
middleware-definitions:
  logger-debug:
    type: logger
    config:
      level: DEBUG
      colorize: true
      
  logger-info:
    type: logger
    config:
      level: INFO
      colorize: false
      
  rate-limiter-strict:
    type: rate-limiter
    config:
      requests_per_minute: 10
      burst: 5
      
  rate-limiter-relaxed:
    type: rate-limiter
    config:
      requests_per_minute: 100
      burst: 20
```

**Usage:**
```go
// Framework auto-loads YAML definitions
logger := lokstra_registry.CreateMiddleware("logger-debug")
rateLimiter := lokstra_registry.CreateMiddleware("rate-limiter-strict")

router.Use(logger, rateLimiter)
```

---

## Complete Examples

### Simple Logger Middleware
```go
package main

import (
    "log"
    "time"
    
    "github.com/primadi/lokstra/core/request"
    "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    // Register factory
    lokstra_registry.RegisterMiddlewareFactory("logger", func(cfg map[string]any) any {
        level := cfg["level"].(string)
        
        return func(ctx *request.Context) error {
            start := time.Now()
            
            log.Printf("[%s] → %s %s", level, ctx.R.Method, ctx.R.URL.Path)
            
            err := ctx.Next()
            
            duration := time.Since(start)
            status := ctx.Resp.RespStatusCode
            
            log.Printf("[%s] ← %s %s %d (%v)", level, ctx.R.Method, ctx.R.URL.Path, status, duration)
            
            return err
        }
    })
    
    // Register named instances
    lokstra_registry.RegisterMiddlewareName("logger-debug", "logger",
        map[string]any{"level": "DEBUG"})
    lokstra_registry.RegisterMiddlewareName("logger-info", "logger",
        map[string]any{"level": "INFO"})
}
```

---

### CORS Middleware
```go
func main() {
    lokstra_registry.RegisterMiddlewareFactory("cors", func(cfg map[string]any) any {
        allowOrigin := cfg["allow_origin"].(string)
        allowMethods := cfg["allow_methods"].(string)
        allowHeaders := cfg["allow_headers"].(string)
        
        return func(ctx *request.Context) error {
            ctx.W.Header().Set("Access-Control-Allow-Origin", allowOrigin)
            ctx.W.Header().Set("Access-Control-Allow-Methods", allowMethods)
            ctx.W.Header().Set("Access-Control-Allow-Headers", allowHeaders)
            
            // Handle preflight
            if ctx.R.Method == "OPTIONS" {
                ctx.W.WriteHeader(204)
                return nil
            }
            
            return ctx.Next()
        }
    })
    
    // Strict CORS for production
    lokstra_registry.RegisterMiddlewareName("cors-prod", "cors",
        map[string]any{
            "allow_origin":  "https://example.com",
            "allow_methods": "GET, POST, PUT, DELETE",
            "allow_headers": "Content-Type, Authorization",
        })
    
    // Permissive CORS for development
    lokstra_registry.RegisterMiddlewareName("cors-dev", "cors",
        map[string]any{
            "allow_origin":  "*",
            "allow_methods": "*",
            "allow_headers": "*",
        })
}
```

---

### Authentication Middleware
```go
func main() {
    lokstra_registry.RegisterMiddlewareFactory("auth", func(cfg map[string]any) any {
        secretKey := cfg["secret_key"].(string)
        skipPaths := cfg["skip_paths"].([]string)
        
        return func(ctx *request.Context) error {
            // Skip auth for certain paths
            for _, path := range skipPaths {
                if ctx.R.URL.Path == path {
                    return ctx.Next()
                }
            }
            
            // Check Authorization header
            token := ctx.Req.HeaderParam("Authorization")
            if token == "" {
                return ctx.Api.Unauthorized("Missing authorization token")
            }
            
            // Validate token
            user, err := validateToken(token, secretKey)
            if err != nil {
                return ctx.Api.Unauthorized("Invalid token")
            }
            
            // Store user in context
            ctx.Set("user", user)
            
            return ctx.Next()
        }
    })
    
    lokstra_registry.RegisterMiddlewareName("auth-jwt", "auth",
        map[string]any{
            "secret_key": "your-secret-key",
            "skip_paths": []string{"/login", "/register", "/health"},
        })
}
```

---

### Rate Limiter Middleware
```go
import (
    "sync"
    "time"
)

type rateLimiter struct {
    mu               sync.Mutex
    requestsPerMin   int
    burst            int
    visitors         map[string]*visitor
    cleanupInterval  time.Duration
}

type visitor struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

func main() {
    lokstra_registry.RegisterMiddlewareFactory("rate-limiter", func(cfg map[string]any) any {
        rpm := cfg["requests_per_minute"].(int)
        burst := cfg["burst"].(int)
        
        rl := &rateLimiter{
            requestsPerMin: rpm,
            burst:          burst,
            visitors:       make(map[string]*visitor),
        }
        
        // Start cleanup goroutine
        go rl.cleanupVisitors()
        
        return func(ctx *request.Context) error {
            ip := ctx.R.RemoteAddr
            
            limiter := rl.getLimiter(ip)
            if !limiter.Allow() {
                return ctx.Api.Error(429, "Rate limit exceeded", nil)
            }
            
            return ctx.Next()
        }
    })
    
    // Strict rate limiting
    lokstra_registry.RegisterMiddlewareName("rate-limiter-strict", "rate-limiter",
        map[string]any{
            "requests_per_minute": 10,
            "burst":               5,
        })
    
    // Relaxed rate limiting
    lokstra_registry.RegisterMiddlewareName("rate-limiter-relaxed", "rate-limiter",
        map[string]any{
            "requests_per_minute": 100,
            "burst":               20,
        })
}
```

---

### Validation Middleware
```go
func main() {
    lokstra_registry.RegisterMiddlewareFactory("validator", func(cfg map[string]any) any {
        requiredHeaders := cfg["required_headers"].([]string)
        requiredQueryParams := cfg["required_query_params"].([]string)
        
        return func(ctx *request.Context) error {
            // Validate headers
            for _, header := range requiredHeaders {
                if ctx.Req.HeaderParam(header) == "" {
                    return ctx.Api.BadRequest("Missing required header: " + header)
                }
            }
            
            // Validate query params
            for _, param := range requiredQueryParams {
                if ctx.Req.QueryParam(param) == "" {
                    return ctx.Api.BadRequest("Missing required query param: " + param)
                }
            }
            
            return ctx.Next()
        }
    })
    
    lokstra_registry.RegisterMiddlewareName("api-validator", "validator",
        map[string]any{
            "required_headers":      []string{"Content-Type", "X-API-Key"},
            "required_query_params": []string{},
        })
}
```

---

### Request ID Middleware
```go
import (
    "github.com/google/uuid"
)

func main() {
    lokstra_registry.RegisterMiddlewareFactory("request-id", func(cfg map[string]any) any {
        headerName := cfg["header_name"].(string)
        
        return func(ctx *request.Context) error {
            // Check if request ID already exists
            requestID := ctx.Req.HeaderParam(headerName)
            if requestID == "" {
                // Generate new request ID
                requestID = uuid.New().String()
            }
            
            // Store in context
            ctx.Set("request_id", requestID)
            
            // Add to response headers
            ctx.W.Header().Set(headerName, requestID)
            
            return ctx.Next()
        }
    })
    
    lokstra_registry.RegisterMiddlewareName("req-id", "request-id",
        map[string]any{
            "header_name": "X-Request-ID",
        })
}
```

---

## Middleware Composition

### Chaining Multiple Middleware
```go
func main() {
    // Register factories
    lokstra_registry.RegisterMiddlewareFactory("logger", loggerFactory)
    lokstra_registry.RegisterMiddlewareFactory("cors", corsFactory)
    lokstra_registry.RegisterMiddlewareFactory("auth", authFactory)
    lokstra_registry.RegisterMiddlewareFactory("rate-limiter", rateLimiterFactory)
    
    // Register named instances
    lokstra_registry.RegisterMiddlewareName("logger-info", "logger", loggerCfg)
    lokstra_registry.RegisterMiddlewareName("cors-prod", "cors", corsCfg)
    lokstra_registry.RegisterMiddlewareName("auth-jwt", "auth", authCfg)
    lokstra_registry.RegisterMiddlewareName("rate-limit", "rate-limiter", rateCfg)
    
    // Apply to router
    router := lokstra.NewRouter()
    router.Use(
        lokstra_registry.CreateMiddleware("logger-info"),
        lokstra_registry.CreateMiddleware("cors-prod"),
        lokstra_registry.CreateMiddleware("rate-limit"),
        lokstra_registry.CreateMiddleware("auth-jwt"),
    )
    
    router.GET("/users", handlers.GetUsers)
}
```

---

### Route-Specific Middleware
```go
func main() {
    router := lokstra.NewRouter()
    
    // Global middleware
    router.Use(lokstra_registry.CreateMiddleware("logger-info"))
    
    // Public routes (no auth)
    router.GET("/health", handlers.Health)
    router.POST("/login", handlers.Login)
    
    // Protected routes (with auth)
    protected := router.Group("/api")
    protected.Use(lokstra_registry.CreateMiddleware("auth-jwt"))
    protected.GET("/users", handlers.GetUsers)
    protected.POST("/orders", handlers.CreateOrder)
    
    // Admin routes (with admin check)
    admin := protected.Group("/admin")
    admin.Use(lokstra_registry.CreateMiddleware("admin-check"))
    admin.GET("/stats", handlers.GetStats)
    admin.DELETE("/users/:id", handlers.DeleteUser)
}
```

---

## Best Practices

### 1. Use Factory Pattern for Configurable Middleware
```go
// ✅ Good: Factory with config
lokstra_registry.RegisterMiddlewareFactory("logger", func(cfg map[string]any) any {
    level := cfg["level"].(string)
    return func(ctx *request.Context) error {
        log.Printf("[%s] %s", level, ctx.R.URL.Path)
        return ctx.Next()
    }
})

// 🚫 Avoid: Hardcoded values
lokstra_registry.RegisterMiddleware("logger", func(ctx *request.Context) error {
    log.Printf("[INFO] %s", ctx.R.URL.Path)
    return ctx.Next()
})
```

---

### 2. Name Middleware Instances Descriptively
```go
// ✅ Good: Clear naming
lokstra_registry.RegisterMiddlewareName("logger-debug", "logger", debugCfg)
lokstra_registry.RegisterMiddlewareName("logger-prod", "logger", prodCfg)
lokstra_registry.RegisterMiddlewareName("cors-strict", "cors", strictCfg)

// 🚫 Avoid: Ambiguous naming
lokstra_registry.RegisterMiddlewareName("logger1", "logger", cfg1)
lokstra_registry.RegisterMiddlewareName("logger2", "logger", cfg2)
```

---

### 3. Always Call Next() Unless Terminating
```go
// ✅ Good: Continue chain
return func(ctx *request.Context) error {
    log.Println("Before handler")
    err := ctx.Next()
    log.Println("After handler")
    return err
}

// 🚫 Avoid: Forgetting Next() (breaks middleware chain)
return func(ctx *request.Context) error {
    log.Println("Before handler")
    return nil // Chain stops here!
}
```

---

### 4. Handle Errors Properly
```go
// ✅ Good: Check Next() errors
return func(ctx *request.Context) error {
    if err := ctx.Next(); err != nil {
        log.Printf("Handler error: %v", err)
        return err
    }
    return nil
}

// 🚫 Avoid: Ignoring errors
return func(ctx *request.Context) error {
    ctx.Next() // Error ignored!
    return nil
}
```

---

### 5. Use YAML for Environment-Specific Config
```yaml
# config/development.yaml
middleware-definitions:
  logger:
    type: logger
    config:
      level: DEBUG
      colorize: true
      
  cors:
    type: cors
    config:
      allow_origin: "*"

# config/production.yaml
middleware-definitions:
  logger:
    type: logger
    config:
      level: INFO
      colorize: false
      
  cors:
    type: cors
    config:
      allow_origin: "https://example.com"
```

---

## See Also

- **[lokstra_registry](./lokstra_registry.md)** - Registry API
- **[Service Registration](./service-registration.md)** - Service patterns
- **[Request](../01-core-packages/request.md)** - Request Context
- **[Router](../01-core-packages/router.md)** - Router middleware

---

## Related Guides

- **[Middleware Essentials](../../01-essentials/03-middleware/)** - Middleware basics
- **[Authentication](../../04-guides/auth/)** - Auth patterns
- **[Error Handling](../../04-guides/error-handling/)** - Error patterns
