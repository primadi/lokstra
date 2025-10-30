# Middleware

> Built-in middleware collection for request processing

## Overview

Lokstra provides a comprehensive collection of built-in middleware for common HTTP concerns including security, logging, compression, and error handling. All middleware can be configured programmatically or via YAML configuration.

## Available Middleware

| Middleware | Purpose | Priority |
|------------|---------|----------|
| **[Recovery](./recovery)** | Panic recovery | First |
| **[Request Logger](./request-logger)** | Request logging | Early |
| **[Slow Request Logger](./slow-request-logger)** | Slow request detection | Early |
| **[CORS](./cors)** | Cross-origin handling | Early |
| **[Body Limit](./body-limit)** | Request size protection | Before parsing |
| **[Gzip Compression](./gzip-compression)** | Response compression | Late |
| **[JWT Auth](./jwt-auth)** | Authentication | Before handlers |
| **[Access Control](./access-control)** | Authorization | After JWT Auth |

---

## Quick Start

### Programmatic Configuration

```go
import (
    "github.com/primadi/lokstra/middleware/recovery"
    "github.com/primadi/lokstra/middleware/request_logger"
    "github.com/primadi/lokstra/middleware/cors"
    "github.com/primadi/lokstra/middleware/body_limit"
    "github.com/primadi/lokstra/middleware/gzipcompression"
)

// Apply middleware to router
router.Use(
    recovery.Middleware(&recovery.Config{
        EnableStackTrace: false,
    }),
    request_logger.Middleware(&request_logger.Config{
        EnableColors: true,
    }),
    cors.Middleware([]string{"*"}),
    body_limit.Middleware(&body_limit.Config{
        MaxSize: 10 * 1024 * 1024, // 10MB
    }),
    gzipcompression.Middleware(&gzipcompression.Config{
        MinSize: 1024, // 1KB
    }),
)
```

---

### YAML Configuration

```yaml
middlewares:
  - type: recovery
    params:
      enable_stack_trace: false
      enable_logging: true
  
  - type: request_logger
    params:
      enable_colors: true
      skip_paths: ["/health", "/metrics"]
  
  - type: slow_request_logger
    params:
      threshold: 500  # milliseconds
      enable_colors: true
  
  - type: cors
    params:
      allow_origins: ["*"]
  
  - type: body_limit
    params:
      max_size: 10485760  # 10MB
      skip_on_path: ["/upload/**"]
  
  - type: gzip_compression
    params:
      min_size: 1024
      compression_level: -1
  
  - type: jwtauth
    params:
      validator_service_name: auth_validator
      token_header: Authorization
      token_prefix: "Bearer "
      skip_paths: ["/auth/login", "/auth/register"]
  
  - type: accesscontrol
    params:
      allowed_roles: ["admin", "manager"]
      role_field: role
```

---

## Middleware Order

The order of middleware is critical for proper functionality:

```go
// Recommended order:
router.Use(
    // 1. Recovery - catch all panics from other middleware
    recovery.Middleware(&recovery.Config{}),
    
    // 2. Request Logger - log all requests
    request_logger.Middleware(&request_logger.Config{}),
    
    // 3. Slow Request Logger - detect performance issues
    slow_request_logger.Middleware(&slow_request_logger.Config{
        Threshold: 500 * time.Millisecond,
    }),
    
    // 4. CORS - handle preflight requests early
    cors.Middleware([]string{"*"}),
    
    // 5. Body Limit - protect memory before parsing
    body_limit.Middleware(&body_limit.Config{
        MaxSize: 10 * 1024 * 1024,
    }),
    
    // 6. JWT Auth - authenticate requests
    jwtauth.Middleware(&jwtauth.Config{
        ValidatorServiceName: "auth_validator",
    }),
    
    // 7. Access Control - check permissions
    accesscontrol.Middleware(&accesscontrol.Config{
        AllowedRoles: []string{"admin"},
    }),
    
    // 8. Gzip Compression - compress responses (last)
    gzipcompression.Middleware(&gzipcompression.Config{
        MinSize: 1024,
    }),
)
```

**Why this order?**

1. **Recovery first** - Catches panics from all other middleware
2. **Logging early** - Records all requests, even failed ones
3. **CORS early** - Handles preflight before authentication
4. **Body limit before parsing** - Prevents memory exhaustion
5. **Auth before handlers** - Protects endpoints
6. **Access control after auth** - Requires user info from JWT
7. **Compression last** - Compresses final response

---

## Registry Integration

### Registering Middleware

```go
import (
    "github.com/primadi/lokstra/middleware/recovery"
    "github.com/primadi/lokstra/middleware/cors"
    // ... other middleware
)

func init() {
    // Register all middleware
    recovery.Register()
    cors.Register()
    body_limit.Register()
    gzipcompression.Register()
    request_logger.Register()
    slow_request_logger.Register()
    jwtauth.Register()
    accesscontrol.Register()
}
```

