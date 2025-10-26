# Recovery Middleware

> Panic recovery and error handling

## Overview

Recovery middleware catches panics in request handlers and returns proper error responses instead of crashing the server. This is essential for production stability.

## Import Path

```go
import "github.com/primadi/lokstra/middleware/recovery"
```

---

## Configuration

### Config Type

```go
type Config struct {
    EnableStackTrace bool                                                  // Include stack trace in response
    EnableLogging    bool                                                  // Log panic details
    CustomHandler    func(*request.Context, any, []byte) error // Custom panic handler
}
```

**Fields:**
- `EnableStackTrace` - If `true`, includes stack trace in error response (for debugging, **disable in production**)
- `EnableLogging` - If `true`, logs panic details to console
- `CustomHandler` - Custom function to handle recovered panics (optional)

---

### DefaultConfig

```go
func DefaultConfig() *Config
```

**Returns:**
```go
&Config{
    EnableStackTrace: false, // Disabled for security
    EnableLogging:    true,
    CustomHandler:    nil,
}
```

---

## Usage

### Basic Usage

```go
router.Use(recovery.Middleware(&recovery.Config{
    EnableStackTrace: false,
    EnableLogging:    true,
}))
```

---

### Development Mode

```go
router.Use(recovery.Middleware(&recovery.Config{
    EnableStackTrace: true, // Show stack traces for debugging
    EnableLogging:    true,
}))
```

---

### Production Mode

```go
router.Use(recovery.Middleware(&recovery.Config{
    EnableStackTrace: false, // Hide internal details
    EnableLogging:    true,
}))
```

---

### Custom Panic Handler

```go
router.Use(recovery.Middleware(&recovery.Config{
    EnableStackTrace: false,
    EnableLogging:    true,
    CustomHandler: func(c *request.Context, recovered any, stack []byte) error {
        // Log to external service
        logger.Error("panic", map[string]any{
            "error":      fmt.Sprint(recovered),
            "stack":      string(stack),
            "path":       c.R.URL.Path,
            "method":     c.R.Method,
            "request_id": c.RequestID,
        })
        
        // Return custom error response
        return c.Api.InternalError("An unexpected error occurred")
    },
}))
```

---

## YAML Configuration

```yaml
middlewares:
  - type: recovery
    params:
      enable_stack_trace: false
      enable_logging: true
```

---

## Examples

### Handler that Panics

```go
router.GET("/panic", func(c *request.Context) error {
    panic("something went wrong!")
    return nil
})

// Request: GET /panic
// Response: 500 Internal Server Error
// {
//   "error": "Internal server error: something went wrong!"
// }
```

---

### Nil Pointer Panic

```go
router.GET("/user/:id", func(c *request.Context) error {
    var user *User = nil
    return c.Api.Ok(user.Name) // Panics: nil pointer dereference
})

// Recovery catches panic and returns 500 error
```

---

### Array Index Panic

```go
router.GET("/items", func(c *request.Context) error {
    items := []string{"a", "b", "c"}
    return c.Api.Ok(items[10]) // Panics: index out of range
})

// Recovery catches panic and returns 500 error
```

---

### Integration with Logging Service

```go
type LoggingService struct {
    logger *zap.Logger
}

func (s *LoggingService) LogPanic(ctx *request.Context, recovered any, stack []byte) {
    s.logger.Error("panic recovered",
        zap.String("error", fmt.Sprint(recovered)),
        zap.String("stack", string(stack)),
        zap.String("path", ctx.R.URL.Path),
        zap.String("method", ctx.R.Method),
        zap.String("request_id", ctx.RequestID),
    )
}

// Use in middleware
loggingService := &LoggingService{logger: zapLogger}

router.Use(recovery.Middleware(&recovery.Config{
    EnableStackTrace: false,
    EnableLogging:    false, // Disable default logging
    CustomHandler: func(c *request.Context, recovered any, stack []byte) error {
        loggingService.LogPanic(c, recovered, stack)
        return c.Api.InternalError("An error occurred")
    },
}))
```

---

### Panic with Context

```go
router.GET("/process", func(c *request.Context) error {
    defer func() {
        if r := recover(); r != nil {
            // Handler-specific panic handling
            log.Printf("process failed: %v", r)
            panic(r) // Re-panic to be caught by recovery middleware
        }
    }()
    
    // Code that might panic
    processData()
    
    return c.Api.Ok("success")
})
```

---

### Alert on Critical Panics

```go
router.Use(recovery.Middleware(&recovery.Config{
    EnableStackTrace: false,
    EnableLogging:    true,
    CustomHandler: func(c *request.Context, recovered any, stack []byte) error {
        errorMsg := fmt.Sprint(recovered)
        
        // Alert on critical errors
        if strings.Contains(errorMsg, "database") || 
           strings.Contains(errorMsg, "connection") {
            alertService.SendCritical("Panic: " + errorMsg)
        }
        
        return c.Api.InternalError("Service temporarily unavailable")
    },
}))
```

---

## Behavior

### Default Error Response

