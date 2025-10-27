# Request Logger Middleware

> HTTP request logging with colored output

## Overview

Request Logger middleware logs all incoming HTTP requests with method, path, status code, and duration. It provides colored terminal output for better readability during development.

## Import Path

```go
import "github.com/primadi/lokstra/middleware/request_logger"
```

---

## Configuration

### Config Type

```go
type Config struct {
    EnableColors bool                            // Enable colored output
    SkipPaths    []string                        // Paths to skip logging
    CustomLogger func(format string, args ...any) // Custom logging function
}
```

**Fields:**
- `EnableColors` - If `true`, enables ANSI color codes in output (default: `true`)
- `SkipPaths` - List of paths to skip logging (e.g., health checks)
- `CustomLogger` - Custom logging function (default: `log.Printf`)

---

### DefaultConfig

```go
func DefaultConfig() *Config
```

**Returns:**
```go
&Config{
    EnableColors: true,
    SkipPaths:    []string{},
    CustomLogger: nil, // Uses log.Printf
}
```

---

## Usage

### Basic Usage

```go
router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: true,
}))
```

---

### Skip Health Checks

```go
router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: true,
    SkipPaths: []string{
        "/health",
        "/metrics",
        "/ping",
    },
}))
```

---

### Production (No Colors)

```go
router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: false, // No colors in log files
}))
```

---

### Custom Logger

```go
// Using zap logger
router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: false,
    CustomLogger: func(format string, args ...any) {
        zapLogger.Info(fmt.Sprintf(format, args...))
    },
}))
```

---

## YAML Configuration

```yaml
middlewares:
  - type: request_logger
    params:
      enable_colors: true
      skip_paths:
        - "/health"
        - "/metrics"
```

---

## Output Format

### Colored Output (Terminal)

```
GET /api/users 200 12ms
POST /api/users 201 45ms
PUT /api/users/123 200 23ms
DELETE /api/users/123 204 8ms
GET /api/orders 500 145ms
```

**Colors:**
- Method: Cyan
- Status & Duration: Gray
- Errors highlighted in red (500+)

---

### Plain Output (Log Files)

```
[GET] /api/users - Status: 200 - Duration: 12ms
[POST] /api/users - Status: 201 - Duration: 45ms
[PUT] /api/users/123 - Status: 200 - Duration: 23ms
[DELETE] /api/users/123 - Status: 204 - Duration: 8ms
[GET] /api/orders - Status: 500 - Duration: 145ms
```

---

## Duration Formatting

Durations are automatically formatted for readability:

| Duration | Format | Example |
|----------|--------|---------|
| < 1ms | Microseconds | `850µs` |
| < 1s | Milliseconds | `45ms` |
| ≥ 1s | Seconds | `1.2s` |

---

## Examples

### Development Setup

```go
router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: true,
    SkipPaths:    []string{}, // Log everything
}))
```

---

### Production Setup

```go
router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: false, // No colors in production logs
    SkipPaths: []string{
        "/health",
        "/metrics",
        "/.well-known/**",
    },
}))
```

---

### Skip Static Files

```go
router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: true,
    SkipPaths: []string{
        "/static/**",
        "/assets/**",
        "/favicon.ico",
    },
}))
```

---

### Environment-Based Configuration

```go
cfg := &request_logger.Config{
    EnableColors: os.Getenv("ENV") == "development",
    SkipPaths:    []string{"/health", "/metrics"},
}

if os.Getenv("ENV") == "production" {
    cfg.SkipPaths = append(cfg.SkipPaths, "/.well-known/**")
}

router.Use(request_logger.Middleware(cfg))
```

---

### Structured Logging with Zap

```go
import "go.uber.org/zap"

zapLogger, _ := zap.NewProduction()
defer zapLogger.Sync()

router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: false,
    CustomLogger: func(format string, args ...any) {
        zapLogger.Info(fmt.Sprintf(format, args...))
    },
}))
```

---

### JSON Logging

```go
import "encoding/json"

router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: false,
    CustomLogger: func(format string, args ...any) {
        logEntry := map[string]interface{}{
            "message":   fmt.Sprintf(format, args...),
            "timestamp": time.Now().Format(time.RFC3339),
            "level":     "info",
        }
        json.NewEncoder(os.Stdout).Encode(logEntry)
    },
}))
```

---

### Request ID Tracking

```go
// Custom logger that includes request ID
router.Use(func(c *request.Context) error {
    requestID := c.RequestID
    
    // Store original logger
    originalLogger := request_logger.DefaultConfig().CustomLogger
    
    // Override with request ID
    c.Set("logger", func(format string, args ...any) {
        msg := fmt.Sprintf(format, args...)
        log.Printf("[%s] %s", requestID, msg)
    })
    
    return c.Next()
})

router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: true,
}))
```

---

### Detailed Logging

```go
// Custom logger with more details
router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: false,
    CustomLogger: func(format string, args ...any) {
        log.Printf("[REQUEST] %s | IP: %s | User-Agent: %s",
            fmt.Sprintf(format, args...),
            "client_ip",
            "user_agent",
        )
    },
}))
```

---

## Best Practices

### 1. Place Early in Middleware Chain

```go
// ✅ Good - logs all requests
router.Use(
    recovery.Middleware(&recovery.Config{}),
    request_logger.Middleware(&request_logger.Config{}), // Early
    // ... other middleware
)

// 🚫 Bad - misses some requests
router.Use(
    recovery.Middleware(&recovery.Config{}),
    // ... other middleware
    request_logger.Middleware(&request_logger.Config{}), // Too late
)
```

---

### 2. Skip Health Check Endpoints

