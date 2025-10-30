# Slow Request Logger Middleware

> Detect and log slow requests

## Overview

Slow Request Logger middleware identifies and logs requests that exceed a configurable duration threshold. This helps detect performance issues and bottlenecks.

## Import Path

```go
import "github.com/primadi/lokstra/middleware/slow_request_logger"
```

---

## Configuration

### Config Type

```go
type Config struct {
    Threshold    time.Duration                   // Minimum duration to log
    EnableColors bool                            // Enable colored output
    SkipPaths    []string                        // Paths to skip logging
    CustomLogger func(format string, args ...any) // Custom logging function
}
```

**Fields:**
- `Threshold` - Minimum request duration to log (default: 500ms)
- `EnableColors` - If `true`, enables colored terminal output (default: `true`)
- `SkipPaths` - List of paths to skip logging
- `CustomLogger` - Custom logging function (default: `log.Printf`)

---

### DefaultConfig

```go
func DefaultConfig() *Config
```

**Returns:**
```go
&Config{
    Threshold:    500 * time.Millisecond,
    EnableColors: true,
    SkipPaths:    []string{},
    CustomLogger: nil, // Uses log.Printf
}
```

---

## Usage

### Basic Usage

```go
router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 500 * time.Millisecond, // Log requests > 500ms
}))
```

---

### Strict Threshold

```go
// Log requests slower than 200ms
router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 200 * time.Millisecond,
}))
```

---

### Relaxed Threshold

```go
// Only log very slow requests (> 2 seconds)
router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 2 * time.Second,
}))
```

---

### Skip Specific Paths

```go
router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 500 * time.Millisecond,
    SkipPaths: []string{
        "/health",
        "/metrics",
        "/long-running/**", // Expect slow responses
    },
}))
```

---

## YAML Configuration

```yaml
middlewares:
  - type: slow_request_logger
    params:
      threshold: 500  # milliseconds
      enable_colors: true
      skip_paths:
        - "/health"
        - "/metrics"
```

**Note:** In YAML, threshold is in **milliseconds** (integer), automatically converted to `time.Duration`.

---

## Output Format

### Colored Output (Terminal)

```
[SLOW REQUEST] GET /api/search - Status: 200 - Duration: 1.2s (threshold: 500ms)
[SLOW REQUEST] POST /api/process - Status: 200 - Duration: 850ms (threshold: 500ms)
[SLOW REQUEST] GET /api/report - Status: 200 - Duration: 3.5s (threshold: 500ms)
```

**Colors:**
- Yellow: Slow (1x-2x threshold)
- Red: Very slow (>2x threshold)

---

### Plain Output (Log Files)

```
[SLOW REQUEST] [GET] /api/search - Status: 200 - Duration: 1.2s (threshold: 500ms)
[SLOW REQUEST] [POST] /api/process - Status: 200 - Duration: 850ms (threshold: 500ms)
[SLOW REQUEST] [GET] /api/report - Status: 200 - Duration: 3.5s (threshold: 500ms)
```

---

## Severity Levels

Middleware automatically highlights extra-slow requests:

| Duration | Color | Severity |
|----------|-------|----------|
| < Threshold | Not logged | Normal |
| 1x-2x Threshold | Yellow | Slow |
| > 2x Threshold | Red | Very Slow |

**Example with 500ms threshold:**
- 450ms → Not logged
- 700ms → Yellow (slow)
- 1200ms → Red (very slow)

---

## Examples

### Development Monitoring

```go
// Strict threshold for development
router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold:    200 * time.Millisecond, // Stricter in dev
    EnableColors: true,
}))
```

---

### Production Monitoring

```go
// Relaxed threshold for production
router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold:    1 * time.Second, // More lenient in prod
    EnableColors: false,           // No colors in logs
}))
```

---

### Environment-Based Configuration

```go
var threshold time.Duration

if os.Getenv("ENV") == "development" {
    threshold = 200 * time.Millisecond // Strict
} else {
    threshold = 1 * time.Second // Relaxed
}

router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold:    threshold,
    EnableColors: os.Getenv("ENV") == "development",
}))
```

---

### Skip Long-Running Operations

```go
router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 500 * time.Millisecond,
    SkipPaths: []string{
        "/export/**",   // Exports are expected to be slow
        "/import/**",   // Imports are expected to be slow
        "/batch/**",    // Batch operations
        "/reports/**",  // Report generation
    },
}))
```

---

### Alert on Slow Requests

```go
router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 1 * time.Second,
    CustomLogger: func(format string, args ...any) {
        msg := fmt.Sprintf(format, args...)
        log.Println(msg)
        
        // Alert if very slow
        if strings.Contains(msg, "Duration: 3") || 
           strings.Contains(msg, "Duration: 4") {
            alertService.SendWarning("Very slow request: " + msg)
        }
    },
}))
```

---

### Track Slow Endpoints

```go
var slowPaths = make(map[string]int)
var mu sync.Mutex

router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 500 * time.Millisecond,
    CustomLogger: func(format string, args ...any) {
        log.Printf(format, args...)
        
        // Extract path and count
        // (parse from format string)
        mu.Lock()
        slowPaths[path]++
        mu.Unlock()
    },
}))

// Periodic report
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        mu.Lock()
        log.Printf("Slow paths summary: %+v", slowPaths)
        slowPaths = make(map[string]int) // Reset
        mu.Unlock()
    }
}()
```

---

### Integration with Metrics

```go
import "github.com/prometheus/client_golang/prometheus"

var slowRequestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "slow_requests_total",
        Help: "Total number of slow requests",
    },
    []string{"method", "path"},
)

router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 500 * time.Millisecond,
    CustomLogger: func(format string, args ...any) {
        log.Printf(format, args...)
        
        // Increment Prometheus counter
        slowRequestsTotal.WithLabelValues(method, path).Inc()
    },
}))
```