---

### Using from Registry

```go
// Get middleware by name
recoveryMw := lokstra_registry.GetMiddleware("recovery", map[string]any{
    "enable_stack_trace": false,
})

corsMw := lokstra_registry.GetMiddleware("cors", map[string]any{
    "allow_origins": []string{"*"},
})

// Apply to router
router.Use(recoveryMw, corsMw)
```

---

## Common Patterns

### Production Setup

```go
// Production-optimized configuration
router.Use(
    recovery.Middleware(&recovery.Config{
        EnableStackTrace: false, // Hide stack traces
        EnableLogging:    true,
    }),
    request_logger.Middleware(&request_logger.Config{
        EnableColors: false, // No colors in logs
        SkipPaths:    []string{"/health", "/metrics"},
    }),
    slow_request_logger.Middleware(&slow_request_logger.Config{
        Threshold: 1 * time.Second,
        EnableColors: false,
    }),
    cors.Middleware([]string{
        "https://app.example.com",
        "https://admin.example.com",
    }),
    body_limit.Middleware(&body_limit.Config{
        MaxSize: 5 * 1024 * 1024, // 5MB max
    }),
    gzipcompression.Middleware(&gzipcompression.Config{
        MinSize:          1024,
        CompressionLevel: gzip.BestSpeed,
    }),
)
```

---

### Development Setup

```go
// Development-friendly configuration
router.Use(
    recovery.Middleware(&recovery.Config{
        EnableStackTrace: true, // Show stack traces for debugging
        EnableLogging:    true,
    }),
    request_logger.Middleware(&request_logger.Config{
        EnableColors: true, // Colored terminal output
        SkipPaths:    []string{},
    }),
    slow_request_logger.Middleware(&slow_request_logger.Config{
        Threshold:    200 * time.Millisecond, // Lower threshold
        EnableColors: true,
    }),
    cors.Middleware([]string{"*"}), // Allow all origins
)
```

---

### Selective Middleware

```go
// Global middleware
router.Use(
    recovery.Middleware(&recovery.Config{}),
    request_logger.Middleware(&request_logger.Config{}),
)

// Protected group with authentication
apiGroup := router.Group("/api")
apiGroup.Use(
    jwtauth.Middleware(&jwtauth.Config{
        ValidatorServiceName: "auth_validator",
    }),
)

// Admin-only group
adminGroup := apiGroup.Group("/admin")
adminGroup.Use(
    accesscontrol.RequireAdmin(),
)
```

---

### Custom Error Messages

```go
router.Use(
    body_limit.Middleware(&body_limit.Config{
        MaxSize:    10 * 1024 * 1024,
        Message:    "File too large. Maximum size is 10MB",
        StatusCode: http.StatusRequestEntityTooLarge,
    }),
    jwtauth.Middleware(&jwtauth.Config{
        ValidatorServiceName: "auth_validator",
        ErrorMessage:         "Invalid or expired session. Please login again",
    }),
    accesscontrol.Middleware(&accesscontrol.Config{
        AllowedRoles: []string{"admin", "manager"},
        ErrorMessage: "You don't have permission to access this resource",
    }),
)
```

---

### Path-Based Configuration

```go
router.Use(
    body_limit.Middleware(&body_limit.Config{
        MaxSize: 1 * 1024 * 1024, // 1MB default
        SkipOnPath: []string{
            "/upload/**",   // Skip limit for uploads
            "/import/**",   // Skip for imports
        },
    }),
    jwtauth.Middleware(&jwtauth.Config{
        ValidatorServiceName: "auth_validator",
        SkipPaths: []string{
            "/auth/login",
            "/auth/register",
            "/health",
            "/public/**",
        },
    }),
    request_logger.Middleware(&request_logger.Config{
        SkipPaths: []string{
            "/health",
            "/metrics",
            "/.well-known/**",
        },
    }),
)
```

---

## Performance Considerations

### Middleware Overhead

| Middleware | Overhead | Notes |
|------------|----------|-------|
| Recovery | ~50ns | Minimal, deferred only |
| Request Logger | ~1-5μs | Time recording + formatting |
| CORS | ~500ns | Header checks |
| Body Limit | ~100ns | Wrapper allocation |
| Gzip | ~50-500μs | Depends on response size |
| JWT Auth | ~1-10ms | Token validation + DB lookup |
| Access Control | ~100ns | Role check only |

---

### Optimization Tips

**1. Skip unnecessary paths:**
```go
request_logger.Middleware(&request_logger.Config{
    SkipPaths: []string{"/health", "/metrics"}, // Skip frequent health checks
})
```

**2. Use appropriate compression:**
```go
gzipcompression.Middleware(&gzipcompression.Config{
    MinSize: 1024, // Don't compress small responses
    ExcludedContentTypes: []string{
        "image/jpeg",
        "video/mp4", // Already compressed
    },
})
```

