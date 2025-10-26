# Lokstra Middleware Collection

This directory contains various middleware implementations for the Lokstra framework.

## Available Middleware

### 1. Body Limit (`body_limit/`)
Limits the size of request bodies to prevent memory exhaustion attacks.

**Features:**
- Configurable max body size
- Path-based skip patterns (supports `*` and `**` wildcards)
- Optional skip for large payloads
- Two-layer protection: ContentLength check + runtime enforcement

**Usage:**
```go
router.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 10 * 1024 * 1024, // 10MB
    SkipOnPath: []string{"/upload/**"},
}))
```

---

### 2. CORS (`cors/`)
Handles Cross-Origin Resource Sharing (CORS) requests.

**Features:**
- Allow specific origins or all origins (`*`)
- Automatic preflight handling
- Credentials support

**Usage:**
```go
router.Use(cors.Middleware([]string{"https://example.com"}))
// Or allow all:
router.Use(cors.Middleware([]string{"*"}))
```

---

### 3. Gzip Compression (`gzipcompression/`)
Compresses HTTP responses using gzip encoding to reduce bandwidth.

**Features:**
- Minimum size threshold (don't compress small responses)
- Configurable compression level
- Exclude specific content types (images, videos, etc.)
- Automatic client capability detection

**Usage:**
```go
router.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize: 1024,                  // 1KB minimum
    CompressionLevel: gzip.BestSpeed,
    ExcludedContentTypes: []string{"image/jpeg", "video/mp4"},
}))
```

---

### 4. Recovery (`recovery/`)
Recovers from panics and returns proper error responses instead of crashing.

**Features:**
- Catches panics in handlers
- Logs stack traces
- Configurable stack trace in response (disable in production)
- Custom panic handler support

**Usage:**
```go
router.Use(recovery.Middleware(&recovery.Config{
    EnableStackTrace: false, // Disable in production
    EnableLogging: true,
}))
```

**Best Practice:** Place as the **first middleware** to catch panics from all other middleware.

---

### 5. Request Logger (`request_logger/`)
Logs all incoming HTTP requests with method, path, status, and duration.

**Features:**
- Colored terminal output
- Duration formatting (µs, ms, s)
- Skip specific paths (e.g., `/health`)
- Custom logger support

**Usage:**
```go
router.Use(request_logger.Middleware(&request_logger.Config{
    EnableColors: true,
    SkipPaths: []string{"/health", "/metrics"},
}))
```

**Example Output:**
```
GET /api/users - Status: 200 - Duration: 45ms
POST /api/create - Status: 201 - Duration: 123ms
```

---

### 6. Slow Request Logger (`slow_request_logger/`)
Logs only slow requests that exceed a configurable threshold.

**Features:**
- Configurable threshold duration
- Highlights extra-slow requests (2x threshold)
- Colored output with severity indication
- Skip specific paths

**Usage:**
```go
router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 500 * time.Millisecond, // Log requests > 500ms
    EnableColors: true,
}))
```

**Example Output:**
```
[SLOW REQUEST] [GET] /api/search - Status: 200 - Duration: 1.2s (threshold: 500ms)
```

---

## Middleware Order Best Practices

Recommended order for optimal performance and safety:

```go
// 1. Recovery - catch all panics
router.Use(recovery.Middleware(&recovery.Config{
    EnableStackTrace: false,
}))

// 2. Request Logger - log all requests
router.Use(request_logger.Middleware(&request_logger.Config{
    SkipPaths: []string{"/health"},
}))

// 3. Slow Request Logger - detect performance issues
router.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 500 * time.Millisecond,
}))

// 4. CORS - handle preflight early
router.Use(cors.Middleware([]string{"*"}))

// 5. Body Limit - protect memory early
router.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 10 * 1024 * 1024,
}))

// 6. Gzip Compression - compress responses
router.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize: 1024,
}))

// Your routes here...
```

---

## Factory Functions for YAML Config

All middleware support factory functions for YAML-based configuration:

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
  
  - type: body_limit
    params:
      max_size: 10485760  # 10MB in bytes
      skip_on_path: ["/upload/**"]
  
  - type: cors
    params:
      allow_origins: ["*"]
  
  - type: gzip_compression
    params:
      min_size: 1024
      compression_level: -1  # default
```

---

## Testing

All middleware include comprehensive unit tests:

```bash
# Test all middleware
go test ./middleware/...

# Test specific middleware
go test ./middleware/gzipcompression
go test ./middleware/recovery
go test ./middleware/request_logger
go test ./middleware/slow_request_logger
go test ./middleware/body_limit
go test ./middleware/cors
```

---

## Adding Custom Middleware

To create a new middleware, follow this pattern:

1. Create a new folder: `middleware/your_middleware/`
2. Create main file: `your_middleware.go`
3. Implement:
   - `Config` struct with configuration options
   - `defaultConfig()` function
   - Main middleware function returning `request.HandlerFunc`
   - Factory function for YAML support
   - Register function for registry
4. Create test file: `your_middleware_test.go`
5. Add documentation to this README

**Example Structure:**
```go
package your_middleware

func YourMiddleware(cfg *Config) request.HandlerFunc {
    return request.HandlerFunc(func(c *request.Context) error {
        // Pre-processing
        
        err := c.Next() // Call next handler
        
        // Post-processing
        
        return err
    })
}
```