---

### Structured Logging

```go
import "go.uber.org/zap"

zapLogger, _ := zap.NewProduction()

router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold:    500 * time.Millisecond,
    EnableColors: false,
    CustomLogger: func(format string, args ...any) {
        // Parse format string to extract fields
        zapLogger.Warn("slow request",
            zap.String("method", method),
            zap.String("path", path),
            zap.Int("status", statusCode),
            zap.Duration("duration", duration),
            zap.Duration("threshold", threshold),
        )
    },
}))
```

---

## Best Practices

### 1. Use with Regular Request Logger

```go
// ✅ Good - both loggers for complete picture
router.Use(
    request_logger.Middleware(&request_logger.Config{
        SkipPaths: []string{"/health"},
    }),
    slow_request_logger.Middleware(&slow_request_logger.Config{
        Threshold: 500 * time.Millisecond,
    }),
)

// 🚫 Bad - only slow requests logged
router.Use(
    slow_request_logger.Middleware(&slow_request_logger.Config{
        Threshold: 500 * time.Millisecond,
    }),
)
```

---

### 2. Set Appropriate Threshold

```go
// ✅ Good - realistic threshold
slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 500 * time.Millisecond, // Reasonable
})

// 🚫 Bad - too strict, logs everything
slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 10 * time.Millisecond, // Too low
})
```

---

### 3. Skip Expected Slow Operations

```go
// ✅ Good - skip intentionally slow endpoints
slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 500 * time.Millisecond,
    SkipPaths: []string{
        "/export/**",
        "/import/**",
        "/reports/**",
    },
})

// 🚫 Bad - alerts on expected slow operations
slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 500 * time.Millisecond,
    SkipPaths: []string{}, // Logs exports/imports
})
```

---

### 4. Use Different Thresholds for Different Endpoints

```go
// ✅ Good - endpoint-specific thresholds
apiRouter := router.Group("/api")
apiRouter.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 200 * time.Millisecond, // Strict for API
}))

reportsRouter := router.Group("/reports")
reportsRouter.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 5 * time.Second, // Relaxed for reports
}))
```

---

### 5. Monitor and Adjust Threshold

```go
// ✅ Good - start strict, adjust based on metrics
// Week 1: 200ms → too many logs
// Week 2: 500ms → reasonable
// Week 3: 500ms → keep monitoring
```

---

## Performance

**Overhead:** ~1-2μs per request (time recording only)

**Impact:** Negligible - only logs when threshold is exceeded

---

## Common Use Cases

### API Performance Monitoring

```go
// Monitor API response times
apiRouter := router.Group("/api")
apiRouter.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 300 * time.Millisecond,
}))
```

---

### Database Query Optimization

```go
// Identify slow database queries
router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 500 * time.Millisecond,
    CustomLogger: func(format string, args ...any) {
        msg := fmt.Sprintf(format, args...)
        log.Println(msg)
        
        // Log query details if available
        if ctx := getCurrentContext(); ctx != nil {
            log.Printf("Query: %s", ctx.Get("sql_query"))
        }
    },
}))
```

---

### Third-Party Service Monitoring

```go
// Track slow external API calls
proxyRouter := router.Group("/proxy")
proxyRouter.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 1 * time.Second,
}))
```

---

### SLA Monitoring

```go
// Alert if SLA is breached
router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 2 * time.Second, // SLA: 2s
    CustomLogger: func(format string, args ...any) {
        msg := fmt.Sprintf(format, args...)
        log.Println(msg)
        
        // Alert SLA breach
        alertService.SendCritical("SLA breached: " + msg)
    },
}))
```

---

## Testing

### Test Slow Request Detection

```go
func TestSlowRequest(t *testing.T) {
    var logged bool
    
    router := lokstra.NewRouter()
    router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
        Threshold: 100 * time.Millisecond,
        CustomLogger: func(format string, args ...any) {
            logged = true
        },
    }))
    
    router.GET("/slow", func(c *request.Context) error {
        time.Sleep(200 * time.Millisecond) // Slow handler
        return c.Api.Ok("done")
    })
    
    req := httptest.NewRequest("GET", "/slow", nil)
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.True(t, logged, "Slow request should be logged")
}
```

---

### Test Fast Request (Not Logged)

```go
func TestFastRequest(t *testing.T) {
    var logged bool
    
    router := lokstra.NewRouter()
    router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
        Threshold: 500 * time.Millisecond,
        CustomLogger: func(format string, args ...any) {
            logged = true
        },
    }))
    
    router.GET("/fast", func(c *request.Context) error {
        return c.Api.Ok("done") // Fast handler
    })
    
    req := httptest.NewRequest("GET", "/fast", nil)
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.False(t, logged, "Fast request should not be logged")
}
```

---

## Comparison with Request Logger

| Feature | Request Logger | Slow Request Logger |
|---------|----------------|---------------------|
| **Logs** | All requests | Only slow requests |
| **Purpose** | General monitoring | Performance issues |
| **Overhead** | ~1-5μs per request | ~1-2μs per request |
| **Output** | All methods/paths | Only slow ones |
| **Use Case** | Access logs | Performance debugging |

**Recommendation:** Use **both** for comprehensive monitoring.

---

## See Also

- **[Request Logger](./request-logger)** - General request logging
- **[Recovery](./recovery)** - Panic recovery
- **[Request](../01-core-packages/request)** - Request handling

---

## Related Guides

- **[Performance Optimization](../../04-guides/performance/)** - Performance tips
- **[Monitoring](../../04-guides/monitoring/)** - Monitoring setup
- **[Troubleshooting](../../04-guides/troubleshooting/)** - Debugging slow requests