**3. Set realistic body limits:**
```go
body_limit.Middleware(&body_limit.Config{
    MaxSize: 5 * 1024 * 1024, // Lower limit = better protection
})
```

**4. Enable gzip only when beneficial:**
```go
// Don't compress API responses < 1KB
gzipcompression.Middleware(&gzipcompression.Config{
    MinSize: 1024,
})
```

---

## Testing

### Testing with Middleware

```go
func TestHandlerWithMiddleware(t *testing.T) {
    // Create router with middleware
    router := lokstra.NewRouter()
    router.Use(
        recovery.Middleware(&recovery.Config{}),
        request_logger.Middleware(&request_logger.Config{}),
    )
    
    // Add test handler
    router.GET("/test", func(c *request.Context) error {
        return c.Api.Ok("success")
    })
    
    // Test request
    req := httptest.NewRequest("GET", "/test", nil)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, 200, rec.Code)
}
```

---

### Mocking JWT Auth

```go
func TestProtectedEndpoint(t *testing.T) {
    // Create mock validator
    mockValidator := &MockValidator{
        ValidateFunc: func(ctx context.Context, token string) (*auth.TokenClaims, error) {
            return &auth.TokenClaims{
                UserID: "user123",
            }, nil
        },
    }
    
    // Register mock
    lokstra_registry.RegisterService("auth_validator", mockValidator)
    
    // Create router with JWT middleware
    router := lokstra.NewRouter()
    router.Use(jwtauth.Middleware(&jwtauth.Config{
        ValidatorServiceName: "auth_validator",
    }))
    
    // Test with token
    req := httptest.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer test-token")
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, 200, rec.Code)
}
```

---

## Creating Custom Middleware

### Basic Pattern

```go
package mymiddleware

import "github.com/primadi/lokstra/core/request"

type Config struct {
    Option1 string
    Option2 int
}

func Middleware(cfg *Config) request.HandlerFunc {
    return request.HandlerFunc(func(c *request.Context) error {
        // Pre-processing
        
        // Call next handler
        err := c.Next()
        
        // Post-processing
        
        return err
    })
}
```

---

### With Factory Function

```go
func MiddlewareFactory(params map[string]any) request.HandlerFunc {
    cfg := &Config{
        Option1: utils.GetValueFromMap(params, "option1", "default"),
        Option2: utils.GetValueFromMap(params, "option2", 10),
    }
    return Middleware(cfg)
}

func Register() {
    lokstra_registry.RegisterMiddlewareFactory("mymiddleware", MiddlewareFactory,
        lokstra_registry.AllowOverride(true))
}
```

---

## Best Practices

### 1. Always Use Recovery First
```go
// ✅ Good
router.Use(
    recovery.Middleware(&recovery.Config{}), // First
    // ... other middleware
)

// 🚫 Bad
router.Use(
    cors.Middleware([]string{"*"}),
    recovery.Middleware(&recovery.Config{}), // Too late
)
```

---

### 2. Skip Health Checks from Logging
```go
// ✅ Good
request_logger.Middleware(&request_logger.Config{
    SkipPaths: []string{"/health", "/metrics"},
})

// 🚫 Bad - logs every health check
request_logger.Middleware(&request_logger.Config{})
```

---

### 3. Use Appropriate Body Limits
```go
// ✅ Good - different limits for different endpoints
body_limit.Middleware(&body_limit.Config{
    MaxSize: 1 * 1024 * 1024, // 1MB default
    SkipOnPath: []string{"/upload/**"}, // Higher limit elsewhere
})

// 🚫 Bad - one size for all
body_limit.Middleware(&body_limit.Config{
    MaxSize: 100 * 1024 * 1024, // 100MB everywhere
})
```

---

### 4. Disable Stack Traces in Production
```go
// ✅ Good
recovery.Middleware(&recovery.Config{
    EnableStackTrace: os.Getenv("ENV") == "development",
})

// 🚫 Bad - exposes internal details
recovery.Middleware(&recovery.Config{
    EnableStackTrace: true,
})
```

---

### 5. Configure CORS Properly
```go
// ✅ Good - specific origins in production
allowedOrigins := []string{"*"}
if os.Getenv("ENV") == "production" {
    allowedOrigins = []string{
        "https://app.example.com",
        "https://admin.example.com",
    }
}
cors.Middleware(allowedOrigins)

// 🚫 Bad - wildcard in production
cors.Middleware([]string{"*"})
```

---

## See Also

- **[Router](../01-core-packages/router)** - Router configuration
- **[Request](../01-core-packages/request)** - Request handling
- **[Response](../01-core-packages/response)** - Response formatting
- **[Registry](../02-registry/)** - Middleware registration

---

## Related Guides

- **[Security Best Practices](../../04-guides/security/)** - Security patterns
- **[Performance Optimization](../../04-guides/performance/)** - Performance tips
- **[Testing Middleware](../../04-guides/testing/)** - Testing strategies