```go
// ✅ Good - prevents log spam
request_logger.Middleware(&request_logger.Config{
    SkipPaths: []string{"/health", "/metrics"},
})

// 🚫 Bad - logs every health check
request_logger.Middleware(&request_logger.Config{
    SkipPaths: []string{},
})
```

---

### 3. Disable Colors in Production

```go
// ✅ Good - environment-aware
request_logger.Middleware(&request_logger.Config{
    EnableColors: os.Getenv("ENV") == "development",
})

// 🚫 Bad - colors in log files
request_logger.Middleware(&request_logger.Config{
    EnableColors: true, // Breaks log parsing
})
```

---

### 4. Use Structured Logging in Production

```go
// ✅ Good - machine-readable logs
zapLogger, _ := zap.NewProduction()
request_logger.Middleware(&request_logger.Config{
    CustomLogger: func(format string, args ...any) {
        zapLogger.Info(fmt.Sprintf(format, args...))
    },
})

// 🚫 Bad - unstructured logs
request_logger.Middleware(&request_logger.Config{
    CustomLogger: log.Printf,
})
```

---

### 5. Skip Low-Value Endpoints

```go
// ✅ Good - focus on important requests
request_logger.Middleware(&request_logger.Config{
    SkipPaths: []string{
        "/health",
        "/metrics",
        "/favicon.ico",
        "/.well-known/**",
        "/static/**",
    },
})
```

---

## Performance

**Overhead:** ~1-5μs per request

**Components:**
- Time recording: ~500ns
- Formatting: ~500ns-1μs
- Logging I/O: ~1-3μs

**Impact:** Negligible for most applications

---

## Integration Examples

### With Prometheus Metrics

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "path", "status"},
    )
)

// Custom logger that also records metrics
router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: true,
    CustomLogger: func(format string, args ...any) {
        log.Printf(format, args...)
        // Extract method, path, status, duration from format
        // Record to Prometheus
    },
}))
```

---

### With ELK Stack

```go
import "github.com/sirupsen/logrus"

logrus.SetFormatter(&logrus.JSONFormatter{})

router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: false,
    CustomLogger: func(format string, args ...any) {
        logrus.WithFields(logrus.Fields{
            "service": "api",
            "type":    "http_request",
        }).Info(fmt.Sprintf(format, args...))
    },
}))
```

---

### With Datadog

```go
import "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: false,
    CustomLogger: func(format string, args ...any) {
        // Log to Datadog
        span, _ := tracer.StartSpanFromContext(context.Background(), "http.request")
        defer span.Finish()
        
        log.Printf(format, args...)
    },
}))
```

---

## Testing

### Test Logging Output

```go
func TestRequestLogger(t *testing.T) {
    // Capture log output
    var buf bytes.Buffer
    
    router := lokstra.NewRouter()
    router.Use(request_logger.Middleware(&request_logger.Config{
        EnableColors: false,
        CustomLogger: func(format string, args ...any) {
            fmt.Fprintf(&buf, format, args...)
        },
    }))
    
    router.GET("/test", func(c *request.Context) error {
        return c.Api.Ok("success")
    })
    
    req := httptest.NewRequest("GET", "/test", nil)
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    logOutput := buf.String()
    assert.Contains(t, logOutput, "GET")
    assert.Contains(t, logOutput, "/test")
    assert.Contains(t, logOutput, "200")
}
```

---

### Test Skip Paths

```go
func TestSkipPaths(t *testing.T) {
    var logged bool
    
    router := lokstra.NewRouter()
    router.Use(request_logger.Middleware(&request_logger.Config{
        SkipPaths: []string{"/health"},
        CustomLogger: func(format string, args ...any) {
            logged = true
        },
    }))
    
    router.GET("/health", func(c *request.Context) error {
        return c.Api.Ok("ok")
    })
    
    req := httptest.NewRequest("GET", "/health", nil)
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.False(t, logged, "Health check should not be logged")
}
```

---

## Common Patterns

### Request Count by Path

```go
var requestCounts = make(map[string]int)
var mu sync.Mutex

router.Use(request_logger.Middleware(&request_logger.Config{
    CustomLogger: func(format string, args ...any) {
        log.Printf(format, args...)
        
        // Extract path from format
        // Increment counter
        mu.Lock()
        requestCounts[path]++
        mu.Unlock()
    },
}))
```

---

### Error Request Alerting

```go
router.Use(request_logger.Middleware(&request_logger.Config{
    CustomLogger: func(format string, args ...any) {
        log.Printf(format, args...)
        
        // Check if status is 500+
        if statusCode >= 500 {
            alertService.SendAlert("Server error: " + fmt.Sprintf(format, args...))
        }
    },
}))
```

---

### Request Rate Limiting

```go
var requestTimes []time.Time
var mu sync.Mutex

router.Use(request_logger.Middleware(&request_logger.Config{
    CustomLogger: func(format string, args ...any) {
        log.Printf(format, args...)
        
        mu.Lock()
        requestTimes = append(requestTimes, time.Now())
        
        // Keep only last minute
        cutoff := time.Now().Add(-time.Minute)
        for len(requestTimes) > 0 && requestTimes[0].Before(cutoff) {
            requestTimes = requestTimes[1:]
        }
        
        if len(requestTimes) > 1000 {
            log.Println("ALERT: Request rate exceeds 1000/min")
        }
        mu.Unlock()
    },
}))
```

---

## See Also

- **[Slow Request Logger](./slow-request-logger.md)** - Slow request detection
- **[Recovery](./recovery.md)** - Panic recovery
- **[Request](../01-core-packages/request.md)** - Request handling

---

## Related Guides

- **[Monitoring](../../04-guides/monitoring/)** - Monitoring setup
- **[Logging](../../04-guides/logging/)** - Logging best practices
- **[Observability](../../04-guides/observability/)** - Observability patterns