When panic is caught, default response is:

```json
{
  "error": "Internal server error: <panic message>"
}
```

**Status Code:** 500 (Internal Server Error)

---

### Console Logging

When `EnableLogging: true`, logs to console:

```
[PANIC RECOVERY] something went wrong!
goroutine 1 [running]:
runtime/debug.Stack()
    /usr/local/go/src/runtime/debug/stack.go:24 +0x65
github.com/primadi/lokstra/middleware/recovery.Middleware.func1.1()
    /app/middleware/recovery/recovery.go:47 +0x65
panic({0x1234567, 0xc000123456})
    /usr/local/go/src/runtime/panic.go:890 +0x262
...
```

---

## Best Practices

### 1. Place Recovery First

```go
// ✅ Good - catches panics from all middleware
router.Use(
    recovery.Middleware(&recovery.Config{}), // First
    request_logger.Middleware(&request_logger.Config{}),
    // ... other middleware
)

// 🚫 Bad - panics in CORS won't be caught
router.Use(
    cors.Middleware([]string{"*"}),
    recovery.Middleware(&recovery.Config{}), // Too late
)
```

---

### 2. Disable Stack Traces in Production

```go
// ✅ Good
router.Use(recovery.Middleware(&recovery.Config{
    EnableStackTrace: os.Getenv("ENV") == "development",
}))

// 🚫 Bad - exposes internal implementation
router.Use(recovery.Middleware(&recovery.Config{
    EnableStackTrace: true,
}))
```

---

### 3. Always Enable Logging

```go
// ✅ Good - helps debugging
router.Use(recovery.Middleware(&recovery.Config{
    EnableLogging: true,
}))

// 🚫 Bad - panics go unnoticed
router.Use(recovery.Middleware(&recovery.Config{
    EnableLogging: false,
}))
```

---

### 4. Use Custom Handler for Structured Logging

```go
// ✅ Good - structured logs with context
router.Use(recovery.Middleware(&recovery.Config{
    CustomHandler: func(c *request.Context, recovered any, stack []byte) error {
        structuredLogger.Error(map[string]any{
            "error":      fmt.Sprint(recovered),
            "path":       c.R.URL.Path,
            "method":     c.R.Method,
            "request_id": c.RequestID,
            "user_id":    c.Get("user_id"),
        })
        return c.Api.InternalError("Error occurred")
    },
}))

// 🚫 Bad - basic logging loses context
router.Use(recovery.Middleware(&recovery.Config{
    EnableLogging: true,
}))
```

---

### 5. Monitor Panic Frequency

```go
var panicCounter = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_panics_total",
        Help: "Total number of panics recovered",
    },
    []string{"path", "method"},
)

router.Use(recovery.Middleware(&recovery.Config{
    CustomHandler: func(c *request.Context, recovered any, stack []byte) error {
        panicCounter.WithLabelValues(c.R.URL.Path, c.R.Method).Inc()
        log.Printf("[PANIC] %v", recovered)
        return c.Api.InternalError("Error occurred")
    },
}))
```

---

## Common Panic Scenarios

### Nil Pointer Dereference

```go
// Handler
func GetUser(c *request.Context) error {
    var user *User = nil // Forgot to initialize
    return c.Api.Ok(user.Name) // Panic!
}

// Recovery catches: "runtime error: invalid memory address or nil pointer dereference"
```

---

### Index Out of Range

```go
func GetItem(c *request.Context) error {
    items := []string{"a", "b", "c"}
    return c.Api.Ok(items[10]) // Panic!
}

// Recovery catches: "runtime error: index out of range [10] with length 3"
```

---

### Type Assertion

```go
func ProcessData(c *request.Context) error {
    data := c.Get("data")
    str := data.(string) // Panic if data is not string!
    return c.Api.Ok(str)
}

// Recovery catches: "interface conversion: interface {} is int, not string"
```

---

### Division by Zero

```go
func Calculate(c *request.Context) error {
    x := 10
    y := 0
    result := x / y // Panic!
    return c.Api.Ok(result)
}

// Recovery catches: "runtime error: integer divide by zero"
```

---

### Map Concurrent Access

```go
var cache = make(map[string]string)

func CacheGet(c *request.Context) error {
    // Panic if another goroutine writes to cache concurrently
    value := cache["key"]
    return c.Api.Ok(value)
}

// Recovery catches: "fatal error: concurrent map read and map write"
```

---

## Performance

**Overhead:** ~50ns per request (deferred function only)

**Impact:** Negligible - panic recovery uses Go's built-in defer/recover mechanism which has minimal overhead when no panic occurs.

---

## See Also

- **[Request Logger](./request-logger.md)** - Request logging
- **[Request](../01-core-packages/request.md)** - Request handling
- **[Response](../01-core-packages/response.md)** - Response formatting

---

## Related Guides

- **[Error Handling](../../04-guides/error-handling/)** - Error handling patterns
- **[Monitoring](../../04-guides/monitoring/)** - Monitoring and alerting
- **[Production Setup](../../04-guides/production/)** - Production configuration
